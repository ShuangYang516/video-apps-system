package proto

import (
	"image/color"
	"time"

	"gopkg.in/mgo.v2/bson"
)

//TODO 根据算法定义暂定
const (
	TrafficDetectClassNone = iota
	TrafficDetectClassPerson
	TrafficDetectClassNormalNonMotor
	TrafficDetectClassEle
	TrafficDetectClassMeituan
)

// 注意：增加class，请同时更新此map !!!
var TrafficDetectClassMap = map[string][]int{
	EventTypeNonMotor: []int{
		TrafficDetectClassPerson,
		TrafficDetectClassNormalNonMotor,
		TrafficDetectClassEle,
		TrafficDetectClassMeituan,
	},
	EventTypeVehicle: []int{},
}

func MapTrafficDetectClass(t int) string {
	switch t {
	case 0:
		return ""
	case TrafficDetectClassPerson:
		return "行人"
	case TrafficDetectClassNormalNonMotor:
		return "普通车辆"
	case TrafficDetectClassEle:
		return "饿了么"
	case TrafficDetectClassMeituan:
		return "美团"
	}
	return "其他"
}

func MapTrafficDetectClassColor(t int) color.Color {
	switch t {
	case TrafficDetectClassEle:
		return color.NRGBA{0, 0, 255, 255}
	case TrafficDetectClassNormalNonMotor:
		return color.NRGBA{50, 205, 50, 255}
	case TrafficDetectClassMeituan:
		return color.NRGBA{255, 255, 0, 255}
	case TrafficDetectClassNone:
		return color.NRGBA{255, 0, 0, 255}
	}

	return color.NRGBA{50, 205, 50, 255}
}

const (
	StatusInit     = "init"
	StatusFinished = "finished"
)

// 打标操作状态
const (
	MarkingInit    = "init"
	MarkingIllegal = "illegal"
	MarkingDiscard = "discard"
)

// // 检测区域划线
// type CrossLine struct {
// 	X1 int `json:"x1"`
// 	Y1 int `json:"y1"`
// 	X2 int `json:"x2"`
// 	Y2 int `json:"y2"`
// }

// type CrossConfig struct {
// 	ID        int       `json:"id"`
// 	Type      int       `json:"type"`
// 	Direction int       `json:"direction"`
// 	Startline CrossLine `json:"startline"`
// 	Endline   CrossLine `json:"endline"`
// }

type DeeperInfo interface{}

type IndexData struct {
	Hour int    `json:"-" bson:"hour"` //发生时间，24小时制 ，用于统计
	Date string `json:"-" bson:"date"` //发生时间日期 2006-15-04，用于统计
}

type Snapshot struct {
	FeatureURI     string   `json:"featureUri" bson:"featureUri"`   // 特写图片
	SnapshotURI    string   `json:"snapshotUri" bson:"snapshotUri"` // 截帧图片
	SnapshotURIRaw string   `json:"snapshotUriRaw" bson:"snapshotUriRaw"`
	Pts            [][2]int `json:"pts" bson:"pts"`             // 检测目标坐标
	SizeRatio      float64  `json:"sizeRatio" bson:"sizeRatio"` // 缩放比例
	Score          float64  `json:"score" bson:"score"`
	Class          int      `json:"class" bson:"class"`           //类别（非机动车-外卖类型，饿了么/美团等）
	Label          string   `json:"label" bson:"label"`           // 标牌（非机动车-车牌）
	LabelScore     float64  `json:"labelScore" bson:"labelScore"` // 标牌置信度
}

// 打标相关字段
type Mark struct {
	Marking       string `json:"marking" bson:"marking"`             // 打标，init/illegal/discard
	IsClassEdit   bool   `json:"isClassEdit" bson:"isClassEdit"`     // 类别是否被编辑
	IsLabelEdit   bool   `json:"isLabelEdit" bson:"isLabelEdit"`     // 标牌是否被编辑
	DiscardReason string `json:"discardReason" bson:"discardReason"` // 作废原因
}

//交通事件检测消息
type TrafficEventMsg struct {
	ID         bson.ObjectId `json:"id,omitempty" bson:"_id,omitempty"`
	Region     string        `json:"region" bson:"region"`        // 区域，xuhui | pudong
	EventID    string        `json:"eventId" bson:"eventId"`      // 事件ID ，全局唯一
	EventType  int           `json:"eventType" bson:"eventType" ` // 事件类型
	Address    string        `json:"address" bson:"address"`      // 流地址
	CameraID   string        `json:"cameraId" bson:"cameraId"`    // 设备ID
	Snapshot   []Snapshot    `json:"snapshot" bson:"snapshot"`
	Zone       interface{}   `json:"zone" bson:"zone"` //划线或监测区域配置
	DeeperInfo DeeperInfo    `json:"deeperInfo" bson:"deeperInfo"`

	StartTime time.Time `json:"startTime" bson:"startTime"` //事件起始时间
	EndTime   time.Time `json:"endTime" bson:"endTime"`     //事件结束时间
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"` // 数据产生时间
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"` // 数据更新时间

	Status    string    `json:"status" bson:"status"` // 状态，init/finished
	Mark      Mark      `json:"mark" bson:"mark"`
	IndexData IndexData `json:"-" bson:"indexData"` // 索引
}
