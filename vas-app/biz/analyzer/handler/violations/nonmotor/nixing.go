package nonmotor

import (
	"context"

	log "qiniupkg.com/x/log.v7"

	vio "qiniu.com/vas-app/biz/analyzer/handler/violations"
	"qiniu.com/vas-app/biz/proto"
)

type NiXingViolationConfig struct {
	Timeout int //超时时间，超过多少秒没有同一id的目标认为结束
}

// 实线变道
type NiXingViolation struct {
	VioEventType int
	ctx          context.Context
	logger       *log.Logger
	cancel       context.CancelFunc
	impl         *ChuangHongDengViolation // 复用大弯小转的判断逻辑
}

func NewNiXingViolation(ctx context.Context, cfg *NiXingViolationConfig) *NiXingViolation {
	sv := &NiXingViolation{
		VioEventType: proto.EventTypeNonMotorNiXing,
		ctx:          ctx,
		logger:       log.Std,
		impl: NewChuangHongDengViolation(
			ctx, &ChuangHongDengViolationConfig{
				Timeout: cfg.Timeout,
			},
		),
	}
	sv.ctx, sv.cancel = context.WithCancel(ctx)
	// tricky, 将 impl 中的判断 violation type 改掉
	sv.impl.VioEventType = sv.VioEventType

	return sv
}

func (sv *NiXingViolation) Handle(frameData *proto.WaimaiModelData, picData *proto.ImageBody) (events []*vio.ViolationEvent, err error) {
	return sv.impl.Handle(frameData, picData)
}

func (sv *NiXingViolation) Release() (err error) {
	sv.logger.Debug("NiXingViolation Release")
	err = sv.impl.Release()
	if err != nil {
		sv.logger.Warnf("NiXingViolation impl Release")
		return
	}
	sv.cancel()

	return nil
}
