package proto

import (
	baseProto "qiniu.com/vas-app/biz/proto"
)

type PostEventsReq struct {
	baseProto.TrafficEventMsg
}

type GetEventsReq struct {
	Page              int  `json:"page"`
	PerPage           int  `json:"per_page"`
	IsClassFaceFilter bool `json:"isClassFaceFilter"`
	GetEventsCommonReq
}

type GetEventsCommonReq struct {
	Start      int      `json:"start"`
	End        int      `json:"end"`
	EventId    string   `json:"eventId"`
	CameraIDs  []string `json:"cameraIds"`
	Type       string   `json:"type"`
	EventTypes []int    `json:"eventTypes"`
	Classes    []int    `json:"classes"`
	Marking    string   `json:"marking"` // 打标

	// 标牌
	Label      string  `json:"label"`
	HasLabel   int     `json:"hasLabel"` // 1:包含标牌  2:不含标牌
	LabelScore float64 `json:"labelScore"`

	// 人脸
	HasFace    int     `json:"hasFace"` // 1:包含人脸  2:不含人脸
	Name       string  `json:"name"`
	IDCard     string  `json:"idCard"`
	Similarity float64 `json:"similarity"`
}

type GetEventsData struct {
	baseProto.TrafficEventMsg
	EventTypeStr string   `json:"eventTypeStr"`
	ClassStr     []string `json:"classStr"`
	Faces        []People `json:"faces,omitempty"`
}

type People struct {
	Name         string  `json:"name"`
	Nation       string  `json:"nation"`
	Sex          string  `json:"sex"` //  1: Male, 0: Female
	IDCard       string  `json:"idCard"`
	Similarity   float64 `json:"similarity"`
	FaceImageUrl string  `json:"faceImageUrl"`
}

type GetEventsAnalysisHourlyReq struct {
	CameraID  string `json:"cameraId"`
	Date      string `json:"date"`
	Type      string `json:"type"`
	EventType int    `json:"eventType"`
}

type GetEventsAnalysisDailyReq struct {
	CameraID  string `json:"cameraId"`
	Class     int    `json:"class"`
	From      string `json:"from"`
	To        string `json:"to"`
	Type      string `json:"type"`
	EventType int    `json:"eventType"`
}

type GetEventsAnalysisClassReq struct {
	CameraID  string `json:"cameraId"`
	From      string `json:"from"`
	To        string `json:"to"`
	Type      string `json:"type"`
	EventType int    `json:"eventType"`
}

type EventsAnalysisClassData struct {
	Class     int `json:"class"`
	EventsNum int `json:"eventsNum"`
}

type GetEventsAnalysisTypeReq struct {
	CameraID  string `json:"cameraId"`
	From      string `json:"from"`
	To        string `json:"to"`
	Type      string `json:"type"`
	EventType int    `json:"eventType"`
}

type EventsAnalysisTypeData struct {
	EventType int `json:"eventType"`
	EventsNum int `json:"eventsNum"`
}

type GetEventsAnalysisCameraReq struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Type      string `json:"type"`
	EventType int    `json:"eventType"`
}

type EventsAnalysisCameraData struct {
	CameraName string `json:"cameraName"`
	EventsNum  int    `json:"eventsNum"`
}

// type GetEventsExportReq struct {
// 	CameraID       string `json:"cameraId"`
// 	From           string `json:"from"`
// 	To             string `json:"to"`
// 	Class int    `json:"class"`
// }

type EventsAnalysisHourlyData struct {
	Hour  int `json:"hour" bson:"hour"`
	Count int `json:"count" bson:"count"`
}

type EventsAnalysisDailyData struct {
	Date  string `json:"date" bson:"date"`
	Count int    `json:"count" bson:"count"`
}

// type EventsAnalysisClassData struct {
// 	Class string `json:"class" bson:"class"`
// 	EventsNum      int    `json:"eventsNum" bson:"count"`
// }

type GetEventsExportReq struct {
	GetEventsCommonReq
	Limit int      `json:"limit"`
	Ids   []string `json:"ids"`
}

type GetEventsExportPrivateReq struct {
	GetEventsExportReq
}

type MultiRecordsIdsReq struct {
	Ids []string `json:"ids"`
}

type MultiRecordsProcessResult struct {
	Success []string `json:"success,omitempty"`
	Fail    []string `json:"fail,omitempty"`
}

type PutEventsReq struct {
	CmdArgs []string `json:"-"`
	Class   int      `json:"class,omitempty"`
	Label   string   `json:"label,omitempty"`
}

type GetEventTypeEnumsReq struct {
	Type string `json:"type"`
}

type GetEventClassEnumsReq struct {
	Type string `json:"type"`
}

type EnumsResp struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}
