package vehicle

import (
	"context"
	"sync"
	"time"

	log "qiniupkg.com/x/log.v7"

	vio "qiniu.com/vas-app/biz/analyzer/handler/violations"
	"qiniu.com/vas-app/biz/proto"
)

type DawanxiaozhuanViolationConfig struct {
	Timeout int //超时时间，超过多少秒没有同一id的目标认为结束
}

// 大弯小转
type DawanxiaozhuanViolation struct {
	// 需要处理的违法类型
	// tricky，可以从外部修改，因为存在多个违法行为的取证逻辑是一致的，只是违法类型不同，可以复用
	VioEventType int

	objVioEvent map[int]*vio.ViolationEvent            // 目标 ID -> 当前的违法事件
	labelSnap   map[string]*vio.ViolationEventSnapshot // 车牌 -> 当时的取证图片

	cfg    *DawanxiaozhuanViolationConfig
	logger *log.Logger
	ctx    context.Context
	mutex  sync.RWMutex
	cancel context.CancelFunc
}

func NewDawanxiaozhuanViolation(ctx context.Context, cfg *DawanxiaozhuanViolationConfig) *DawanxiaozhuanViolation {
	v := &DawanxiaozhuanViolation{
		VioEventType: proto.EventTypeVehicleDaWanXiaoZhuan,
		objVioEvent:  make(map[int]*vio.ViolationEvent),
		labelSnap:    make(map[string]*vio.ViolationEventSnapshot),
		logger:       log.Std,
		ctx:          ctx,
		cfg:          cfg,
	}
	v.ctx, v.cancel = context.WithCancel(ctx)
	v.backgroundWork()

	return v
}

func (dv *DawanxiaozhuanViolation) Handle(frameData *proto.VehicleModelData, picData *proto.ImageBody) (event *vio.ViolationEvent, err error) {
	log := dv.logger
	now := time.Now()
	dv.mutex.Lock()
	defer dv.mutex.Unlock()
	for _, v := range frameData.Boxes {
		if v.ViolationType != proto.EventTypeNone && v.ViolationType != dv.VioEventType {
			continue
		}
		log.Infof("box data: %+v", v)
		// 做适配，算法只会传 frame_idx == 0/2/3
		if v.ViolationFrameIdx == vio.VehicleEmptyIdx {
			v.ViolationFrameIdx = vio.VehicleFirstIdx
		}

		// 第一帧，需要保证有车牌
		if v.ViolationFrameIdx == vio.VehicleFirstIdx && v.PlateContent == "" {
			// too many logs
			// log.Warnf("1st frame has no plate content, ignore")
			continue
		}
		// 第二帧，需要保证既有 ID，又有车牌
		if v.ViolationFrameIdx == vio.VehicleSecondIdx &&
			(v.ID < 0 || v.PlateContent == "") {
			log.Warnf("2nd frame has no plate content or id, ignore")
			continue
		}
		// 第三帧，需要保证有 ID
		if v.ViolationFrameIdx == vio.VehicleThirdIdx && v.ID < 0 {
			log.Warnf("3rd frame has no id, ignore")
			continue
		}

		snap := vio.ViolationEventSnapshot{
			RawData: picData,
			Pts:     v.Pts,
			Label:   v.PlateContent,
			Tz:      now,
		}

		// 保存第一次出现某个车牌时对应的帧信息
		if v.ViolationFrameIdx == vio.VehicleFirstIdx {
			if _, ok := dv.labelSnap[v.PlateContent]; !ok {
				dv.labelSnap[v.PlateContent] = &snap
				log.Debugf("save plate %s for snapshot", v.PlateContent)
			}
			continue
		}

		if vioEvent, ok := dv.objVioEvent[v.ID]; ok && vioEvent != nil {
			vioEvent.EndTime = now
			switch v.ViolationFrameIdx {
			case vio.VehicleFirstIdx:
				// skip
			case vio.VehicleSecondIdx:
				if vioEvent.Snapshots[1] == nil {
					vioEvent.Snapshots[1] = &snap
				}
			case vio.VehicleThirdIdx:
				if vioEvent.Snapshots[2] == nil {
					vioEvent.Snapshots[2] = &snap
				}
			default:
				log.Println("not supported violation frame idx:", v.ViolationFrameIdx)
			}

			if vioEvent.Snapshots[1] != nil && vioEvent.Snapshots[2] != nil {
				vioEvent.EndTime = now
				// 找到出现车牌的 snap 作为第一张
				vioEvent.Snapshots[0] = dv.labelSnap[vioEvent.Snapshots[1].Label]
				if vioEvent.Snapshots[0] == nil {
					log.Warnf("cannot find first snap, obj id: %d, label: %s",
						vioEvent.ID, vioEvent.Snapshots[1].Label)
					continue
				}
				delete(dv.objVioEvent, v.ID)
				delete(dv.labelSnap, vioEvent.Snapshots[1].Label)

				// 设置返回的违法事件
				event = vioEvent
			}
		} else {
			vioEvent := &vio.ViolationEvent{
				ID:            v.ID,
				ViolationType: v.ViolationType,
				StartTime:     now,
				EndTime:       now,
				Snapshots:     make([]*vio.ViolationEventSnapshot, 3, 3), // 创建一个三个元素的数组
			}

			switch v.ViolationFrameIdx {
			case vio.VehicleFirstIdx:
				// skip
			case vio.VehicleSecondIdx:
				vioEvent.Snapshots[1] = &snap
			case vio.VehicleThirdIdx:
				vioEvent.Snapshots[2] = &snap
			default:
				log.Println("not supported violation frame idx:", v.ViolationFrameIdx)
			}
			dv.objVioEvent[v.ID] = vioEvent
			log.Debugf("save vioEvent, v.ID:%d, event:%+v", v.ID, vioEvent)
		}
	}

	return

}

func (dv *DawanxiaozhuanViolation) Release() (err error) {
	dv.logger.Debug("DawanxiaozhuanViolation Release")
	dv.cancel()

	return nil
}

func (dv *DawanxiaozhuanViolation) backgroundWork() {
	xlog := dv.logger
	//start timeout worker
	go func() {
	Loop:
		for {
			select {
			case <-dv.ctx.Done():
				xlog.Println("stop dawanxiaozhuan handler event go routine ")
				break Loop
			case <-time.After(time.Second * time.Duration(dv.cfg.Timeout)):
				now := time.Now()
				removeList := []int{}
				snapRemoveList := []string{}

				dv.mutex.RLock()
				for k, v := range dv.objVioEvent {
					if v == nil || now.Sub(v.StartTime) > time.Second*time.Duration(dv.cfg.Timeout) {
						removeList = append(removeList, k)
						v = nil
					}
				}
				for k, v := range dv.labelSnap {
					if v == nil || now.Sub(v.Tz) > time.Second*time.Duration(dv.cfg.Timeout) {
						snapRemoveList = append(snapRemoveList, k)
					}
				}
				dv.mutex.RUnlock()

				if len(removeList) != 0 {
					dv.mutex.Lock()
					xlog.Println("removing map keys :", removeList)
					for _, k := range removeList {
						delete(dv.objVioEvent, k)
					}
					dv.mutex.Unlock()
				}
				if len(snapRemoveList) != 0 {
					dv.mutex.Lock()
					for _, k := range snapRemoveList {
						delete(dv.labelSnap, k)
					}
					dv.mutex.Unlock()
				}
			}
		}
	}()
}
