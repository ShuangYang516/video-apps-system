package dao

import (
	"context"
	"time"

	mgoutil "github.com/qiniu/db/mgoutil.v3"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"qiniu.com/vas-app/biz/deeper/proto"
	gproto "qiniu.com/vas-app/biz/proto"
)

/* config info */
const (
	mgoSessionPoolLimit = 100
)

/* status info */
const (
	StatusWaiting  = gproto.StatusInit
	StatusDoing    = "doing"
	StatusFinished = gproto.StatusFinished
)

type JobDaoConfig struct {
	Mgo mgoutil.Config `json:"mgo"`

	JobRateSecond    time.Duration `json:"job_rate_second"`
	JobTimeoutSecond time.Duration `json:"job_timeout_second"`
}

type JobDao struct {
	JobDaoConfig
	coll mgoutil.Collection
}

func NewJobDao(config JobDaoConfig) (dao *JobDao, err error) {
	var (
		colls = struct {
			Jobs mgoutil.Collection `coll:"vas_event"`
		}{}
	)
	if config.JobRateSecond <= 0 {
		config.JobRateSecond = 1
	}
	if config.JobTimeoutSecond <= 5 {
		config.JobTimeoutSecond = 5
	}
	sess, err := mgoutil.Open(&colls, &config.Mgo)
	if err != nil {
		return
	}

	sess.SetPoolLimit(mgoSessionPoolLimit)
	colls.Jobs.EnsureIndexes("status")

	dao = &JobDao{
		JobDaoConfig: config,
		coll:         colls.Jobs,
	}
	return
}

func (dao *JobDao) acquire() (job *proto.JobInMgo, err error) {
	coll := dao.coll.CopySession()
	defer coll.CloseSession()
	now := time.Now()
	changeM := mgo.Change{
		Update:    bson.M{"$set": bson.M{"updatedAt": now, "status": StatusDoing}},
		ReturnNew: true,
	}
	job = &proto.JobInMgo{}
	for _, findM := range []bson.M{
		bson.M{"status": StatusDoing, "updatedAt": bson.M{"$lte": now.Add(-dao.JobTimeoutSecond * time.Second)}},
		bson.M{"status": StatusWaiting},
	} {
		_, err = coll.Find(findM).Apply(changeM, job)
		if err == nil {
			// 找到直接返回
			return job, nil
		} else if err != mgo.ErrNotFound {
			// 其他报错直接返回
			return nil, err
		}
		// 未找到则执行下一个查询, 并重置job
	}
	if err == mgo.ErrNotFound {
		return nil, nil
	}
	return
}

func (dao *JobDao) patch(ctx context.Context, job *proto.JobInMgo, deeperInfo *proto.DeeperInfo) (err error) {
	coll := dao.coll.CopySession()
	job.UpdatedAt = time.Now()
	data := bson.M{
		"updatedAt": job.UpdatedAt,
		"status":    job.Status,
	}
	if deeperInfo != nil {
		data["deeperInfo"] = job.DeeperInfo
	}
	err = coll.Update(
		bson.M{"_id": job.ID},
		bson.M{"$set": data},
	)
	coll.CloseSession()
	return
}

func (dao *JobDao) insert(ctx context.Context, job *proto.JobInMgo) (err error) {
	coll := dao.coll.CopySession()
	err = coll.Insert(job)
	coll.CloseSession()
	return
}

func (dao *JobDao) Consume(ctx context.Context) <-chan *proto.JobInMgo {
	jobsChan := make(chan *proto.JobInMgo) // 长度为0, 默认阻塞
	go func(c chan<- *proto.JobInMgo) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			// 尝试找出超时的job
			job, err := dao.acquire()
			if err != nil {
				// 稍等几分钟重试
				time.Sleep(time.Minute)
				job, err = dao.acquire()
			}

			if err != nil {
				close(c)
			} else {
				if job != nil {
					c <- job
				}
				time.Sleep(1.0 / dao.JobRateSecond * time.Second)
			}
		}
	}(jobsChan)
	return jobsChan
}

func (dao *JobDao) Heartbeat(ctx context.Context, job *proto.JobInMgo) (err error) {
	job.Status = StatusDoing
	return dao.patch(ctx, job, nil)
}

func (dao *JobDao) Finish(ctx context.Context, job *proto.JobInMgo) (err error) {
	job.Status = StatusFinished
	return dao.patch(ctx, job, &job.DeeperInfo)
}
