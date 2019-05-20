package dao

import (
	"time"

	mgoutil "github.com/qiniu/db/mgoutil.v3"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"qiniu.com/vas-app/biz/proto"
)

type EventDao interface {
	Insert(v proto.TrafficEventMsg) error
	// Update(id string, updates map[string]interface{}) error
	// Delete(id string) (err error)
	// GetList(query map[string]interface{}, page int, size int) ([]proto.TrafficEventMsg, error)
}

func NewEventDao(eventColl *mgoutil.Collection) EventDao {

	if eventColl != nil && eventColl.Collection != nil {
		_ = eventColl.EnsureIndex(mgo.Index{Key: []string{"eventId"}, Unique: true})
		eventColl.EnsureIndexes("eventType", "eventId", "cameraId", "startTime", "endTime",
			"createdAt", "status", "mark.marking", "indexData.hour", "indexData.date", "snapshot.class", "snapshot.label", "snapshot.labelScore",
			"deeperInfo.face.people.infos.score",
			"deeperInfo.face.people.infos.yituPeople.name",
			"deeperInfo.face.people.infos.yituPeople.idCard",
			"deeperInfo.face.people.infos.yituPeople.similarity",
		)
	}

	return &EventColl{eventColl: eventColl}
}

type EventColl struct {
	eventColl *mgoutil.Collection
}

func (d *EventColl) Insert(v proto.TrafficEventMsg) error {
	c := d.eventColl.CopySession()
	defer c.CloseSession()
	now := time.Now()
	v.CreatedAt = now
	v.UpdatedAt = now
	v.IndexData.Hour = v.StartTime.Hour()
	v.IndexData.Date = v.StartTime.Format("2006-01-02")
	return c.Insert(v)
}

func (d *EventColl) Update(id string, updates map[string]interface{}) error {
	c := d.eventColl.CopySession()
	defer c.CloseSession()

	err := c.UpdateId(bson.ObjectIdHex(id), bson.M{"$set": updates})
	return err
}

func (d *EventColl) Delete(id string) (err error) {
	c := d.eventColl.CopySession()
	defer c.CloseSession()
	err = c.RemoveId(bson.ObjectIdHex(id))
	return err
}

func (d *EventColl) GetList(query map[string]interface{}, page int, size int) ([]proto.TrafficEventMsg, error) {
	c := d.eventColl.CopySession()
	defer c.CloseSession()
	if size < 0 {
		size = 10
	}
	if page < 0 {
		page = 1
	}
	list := make([]proto.TrafficEventMsg, size)
	err := c.Find(query).Skip((page - 1) * size).Limit(size).Sort("-timeStamp").All(&list)
	return list, err
}

type MockEventDao struct{}

func (MockEventDao) Insert(v proto.TrafficEventMsg) error {
	return nil
}
