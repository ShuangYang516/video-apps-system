package deeper

import (
	"context"

	restrpc "github.com/qiniu/http/restrpc.v1"
)

type DeeperService struct {
}

func NewDeeperService() *DeeperService {
	return &DeeperService{}
}

type ReportReq struct {
}

type ReportResp struct {
}

func (service *DeeperService) GetReport(ctx context.Context, req *ReportReq, env *restrpc.Env) (resp *ReportResp, err error) {
	return
}
