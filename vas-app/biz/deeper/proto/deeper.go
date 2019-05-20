package proto

import (
	"time"

	"gopkg.in/mgo.v2/bson"

	gproto "qiniu.com/vas-app/biz/proto"
)

type TrafficEvent struct {
	EventType int               `json:"eventType" bson:"eventType"`
	Snapshot  []gproto.Snapshot `json:"snapshot" bson:"snapshot"`
}

type JobInMgo struct {
	ID bson.ObjectId `json:"_id,omitempty" bson:"_id,omitempty"`

	TrafficEvent `json:",inline" bson:",inline"`

	DeeperInfo `json:"deeperInfo,omitempty" bson:"deeperInfo,omitempty"`

	CreatedAt time.Time `json:"createdAt" bson:"createdAt"` // 数据产生时间
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"` // 数据更新时间

	Status string `json:"status" bson:"status" bson:"status" bson:"status"`
}

type DeeperFacePeopleInfo struct {
	PTS         [4][2]int `json:"pts" bson:"pts"`
	Score       float64   `json:"score" bson:"score"`
	Orientation string    `json:"orientation" bson:"orientation"`
	Quality     string    `json:"quality" bson:"quality"`
	FaceUri     []byte    `json:"faceUri" bson:"faceUri"`

	YituPeople []YituSearchResultItem `json:"yituPeople" bson:"yituPeople"` // 保存从依图搜索出来的结果
}

type DeeperFaceInfoArray struct {
	Infos      []DeeperFacePeopleInfo `json:"infos,omitempty" bson:"infos,omitempty"`
	FeatureURI string                 `json:"featureUri,omitempty" bson:"featureUri,omitempty"`
	FaceInfo   string                 `json:"faceInfo,omitempty" bson:"faceInfo,omitempty"`
}

type DeeperFaceInfo struct {
	People    []DeeperFaceInfoArray `json:"people,omitempty" bson:"people,omitempty"` // 顺序和Snapshot保持一致, snapshot与之一一对应
	ErrorInfo string                `json:"errorInfo,omitempty" bson:"errorInfo,omitempty"`
}

type DeeperInfo struct {
	Face DeeperFaceInfo `json:"face,omitempty" bson:"face,omitempty"`
}
