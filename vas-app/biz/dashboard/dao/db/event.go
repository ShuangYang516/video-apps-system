package db

import (
	"time"

	log "qiniupkg.com/x/log.v7"

	mgoutil "github.com/qiniu/db/mgoutil.v3"
	"gopkg.in/mgo.v2/bson"
	"qiniu.com/vas-app/biz/dashboard/dao/proto"
	baseProto "qiniu.com/vas-app/biz/proto"
)

type EventDao interface {
	Insert(v baseProto.TrafficEventMsg) error
	Update(id string, updates map[string]interface{}) error
	Delete(id string) (err error)
	GetByID(id string) (baseProto.TrafficEventMsg, error)
	GetList(query map[string]interface{}, page int, perPage int) ([]baseProto.TrafficEventMsg, int, error)
	CountByHour(query map[string]interface{}) ([]proto.EventsAnalysisHourlyData, error)
	CountByDate(query map[string]interface{}, from, to time.Time) ([]proto.EventsAnalysisDailyData, error)
	CountClassByDate(query map[string]interface{}, from, to time.Time) (interface{}, error)
	CountEventTypeByDate(query map[string]interface{}, from, to time.Time) (interface{}, error)
	CountCameraByDate(query map[string]interface{}, from, to time.Time) (interface{}, error)
	UpdateMarking(id string, marking string) error
}

func NewEventDao() (EventDao, error) {
	c := colls.Event.CopySession()
	defer c.CloseSession()

	// c.EnsureIndexes("eventId", "eventType", "cameraId", "snapshot.class", "indexData.hour", "indexData.date", "createdAt")

	return &EventColl{event: colls.Event}, nil
}

type EventColl struct {
	event mgoutil.Collection
}

func (d *EventColl) Insert(v baseProto.TrafficEventMsg) error {
	c := d.event.CopySession()
	defer c.CloseSession()

	now := time.Now()
	v.CreatedAt = now
	v.UpdatedAt = now
	v.IndexData.Hour = v.CreatedAt.Hour()
	v.IndexData.Date = v.CreatedAt.Format("2006-01-02")
	return c.Insert(v)
}

func (d *EventColl) Update(id string, updates map[string]interface{}) error {
	c := d.event.CopySession()
	defer c.CloseSession()

	now := time.Now()
	updates["updatedAt"] = now

	err := c.UpdateId(bson.ObjectIdHex(id), bson.M{"$set": updates})
	return err
}

func (d *EventColl) Delete(id string) (err error) {
	c := d.event.CopySession()
	defer c.CloseSession()
	err = c.RemoveId(bson.ObjectIdHex(id))
	return err
}

func (d *EventColl) GetByID(id string) (baseProto.TrafficEventMsg, error) {
	c := d.event.CopySession()
	defer c.CloseSession()
	var ret baseProto.TrafficEventMsg
	err := c.Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&ret)
	return ret, err
}

func (d *EventColl) GetList(query map[string]interface{}, page int, perPage int) ([]baseProto.TrafficEventMsg, int, error) {
	c := d.event.CopySession()
	defer c.CloseSession()

	list := make([]baseProto.TrafficEventMsg, 0)
	log.Println("query:", query)

	total, err := c.Find(query).Count()
	if err != nil {
		return list, 0, err
	}

	err = c.Find(query).Skip((page - 1) * perPage).Limit(perPage).Sort("-createdAt").All(&list)
	return list, total, err
}

func (d *EventColl) CountByHour(query map[string]interface{}) ([]proto.EventsAnalysisHourlyData, error) {
	c := d.event.CopySession()
	defer c.CloseSession()
	result := make([]proto.EventsAnalysisHourlyData, 0)

	err := c.Pipe([]bson.M{
		{"$match": bson.M(query)},
		{"$group": bson.M{
			"_id": "$indexData.hour", "count": bson.M{"$sum": 1},
		}},
		{"$project": bson.M{
			"_id":   0,
			"hour":  "$_id",
			"count": "$count",
		}},
	}).All(&result)
	log.Println(result)
	return result, err
}

func (d *EventColl) CountByDate(query map[string]interface{}, from, to time.Time) ([]proto.EventsAnalysisDailyData, error) {
	c := d.event.CopySession()
	defer c.CloseSession()
	result := make([]proto.EventsAnalysisDailyData, 0)
	if query == nil {
		query = map[string]interface{}{}
	}
	q := bson.M(query)
	q["createdAt"] = bson.M{"$gte": from, "$lte": to}

	err := c.Pipe([]bson.M{
		{"$match": q},
		{"$group": bson.M{
			"_id": "$indexData.date", "count": bson.M{"$sum": 1},
		},
		},
		{"$project": bson.M{
			"_id":   0,
			"date":  "$_id",
			"count": "$count",
		}},
	}).All(&result)
	log.Println(result)
	return result, err
}

func (d *EventColl) CountClassByDate(query map[string]interface{}, from, to time.Time) (interface{}, error) {
	c := d.event.CopySession()
	defer c.CloseSession()
	result := make([]bson.M, 0)
	if query == nil {
		query = map[string]interface{}{}
	}
	q := bson.M(query)
	q["createdAt"] = bson.M{"$gte": from, "$lte": to}

	err := c.Pipe([]bson.M{
		{"$match": q},
		{"$unwind": "$snapshot"},
		{"$group": bson.M{
			"_id": "$snapshot.class", "count": bson.M{"$sum": 1},
		},
		},
		{"$project": bson.M{
			"_id":       0,
			"class":     "$_id",
			"eventsNum": "$count",
		}},
	}).All(&result)
	log.Println(result)

	key := "class"
	ret := []bson.M{}
	for _, item := range result {
		// item[key] = "未知"
		_, ok := item[key]
		if !ok {
			continue
		}

		classInt, ok := item[key].(int)
		if !ok {
			continue
		}

		if classInt == 0 {
			continue
		}

		item[key] = baseProto.MapTrafficDetectClass(classInt)
		ret = append(ret, item)
	}

	return ret, err
}

func (d *EventColl) CountEventTypeByDate(query map[string]interface{}, from, to time.Time) (interface{}, error) {
	c := d.event.CopySession()
	defer c.CloseSession()
	result := make([]bson.M, 0)
	if query == nil {
		query = map[string]interface{}{}
	}
	q := bson.M(query)
	q["createdAt"] = bson.M{"$gte": from, "$lte": to}

	err := c.Pipe([]bson.M{
		{"$match": q},
		{"$group": bson.M{
			"_id": "$eventType", "count": bson.M{"$sum": 1},
		},
		},
		{"$project": bson.M{
			"_id":       0,
			"eventType": "$_id",
			"eventsNum": "$count",
		}},
	}).All(&result)
	log.Println(result)

	key := "eventType"
	for _, item := range result {
		// item[key] = "未知"
		_, ok := item[key]
		if !ok {
			continue
		}

		typeInt, ok := item[key].(int)
		if !ok {
			continue
		}

		item[key] = baseProto.MapEventType(typeInt)
	}

	return result, err
}

func (d *EventColl) CountCameraByDate(query map[string]interface{}, from, to time.Time) (interface{}, error) {
	c := d.event.CopySession()
	defer c.CloseSession()
	result := make([]bson.M, 0)
	if query == nil {
		query = map[string]interface{}{}
	}
	q := bson.M(query)
	q["createdAt"] = bson.M{"$gte": from, "$lte": to}

	err := c.Pipe([]bson.M{
		{"$match": q},
		{"$group": bson.M{
			"_id": "$cameraId", "count": bson.M{"$sum": 1},
		},
		},
		{"$project": bson.M{
			"_id":       0,
			"cameraId":  "$_id",
			"eventsNum": "$count",
		}},
	}).All(&result)
	log.Println(result)

	return result, err
}

func (d *EventColl) UpdateMarking(id string, marking string) error {
	c := d.event.CopySession()
	defer c.CloseSession()

	now := time.Now()
	updates := bson.M{
		"mark.marking": marking,
		"updatedAt":    now,
	}

	err := c.Update(bson.D{{"_id", bson.ObjectIdHex(id)}}, bson.M{"$set": updates})
	return err
}
