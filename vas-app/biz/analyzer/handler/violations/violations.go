package violations

import (
	"time"

	"qiniu.com/vas-app/biz/proto"
)

type ViolationEventSnapshot struct {
	RawData *proto.ImageBody
	Pts     [][2]int // 目标的坐标点
	Label   string   // 车牌
	Tz      time.Time
	// FrameIdx int
	LabelScore  float64 //车牌score
	ObjectClass int     //类别，非机动车用
}

type ViolationEvent struct {
	ID            int // 目标 ID
	ViolationType int
	Snapshots     []*ViolationEventSnapshot
	StartTime     time.Time
	EndTime       time.Time
	FrameCount    int
}

type ViolationHandler interface {
	// event == nil 表示没有检测出交通违法事件
	Handle(frameData *proto.VehicleModelData, picData *proto.ImageBody) (event *ViolationEvent, err error)
	Release() error
}
