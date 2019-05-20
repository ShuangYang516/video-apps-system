package handler

import (
	"qiniu.com/vas-app/biz/proto"
)

type Handler interface {
	// Init() error
	// CanHandle(modelKey string) bool
	Handle(data interface{}, body *proto.ImageBody) error
	Release() error
}

const (
	NonMotorHandlerModelKey     = "non_motor"
	VehicleHandlerModelKey      = "vehicle"
	WaimaiHandlerModelKey       = "waimai"
	ConstructionHandlerModelKey = "construction"
)
