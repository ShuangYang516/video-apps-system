package db

import (
	"time"

	log "qiniupkg.com/x/log.v7"

	mgoutil "github.com/qiniu/db/mgoutil.v3"
	"gopkg.in/mgo.v2/bson"
	baseProto "qiniu.com/vas-app/biz/proto"
)

type TaskDao interface {
	Insert(v baseProto.Task) error
	Update(id string, updates map[string]interface{}) error
	UpdateStatus(id string, status string) error
	Delete(id string) (err error)
	GetByID(id string) (baseProto.Task, error)
	GetList(query map[string]interface{}, page int, perPage int) ([]baseProto.Task, int, error)
}

func NewTaskDao() (TaskDao, error) {
	c := colls.Task.CopySession()
	defer c.CloseSession()

	return &TaskColl{task: colls.Task}, nil
}

type TaskColl struct {
	task mgoutil.Collection
}

func (d *TaskColl) Insert(v baseProto.Task) error {
	c := d.task.CopySession()
	defer c.CloseSession()

	v.Ver = 0

	now := time.Now()
	v.CreatedAt = now
	v.UpdatedAt = now
	return c.Insert(v)
}

func (d *TaskColl) Update(id string, updates map[string]interface{}) error {
	c := d.task.CopySession()
	defer c.CloseSession()

	now := time.Now()
	updates["updatedAt"] = now
	updates["ver"] = 0

	err := c.Update(bson.D{{"id", id}}, bson.M{"$set": updates})
	return err
}

func (d *TaskColl) UpdateStatus(id string, status string) error {
	c := d.task.CopySession()
	defer c.CloseSession()

	now := time.Now()
	updates := bson.M{
		"status": status,
		"ver": 0,
		"updatedAt": now,
	}

	err := c.Update(bson.D{{"id", id}}, bson.M{"$set": updates})
	return err
}

func (d *TaskColl) Delete(id string) (err error) {
	c := d.task.CopySession()
	defer c.CloseSession()
	err = c.Remove(bson.D{{"id", id}})
	return err
}

func (d *TaskColl) GetByID(id string) (baseProto.Task, error) {
	c := d.task.CopySession()
	defer c.CloseSession()
	var ret baseProto.Task
	err := c.Find(bson.M{"id": id}).One(&ret)
	return ret, err
}

func (d *TaskColl) GetList(query map[string]interface{}, page int, perPage int) ([]baseProto.Task, int, error) {
	c := d.task.CopySession()
	defer c.CloseSession()

	list := make([]baseProto.Task, 0)
	log.Println("query:", query)

	total, err := c.Find(query).Count()
	if err != nil {
		return list, 0, err
	}

	err = c.Find(query).Skip((page - 1) * perPage).Limit(perPage).Sort("-createdAt").All(&list)
	return list, total, err
}
