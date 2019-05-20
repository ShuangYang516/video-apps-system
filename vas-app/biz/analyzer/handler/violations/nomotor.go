package violations

import (
	"qiniu.com/vas-app/biz/proto"
)

// type WaimaiEventSnapshot struct {
// 	RawData    *proto.ImageBody
// 	Pts        [][2]int // 目标的坐标点
// 	Label      string   // 车牌
// 	LabelScore float64
// 	Tz         time.Time
// 	// FrameIdx int
// }

// type WaimaiViolationEvent struct {
// 	ID            int // 目标 ID
// 	ViolationType int
// 	Snapshots     []*ViolationEventSnapshot
// 	StartTime     time.Time
// 	EndTime       time.Time
// }

type NonMotorViolationHandler interface {
	// event == nil 表示没有检测出交通违法事件
	Handle(frameData *proto.WaimaiModelData, picData *proto.ImageBody) (event []*ViolationEvent, err error)
	Release() error
}
