package service

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	rpcutil "github.com/qiniu/http/rpcutil.v1"
	"golang.org/x/time/rate"
	"gopkg.in/mgo.v2/bson"
	"qiniu.com/vas-app/biz/dashboard/dao/proto"
	"qiniu.com/vas-app/biz/export/cmd"
	"qiniu.com/vas-app/util"
	log "qiniupkg.com/x/log.v7"

	baseProto "qiniu.com/vas-app/biz/proto"
)

func (s *DashboardService) PostEvents(req *proto.PostEventsReq) (ret *proto.CommonRes, err error) {
	log.Debugf("PostEvents:%v", req)

	err = s.eventDao.Insert(req.TrafficEventMsg)
	if err != nil {
		log.Println(err)
		return retDefault(nil, err)
	}
	return retDefault(nil, err)
}

func (s *DashboardService) GetEvents(req *proto.GetEventsReq) (ret *proto.CommonRes, err error) {
	log.Debugf("GetEvents:%v", req)
	query := map[string]interface{}{}
	s.getEventsCommonQuery(req.GetEventsCommonReq, query)

	// 用于徐汇大屏筛选
	if req.IsClassFaceFilter {
		query["$or"] = []bson.M{
			bson.M{
				"snapshot.class": bson.M{
					"$in": []int{baseProto.TrafficDetectClassEle, baseProto.TrafficDetectClassMeituan},
				},
			},
			bson.M{
				"deeperInfo.face.people.infos.score": bson.M{
					"$gte": s.config.Deeper.FaceScore,
				},
				"deeperInfo.face.people.infos.yituPeople.similarity": bson.M{
					"$gte": s.config.Deeper.Similarity,
				},
			},
		}
	}

	perPage := req.PerPage
	if perPage <= 0 {
		perPage = 10
	}
	page := req.Page
	if page <= 0 {
		page = 1
	}

	list, total, err := s.eventDao.GetList(query, page, perPage)
	if err != nil {
		log.Println(err)
		return retDefault(nil, err)
	}

	data := proto.PageData{
		Page:       page,
		PerPage:    perPage,
		TotalPage:  (total + perPage - 1) / perPage,
		TotalCount: total,
	}

	result := make([]proto.GetEventsData, len(list))
	for index, item := range list {
		result[index].TrafficEventMsg = item
		result[index].EventTypeStr = baseProto.MapEventType(item.EventType)

		if item.Snapshot == nil || len(item.Snapshot) == 0 {
			continue
		}

		for _, snapshot := range item.Snapshot {
			result[index].ClassStr = append(result[index].ClassStr, baseProto.MapTrafficDetectClass(snapshot.Class))
		}

		if item.Status == baseProto.StatusFinished {
			result[index].Faces = util.ProcessNonMotorDeepInfo(item.EventType, item.DeeperInfo)
		}
	}
	data.Content = result

	return retDefault(data, err)
}

func (s *DashboardService) PutEvents_(req *proto.PutEventsReq) (ret *proto.CommonRes, err error) {
	log.Debugf("PutEvents_:%v", req)
	args := req.CmdArgs
	if !validateArgsLength(args, 1) {
		log.Debugf("args:%v", args)
		return nil, errors.New("invalid request args")
	}

	idArg := args[0]

	updates := make(map[string]interface{})
	if req.Class != 0 {
		updates["snapshot.0.class"] = req.Class
		updates["mark.isClassEdit"] = true
	}
	if req.Label != "" {
		updates["snapshot.0.label"] = req.Label
		updates["mark.isLabelEdit"] = true
	}

	err = s.eventDao.Update(idArg, updates)
	if err != nil {
		log.Error(err)
		return retDefault(nil, err)
	}
	retData, err := s.eventDao.GetByID(idArg)
	return retDefault(retData, err)
}

func (s *DashboardService) DeleteEvents(req *proto.MultiRecordsIdsReq) (ret *proto.CommonRes, err error) {
	log.Debugf("DeleteEvents:%v", req)

	if len(req.Ids) == 0 {
		return nil, errors.New("invalid request args")
	}

	result := &proto.MultiRecordsProcessResult{
		Success: []string{},
		Fail:    []string{},
	}

	for _, id := range req.Ids {
		err = s.eventDao.Delete(id)
		if err != nil {
			log.Errorf("s.eventDao.Delete(%s): %v", id, err)
			result.Fail = append(result.Fail, id)
			continue
		}

		result.Success = append(result.Success, id)
	}

	return retDefault(result, err)
}

func (s *DashboardService) parseDate(strDate string) (time.Time, error) {
	baseFormat := "2006-01-02"
	if strDate == "" {
		strDate = time.Now().Format(baseFormat)
	}
	t, err := time.ParseInLocation(baseFormat, strDate, time.Local)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func (s *DashboardService) GetEventsAnalysisHourly(req *proto.GetEventsAnalysisHourlyReq) (ret *proto.CommonRes, err error) {
	log.Debugf("GetEventsAnalysisHourly:%v", req)
	t, err := s.parseDate(req.Date)
	if err != nil {
		err = errors.New(fmt.Sprintln("invalid date format:", err))
		log.Println(err)
		return retDefault(nil, err)
	}
	query := map[string]interface{}{
		"indexData.date": t.Format("2006-01-02"),
	}
	if req.CameraID != "" {
		query["cameraId"] = req.CameraID
	}
	typo := req.Type
	if typo == baseProto.EventTypeVehicle {
		query["eventType"] = bson.M{
			"$in": baseProto.EventTypeMap[baseProto.EventTypeVehicle],
		}
	} else if typo == baseProto.EventTypeNonMotor {
		query["eventType"] = bson.M{
			"$in": baseProto.EventTypeMap[baseProto.EventTypeNonMotor],
		}
	}
	if req.EventType != 0 {
		query["eventType"] = req.EventType
	}

	data, err := s.eventDao.CountByHour(query)
	arr := make([]int, 24)
	for _, d := range data {
		if d.Hour < 24 {
			arr[d.Hour] = d.Count
		}
	}
	return retDefault(arr, err)
}

func (s *DashboardService) GetEventsAnalysisDaily(req *proto.GetEventsAnalysisDailyReq) (ret *proto.CommonRes, err error) {
	log.Debugf("GetEventsAnalysisDaily:%v", req)
	from, err := s.parseDate(req.From)
	if err != nil {
		err = errors.New(fmt.Sprintln("invalid date format:", err))
		log.Println(err)
		return retDefault(nil, err)
	}
	to, err := s.parseDate(req.To)
	if err != nil {
		err = errors.New(fmt.Sprintln("invalid date format:", err))
		log.Println(err)
		return retDefault(nil, err)
	}

	to = to.Add(time.Duration(time.Hour * 24))

	query := map[string]interface{}{}

	if req.CameraID != "" {
		query["cameraId"] = req.CameraID
	}
	if req.Class != 0 {
		query["snapshot.class"] = req.Class
	}
	typo := req.Type
	if typo == baseProto.EventTypeVehicle {
		query["eventType"] = bson.M{
			"$in": baseProto.EventTypeMap[baseProto.EventTypeVehicle],
		}
	} else if typo == baseProto.EventTypeNonMotor {
		query["eventType"] = bson.M{
			"$in": baseProto.EventTypeMap[baseProto.EventTypeNonMotor],
		}
	}
	if req.EventType != 0 {
		query["eventType"] = req.EventType
	}

	data, err := s.eventDao.CountByDate(query, from, to)

	m := make(map[string]int)
	for _, d := range data {
		m[d.Date] = d.Count
	}

	return retDefault(m, err)
}

func (s *DashboardService) GetEventsAnalysisClass(req *proto.GetEventsAnalysisClassReq) (ret *proto.CommonRes, err error) {
	log.Debugf("GetEventsAnalysisClass:%v", req)
	from, err := s.parseDate(req.From)
	if err != nil {
		err = errors.New(fmt.Sprintln("invalid date format:", err))
		log.Println(err)
		return retDefault(nil, err)
	}

	to, err := s.parseDate(req.To)
	if err != nil {
		err = errors.New(fmt.Sprintln("invalid date format:", err))
		log.Println(err)
		return retDefault(nil, err)
	}

	to = to.Add(time.Duration(time.Hour * 24))
	log.Println(from, to)
	query := map[string]interface{}{}

	if req.CameraID != "" {
		query["cameraId"] = req.CameraID
	}
	typo := req.Type
	if typo == baseProto.EventTypeVehicle {
		query["eventType"] = bson.M{
			"$in": baseProto.EventTypeMap[baseProto.EventTypeVehicle],
		}
	} else if typo == baseProto.EventTypeNonMotor {
		query["eventType"] = bson.M{
			"$in": baseProto.EventTypeMap[baseProto.EventTypeNonMotor],
		}
	}
	if req.EventType != 0 {
		query["eventType"] = req.EventType
	}

	data, err := s.eventDao.CountClassByDate(query, from, to)

	return retDefault(data, err)
}

func (s *DashboardService) GetEventsAnalysisType(req *proto.GetEventsAnalysisTypeReq) (ret *proto.CommonRes, err error) {
	log.Debugf("GetEventsAnalysisType:%v", req)
	from, err := s.parseDate(req.From)
	if err != nil {
		err = errors.New(fmt.Sprintln("invalid date format:", err))
		log.Println(err)
		return retDefault(nil, err)
	}

	to, err := s.parseDate(req.To)
	if err != nil {
		err = errors.New(fmt.Sprintln("invalid date format:", err))
		log.Println(err)
		return retDefault(nil, err)
	}

	to = to.Add(time.Duration(time.Hour * 24))
	// log.Println(from, to)
	query := map[string]interface{}{}

	if req.CameraID != "" {
		query["cameraId"] = req.CameraID
	}
	typo := req.Type
	if typo == baseProto.EventTypeVehicle {
		query["eventType"] = bson.M{
			"$in": baseProto.EventTypeMap[baseProto.EventTypeVehicle],
		}
	} else if typo == baseProto.EventTypeNonMotor {
		query["eventType"] = bson.M{
			"$in": baseProto.EventTypeMap[baseProto.EventTypeNonMotor],
		}
	}
	if req.EventType != 0 {
		query["eventType"] = req.EventType
	}

	data, err := s.eventDao.CountEventTypeByDate(query, from, to)

	return retDefault(data, err)
}

func (s *DashboardService) GetEventsAnalysisCamera(req *proto.GetEventsAnalysisCameraReq) (ret *proto.CommonRes, err error) {
	log.Debugf("GetEventsAnalysisCamera:%v", req)
	from, err := s.parseDate(req.From)
	if err != nil {
		err = errors.New(fmt.Sprintln("invalid date format:", err))
		log.Println(err)
		return retDefault(nil, err)
	}

	to, err := s.parseDate(req.To)
	if err != nil {
		err = errors.New(fmt.Sprintln("invalid date format:", err))
		log.Println(err)
		return retDefault(nil, err)
	}

	to = to.Add(time.Duration(time.Hour * 24))
	// log.Println(from, to)
	query := map[string]interface{}{}

	typo := req.Type
	if typo == baseProto.EventTypeVehicle {
		query["eventType"] = bson.M{
			"$in": baseProto.EventTypeMap[baseProto.EventTypeVehicle],
		}
	} else if typo == baseProto.EventTypeNonMotor {
		query["eventType"] = bson.M{
			"$in": baseProto.EventTypeMap[baseProto.EventTypeNonMotor],
		}
	}
	if req.EventType != 0 {
		query["eventType"] = req.EventType
	}

	resultInterface, err := s.eventDao.CountCameraByDate(query, from, to)

	result, ok := resultInterface.([]bson.M)
	if !ok {
		err = errors.New(fmt.Sprintf("resultInterface.([]bson.M) error: %v", resultInterface))
		log.Println(err)
		return retDefault(nil, err)
	}

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
		camerasMap[subId.CameraId] = subId.Name
	}

	data := make(cameraData, len(result))

	keyCamera := "cameraId"
	keyNum := "eventsNum"
	for i, item := range result {
		_, ok := item[keyCamera]
		if !ok {
			continue
		}

		_, ok = item[keyNum]
		if !ok {
			continue
		}

		cameraId, ok := item[keyCamera].(string)
		if !ok {
			continue
		}

		data[i] = proto.EventsAnalysisCameraData{
			CameraName: camerasMap[cameraId],
			EventsNum:  item[keyNum].(int),
		}
	}

	sort.Sort(sort.Reverse(data))

	return retDefault(data, err)
}

func (s *DashboardService) GetEventsExport(req *proto.GetEventsExportReq, env *rpcutil.Env) {
	log.Debugf("GetEventsExport:%v", env.Req)
	var (
		list []baseProto.TrafficEventMsg
		err  error
	)

	if req.Ids != nil && req.Ids[0] != "" {
		list = make([]baseProto.TrafficEventMsg, 0)
		for _, id := range req.Ids {
			event, err := s.eventDao.GetByID(id)
			if err != nil {
				log.Printf("s.eventDao.GetGetByIDList(%s): %v\n", id, err)
				continue
			}
			list = append(list, event)
		}
	} else {
		if req.Limit == 0 {
			req.Limit = 100000
		}
		query := map[string]interface{}{}
		s.getEventsCommonQuery(req.GetEventsCommonReq, query)

		list, _, err = s.eventDao.GetList(query, 1, req.Limit)
		if err != nil {
			log.Printf("s.eventDao.GetList(%v): %v\n", query, err)
			env.W.Write([]byte(err.Error()))
			return
		}
	}
	bf := bytes.NewBuffer([]byte{})
	wr := bufio.NewWriter(bf)

	title := strings.Join([]string{"事件类型", "事件ID", "摄像头ID", "发生时间", "抓拍图片", "抓拍背景", "车辆类型"}, ",")

	wr.WriteString(title + "\n")
	for _, v := range list {
		line := strings.Join([]string{
			baseProto.MapEventType(v.EventType),
			v.EventID,
			v.CameraID,
			v.CreatedAt.Format("2006-01-02-15:04:05"),
		},
			",")

		if len(v.Snapshot) > 0 {
			var featureUris, snapshortUris, classes string

			for _, snapshort := range v.Snapshot {
				if snapshort.FeatureURI != "" {
					if featureUris != "" {
						featureUris += ","
					}
					featureUris += snapshort.FeatureURI
				}
				if snapshort.SnapshotURI != "" {
					if snapshortUris != "" {
						snapshortUris += ","
					}
					snapshortUris += snapshort.SnapshotURI
				}
				if snapshort.Class != 0 {
					if classes != "" {
						classes += ","
					}
					classes += baseProto.MapTrafficDetectClass(snapshort.Class)
				}
			}

			line += ","
			line += strings.Join([]string{
				"\"" + featureUris + "\"",
				"\"" + snapshortUris + "\"",
				"\"" + classes + "\"",
			},
				",")

		}
		wr.WriteString(line + "\n")
	}
	wr.Flush()

	filename := "违章数据" + time.Now().Format("20060102150405") + ".csv"
	env.W.Header().Set("Content-Disposition", "attachment;filename="+filename)
	env.W.Header().Set("Content-Type", "text/csv;charset=gbk")
	output, err := util.Utf8ToGbk(bf.Bytes())
	if err != nil {
		log.Println("convert to gbk err:", err)
		output = bf.Bytes()
	}
	env.W.Write(output)
	return
}

// 限流
var limiter = rate.NewLimiter(0.05, 2)

func (s *DashboardService) GetEventsExportEvidence(req *proto.GetEventsExportPrivateReq, env *rpcutil.Env) {
	log.Debugf("GetEventsExportEvidence:%v", env.Req)
	if !limiter.Allow() {
		env.W.Write([]byte("Reach request limiting!"))
		return
	}
	var (
		list []baseProto.TrafficEventMsg
		err  error
	)

	if req.Ids != nil && req.Ids[0] != "" {
		list = make([]baseProto.TrafficEventMsg, 0)
		for _, id := range req.Ids {
			event, err := s.eventDao.GetByID(id)
			if err != nil {
				log.Printf("s.eventDao.GetGetByIDList(%s): %v\n", id, err)
				continue
			}
			list = append(list, event)
		}
	} else {
		if req.Limit == 0 {
			req.Limit = 100000
		}
		query := map[string]interface{}{}
		s.getEventsCommonQuery(req.GetEventsCommonReq, query)

		list, _, err = s.eventDao.GetList(query, 1, req.Limit)
		if err != nil {
			log.Printf("s.eventDao.GetList(%v): %v\n", query, err)
			env.W.Write([]byte(err.Error()))
			return
		}
	}

	filePath := "./"
	parentDir, err := cmd.ProcessEvents(list, s.config.Devices, filePath)
	if err != nil {
		log.Printf("cmd.ProcessEvents(): %v", err)
		env.W.Write([]byte(err.Error()))
		return
	}
	defer func(dir string) {
		os.RemoveAll(dir)
	}(parentDir)

	bf := bytes.NewBuffer([]byte{})

	err = util.Zip(bf, parentDir)
	if err != nil {
		log.Printf("util.Zip(%s): %v", parentDir, err)
		env.W.Write([]byte(err.Error()))
		return
	}

	filename := filepath.Base(parentDir) + ".zip"
	env.W.Header().Set("Content-Disposition", "attachment;filename="+filename)
	env.W.Header().Set("Content-Type", "application/zip")
	env.W.Write(bf.Bytes())
	return
}

func (s *DashboardService) GetEventsExportPrivate(req *proto.GetEventsExportPrivateReq, env *rpcutil.Env) (ret *proto.CommonRes, err error) {
	log.Debugf("GetEventsExportPrivate:%v", env.Req)
	if req.Limit == 0 {
		req.Limit = 100000
	}
	query := map[string]interface{}{}
	s.getEventsCommonQuery(req.GetEventsCommonReq, query)

	if req.Similarity > 0 {
		query["deeperInfo.face.people.infos.yituPeople.similarity"] = bson.M{
			"$gte": req.Similarity,
		}
	}

	list, _, err := s.eventDao.GetList(query, 1, req.Limit)
	if err != nil {
		log.Println(err)
		return retDefault(nil, err)
	}

	return retDefault(list, nil)
}

func (s *DashboardService) PostMarkIllegal(req *proto.MultiRecordsIdsReq) (ret *proto.CommonRes, err error) {
	log.Debugf("MarkIllegal:%v", req)
	return s.updateMarking(req, baseProto.MarkingIllegal)
}

func (s *DashboardService) PostMarkDiscard(req *proto.MultiRecordsIdsReq) (ret *proto.CommonRes, err error) {
	log.Debugf("MarkDiscard:%v", req)
	return s.updateMarking(req, baseProto.MarkingDiscard)
}

func (s *DashboardService) updateMarking(req *proto.MultiRecordsIdsReq, marking string) (ret *proto.CommonRes, err error) {
	if len(req.Ids) == 0 {
		return nil, errors.New("invalid request args")
	}

	result := &proto.MultiRecordsProcessResult{
		Success: []string{},
		Fail:    []string{},
	}

	for _, id := range req.Ids {
		event, err := s.eventDao.GetByID(id)
		if err != nil {
			log.Errorf("s.eventDao.GetByID(%s): %v", id, err)
			result.Fail = append(result.Fail, id)
			continue
		}

		if event.Mark.Marking == marking {
			err = errors.New("event already " + marking)
			log.Error(err)
			result.Fail = append(result.Fail, id)
			continue
		}
		event.Mark.Marking = marking

		err = s.eventDao.UpdateMarking(id, marking)
		if err != nil {
			log.Errorf("s.eventDao.UpdateMarking(%s): %v", id, err)
			result.Fail = append(result.Fail, id)
			continue
		}

		result.Success = append(result.Success, id)
	}

	return retDefault(result, err)
}

func (s *DashboardService) GetEventTypeEnums(req *proto.GetEventTypeEnumsReq) (ret *proto.CommonRes, err error) {
	log.Debugf("GetEventTypeEnums:%v", req)

	data := []*proto.EnumsResp{}
	var eventTypes []int
	typo := req.Type
	if typo == baseProto.EventTypeVehicle {
		eventTypes = baseProto.EventTypeMap[baseProto.EventTypeVehicle]
	} else if typo == baseProto.EventTypeNonMotor {
		eventTypes = baseProto.EventTypeMap[baseProto.EventTypeNonMotor]
	} else {
		eventTypes = baseProto.EventTypeMap[baseProto.EventTypeVehicle]
		eventTypes = append(eventTypes, baseProto.EventTypeMap[baseProto.EventTypeNonMotor]...)
	}

	for _, eventType := range eventTypes {
		data = append(data, &proto.EnumsResp{
			Value: eventType,
			Name:  baseProto.MapEventType(eventType),
		})
	}
	return retDefault(data, nil)
}

func (s *DashboardService) GetEventClassEnums(req *proto.GetEventTypeEnumsReq) (ret *proto.CommonRes, err error) {
	log.Debugf("GetEventClassEnums:%v", req)

	data := []*proto.EnumsResp{}
	var eventClasses []int
	typo := req.Type
	if typo == baseProto.EventTypeVehicle {
		eventClasses = baseProto.TrafficDetectClassMap[baseProto.EventTypeVehicle]
	} else if typo == baseProto.EventTypeNonMotor {
		eventClasses = baseProto.TrafficDetectClassMap[baseProto.EventTypeNonMotor]
	} else {
		eventClasses = baseProto.TrafficDetectClassMap[baseProto.EventTypeVehicle]
		eventClasses = append(eventClasses, baseProto.TrafficDetectClassMap[baseProto.EventTypeNonMotor]...)
	}

	for _, eventClass := range eventClasses {
		data = append(data, &proto.EnumsResp{
			Value: eventClass,
			Name:  baseProto.MapTrafficDetectClass(eventClass),
		})
	}

	return retDefault(data, nil)
}

func (s *DashboardService) getEventsCommonQuery(req proto.GetEventsCommonReq, query map[string]interface{}) {
	timeQuery := map[string]interface{}{}
	if req.Start != 0 {
		timeQuery["$gte"] = time.Unix(int64(req.Start/1000), 0)
	}
	if req.End != 0 {
		timeQuery["$lte"] = time.Unix(int64(req.End/1000), 0)
	}
	if len(timeQuery) > 0 {
		query["createdAt"] = timeQuery
	}
	if req.EventId != "" {
		query["eventId"] = bson.RegEx{Pattern: strings.TrimSpace(req.EventId), Options: "i"}
	}

	if req.CameraIDs != nil && req.CameraIDs[0] != "" {
		query["cameraId"] = bson.M{
			"$in": req.CameraIDs,
		}
	}

	if req.EventTypes != nil && req.EventTypes[0] != 0 {
		query["eventType"] = bson.M{
			"$in": req.EventTypes,
		}
	} else {
		typo := req.Type
		if typo == baseProto.EventTypeVehicle {
			query["eventType"] = bson.M{
				"$in": baseProto.EventTypeMap[baseProto.EventTypeVehicle],
			}
		} else if typo == baseProto.EventTypeNonMotor {
			query["eventType"] = bson.M{
				"$in": baseProto.EventTypeMap[baseProto.EventTypeNonMotor],
			}
		}
	}
	if req.Classes != nil && req.Classes[0] != 0 {
		query["snapshot.class"] = bson.M{
			"$in": req.Classes,
		}
	}
	if req.Marking != "" {
		query["mark.marking"] = req.Marking

		// if req.Marking == baseProto.MarkingInit {
		// 	query["mark.marking"] = bson.M{
		// 		"$nin": []string{baseProto.MarkingIllegal, baseProto.MarkingDiscard},
		// 	}
		// }
	}

	// 标牌
	if req.HasLabel != 0 {
		if req.HasLabel == 1 {
			query["snapshot.0.label"] = bson.M{
				"$exists": true,
				"$ne":     "",
			}
		} else if req.HasLabel == 2 {
			query["snapshot.0.label"] = bson.M{
				"$eq": "",
			}
		}
	}
	if req.HasLabel != 2 && req.Label != "" {
		query["snapshot.label"] = bson.RegEx{Pattern: strings.TrimSpace(req.Label), Options: "i"}
	}
	if req.LabelScore > 0 {
		query["snapshot.labelScore"] = bson.M{
			"$gte": req.LabelScore,
		}
	}

	// 人脸
	if req.HasFace != 0 {
		if req.HasFace == 1 {
			query["deeperInfo.face.people.infos.yituPeople.0"] = bson.M{
				"$exists": true,
			}
		} else if req.HasFace == 2 {
			query["deeperInfo.face.people.infos.yituPeople.0"] = bson.M{
				"$exists": false,
			}
		}
	}
	if req.HasFace != 2 && req.Name != "" {
		query["deeperInfo.face.people.infos.yituPeople.name"] = bson.RegEx{Pattern: strings.TrimSpace(req.Name), Options: "i"}
	}
	if req.HasFace != 2 && req.IDCard != "" {
		query["deeperInfo.face.people.infos.yituPeople.idCard"] = bson.RegEx{Pattern: strings.TrimSpace(req.IDCard), Options: "i"}
	}
	if req.Similarity > 0 {
		query["deeperInfo.face.people.infos.yituPeople.similarity"] = bson.M{
			"$gte": req.Similarity,
		}
	}
}

type cameraData []proto.EventsAnalysisCameraData

func (p cameraData) Len() int {
	return len(p)
}

func (p cameraData) Less(i, j int) bool {
	return p[i].EventsNum < p[j].EventsNum
}

func (p cameraData) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
