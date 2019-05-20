package nonmotor

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/qiniu/log.v1"
	vio "qiniu.com/vas-app/biz/analyzer/handler/violations"
	"qiniu.com/vas-app/biz/proto"
)

type ZhanRenXingDaoViolationConfig struct {
	Timeout int
}

type ZhanRenXingDaoViolation struct {
	// 需要处理的违法类型
	// tricky，可以从外部修改，因为存在多个违法行为的取证逻辑是一致的，只是违法类型不同，可以复用
	VioEventType int

	objVioEvent map[int]*vio.ViolationEvent // 目标 ID -> 当前的违法事件

	cfg    *ZhanRenXingDaoViolationConfig
	xlog   *log.Logger
	ctx    context.Context
	mutex  sync.RWMutex
	cancel context.CancelFunc
	saveCh chan *vio.ViolationEvent
}

func NewZhanRenXingDaoViolation(ctx context.Context, cfg *ZhanRenXingDaoViolationConfig) *ZhanRenXingDaoViolation {
	xlog, ok := ctx.Value("xlog").(*log.Logger)
	if !ok {
		xlog = log.Std
		xlog.Error("Get context log error !")
	}
	v := &ZhanRenXingDaoViolation{
		VioEventType: proto.EventTypeNonMotorZhanRenXingDao,
		objVioEvent:  make(map[int]*vio.ViolationEvent),
		xlog:         xlog,
		ctx:          ctx,
		cfg:          cfg,
		saveCh:       make(chan *vio.ViolationEvent, 10),
	}
	v.ctx, v.cancel = context.WithCancel(ctx)
	v.backgroundWork()

	return v
}
func (v *ZhanRenXingDaoViolation) Handle(frameData *proto.WaimaiModelData, picData *proto.ImageBody) (events []*vio.ViolationEvent, err error) {
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
				// info.FrameCount++
				info.Snapshots[2] = snap
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

				if b, err := isBiggerRect(box.Pts, info.Snapshots[0].Pts); b && err == nil {
					xlog.Println("bigger rect ,replace old data")
					info.Snapshots[0] = snap
				}

				if updatePlate {
					info.Snapshots[0].LabelScore = box.PlateScore
					info.Snapshots[0].Label = box.Plate
				}
			} else {
				v.objVioEvent[box.ID] = &vio.ViolationEvent{
					ID:            box.ID,
					ViolationType: v.VioEventType,
					StartTime:     now,
					EndTime:       now,
					Snapshots: []*vio.ViolationEventSnapshot{
						snap, //最大的一张
						snap, // 第一张 ,ignore label value of snap 1 ,2
						snap, // 最后一张
					},
				}
			}
		}
		v.mutex.Unlock()
	}

	//todo
	select {
	case newEvent := <-v.saveCh:
		events = append(events, newEvent)
		return events, nil
	default:
	}
	return nil, nil
}

func (zv *ZhanRenXingDaoViolation) backgroundWork() {
	xlog := zv.xlog
	//start timeout worker
	go func() {
	Loop:
		for {
			select {
			case <-zv.ctx.Done():
				xlog.Println("stop zhanrenxingdao handler event go routine ")
				break Loop
			case <-time.After(time.Second * time.Duration(zv.cfg.Timeout)):
				now := time.Now()
				removeList := []int{}

				zv.mutex.RLock()
				for k, v := range zv.objVioEvent {
					if now.Sub(v.StartTime) > time.Second*time.Duration(zv.cfg.Timeout) {
						removeList = append(removeList, k)
						zv.saveCh <- v
					}
				}

				zv.mutex.RUnlock()

				if len(removeList) != 0 {
					zv.mutex.Lock()
					xlog.Println("removing map keys :", removeList)
					for _, k := range removeList {
						delete(zv.objVioEvent, k)
					}
					zv.mutex.Unlock()
				}
			}
		}
	}()
}

func (v *ZhanRenXingDaoViolation) Release() (err error) {
	v.xlog.Debug("ZhanRenXingDaoViolation Release")
	v.cancel()
	close(v.saveCh)
	return nil
}

func isBiggerRect(a, b [][2]int) (bool, error) {
	if len(a) != 2 || len(b) != 2 {
		return false, errors.New("invalid pts")
	}

	return abs((a[1][0]-a[0][0])*(a[1][1]*a[0][1])) > abs((b[1][0]-b[0][0])*(b[1][1]*b[0][1])), nil
}

func abs(input int) int {
	if input < 0 {
		return -input
	}
	return input
}
