package proto

import (
	"context"
	"time"
)

const (
	STATUS_ON  = "ON"
	STATUS_OFF = "OFF"
)

type Task struct {
	ID         string `json:"id" bson:"id"`             // 日志唯一ID
	Region     string `json:"region" bson:"region"`     // 区域，xuhui | pudong
	CameraID   string `json:"cameraId" bson:"cameraId"` // 为空则以StreamAddr为准
	StreamAddr string `json:"streamAddr" bson:"streamAddr"`
	Location   string `json:"location" bson:"location"`

	//====status====
	Status     string             `json:"status" bson:"status"`
	CreatedAt  time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt  time.Time          `json:"updatedAt" bson:"updatedAt"`
	Ver        int64              `json:"ver" bson:"ver"`
	Worker     interface{}        `json:"worker" bson:"worker"`
	CancelFunc context.CancelFunc `json:"-" bson:"-"` // for cancel
	//==============

	LastErrorMsg  string                 `json:"lastErrorMsg" bson:"lastErrorMsg"`
	LastErrorTime time.Time              `json:"lastErrorTime" bson:"lastErrorTime"`
	RestartTimes  int                    `json:"restartTimes" bson:"restartTimes"`
	Config        map[string]interface{} `json:"config" bson:"config"`
}
