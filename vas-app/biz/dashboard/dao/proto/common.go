package proto

type CommonRes struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type PageData struct {
	Content    interface{} `json:"content"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	TotalPage  int         `json:"total_page"`
	TotalCount int         `json:"total_count"`
}

type CmdArgsReq struct {
	CmdArgs []string `json:"-"`
}
