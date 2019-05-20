package dao

import (
	"context"
	"testing"
	"time"

	mgoutil "github.com/qiniu/db/mgoutil.v3"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2/bson"

	"qiniu.com/vas-app/biz/deeper/proto"
)

var daoConfig = JobDaoConfig{
	Mgo: mgoutil.Config{
		Host: "100.100.62.248:27017",
		DB:   "vas-app_test",
		Mode: "strong",
	},
	JobRateSecond:    2,
	JobTimeoutSecond: 10,
}

// FIXME
func testDaoDB(t *testing.T) {
	assertion := assert.New(t)
	ctx := context.TODO()

	dao, err := NewJobDao(daoConfig)
	if !assertion.NoError(err) {
		return
	}
	jobInMgo := proto.JobInMgo{
		ID:        bson.NewObjectId(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    StatusWaiting,
	}
	err = dao.insert(ctx, &jobInMgo)
	if !assertion.NoError(err) {
		return
	}

	job, err := dao.acquire()
	if !assertion.NoError(err) {
		return
	}
	if !assertion.Equal(StatusDoing, job.Status) {
		return
	}

	job.Status = StatusFinished
	err = dao.patch(ctx, job, &job.DeeperInfo)
	if !assertion.NoError(err) {
		return
	}
	assertion.Equal(StatusFinished, job.Status)
	assertion.Empty(job.DeeperInfo.Face.ErrorInfo)

	job.DeeperInfo.Face = proto.DeeperFaceInfo{
		ErrorInfo: "test bug",
	}
	job.Status = StatusFinished
	err = dao.patch(ctx, job, &job.DeeperInfo)
	if !assertion.NoError(err) {
		return
	}
	assertion.Equal(StatusFinished, job.Status)
	assertion.NotEmpty(job.DeeperInfo.Face.ErrorInfo)
}

// FIXME
func testDaoMQ(t *testing.T) {
	assertion := assert.New(t)
	ctx := context.TODO()

	dao, err := NewJobDao(daoConfig)
	if !assertion.NoError(err) {
		return
	}
	jobInMgo := proto.JobInMgo{
		ID:        bson.NewObjectId(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    StatusWaiting,
	}
	err = dao.insert(ctx, &jobInMgo)
	if !assertion.NoError(err) {
		return
	}

	c := dao.Consume(context.TODO())
	if !assertion.NoError(err) {
		return
	}

	job := <-c
	if !assertion.Equal(StatusDoing, job.Status) {
		return
	}

	dao.Finish(ctx, job)
	if !assertion.NoError(err) {
		return
	}
	assertion.Equal(StatusFinished, job.Status)
}
