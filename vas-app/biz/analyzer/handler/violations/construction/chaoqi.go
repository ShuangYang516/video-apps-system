package construction

import (
	"context"
	"time"

	log "qiniupkg.com/x/log.v7"

	vio "qiniu.com/vas-app/biz/analyzer/handler/violations"
	"qiniu.com/vas-app/biz/proto"
)

type ConstructionChaoqiViolationConfig struct {
	EndTime time.Time // 截止时间
}

// 施工超期
type ConstructionChaoqiViolation struct {
	VioEventType int
	cfg          *ConstructionChaoqiViolationConfig
	ctx          context.Context
	logger       *log.Logger
	cancel       context.CancelFunc
	lastWarning  *time.Time
}

func NewConstructionChaoqiViolation(ctx context.Context, cfg *ConstructionChaoqiViolationConfig) *ConstructionChaoqiViolation {
	sv := &ConstructionChaoqiViolation{
		VioEventType: proto.EventTypeConstructionChaoqi,
		ctx:          ctx,
		logger:       log.Std,
		cfg:          cfg,
	}
	sv.ctx, sv.cancel = context.WithCancel(ctx)
	return sv
}

func (sv *ConstructionChaoqiViolation) Handle(frameData *proto.ConstructionModelData, picData *proto.ImageBody) (events []*vio.ViolationEvent, err error) {
	now := time.Now()

	if frameData.State != proto.ConstructionStatusWorking {
		return
	}

	// 一小时报警一次
	if sv.lastWarning != nil && sv.lastWarning.Day() == now.Day() && sv.lastWarning.Hour() == now.Hour() {
		return
	}

	if now.After(sv.cfg.EndTime) {
		env := &vio.ViolationEvent{
			ViolationType: sv.VioEventType,
			Snapshots: []*vio.ViolationEventSnapshot{
				&vio.ViolationEventSnapshot{
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

func (sv *ConstructionChaoqiViolation) Release() (err error) {
	sv.logger.Debug("ConstructionChaoqiViolation Release")
	sv.cancel()

	return nil
}
