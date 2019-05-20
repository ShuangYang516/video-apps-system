package proto

import (
	"encoding/json"
	"log"
	"time"
)

// 施工区域
type Area []int

// 多个施工区域
type Areas []Area

type ConstructionConfig struct {
	ConstructionOn    bool      `json:"construction_on"`
	AvailableTimes    []int     `json:"available_times"`
	EndTime           time.Time `json:"end_time"`
	ConstructionAreas Areas     `json:"shigongAreas"`
}

func NewConstructionConfigFromMap(data map[string]interface{}) (*ConstructionConfig, error) {
	retv := &ConstructionConfig{}
	buffer, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	err = json.Unmarshal(buffer, &retv)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return retv, nil
}

// 施工状态
type ConstructionStatus int

const (
	ConstructionStatusUndefine   = -1
	ConstructionStatusWorking    = 0
	ConstructionStatusNonworking = 1
	ConstructionStatusUnknow     = 2
)

//模型返回数据结构
type ConstructionModelData struct {
	State ConstructionStatus `json:"state"`
}
