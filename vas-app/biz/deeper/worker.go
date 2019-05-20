package deeper

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/imdario/mergo"
	xlog "github.com/qiniu/xlog.v1"

	"qiniu.com/vas-app/biz/deeper/proto"
)

type Handler interface {
	Handle(ctx context.Context, input proto.TrafficEvent) (output proto.DeeperInfo)
}

type MQ interface {
	Consume(context.Context) <-chan *proto.JobInMgo
	Heartbeat(context.Context, *proto.JobInMgo) error
	Finish(context.Context, *proto.JobInMgo) error
}

type WorkerConfig struct {
	PoolSize        int           `json:"pool_size"`
	HeartbeatSecond time.Duration `json:"heartbeat_second"`
}

type Worker struct {
	WorkerConfig
	mq       MQ
	handlers []Handler

	heartbeatMap sync.Map
	stopFunc     context.CancelFunc
}

func NewWorker(config WorkerConfig, mq MQ) *Worker {
	return &Worker{
		WorkerConfig: config,
		mq:           mq,
		handlers:     []Handler{},
	}
}

// 注册handler
func (w *Worker) RegisterHandler(handler Handler) {
	w.handlers = append(w.handlers, handler)
	return
}

// 启动worker
func (w *Worker) Start() {
	ctx := context.Background()
	xl := xlog.NewWith("main")

	ctx, w.stopFunc = context.WithCancel(ctx)
	go w.crontHeartbeat(ctx, w.HeartbeatSecond)
	c := w.mq.Consume(ctx)
	for i := 0; i < w.PoolSize; i++ {
		go func(ctx context.Context, c <-chan *proto.JobInMgo, id int) {
			xl.Infof("Worker[%d] started", id)
			xl := xlog.NewWith(fmt.Sprintf("worker %d", id))
			for jobInMgo := range c {
				jobID := jobInMgo.ID.Hex()
				xl.Infof("Handling job %s", jobID)
				w.heartbeatMap.Store(jobID, jobInMgo)
				for _, handler := range w.handlers {
					deeperInfo := handler.Handle(ctx, jobInMgo.TrafficEvent)
					mergo.Merge(&jobInMgo.DeeperInfo, deeperInfo)
				}
				w.heartbeatMap.Delete(jobID)
				w.mq.Finish(ctx, jobInMgo)
				xl.Infof("Finished job %s", jobID)
			}
			xl.Infof("Worker[%d] stopped", id)
		}(ctx, c, i)
	}
	return
}

// 停止worker
func (w *Worker) Stop() {
	xl := xlog.NewWith("main")

	w.stopFunc()

	xl.Infof("Stopping worker....")
	for {
		hasJobPending := false
		w.heartbeatMap.Range(func(k, v interface{}) bool {
			xl.Debugf("Waiting job %s", k)
			if v != nil {
				hasJobPending = true
				return false
			}
			return true
		})
		if hasJobPending {
			time.Sleep(time.Second)
		} else {
			break
		}
	}
	return
}

// 定期发送心跳
func (w *Worker) crontHeartbeat(ctx context.Context, heartbeatGapSecond time.Duration) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		w.heartbeatMap.Range(func(k, v interface{}) bool {
			if v != nil {
				w.mq.Heartbeat(ctx, v.(*proto.JobInMgo))
			}
			return true
		})
		time.Sleep(time.Second * heartbeatGapSecond)
	}
}
