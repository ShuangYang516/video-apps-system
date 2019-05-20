package nonmotor

import (
	"context"
	"sync"
	"time"

	"github.com/qiniu/log.v1"
	vio "qiniu.com/vas-app/biz/analyzer/handler/violations"

	"qiniu.com/vas-app/biz/proto"
)

type ChuangHongDengViolationConfig struct {
	Timeout int
}

type ChuangHongDengViolation struct {
	// 需要处理的违法类型
	// tricky，可以从外部修改，因为存在多个违法行为的取证逻辑是一致的，只是违法类型不同，可以复用
	VioEventType int

	objVioEvent map[int]*vio.ViolationEvent // 目标 ID -> 当前的违法事件

	cfg    *ChuangHongDengViolationConfig
	xlog   *log.Logger
	ctx    context.Context
	mutex  sync.RWMutex
	cancel context.CancelFunc
}

func NewChuangHongDengViolation(ctx context.Context, cfg *ChuangHongDengViolationConfig) *ChuangHongDengViolation {
	xlog, ok := ctx.Value("xlog").(*log.Logger)
	if !ok {
		xlog = log.Std
		xlog.Error("Get context log error !")
	}
	v := &ChuangHongDengViolation{
		VioEventType: proto.EventTypeNonMotorChuangHongDeng,
		objVioEvent:  make(map[int]*vio.ViolationEvent),
		xlog:         xlog,
		ctx:          ctx,
		cfg:          cfg,
	}
	v.ctx, v.cancel = context.WithCancel(ctx)
	v.backgroundWork()

	return v
}

func (v *ChuangHongDengViolation) Handle(frameData *proto.WaimaiModelData, picData *proto.ImageBody) (events []*vio.ViolationEvent, err error) {
	xlog := v.xlog
	if picData != nil {
		v.mutex.Lock()
		now := time.Now()
		for _, box := range frameData.Boxes {
			xlog.Println(box)
			match := false
			for _, vioId := range box.Violation {
				if vioId == v.VioEventType {
					match = true
				}
			}
			if !match {
				continue
			}

			snap := &vio.ViolationEventSnapshot{
				RawData:     picData,
				Pts:         box.Pts,
				Label:       box.Plate,
				LabelScore:  box.PlateScore,
				ObjectClass: box.Class,
			}

			if info, ok := v.objVioEvent[box.ID]; ok && info != nil {
				info.EndTime = now

				if box.CrossSubLineID == 0 && info.Snapshots[1] == nil {
					info.Snapshots[1] = snap
				} else if box.CrossSubLineID == 1 && info.Snapshots[2] == nil {
					info.Snapshots[2] = snap
				}

				currentLabel := info.Snapshots[0].Label
				currentLabelScore := info.Snapshots[0].LabelScore

				updatePlate := false
				// 暂定5位数
				if len(box.Plate) >= 5 && (len(currentLabel) < 5 || box.PlateScore > currentLabelScore) {
					updatePlate = true
				}
				if len(box.Plate) < 5 && len(box.Plate) > len(currentLabel) {
					updatePlate = true
				}

				if updatePlate {
					info.Snapshots[0].LabelScore = box.PlateScore
					info.Snapshots[0].Label = box.Plate
				}

				if b, err := isBiggerRect(box.Pts, info.Snapshots[0].Pts); b && err == nil {
					xlog.Println("bigger rect ,replace old data")
					info.Snapshots[0] = snap
				}

				if info.Snapshots[0] != nil && info.Snapshots[1] != nil && info.Snapshots[2] != nil {
					events = append(events, info)
					delete(v.objVioEvent, box.ID)
				}

			} else {
				bestSnap := snap
				var firstLineSnap, secondLineSnap *vio.ViolationEventSnapshot

				if box.CrossSubLineID == 0 {
					firstLineSnap = snap
				} else if box.CrossSubLineID == 1 {
					secondLineSnap = snap
				}

				v.objVioEvent[box.ID] = &vio.ViolationEvent{
					ID:            box.ID,
					ViolationType: v.VioEventType,
					StartTime:     now,
					EndTime:       now,
					Snapshots: []*vio.ViolationEventSnapshot{
						bestSnap,       //最大的一张
						firstLineSnap,  // sub cross line = 0
						secondLineSnap, // sub cross line = 1
					},
				}
			}
		}
		v.mutex.Unlock()
	}
	return events, nil
}

func (cv *ChuangHongDengViolation) backgroundWork() {
	xlog := cv.xlog
	//start timeout worker
	go func() {
	Loop:
		for {
			select {
			case <-cv.ctx.Done():
				xlog.Println("stop zhanrenxingdao handler event go routine ")
				break Loop
			case <-time.After(time.Second * time.Duration(cv.cfg.Timeout)):
				now := time.Now()
				removeList := []int{}

				cv.mutex.RLock()
				for k, v := range cv.objVioEvent {
					if now.Sub(v.StartTime) > time.Second*time.Duration(cv.cfg.Timeout) {
						removeList = append(removeList, k)
					}
				}

				cv.mutex.RUnlock()

				if len(removeList) != 0 {
					cv.mutex.Lock()
					xlog.Println("removing map keys :", removeList)
					for _, k := range removeList {
						delete(cv.objVioEvent, k)
					}
					cv.mutex.Unlock()
				}
			}
		}
	}()
}

func (v *ChuangHongDengViolation) Release() (err error) {
	v.xlog.Debug("ChuangHongDengViolation Release")
	v.cancel()

	return nil
}
