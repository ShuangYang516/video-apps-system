package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	xlog "github.com/qiniu/xlog.v1"
	log "qiniupkg.com/x/log.v7"

	"qiniu.com/vas-app/biz/analyzer/client"
	"qiniu.com/vas-app/biz/analyzer/dao"
	vio "qiniu.com/vas-app/biz/analyzer/handler/violations"
	vv "qiniu.com/vas-app/biz/analyzer/handler/violations/vehicle"
	"qiniu.com/vas-app/biz/proto"
	"qiniu.com/vas-app/util"
)

type VehicleSnapshot struct {
	body     *proto.ImageBody
	jpgData  []byte
	pts      [][2]int
	frameIdx int
	label    string // 车牌
	t        time.Time
}

type VehicleInfo struct {
	id             int
	msg            *proto.TrafficEventMsg
	firstSnapshot  *VehicleSnapshot
	secondSnapshot *VehicleSnapshot
	thirdSnapshot  *VehicleSnapshot
}

type VehicleHandlerConfig struct {
	Fileserver string `json:"fileserver"`
}

type VehicleHandler struct {
	xlog     *log.Logger
	task     *proto.Task
	config   *VehicleHandlerConfig
	cancel   context.CancelFunc
	eventDao dao.EventDao
	fs       client.IFileServer
	uploadCh chan *VehicleInfo
	vhs      map[int]vio.ViolationHandler
}

func NewVehicleHandler(ctx context.Context, eventDao dao.EventDao, task *proto.Task, fileserver string) (handler *VehicleHandler, err error) {
	xlog, ok := ctx.Value("xlog").(*log.Logger)
	if !ok {
		xlog = log.Std
		xlog.Error("Get context log error !")
	}
	config := &VehicleHandlerConfig{
		Fileserver: fileserver,
	}

	var fs client.IFileServer
	fs, err = client.NewFileserver(fileserver)
	if err != nil {
		xlog.Errorf("failed to create fileserver instance:%v", err)
		return nil, err
	}

	// 添加各种机动车违法行为判断的 handler
	violationHandlers := map[int]vio.ViolationHandler{
		// 大弯小转
		proto.EventTypeVehicleDaWanXiaoZhuan: vv.NewDawanxiaozhuanViolation(
			ctx,
			&vv.DawanxiaozhuanViolationConfig{
				Timeout: 60,
			},
		),
		// 实线变道
		proto.EventTypeVehiclShiXianBianDao: vv.NewShixianbiandaoViolation(
			ctx,
			&vv.ShixianbiandaoViolationConfig{
				Timeout: 60,
			},
		),
		// 不按导向线行驶
		proto.EventTypeVehicleBuAnDaoXiangXianXiangShi: vv.NewBuandaoxiangxianxingshiViolation(
			ctx,
			&vv.BuandaoxiangxianxingshiViolationConfig{
				Timeout: 60,
			},
		),
		// 网格线停车
		proto.EventTypeVehicleWangGeXianTingChe: vv.NewWanggexiantingcheViolation(
			ctx,
			&vv.WanggexiantingcheViolationConfig{
				Timeout: 60,
			},
		),
		// 不礼让行人
		proto.EventTypeVehicleBuLiRangXingRen: vv.NewBulirangxingrenViolation(
			ctx,
			&vv.BulirangxingrenViolationConfig{
				Timeout: 60,
			},
		),

		// ...
	}

	handler = &VehicleHandler{
		xlog:     xlog,
		task:     task,
		config:   config,
		eventDao: eventDao,
		fs:       fs,
		uploadCh: make(chan *VehicleInfo, 10),
		vhs:      violationHandlers,
	}

	ctx1, cancel := context.WithCancel(ctx)
	handler.cancel = cancel
	handler.init(ctx1)
	return handler, nil
}

func (h *VehicleHandler) CanHandle(modelKey string) bool {
	return modelKey == VehicleHandlerModelKey
}

func (h *VehicleHandler) init(ctx context.Context) {
	xlog := h.xlog
	//upload goroutine pool
	for i := 0; i < 10; i++ {
		go func(routineNum int) {
		Loop:
			for {
				select {
				case <-ctx.Done():
					xlog.Println("stop vehicle handler event go routine :", routineNum)
					break Loop
				case info := <-h.uploadCh:
					xlog.Println("upload chan received:", info.msg)

					msg, err := h.postProcess(info)
					if err != nil {
						xlog.Infof("postPorcess failed, err: %v", err)
						continue
					}

					err = h.eventDao.Insert(*msg)
					if err != nil {
						xlog.Println("insert msg to db error:", err)
						continue
					}
					xlog.Println("insert db:", *msg)
				}
			}
		}(i)
	}

}

func (h *VehicleHandler) postProcess(info *VehicleInfo) (msg *proto.TrafficEventMsg, err error) {
	xlog := h.xlog
	uid := util.GenUuid()

	info.firstSnapshot.jpgData, err = h.parseImage(info.firstSnapshot)
	if err != nil {
		xlog.Println("parse first image data error:", err)
		return
	}
	info.secondSnapshot.jpgData, err = h.parseImage(info.secondSnapshot)
	if err != nil {
		xlog.Println("parse second image data error:", err)
		return
	}
	info.thirdSnapshot.jpgData, err = h.parseImage(info.thirdSnapshot)
	if err != nil {
		xlog.Println("parse third image data error:", err)
		return
	}
	//原图上传
	_, err = h.fs.Save(fmt.Sprintf("%s_%d_raw.jpg", uid, info.firstSnapshot.frameIdx), info.firstSnapshot.jpgData)
	if err != nil {
		xlog.Println("upload raw image  error:", err)
		return
	}
	_, err = h.fs.Save(fmt.Sprintf("%s_%d_raw.jpg", uid, info.secondSnapshot.frameIdx), info.secondSnapshot.jpgData)
	if err != nil {
		xlog.Println("upload raw image  error:", err)
		return
	}
	_, err = h.fs.Save(fmt.Sprintf("%s_%d_raw.jpg", uid, info.thirdSnapshot.frameIdx), info.thirdSnapshot.jpgData)
	if err != nil {
		xlog.Println("upload raw image error:", err)
		return
	}
	// 第一张作为特写图
	rects := []int{}
	for _, i := range info.firstSnapshot.pts {
		for _, j := range i {
			rects = append(rects, j)
		}
	}
	featureImg, err := util.ExtractImage(info.firstSnapshot.jpgData, rects)
	if err != nil {
		xlog.Println("extract image error:", err)
		return
	}
	firstFeatureURL, err := h.fs.Save(fmt.Sprintf("%s_%d_feature.jpg", uid, info.firstSnapshot.frameIdx), featureImg)
	if err != nil {
		xlog.Println("upload raw image error:", err)
		return
	}

	//画框和违法行为标记
	firstImageUrl, err := h.drawAndUpload(
		uid, info.firstSnapshot.jpgData, info.firstSnapshot.pts,
		info.firstSnapshot.frameIdx, info.msg.EventType, 0)
	if err != nil {
		xlog.Warnf("drawAndUpload failed, err: %v", err)
		return
	}

	secondImageUrl, err := h.drawAndUpload(
		uid, info.secondSnapshot.jpgData, info.secondSnapshot.pts,
		info.secondSnapshot.frameIdx, info.msg.EventType, 0)
	if err != nil {
		xlog.Warnf("drawAndUpload failed, err: %v", err)
		return
	}

	thirdImageUrl, err := h.drawAndUpload(
		uid, info.thirdSnapshot.jpgData, info.thirdSnapshot.pts,
		info.thirdSnapshot.frameIdx, info.msg.EventType, 0)
	if err != nil {
		xlog.Warnf("drawAndUpload failed, err: %v", err)
		return
	}

	msg = info.msg
	msg.EventID = fmt.Sprintf("%s-%d-%d", uid, info.id, msg.EventType)
	msg.Snapshot = []proto.Snapshot{
		proto.Snapshot{
			SnapshotURI: firstImageUrl,
			FeatureURI:  firstFeatureURL,
			Pts:         info.firstSnapshot.pts,
			Label:       info.firstSnapshot.label,
		},
		proto.Snapshot{
			SnapshotURI: secondImageUrl,
			Pts:         info.secondSnapshot.pts,
			Label:       info.secondSnapshot.label,
		},
		proto.Snapshot{
			SnapshotURI: thirdImageUrl,
			Pts:         info.thirdSnapshot.pts,
		},
	}
	msg.Status = proto.StatusInit
	msg.Mark = proto.Mark{
		Marking: proto.MarkingInit,
	}

	return msg, nil

}

func (h *VehicleHandler) parseImage(snap *VehicleSnapshot) (jpgData []byte, err error) {
	img, err := ConvertGBRData2Image(snap.body.Body, snap.body.Width, snap.body.Height)
	if err != nil {
		xlog.Error("convert image error:", err)
		return jpgData, err
	}
	body, err := GetJpgData(img)
	if err != nil {
		xlog.Error("get jpg image data error:", err)
		return jpgData, err
	}
	return body, nil
}

func (h *VehicleHandler) Handle(data interface{}, body *proto.ImageBody) (err error) {
	now := time.Now()
	xlog := h.xlog

	if body == nil {
		xlog.Info("empty body ,skip")
		return nil
	}
	bData, err := json.Marshal(data)
	if err != nil {
		xlog.Error(err)
		return err
	}
	var mData proto.VehicleModelData
	err = json.Unmarshal(bData, &mData)
	if err != nil {
		xlog.Error(err)
		return err
	}
	xlog.Info(mData)

	wg := sync.WaitGroup{}
	for vt, vh := range h.vhs {
		// 并发处理各种违法 handler
		wg.Add(1)
		go func(vt int, vh vio.ViolationHandler) {
			defer wg.Done()
			xlog.Infof("handle for violation type[%v]", vt)
			event, err := vh.Handle(&mData, body)
			if err != nil {
				xlog.Warnf("handler failed, violation:%v, err:%v", vt, err)
				return
			}
			if event == nil {
				xlog.Debugf("no event triggered in violation handler, violation:%v, ignore", vt)
				return
			}
			if len(event.Snapshots) <= 2 {
				xlog.Warnf("vehicle violation event should have at least 3 snapshots, have")
				return
			}

			info := &VehicleInfo{
				id: event.ID,
				msg: &proto.TrafficEventMsg{
					EventType: event.ViolationType,
					StartTime: now,
					EndTime:   now,
					Address:   h.task.StreamAddr,
					CameraID:  h.task.CameraID,
				},
				firstSnapshot: &VehicleSnapshot{
					body:     event.Snapshots[0].RawData,
					pts:      event.Snapshots[0].Pts,
					label:    event.Snapshots[0].Label,
					t:        event.Snapshots[0].Tz,
					frameIdx: vio.VehicleFirstIdx,
				},
				secondSnapshot: &VehicleSnapshot{
					body:     event.Snapshots[1].RawData,
					pts:      event.Snapshots[1].Pts,
					label:    event.Snapshots[1].Label,
					t:        event.Snapshots[1].Tz,
					frameIdx: vio.VehicleSecondIdx,
				},
				thirdSnapshot: &VehicleSnapshot{
					body:     event.Snapshots[2].RawData,
					pts:      event.Snapshots[2].Pts,
					label:    event.Snapshots[2].Label,
					t:        event.Snapshots[2].Tz,
					frameIdx: vio.VehicleThirdIdx,
				},
			}
			h.saveMsg(info)
		}(vt, vh)
	} // range vhs

	wg.Wait()

	return nil
}

func (h *VehicleHandler) saveMsg(info *VehicleInfo) {
	h.uploadCh <- info
}

func (h *VehicleHandler) drawAndUpload(uid string, jpgData []byte, pts [][2]int, frameIdx int, eventType int, class int) (uploadURL string, err error) {
	body, err := drawDetectInfo(jpgData, pts, eventType, class)
	if err != nil {
		h.xlog.Println("draw image tag info error:", err)
		return
	}
	uploadURL, err = h.fs.Save(fmt.Sprintf("%s_%d.jpg", uid, frameIdx), body)
	if err != nil {
		h.xlog.Println("upload image error:", err)
		return
	}
	return uploadURL, nil
}

func (h *VehicleHandler) Release() error {
	h.cancel()
	h.xlog.Println("vehicle handler release called")

	for vt, vh := range h.vhs {
		vh.Release()
		h.xlog.Printf("violation[%v] handler released", vt)
	}

	return nil
}
