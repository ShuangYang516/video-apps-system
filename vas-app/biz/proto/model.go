package proto

//算法模型部分配置，目前有需要的才加上
type ModelConfig struct {
	// BannerDetectOn  bool `json:"banner_detect_on"`
	// CrowdCountOn    bool `json:"crowd_count_on"`
	// HeadCountOn     bool `json:"head_count_on"`
	// HeadTotalOn     bool `json:"head_total_on"`
	// FightClassfiyOn bool `json:"fight_classify_on"`
	NonMotorOn     bool `json:"non_motor_on"`
	VehicleOn      bool `json:"vehicle_on"`
	ConstructionOn bool `json:"construction_on"`

	WaimaiOn         bool     `json:"waimai_on"`
	NoLineConfig     bool     `json:"no_line_config" bson:"no_line_config"`       //是否配置划线，无配线会出发占用非机动车道检测，如有配线是逆行或者闯红灯
	WaimaiDetectZone [][2]int `json:"waimai_detect_zone"`                         //非模型参数，业务解析，外卖检测区域,左上角和右下角坐标
	WaimaiZyjdcdOn   bool     `json:"waimai_zyjdcd_on" bson:"waimai_zyjdcd_on"`   //非模型参数，业务解析，占用机动车道，此时no_line_config=true
	WaimaiCjfqOn     bool     `json:"waimai_cjfq_on" bson:"waimai_cjfq_on"`       //非模型参数，业务解析，闯禁非区，此时no_line_config=true
	EventInitStatus  string   `json:"event_init_status" bson:"event_init_status"` //非模型参数，业务解析，输出事件的初始状态
	// PersonDetectOn  bool `json:"person_detect_on"` //目前为无人机
	// RetentionOn     bool `json:"retention_on"`     //非模型参数，业务解析，人员滞留
	// OverfenceOn     bool `json:"overfence_on"`     //非模型参数，业务解析，翻越栏杆

	// EventUploadCoolTime int      `json:"event_upload_cool_time"` //事件上传冷却时间，模型的配置会覆盖workermgnt的配置
	// FightReportCoolTime int      `json:"fight_report_cool_time"` //打架上报冷却时间
	// DroneReportCoolTime int      `json:"drone_report_cool_time"` //打架上报冷却时间
	// ResetTime           string   `json:"reset_time"`             //非模型参数，业务解析,15:04:00,定时重设时间
	// ResetDay            int      `json:"reset_day"`              //非模型参数，业务解析,0-7

}

//inference 接口返回结构
type InferenceMessageData struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}
