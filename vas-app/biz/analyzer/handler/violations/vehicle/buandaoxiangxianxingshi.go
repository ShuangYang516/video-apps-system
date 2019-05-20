package vehicle

import (
	"context"

	log "qiniupkg.com/x/log.v7"

	vio "qiniu.com/vas-app/biz/analyzer/handler/violations"
	"qiniu.com/vas-app/biz/proto"
)

type BuandaoxiangxianxingshiViolationConfig struct {
	Timeout int //超时时间，超过多少秒没有同一id的目标认为结束
}

// 不按导向线行驶
type BuandaoxiangxianxingshiViolation struct {
	VioEventType int
	ctx          context.Context
	logger       *log.Logger
	cancel       context.CancelFunc
	impl         *DawanxiaozhuanViolation // 复用大弯小转的判断逻辑
}

func NewBuandaoxiangxianxingshiViolation(ctx context.Context, cfg *BuandaoxiangxianxingshiViolationConfig) *BuandaoxiangxianxingshiViolation {
	sv := &BuandaoxiangxianxingshiViolation{
		VioEventType: proto.EventTypeVehicleBuAnDaoXiangXianXiangShi,
		ctx:          ctx,
		logger:       log.Std,
		impl: NewDawanxiaozhuanViolation(
			ctx, &DawanxiaozhuanViolationConfig{
				Timeout: cfg.Timeout,
			},
		),
	}
	sv.ctx, sv.cancel = context.WithCancel(ctx)
	// tricky, 将 impl 中的判断 violation type 改掉
	sv.impl.VioEventType = sv.VioEventType

	return sv
}

func (sv *BuandaoxiangxianxingshiViolation) Handle(frameData *proto.VehicleModelData, picData *proto.ImageBody) (event *vio.ViolationEvent, err error) {
	return sv.impl.Handle(frameData, picData)
}

func (sv *BuandaoxiangxianxingshiViolation) Release() (err error) {
	sv.logger.Debug("BuandaoxiangxianxingshiViolation Release")
	err = sv.impl.Release()
	if err != nil {
		sv.logger.Warnf("BuandaoxiangxianxingshiViolation impl Release")
		return
	}
	sv.cancel()

	return nil
}
