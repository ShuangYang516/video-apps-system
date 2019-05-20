package proto

// "cross_line_id":0
// {"waimai":{"boxes":[{"score":0.9991477727890016,"class":3,"id":13,"violation":[5],"pts":[[955.0,563.0],[1054.0,740.0]]}]}}
type NonmotorModelBox struct {
	Score       float64  `json:"score"`
	Class       int      `json:"class"`
	Violation   []int    `json:"violation"`
	Pts         [][2]int `json:"pts"`
	ID          int      `json:"id"`
	CrossLineID int      `json:"cross_line_id"`
}

//模型返回数据结构
type NonmotorModelData struct {
	Boxes []NonmotorModelBox `json:"boxes"`
}
