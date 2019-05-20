package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/qiniu/log.v1"
	xlog "github.com/qiniu/xlog.v1"
	"qiniu.com/vas-app/biz/analyzer/client"
	"qiniu.com/vas-app/biz/analyzer/dao"
	vio "qiniu.com/vas-app/biz/analyzer/handler/violations"
	sg "qiniu.com/vas-app/biz/analyzer/handler/violations/construction"
	"qiniu.com/vas-app/biz/proto"
	"qiniu.com/vas-app/util"
)

type constructionInfo struct {
	id int

	eventType int
	startTime time.Time
	endTime   time.Time

	firstImage snapshot
}

type ConstructionHandlerConfig struct {
	Fileserver string
}

type ConstructionHandler struct {
	xlog        *log.Logger
	task        *proto.Task
	modelConfig *proto.ConstructionConfig
	config      *ConstructionHandlerConfig
	cancel      context.CancelFunc
	eventDao    dao.EventDao
	fs          client.IFileServer
	vhs         map[int]vio.ConstructionViolationHandler
	uploadCh    chan *constructionInfo
}

func NewConstructionHandler(ctx context.Context, eventDao dao.EventDao, task *proto.Task, config *ConstructionHandlerConfig) (handler *ConstructionHandler, err error) {
	xlog, ok := ctx.Value("xlog").(*log.Logger)
	if !ok {
		xlog = log.Std
		xlog.Warn("Get context log error !")
	}

	fs, err := client.NewFileserver(config.Fileserver)
	if err != nil {
		xlog.Errorf("failed to create fileserver instance:%v", err)
		return nil, err
	}

	modelConfig, err := proto.NewConstructionConfigFromMap(task.Config)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// 添加各种机动车违法行为判断的 handler
	violationHandlers := map[int]vio.ConstructionViolationHandler{
		// 施工超时
		proto.EventTypeConstructionChaoshi: sg.NewConstructionChaoshiViolation(
			ctx,
			&sg.ConstructionChaoshiViolationConfig{
				Available: modelConfig.AvailableTimes,
			},
		),
		// 施工期
		proto.EventTypeConstructionChaoqi: sg.NewConstructionChaoqiViolation(
			ctx,
			&sg.ConstructionChaoqiViolationConfig{
				EndTime: modelConfig.EndTime,
			},
		),
	}

	handler = &ConstructionHandler{
		xlog:        xlog,
		task:        task,
		modelConfig: modelConfig,
		config:      config,
		eventDao:    eventDao,
		fs:          fs,
		vhs:         violationHandlers,
	}
	ctx1, cancel := context.WithCancel(ctx)
	handler.cancel = cancel
	handler.init(ctx1)
	return handler, nil
}

func (h *ConstructionHandler) startMsgSaver(ctx context.Context, saverCh chan *constructionInfo) {
	xlog := h.xlog
	//save msg and image data
	for i := 0; i < 10; i++ {

		go func(routineNum int) {
		Loop:
			for {
				select {
				case <-ctx.Done():
					xlog.Println("stop construction handler event go routine :", routineNum)
					break Loop
				case msg := <-saverCh:
					{
						v, err := h.postProcess(msg)
						if err != nil {
							xlog.Println(err)
							continue
						}
						err = h.eventDao.Insert(*v)
						if err != nil {
							xlog.Println("insert msg to db error:", err)
							continue
						}
						xlog.Println("insert db:", v)
					}
				}
			}
			xlog.Println("stop save msg and image data go routine: ", routineNum)
		}(i)
	}

}

func (h *ConstructionHandler) postProcess(info *constructionInfo) (msg *proto.TrafficEventMsg, err error) {
	xlog := h.xlog
	uid := util.GenUuid()
	{
		err = h.parseImage(&info.firstImage)
		if err != nil {
			xlog.Println("parse first image data error:", err)
			return nil, err
		}
	}

	firstImgUrlRaw, err := h.fs.Save(uid+"_1_raw.jpg", info.firstImage.jpgData)
	if err != nil {
		xlog.Println("upload first image error:", err)
		return nil, err
	}

	msg = &proto.TrafficEventMsg{
		EventID:   uid,
		EventType: info.eventType,
		StartTime: info.startTime,
		CameraID:  h.task.CameraID,
		Address:   h.task.StreamAddr,
		EndTime:   info.endTime,
		Status:    proto.StatusFinished,
		Mark: proto.Mark{
			Marking: proto.MarkingInit,
		},
		Snapshot: []proto.Snapshot{
			proto.Snapshot{
				SnapshotURIRaw: firstImgUrlRaw,
			},
		},
	}

	return msg, nil
}

func (h *ConstructionHandler) drawAndUpload(uid string, jpgData []byte, pts [][2]int, frameIdx int, eventType int, class int) (uploadURL string, err error) {
	uploadURL, err = h.fs.Save(fmt.Sprintf("%s_%d.jpg", uid, frameIdx), jpgData)
	if err != nil {
		h.xlog.Println("upload image error:", err)
		return "", err
	}
	return uploadURL, nil
}

func (h *ConstructionHandler) init(ctx context.Context) {
	var saverCh = make(chan *constructionInfo, 20)
	h.startMsgSaver(ctx, saverCh)
	h.uploadCh = saverCh
}

func (h *ConstructionHandler) Handle(data interface{}, imageData *proto.ImageBody) (err error) {
	xlog := h.xlog

	if imageData == nil {
		xlog.Println("empty body ,skip")
		return nil
	}
	bData, err := json.Marshal(data)
	if err != nil {
		xlog.Println(err)
		return err
	}
	var wmData proto.ConstructionModelData
	err = json.Unmarshal(bData, &wmData)
	if err != nil {
		xlog.Println(err)
		return err
	}
	xlog.Println(wmData)

	for vt, vh := range h.vhs {
		xlog.Infof("handle for violation type[%v]", vt)
		events, err := vh.Handle(&wmData, imageData)
		if err != nil {
			xlog.Warnf("handler failed, violation:%v, err:%v", vt, err)
			continue
		}
		if len(events) == 0 {
			xlog.Debugf("no event triggered in violation handler, violation:%v, ignore", vt)
			continue
		}
		for _, event := range events {
			info := &constructionInfo{
				id:        event.ID,
				eventType: event.ViolationType,
				startTime: event.StartTime,
				endTime:   event.EndTime,
				firstImage: snapshot{
					body: event.Snapshots[0].RawData,
				},
			}
			h.saveMsg(info)
		}
	}

	return nil
}
func (h *ConstructionHandler) saveMsg(info *constructionInfo) {
	h.uploadCh <- info
}

func (h *ConstructionHandler) Release() error {
	h.cancel()
	for _, vh := range h.vhs {
		vh.Release()
	}
	h.xlog.Println("construction handler release called")
	return nil
}

func (h *ConstructionHandler) parseImage(snap *snapshot) error {
	img, err := ConvertGBRData2Image(snap.body.Body, snap.body.Width, snap.body.Height)
	if err != nil {
		xlog.Error("convert image error:", err)
		return err
	}
	body, err := GetJpgData(img)
	if err != nil {
		xlog.Error("get jpg image data error:", err)
		return err
	}
	snap.jpgData = body
	return nil
}
