package client

import (
	baseProto "qiniu.com/vas-app/biz/proto"
)

type GetEventReq struct {
	Start int `json:"start"`
	End   int `json:"end"`

	Type string `json:"type"`

	CameraId  string `json:"cameraId"`
	Class     int    `json:"class"`
	EventType int    `json:"eventType"`
	Limit     int    `json:"limit"`

	Marking string `json:"marking"` // 打标

	// 标牌
	HasLabel   int     `json:"hasLabel"` // 1:包含标牌  2:不含标牌
	LabelScore float64 `json:"labelScore"`

	// 人脸
	HasFace    int     `json:"hasFace"` // 1:包含人脸  2:不含人脸
	Similarity float64 `json:"similarity"`
}

type GetEventResp struct {
	Code int                         `json:"code"`
	Msg  string                      `json:"msg"`
	Data []baseProto.TrafficEventMsg `json:"data"`
}
