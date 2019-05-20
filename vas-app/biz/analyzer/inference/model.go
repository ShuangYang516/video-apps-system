package inference

type EvalRequest struct {
	OP      string      `json:"op,omitempty"`
	Cmd     string      `json:"-"`
	Version *string     `json:"-"`
	Data    Resource    `json:"data"`
	Params  interface{} `json:"params,omitempty"`
}

type Resource struct {
	URI       STRING      `json:"uri"`
	Attribute interface{} `json:"attribute,omitempty"`
}

type EvalResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Result  interface{} `json:"result,omitempty"`
}

type GroupEvalRequest struct {
	OP      string      `json:"op,omitempty"`
	Cmd     string      `json:"-"`
	Version *string     `json:"-"`
	Data    []Resource  `json:"data"`
	Params  interface{} `json:"params,omitempty"`
}

type STRING string

func (u *STRING) String() string {
	return string(*u)
}
