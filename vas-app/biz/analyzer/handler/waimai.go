package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/qiniu/log.v1"
	xlog "github.com/qiniu/xlog.v1"
	"qiniu.com/vas-app/biz/analyzer/client"
	"qiniu.com/vas-app/biz/analyzer/dao"
	vio "qiniu.com/vas-app/biz/analyzer/handler/violations"
	vn "qiniu.com/vas-app/biz/analyzer/handler/violations/nonmotor"
	"qiniu.com/vas-app/biz/proto"
	"qiniu.com/vas-app/util"
)

type waimaiValue struct {
	id int

	eventType  int
	startTime  time.Time
	endTime    time.Time
	plate      string
	plateScore float64

	firstImage  snapshot
	secondImage snapshot
	thirdImage  snapshot
}

type WaimaiHandlerConfig struct {
	Fileserver      string
	Timeout         int //超时时间，超过多少秒没有同一id的人认为结束
	DetectZone      [][2]int
	EventInitStatus string
}

//算法区域人员总数
type WaimaiHandler struct {
	xlog        *log.Logger
	task        *proto.Task
	modelConfig *proto.ModelConfig
	config      *WaimaiHandlerConfig
	cancel      context.CancelFunc
	mutex       sync.RWMutex
	eventDao    dao.EventDao
	fs          client.IFileServer
	vhs         map[int]vio.NonMotorViolationHandler
	uploadCh    chan *waimaiValue
}

func NewWaimaiHandler(ctx context.Context, eventDao dao.EventDao, task *proto.Task, config *WaimaiHandlerConfig) (handler *WaimaiHandler, err error) {
	xlog, ok := ctx.Value("xlog").(*log.Logger)
	if !ok {
		xlog = log.Std
		xlog.Error("Get context log error !")
	}

	fs, err := client.NewFileserver(config.Fileserver)
	if err != nil {
		xlog.Errorf("failed to create fileserver instance:%v", err)
		return nil, err
	}

	var mc proto.ModelConfig
	mcData, err := json.Marshal(task.Config)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	err = json.Unmarshal(mcData, &mc)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if mc.EventInitStatus == "" {
		config.EventInitStatus = proto.StatusInit
	} else {
		config.EventInitStatus = mc.EventInitStatus
	}

	// 添加各种机动车违法行为判断的 handler
	violationHandlers := map[int]vio.NonMotorViolationHandler{
		// 闯红灯
		proto.EventTypeNonMotorChuangHongDeng: vn.NewChuangHongDengViolation(
			ctx,
			&vn.ChuangHongDengViolationConfig{
				Timeout: 20,
			},
		),
		//占人行道
		proto.EventTypeNonMotorZhanRenXingDao: vn.NewZhanRenXingDaoViolation(
			ctx,
			&vn.ZhanRenXingDaoViolationConfig{
				Timeout: 20,
			},
		),
		//逆行
		proto.EventTypeNonMotorNiXing: vn.NewNiXingViolation(
			ctx,
			&vn.NiXingViolationConfig{
				Timeout: 20,
			},
		),
	}

	handler = &WaimaiHandler{
		xlog:        xlog,
		task:        task,
		modelConfig: &mc,
		config:      config,
		mutex:       sync.RWMutex{},
		eventDao:    eventDao,
		fs:          fs,
		vhs:         violationHandlers,
	}
	ctx1, cancel := context.WithCancel(ctx)
	handler.cancel = cancel
	handler.init(ctx1)
	return handler, nil
}

func (h *WaimaiHandler) startMsgSaver(ctx context.Context, saverCh chan *waimaiValue) {
	xlog := h.xlog
	//save msg and image data
	for i := 0; i < 10; i++ {

		go func(routineNum int) {
		Loop:
			for {
				select {
				case <-ctx.Done():
					xlog.Println("stop waimai handler event go routine :", routineNum)
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

func (h *WaimaiHandler) postProcess(info *waimaiValue) (msg *proto.TrafficEventMsg, err error) {
	xlog := h.xlog
	uid := util.GenUuid()
	{
		err = h.parseImage(&info.firstImage)
		if err != nil {
			xlog.Println("parse first image data error:", err)
			return nil, err
		}
		err = h.parseImage(&info.secondImage)
		if err != nil {
			xlog.Println("parse second image data error:", err)
			return nil, err
		}
		err = h.parseImage(&info.thirdImage)
		if err != nil {
			xlog.Println("parse third image data error:", err)
			return nil, err
		}
	}

	firstImgUrlRaw, err := h.fs.Save(uid+"_1_raw.jpg", info.firstImage.jpgData)
	if err != nil {
		xlog.Println("upload first image error:", err)
		return nil, err
	}
	secondImgUrlRaw, err := h.fs.Save(uid+"_2_raw.jpg", info.secondImage.jpgData)
	if err != nil {
		xlog.Println("upload second snapshot raw image error:", err)
		return nil, err
	}
	thirdImgUrlRaw, err := h.fs.Save(uid+"_3_raw.jpg", info.thirdImage.jpgData)
	if err != nil {
		xlog.Println("upload third snapshot raw image error:", err)
		return nil, err
	}

	//extract image
	rects := []int{}
	for _, i := range info.firstImage.pts {
		for _, j := range i {
			rects = append(rects, j)
		}
	}
	featureImg, err := util.ExtractImage(info.firstImage.jpgData, rects)
	if err != nil {
		xlog.Println("extract image error:", err)
		return nil, err
	}

	featureUrl, err := h.fs.Save(uid+"_feature.jpg", featureImg)
	if err != nil {
		xlog.Println("upload feature image error:", err)
		return nil, err
	}

	//画框和违法行为标记
	firstImgUrl, err := h.drawAndUpload(uid, info.firstImage.jpgData, info.firstImage.pts, 1, info.eventType, info.firstImage.class)
	if err != nil {
		return nil, err
	}
	secondImgUrl, err := h.drawAndUpload(uid, info.secondImage.jpgData, info.secondImage.pts, 2, info.eventType, info.secondImage.class)
	if err != nil {
		return nil, err
	}
	thirdImgUrl, err := h.drawAndUpload(uid, info.thirdImage.jpgData, info.thirdImage.pts, 3, info.eventType, info.secondImage.class)
	if err != nil {
		return nil, err
	}

	msg = &proto.TrafficEventMsg{
		EventID:   uid,
		EventType: info.eventType,
		StartTime: info.startTime,
		CameraID:  h.task.CameraID,
		Address:   h.task.StreamAddr,
		EndTime:   info.endTime,
		Status:    h.config.EventInitStatus,
		Mark: proto.Mark{
			Marking: proto.MarkingInit,
		},
		Snapshot: []proto.Snapshot{
			proto.Snapshot{
				SnapshotURI:    firstImgUrl,
				SnapshotURIRaw: firstImgUrlRaw,
				FeatureURI:     featureUrl,
				Pts:            info.firstImage.pts,
				Class:          info.firstImage.class,
				Label:          info.plate,
				LabelScore:     info.plateScore,
			},
			proto.Snapshot{
				SnapshotURI:    secondImgUrl,
				SnapshotURIRaw: secondImgUrlRaw,
				Pts:            info.secondImage.pts,
				Class:          info.firstImage.class,
				Label:          info.plate,
				LabelScore:     info.plateScore,
			},
			proto.Snapshot{
				SnapshotURI:    thirdImgUrl,
				SnapshotURIRaw: thirdImgUrlRaw,
				Pts:            info.thirdImage.pts,
				Class:          info.firstImage.class,
				Label:          info.plate,
				LabelScore:     info.plateScore,
			},
		},
	}

	return msg, nil
}

func (h *WaimaiHandler) drawAndUpload(uid string, jpgData []byte, pts [][2]int, frameIdx int, eventType int, class int) (uploadURL string, err error) {
	body, err := drawDetectInfo(jpgData, pts, eventType, class)
	if err != nil {
		h.xlog.Println("draw image tag info error:", err)
		return "", err
	}
	uploadURL, err = h.fs.Save(fmt.Sprintf("%s_%d.jpg", uid, frameIdx), body)
	if err != nil {
		h.xlog.Println("upload image error:", err)
		return "", err
	}
	return uploadURL, nil
}

func (h *WaimaiHandler) init(ctx context.Context) {
	var saverCh = make(chan *waimaiValue, 20)
	h.startMsgSaver(ctx, saverCh)
	h.uploadCh = saverCh
}

func (h *WaimaiHandler) Handle(data interface{}, imageData *proto.ImageBody) (err error) {
	xlog := h.xlog

	bData, err := json.Marshal(data)
	if err != nil {
		xlog.Println(err)
		return err
	}
	var wmData proto.WaimaiModelData
	err = json.Unmarshal(bData, &wmData)
	if err != nil {
		xlog.Println(err)
		return err
	}
	xlog.Println(wmData)

	wg := sync.WaitGroup{}
	for vt, vh := range h.vhs {
		// 并发处理各种违法 handler
		wg.Add(1)
		go func(vt int, vh vio.NonMotorViolationHandler) {
			defer wg.Done()
			xlog.Infof("handle for violation type[%v]", vt)
			events, err := vh.Handle(&wmData, imageData)
			if err != nil {
				xlog.Warnf("handler failed, violation:%v, err:%v", vt, err)
				return
			}
			if len(events) == 0 {
				xlog.Debugf("no event triggered in violation handler, violation:%v, ignore", vt)
				return
			}
			for _, event := range events {
				if len(event.Snapshots) <= 2 {
					xlog.Warnf("nonmotor violation event should have at least 3 snapshots, have")
					continue
				}
				//有些违章由业务层判断，需要额外的判定类型
				eventTypes := h.getEventTypes(event.ViolationType)
				for _, eventType := range eventTypes {
					info := &waimaiValue{
						id:         event.ID,
						eventType:  eventType,
						startTime:  event.StartTime,
						endTime:    event.EndTime,
						plate:      event.Snapshots[0].Label,
						plateScore: event.Snapshots[0].LabelScore,
						firstImage: snapshot{
							body:  event.Snapshots[0].RawData,
							pts:   event.Snapshots[0].Pts,
							class: event.Snapshots[0].ObjectClass,
						},
						secondImage: snapshot{
							body:  event.Snapshots[1].RawData,
							pts:   event.Snapshots[1].Pts,
							class: event.Snapshots[1].ObjectClass,
						},
						thirdImage: snapshot{
							body:  event.Snapshots[2].RawData,
							pts:   event.Snapshots[2].Pts,
							class: event.Snapshots[2].ObjectClass,
						},
					}
					h.saveMsg(info)
				}
			}

		}(vt, vh)
	} // range vhs

	wg.Wait()

	return nil
}
func (h *WaimaiHandler) saveMsg(info *waimaiValue) {
	h.uploadCh <- info
}

func (h *WaimaiHandler) getEventTypes(eventType int) []int {
	arr := []int{}
	if h.modelConfig.WaimaiZyjdcdOn || h.modelConfig.WaimaiCjfqOn {
		if h.modelConfig.WaimaiZyjdcdOn {
			arr = append(arr, proto.EventTypeNonMotorZhanJiDongCheDao)
		}
		if h.modelConfig.WaimaiCjfqOn {
			arr = append(arr, proto.EventTypeNonMotorWeiFanJinLingBiaoZhi)

		}
	} else {
		arr = append(arr, eventType)
	}
	return arr
}

func (h *WaimaiHandler) Release() error {
	h.cancel()
	for _, vh := range h.vhs {
		vh.Release()
	}
	h.xlog.Println("waimai handler release called")
	return nil
}

func (h *WaimaiHandler) parseImage(snap *snapshot) error {
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
