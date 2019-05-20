package vehicle

import (
	"context"
	"math"
	"sync"
	"time"

	log "qiniupkg.com/x/log.v7"

	vio "qiniu.com/vas-app/biz/analyzer/handler/violations"
	"qiniu.com/vas-app/biz/proto"
)

const defaultMaxMovePercent = 0.05
const defaultParkingSec = 5

type WanggexiantingcheViolationConfig struct {
	Timeout        int     // 冷却时间，超过多少秒没有同一id的目标认为结束
	ParkingSec     int     // 停车时长
	MaxMovePercent float64 // 移动距离和车对角线的最大比例
}

// 网格线停车
type WanggexiantingcheViolation struct {
	// 需要处理的违法类型
	VioEventType int

	objVioEvent map[int]*vio.ViolationEvent // 目标 ID -> 当前的违法事件

	cfg    *WanggexiantingcheViolationConfig
	logger *log.Logger
	wg     *sync.WaitGroup
	ctx    context.Context
	mutex  sync.RWMutex
	cancel context.CancelFunc

	now *time.Time
}

func NewWanggexiantingcheViolation(ctx context.Context, cfg *WanggexiantingcheViolationConfig) *WanggexiantingcheViolation {
	defaultCfg := *cfg
	if defaultCfg.ParkingSec == 0 {
		defaultCfg.ParkingSec = defaultParkingSec
	}
	if defaultCfg.MaxMovePercent == 0.0 {
		defaultCfg.MaxMovePercent = defaultMaxMovePercent
	}
	v := &WanggexiantingcheViolation{
		VioEventType: proto.EventTypeVehicleWangGeXianTingChe,
		objVioEvent:  make(map[int]*vio.ViolationEvent),
		logger:       log.Std,
		wg:           &sync.WaitGroup{},
		ctx:          ctx,
		cfg:          &defaultCfg,
	}
	v.ctx, v.cancel = context.WithCancel(ctx)
	v.backgroundWork()

	return v
}

func (dv *WanggexiantingcheViolation) Handle(frameData *proto.VehicleModelData, picData *proto.ImageBody) (event *vio.ViolationEvent, err error) {
	log := dv.logger
	now := time.Now()
	if dv.now != nil {
		now = *dv.now
	}
	dv.mutex.Lock()
	defer dv.mutex.Unlock()
	for _, v := range frameData.Boxes {
		if v.ViolationType != proto.EventTypeNone && v.ViolationType != dv.VioEventType {
			continue
		}
		log.Infof("box data: %+v", v)

		// 第一帧，不需要
		if v.ViolationFrameIdx == vio.VehicleFirstIdx {
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

		if vioEvent, ok := dv.objVioEvent[v.ID]; !ok || vioEvent == nil {
			vioEvent = &vio.ViolationEvent{
				ID:            v.ID,
				ViolationType: v.ViolationType,
				StartTime:     now,
				EndTime:       now,
				Snapshots:     make([]*vio.ViolationEventSnapshot, 3, 3), // 创建一个三个元素的数组
			}
			dv.objVioEvent[v.ID] = vioEvent
			log.Debugf("save vioEvent, v.ID:%d, event:%+v", v.ID, vioEvent)
		}
		if vioEvent, ok := dv.objVioEvent[v.ID]; ok {
			switch v.ViolationFrameIdx {
			case vio.VehicleFirstIdx:
				// skip
			case vio.VehicleSecondIdx:
				if vioEvent.Snapshots[0] == nil {
					vioEvent.Snapshots[0] = &snap
					vioEvent.StartTime = now
					vioEvent.EndTime = now
				} else if vioEvent.Snapshots[1] == nil {
					if checkPos(vioEvent.Snapshots[0], &snap, dv.cfg.MaxMovePercent) {
						if now.Sub(vioEvent.Snapshots[0].Tz) >= time.Duration(dv.cfg.ParkingSec)*time.Second*1/5 &&
							now.Sub(vioEvent.Snapshots[0].Tz) <= time.Duration(dv.cfg.ParkingSec)*time.Second*3/5 {
							vioEvent.Snapshots[1] = &snap
						}
					} else {
						vioEvent.Snapshots[0] = &snap
						vioEvent.StartTime = now
						vioEvent.EndTime = now
						vioEvent.Snapshots[1] = nil
					}
				} else if vioEvent.Snapshots[2] == nil {
					if checkPos(vioEvent.Snapshots[0], &snap, dv.cfg.MaxMovePercent) {
						if now.Sub(vioEvent.Snapshots[0].Tz) >= time.Duration(dv.cfg.ParkingSec)*time.Second {
							vioEvent.EndTime = now
							vioEvent.Snapshots[2] = &snap
							retv := *vioEvent
							return &retv, nil
						}
					} else {
						vioEvent.Snapshots[0] = &snap
						vioEvent.StartTime = now
						vioEvent.EndTime = now
						vioEvent.Snapshots[1] = nil
						vioEvent.Snapshots[2] = nil
					}
				} else {
					// only update endtime
					vioEvent.EndTime = now
				}
			case vio.VehicleThirdIdx:
				delete(dv.objVioEvent, vioEvent.ID)
				log.Debugf("remove vioEvent, v.ID:%d, event:%+v", v.ID, vioEvent)
			default:
				log.Warn("not supported violation frame idx:", v.ViolationFrameIdx)
			}

		}
	}

	return
}

func (dv *WanggexiantingcheViolation) Release() (err error) {
	dv.logger.Debug("WanggexiantingcheViolation Release")
	dv.cancel()
	dv.wg.Wait()
	return nil
}

func (dv *WanggexiantingcheViolation) backgroundWork() {
	xlog := dv.logger
	//start timeout worker
	dv.wg.Add(1)
	go func() {
		defer dv.wg.Done()
		for {
			select {
			case <-dv.ctx.Done():
				xlog.Println("stop Wanggexiantingche handler event go routine ")
				return
			case <-time.After(time.Second * time.Duration(dv.cfg.Timeout)):
				now := time.Now()
				removeList := []int{}

				dv.mutex.RLock()
				for k, v := range dv.objVioEvent {
					if v == nil || now.Sub(v.EndTime) > time.Second*time.Duration(dv.cfg.Timeout) {
						removeList = append(removeList, k)
						v = nil
					}
				}
				dv.mutex.RUnlock()

				if len(removeList) != 0 {
					xlog.Println("removing map keys :", removeList)
					dv.mutex.Lock()
					for _, k := range removeList {
						delete(dv.objVioEvent, k)
					}
					dv.mutex.Unlock()
				}
			}
		}
	}()
}

func checkPos(s1, s2 *vio.ViolationEventSnapshot, maxMovePercent float64) bool {
	p1, p2, m1, m2 := [2]int{}, [2]int{}, [2]int{}, [2]int{}
	for _, v := range s1.Pts {
		p1[0] += v[0]
		p1[1] += v[1]
	}
	for _, v := range s2.Pts {
		p2[0] += v[0]
		p2[1] += v[1]
	}
	p1[0] /= len(s1.Pts)
	p1[1] /= len(s1.Pts)
	p2[0] /= len(s2.Pts)
	p2[1] /= len(s2.Pts)

	for _, v := range s1.Pts {
		if m1[0] > v[0] || m1[0] <= 0 {
			m1[0] = v[0]
		}
		if m2[0] < v[0] || m2[0] <= 0 {
			m2[0] = v[0]
		}
		if m1[1] > v[1] || m1[1] <= 0 {
			m1[1] = v[1]
		}
		if m2[1] < v[1] || m2[1] <= 0 {
			m2[1] = v[1]
		}
	}
	crossLen := math.Sqrt(float64((m1[0]-m2[0])*(m1[0]-m2[0])) + float64((m1[1]-m2[1])*(m1[1]-m2[1])))
	crossLen = math.Max(crossLen, 1)
	moveLen := math.Sqrt(float64((p1[0]-p2[0])*(p1[0]-p2[0])) + float64((p1[1]-p2[1])*(p1[1]-p2[1])))
	return moveLen/crossLen < maxMovePercent
}
