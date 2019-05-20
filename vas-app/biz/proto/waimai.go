package proto

// 非机动车划线
type CrossLine struct {
	X1 int `json:"x1"`
	Y1 int `json:"y1"`
	X2 int `json:"x2"`
	Y2 int `json:"y2"`
}

type CrossConfig struct {
	ID        int       `json:"id"`
	Type      int       `json:"type"`
	Direction int       `json:"direction"`
	Startline CrossLine `json:"startline"`
	Endline   CrossLine `json:"endline"`
}

// "cross_line_id":0
// {"waimai":{"boxes":[{"score":0.9991477727890016,"class":3,"id":13,"violation":[5],"pts":[[955.0,563.0],[1054.0,740.0]]}]}}
type WaimaiModelBox struct {
	Score          float64  `json:"score"`
	Class          int      `json:"class"`
	Violation      []int    `json:"violation"`
	Pts            [][2]int `json:"pts"`
	ID             int      `json:"id"`
	CrossLineID    int      `json:"cross_line_id"`
	CrossSubLineID int      `json:"cross_sub_line_id"`
	Plate          string   `json:"plate"`
	PlateScore     float64  `json:"plate_score"`
}

//模型返回数据结构
type WaimaiModelData struct {
	Boxes []WaimaiModelBox `json:"boxes"`
}

// // 并非全部配置，目前只写了部分上层需要的配置
// type WaimaiModelConfig struct {
// 	WaimaiOn       bool `json:"waimai_on" bson:"waimai_on"`               //如使用必须开启
// 	NoLineConfig   bool `json:"no_line_config" bson:"no_line_config"`     //是否配置划线，无配线会出发占用非机动车道检测，如有配线是逆行或者闯红灯
// 	WaimaiZyjdcdOn bool `json:"waimai_zyjdcd_on" bson:"waimai_zyjdcd_on"` //占用机动车道，此时no_line_config=true
// 	WaimaiCjfqOn   bool `json:"waimai_cjfq_on" bson:"waimai_cjfq_on"`     //闯禁非区，此时no_line_config=true

// }
