package proto

type GetTasksReq struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Start   int `json:"start"`
	End     int `json:"end"`
}

type PutTasksReq struct {
	CmdArgs    []string               `json:"-"`
	Region     string                 `json:"region,omitempty"`
	CameraID   string                 `json:"cameraId,omitempty"`
	StreamAddr string                 `json:"streamAddr,omitempty"`
	Location   string                 `json:"location,omitempty"`
	Status     string                 `json:"status,omitempty"`
	Ver        int64                  `json:"ver,omitempty"`
	Worker     interface{}            `json:"worker,omitempty"`
	Config     map[string]interface{} `json:"config,omitempty"`
}

type GetStreamsReq struct {
	CameraName string `json:"cameraName"`
}

type StreamRet struct {
	TaskId     string `json:"task_id"`
	Stream     string `json:"stream"`
	CameraId   string `json:"camera_id"`
	CameraName string `json:"camera_name"`
}
