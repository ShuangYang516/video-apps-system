package proto

type VehicleModelBox struct {
	Score             float64  `json:"score"`
	ViolationType     int      `json:"violation_type"`
	ViolationFrameIdx int      `json:"violation_frame_idx"`
	Pts               [][2]int `json:"pts"`
	ID                int      `json:"id"`
	FrameCount        int      `json:"frame_count"`
	CrossLineID       int      `json:"cross_line_id"`
	PlateContent      string   `json:"plate_content"`
}

//模型返回数据结构
type VehicleModelData struct {
	Boxes []VehicleModelBox `json:"boxes"`
}

type ImageBody struct {
	Body   []byte
	Width  int
	Height int
}
