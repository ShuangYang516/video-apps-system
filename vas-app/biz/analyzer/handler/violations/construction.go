package violations

import (
	"qiniu.com/vas-app/biz/proto"
)

type ConstructionViolationHandler interface {
	// event == nil 表示没有检测出交通违法事件
	Handle(frameData *proto.ConstructionModelData, picData *proto.ImageBody) (event []*ViolationEvent, err error)
	Release() error
}
