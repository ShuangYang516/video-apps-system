package construction

import (
	"context"
	"time"

	log "qiniupkg.com/x/log.v7"

	vio "qiniu.com/vas-app/biz/analyzer/handler/violations"
	"qiniu.com/vas-app/biz/proto"
)

type ConstructionChaoshiViolationConfig struct {
	Available []int //超时时间 [hhmm,hhmmh,hhmm,hhmm]
}

// 施工超时
type ConstructionChaoshiViolation struct {
	VioEventType int
	cfg          *ConstructionChaoshiViolationConfig
	ctx          context.Context
	logger       *log.Logger
	cancel       context.CancelFunc
	lastWarning  *time.Time
}

func availableTime(t int) bool {
	h := t / 100
	m := t % 100
	return h >= 0 && h <= 24 && m >= 0 && m < 60
}
func NewConstructionChaoshiViolation(ctx context.Context, cfg *ConstructionChaoshiViolationConfig) *ConstructionChaoshiViolation {
	sv := &ConstructionChaoshiViolation{
		VioEventType: proto.EventTypeConstructionChaoshi,
		ctx:          ctx,
		logger:       log.Std,
		cfg:          cfg,
	}
	if len(cfg.Available)/2 != 0 {
		sv.logger.Warn("config.available not pair, %v", cfg.Available)
		cfg.Available = cfg.Available[:len(cfg.Available)/2*2]
	}
	if cfg == nil || len(cfg.Available) == 0 {
		cfg.Available = []int{0, 2400}
		sv.logger.Warn("config.available empty, [], use %v", cfg.Available)
	}
	for i := 0; i < len(cfg.Available); i++ {
		if !availableTime(cfg.Available[i]) {
			sv.logger.Errorf("config.available error, %v, %v", cfg.Available, cfg.Available[i])
			return nil
		}
	}
	for i := 0; i < len(cfg.Available); i += 2 {
		if cfg.Available[i] > cfg.Available[i+1] {
			sv.logger.Warn("config.available error, %v", cfg.Available)
			return nil
		}
	}
	sv.ctx, sv.cancel = context.WithCancel(ctx)
	return sv
}

func (sv *ConstructionChaoshiViolation) Handle(frameData *proto.ConstructionModelData, picData *proto.ImageBody) (events []*vio.ViolationEvent, err error) {
	now := time.Now()
	inow := now.Hour()*100 + now.Second()
	available := false

	if frameData.State != proto.ConstructionStatusWorking {
		return
	}

	// 一小时报警一次
	if sv.lastWarning != nil && sv.lastWarning.Day() == now.Day() && sv.lastWarning.Hour() == now.Hour() {
		return
	}

	for i := 0; i < len(sv.cfg.Available); i += 2 {
		if inow >= sv.cfg.Available[i] && inow <= sv.cfg.Available[i+1] {
			available = true
			break
		}
	}
	if !available {
		env := &vio.ViolationEvent{
			ViolationType: sv.VioEventType,
			Snapshots: []*vio.ViolationEventSnapshot{&vio.ViolationEventSnapshot{
				RawData: picData,
			}},
			StartTime: now,
			EndTime:   now,
		}
		sv.lastWarning = &now
		events = append(events, env)
		return events, nil
	}
	return
}

func (sv *ConstructionChaoshiViolation) Release() (err error) {
	sv.logger.Debug("ConstructionChaoshiViolation Release")
	sv.cancel()

	return nil
}
