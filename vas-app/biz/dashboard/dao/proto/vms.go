package proto

type GetPageDataReq struct {
	CmdArgs []string `json:"-"`
	Page    int      `json:"page"`
	PerPage int      `json:"per_page"`
}

type EnumsInfo struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type SearchSubDeviceInfoResp struct {
	ErrorCode int    `json:"code"`
	Message   string `json:"msg"`
	Data      struct {
		Content []struct {
			ID        string    `json:"id"`
			Type      EnumsInfo `json:"type"`
			Status    EnumsInfo `json:"status"`
			Channel   int       `json:"channel"`
			CreatedAt int       `json:"created_at"`

			DeviceID string `json:"device_id"`
			UserID   string `json:"user_id"`

			Attribute struct {
				Name string `json:"name"`
			} `json:"attribute"`
		} `json:"content"`
		Page       int `json:"page"`
		PerPage    int `json:"per_page"`
		TotalPage  int `json:"total_page"`
		TotalCount int `json:"total_count"`
	} `json:"data"`
}

type GetDeviceResp struct {
	ErrorCode int    `json:"code"`
	Message   string `json:"msg"`
	Data      struct {
		DeviceID string `json:"device_id"`
	} `json:"data"`
}

type SubDeviceRet struct {
	CameraId string `json:"camera_id"`
	Name     string `json:"name"`
}

type GetPlayUrlResp struct {
	//ErrorCode int    `json:"code"`
	//Message   string `json:"msg"`
	Data string `json:"data"`
}

type SnapResp struct {
	Height int    `json:"height"`
	Width  int    `json:"width"`
	Base64 string `json:"base64"`
}

type GetPlayUrlReq struct {
	CmdArgs  []string
	StreamId int    `json:"stream_id"`
	Type     string `json:"type"`
}

type GetCamerasReq struct {
	Name string `json:"name"`
}
