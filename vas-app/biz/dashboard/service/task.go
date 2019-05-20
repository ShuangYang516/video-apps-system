package service

import (
	"errors"
	"fmt"
	"strings"

	"qiniu.com/vas-app/biz/dashboard/dao/proto"
	log "qiniupkg.com/x/log.v7"

	baseProto "qiniu.com/vas-app/biz/proto"
)

func (s *DashboardService) PostTasks(req *baseProto.Task) (ret *proto.CommonRes, err error) {
	log.Debugf("PostTasks:%v", req)

	err = s.taskDao.Insert(*req)
	if err != nil {
		log.Println(err)
		return retDefault(nil, err)
	}
	return retDefault(nil, err)
}

func (s *DashboardService) GetTasks(req *proto.GetTasksReq) (ret *proto.CommonRes, err error) {
	log.Debugf("GetTasks:%v", req)
	query := map[string]interface{}{}

	perPage := req.PerPage
	if perPage <= 0 {
		perPage = 10
	}
	page := req.Page
	if page <= 0 {
		page = 1
	}

	list, total, err := s.taskDao.GetList(query, page, perPage)
	if err != nil {
		log.Println(err)
		return retDefault(nil, err)
	}

	data := proto.PageData{
		Page:       page,
		PerPage:    perPage,
		TotalPage:  (total + perPage - 1) / perPage,
		TotalCount: total,
		Content:    list,
	}

	return retDefault(data, err)
}

func (s *DashboardService) PutTasks_(req *proto.PutTasksReq) (ret *proto.CommonRes, err error) {
	log.Debugf("PutTasks_:%v", req)
	args := req.CmdArgs
	if !validateArgsLength(args, 1) {
		log.Debugf("args:%v", args)
		return nil, errors.New("invalid request args")
	}

	idArg := args[0]
	data, err := obj2Map(req)
	if err != nil {
		log.Error(err)
		return retDefault(nil, err)
	}
	err = s.taskDao.Update(idArg, data)
	if err != nil {
		log.Error(err)
		return retDefault(nil, err)
	}
	retData, err := s.taskDao.GetByID(idArg)
	return retDefault(retData, err)
}

func (s *DashboardService) DeleteTasks_(req *proto.CmdArgsReq) (ret *proto.CommonRes, err error) {
	log.Debugf("DeleteTasks_:%v", req)

	args := req.CmdArgs
	if !validateArgsLength(args, 1) {
		log.Debugf("args:%v", args)
		return retDefault(nil, errors.New("invalid request args"))
	}

	idArg := args[0]
	err = s.taskDao.Delete(idArg)
	return retDefault(nil, err)
}

func (s *DashboardService) PostStartTasks_(req *proto.CmdArgsReq) (ret *proto.CommonRes, err error) {
	log.Debugf("StartTasks_:%v", req)
	return s.updateTasksStatus(req, baseProto.STATUS_ON)
}

func (s *DashboardService) PostStopTasks_(req *proto.CmdArgsReq) (ret *proto.CommonRes, err error) {
	log.Debugf("StopTasks_:%v", req)
	return s.updateTasksStatus(req, baseProto.STATUS_OFF)
}

func (s *DashboardService) updateTasksStatus(req *proto.CmdArgsReq, status string) (ret *proto.CommonRes, err error) {
	args := req.CmdArgs
	if !validateArgsLength(args, 1) {
		log.Debugf("args:%v", args)
		return nil, errors.New("invalid request args")
	}

	idArg := args[0]

	task, err := s.taskDao.GetByID(idArg)
	if err != nil {
		log.Error(err)
		return retDefault(nil, err)
	}

	if task.Status == status {
		return nil, errors.New("task already " + status)
	}
	task.Status = status

	err = s.taskDao.UpdateStatus(task.ID, status)
	if err != nil {
		log.Error(err)
		return retDefault(nil, err)
	}

	return retDefault(task, err)
}

func (s *DashboardService) GetStreams(req *proto.GetStreamsReq) (ret *proto.CommonRes, err error) {
	log.Debugf("GetStreams")
	maxPerPage := 1000

	subIds, err := s.getVmsDeviceSubIds()
	if err != nil {
		log.Error(err)
		return retDefault(nil, err)
	}

	camerasMap := make(map[string]string, 0)
	for _, subId := range subIds {
		if subId == nil {
			continue
		}

		if req.CameraName == "" || strings.Contains(subId.Name, req.CameraName) {
			camerasMap[subId.CameraId] = subId.Name
		}
	}

	tasks, _, err := s.taskDao.GetList(nil, 1, maxPerPage)
	if err != nil {
		log.Println(err)
		return retDefault(nil, err)
	}

	data := make([]*proto.StreamRet, 0)

	for _, task := range tasks {
		if task.Status == baseProto.STATUS_OFF {
			continue
		}

		streamPushAddress := task.Config["stream_push_address"]
		if streamPushAddress == nil {
			continue
		}
		rtmpAddress := streamPushAddress.(string)
		rtmpAddress = strings.TrimPrefix(rtmpAddress, "rtmp://")
		addressArray := strings.Split(rtmpAddress, "/")
		if len(addressArray) != 3 {
			continue
		}

		host := addressArray[0]
		port := "1935"
		if array := strings.Split(host, ":"); len(array) == 2 {
			host = array[0]
			port = array[1]
		}
		app := addressArray[1]
		stream := addressArray[2]

		flvAddress := fmt.Sprintf("http://%s:8088/live?port=%s&app=%s&stream=%s", host, port, app, stream)

		cameraName := ""
		if name, ok := camerasMap[task.CameraID]; ok {
			cameraName = name
		}

		// 筛选条件
		if req.CameraName != "" && cameraName == "" {
			continue
		}

		data = append(data, &proto.StreamRet{
			TaskId:     task.ID,
			Stream:     flvAddress,
			CameraId:   task.CameraID,
			CameraName: cameraName,
		})
	}

	return retDefault(data, err)
}
