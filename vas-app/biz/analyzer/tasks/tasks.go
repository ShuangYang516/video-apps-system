package tasks

import (
	"context"
	"time"

	mgoutil "github.com/qiniu/db/mgoutil.v3"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"qiniu.com/vas-app/biz/proto"
)

type Tasks interface {
	Gain(ctx context.Context) (*proto.Task, error)
	Touch(ctx context.Context, task *proto.Task) error
	Release(ctx context.Context, task *proto.Task) error
}

func NewTasks(tasksColl *mgoutil.Collection,
	due time.Duration, tag string) Tasks {

	if tasksColl != nil && tasksColl.Collection != nil {
		_ = tasksColl.EnsureIndex(mgo.Index{Key: []string{"id"}, Unique: true})
	}

	return &_tasks{
		tasksColl: tasksColl,
		due:       due,
		tag:       tag,
	}
}

type _tasks struct {
	tasksColl *mgoutil.Collection
	due       time.Duration
	tag       string
}

func (t *_tasks) Gain(ctx context.Context) (*proto.Task, error) {
	coll := t.tasksColl.CopySession()
	defer func() {
		_ = coll.CloseSession()
	}()

	task := proto.Task{}
	err := coll.Find(bson.M{"status": proto.STATUS_ON, "updatedAt": bson.M{
		"$lt": time.Now().Add(-1 * t.due)}}).One(&task)
	if err != nil {
		return nil, err
	}

	task.Worker = t.tag // fmt.Sprintf("%s_%d", util.GetLocalIP(), os.Getpid())
	err = t.Touch(ctx, &task)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (t *_tasks) Touch(ctx context.Context, task *proto.Task) error {
	coll := t.tasksColl.CopySession()
	defer func() {
		_ = coll.CloseSession()
	}()

	selector := bson.M{"id": task.ID, "status": proto.STATUS_ON, "ver": task.Ver}
	task.UpdatedAt = time.Now()
	task.Ver++

	return coll.Update(selector, task)
}

func (t *_tasks) Release(ctx context.Context, task *proto.Task) error {
	coll := t.tasksColl.CopySession()
	defer func() {
		_ = coll.CloseSession()
	}()

	selector := bson.M{"id": task.ID, "status": proto.STATUS_ON, "ver": task.Ver}
	task.UpdatedAt = time.Now()
	task.Ver = 0

	return coll.Update(selector, task)
}
