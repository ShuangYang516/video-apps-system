package service

import (
	"encoding/json"

	"qiniu.com/vas-app/biz/dashboard/client"
	"qiniu.com/vas-app/biz/dashboard/dao/db"
	"qiniu.com/vas-app/biz/dashboard/dao/proto"
	"qiniu.com/vas-app/biz/export/cmd"
	log "qiniupkg.com/x/log.v7"
)

type Config struct {
	Deeper  *DeeperConfig               `json:"deeper_config"`
	Vms     *VmsConfig                  `json:"vms"`
	Devices map[string]cmd.DeviceConfig `json:"devices"`
}

type DeeperConfig struct {
	FaceScore  float64 `json:"face_score"`
	Similarity float64 `json:"similarity"`
}

type VmsConfig struct {
	client.VmsConfig
	DeviceId string `json:"deviceId"`
}

func defaultDeeperConfig() *DeeperConfig {
	return &DeeperConfig{
		FaceScore:  0.5,
		Similarity: 0.5,
	}
}

type DashboardService struct {
	eventDao db.EventDao
	taskDao  db.TaskDao
	config   *Config

	vmsClient *client.VmsClient
}

func NewDashboardService(conf *Config) (*DashboardService, error) {
	if conf.Deeper == nil {
		conf.Deeper = defaultDeeperConfig()
	}

	eventDao, err := db.NewEventDao()
	if err != nil {
		log.Errorf("failed to create event dao:%v", err)
		return nil, err
	}

	taskDao, err := db.NewTaskDao()
	if err != nil {
		log.Errorf("failed to create task dao:%v", err)
		return nil, err
	}

	vmsClient := client.NewVmsClient(conf.Vms.VmsConfig)

	service := DashboardService{
		config:    conf,
		eventDao:  eventDao,
		taskDao:   taskDao,
		vmsClient: vmsClient,
	}
	return &service, nil
}

func validateArgsLength(args []string, length int) bool {
	return len(args) == length
}

func retDefault(data interface{}, err error) (*proto.CommonRes, error) {
	if err != nil {
		log.Error(err)
		return retError(1, err.Error())
	}
	return retSuccess(data)
}

func retSuccess(data interface{}) (ret *proto.CommonRes, err error) {
	ret = &proto.CommonRes{
		Code: 0,
		Msg:  "ok",
		Data: data,
	}
	return ret, nil
}

func retError(code int, msg string) (ret *proto.CommonRes, err error) {
	ret = &proto.CommonRes{
		Code: code,
		Msg:  msg,
	}
	return ret, nil
}

func obj2Map(obj interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	var ret map[string]interface{}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
