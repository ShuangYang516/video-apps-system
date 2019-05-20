package service

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gocv.io/x/gocv"

	"qiniu.com/vas-app/util"

	restrpc "github.com/qiniu/http/restrpc.v1"
	"qiniu.com/vas-app/biz/dashboard/dao/proto"
	log "qiniupkg.com/x/log.v7"
)

func returnJson(env *restrpc.Env, bs []byte, err error) {
	if err != nil {
		bs = []byte(fmt.Sprintf(`{"code": 400, "msg": "%s"}`, err.Error()))
	}
	env.W.Header().Set("Content-Type", "application/json")
	env.W.Write(bs)
}

func (s *DashboardService) GetVmsDeviceid() (ret *proto.CommonRes, err error) {
	type Device struct {
		DeviceId string `json:"deviceId"`
	}

	device := Device{
		DeviceId: s.config.Vms.DeviceId,
	}
	return retDefault(device, nil)
}

func (s *DashboardService) GetCameras(req *proto.GetCamerasReq) (ret *proto.CommonRes, err error) {
	data, err := s.getVmsDeviceSubIds()
	if err != nil {
		log.Error(err)
		return retDefault(nil, err)
	}

	if req.Name == "" {
		return retDefault(data, nil)
	}

	result := make([]*proto.SubDeviceRet, 0)

	for _, subDevice := range data {
		if strings.Contains(subDevice.Name, req.Name) {
			result = append(result, subDevice)
		}
	}

	return retDefault(result, nil)
}

func (s *DashboardService) GetVmsDeviceSub(req *proto.GetPageDataReq, env *restrpc.Env) {
	bs, err := s.vmsClient.GetSubDevices(s.config.Vms.DeviceId, req.Page, req.PerPage)
	returnJson(env, bs, err)
}

func (s *DashboardService) GetVmsDeviceChannels(req *proto.CmdArgsReq, env *restrpc.Env) {
	bs, err := s.vmsClient.GetDeviceChannels(s.config.Vms.DeviceId)
	returnJson(env, bs, err)
}

func (s *DashboardService) PostVmsDeviceSub(env *restrpc.Env) {
	bs, err := s.vmsClient.CreateSubDevice(env.Req.Body)
	returnJson(env, bs, err)
}

func (s *DashboardService) PutVmsDeviceSub_(req *proto.CmdArgsReq, env *restrpc.Env) {
	bs, err := s.vmsClient.UpdateSubDevice(req.CmdArgs[0], env.Req.Body)
	returnJson(env, bs, err)
}

func (s *DashboardService) DeleteVmsDeviceSub_(req *proto.CmdArgsReq, env *restrpc.Env) {
	bs, err := s.vmsClient.RemoveSubDevice(req.CmdArgs[0], env.Req.Body)
	returnJson(env, bs, err)
}

func (s *DashboardService) DeleteVmsDeviceSubs(env *restrpc.Env) {
	bs, err := s.vmsClient.RemoveSubDevices(env.Req.Body)
	returnJson(env, bs, err)
}

func (s *DashboardService) GetVmsSubTypeEnums(env *restrpc.Env) {
	bs, err := s.vmsClient.GetTypeEnums()
	returnJson(env, bs, err)
}

func (s *DashboardService) GetVmsSubVendorEnums(env *restrpc.Env) {
	bs, err := s.vmsClient.GetVendorEnums()
	returnJson(env, bs, err)
}

func (s *DashboardService) GetVmsSubDiscoveryProtocol_(req *proto.CmdArgsReq, env *restrpc.Env) {
	bs, err := s.vmsClient.GetDiscoveryProtocol(req.CmdArgs[0])
	returnJson(env, bs, err)
}

func (s *DashboardService) GetVmsSubChanneltypeEnums(env *restrpc.Env) {
	bs, err := s.vmsClient.GetChannelTypeEnums()
	returnJson(env, bs, err)
}

func (s *DashboardService) GetVmsSub_PlayUrl(req *proto.GetPlayUrlReq, env *restrpc.Env) {
	bs, err := s.vmsClient.GetPlayUrl(req.CmdArgs[0], req.StreamId, req.Type)
	returnJson(env, bs, err)
}

func (s *DashboardService) GetVmsSub_Snap(req *proto.CmdArgsReq, env *restrpc.Env) (ret *proto.CommonRes, err error) {
	log.Infof("get snap request %+v", req.CmdArgs[0])
	bs, err := s.vmsClient.GetPlayUrl(req.CmdArgs[0], 0, util.VmsLiveProtocolRTMP)
	if err != nil {
		log.Errorf("fetch play url: %+v", err)
		return retDefault(nil, err)
	}

	var rlt proto.GetPlayUrlResp
	err = json.Unmarshal(bs, &rlt)
	if err != nil {
		log.Errorf("json.Unmarshal(%v): %+v", bs, err)
		return retDefault(nil, err)
	}

	capture, err := gocv.OpenVideoCapture(rlt.Data)
	if err != nil {
		log.Errorf("could not open rtmp %+v: %+v", rlt.Data, err)
		return retDefault(nil, err)
	}
	capture.Grab(25)
	img := gocv.NewMat()
	defer img.Close()
	if ok := capture.Read(&img); !ok {
		log.Errorf("read rtmp %+v: %+v", rlt.Data, ok)
		return retDefault(nil, errors.New("capture rtmp stream failed"))
	}
	b, err := gocv.IMEncode(gocv.JPEGFileExt, img)
	if err != nil {
		log.Errorf("encode image: %+v", err)
		return retDefault(nil, err)
	}
	b64 := base64.StdEncoding.EncodeToString(b)
	// get real url and snap form live stream
	return retDefault(proto.SnapResp{Base64: "data:image/jpeg;base64," + b64, Height: img.Rows(), Width: img.Cols()}, nil)
}

func (s *DashboardService) getVmsDeviceSubIds() (data []*proto.SubDeviceRet, err error) {
	maxPerPage := 1000

	bs, err := s.vmsClient.GetDevice(s.config.Vms.DeviceId)
	if err != nil {
		return
	}

	var deviceResp proto.GetDeviceResp
	err = json.Unmarshal(bs, &deviceResp)
	if err != nil {
		return
	}

	bs, err = s.vmsClient.GetSubDevices(s.config.Vms.DeviceId, 1, maxPerPage)
	if err != nil {
		return
	}

	var subDeviceInfo proto.SearchSubDeviceInfoResp
	err = json.Unmarshal(bs, &subDeviceInfo)
	if err != nil {
		return
	}

	if subDeviceInfo.ErrorCode != 0 {
		err = errors.New(subDeviceInfo.Message)
		return
	}

	data = make([]*proto.SubDeviceRet, 0)

	if len(subDeviceInfo.Data.Content) == 0 {
		return
	}

	for _, content := range subDeviceInfo.Data.Content {
		subDeviceRet := &proto.SubDeviceRet{
			CameraId: fmt.Sprintf("%s_%d", deviceResp.Data.DeviceID, content.Channel),
			Name:     content.Attribute.Name,
		}
		data = append(data, subDeviceRet)
	}

	return
}
