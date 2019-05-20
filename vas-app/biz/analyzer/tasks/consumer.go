package tasks

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	xlog "github.com/qiniu/xlog.v1"
	"qiniu.com/vas-app/biz/proto"
)

const (
	slotDur = time.Second
)

//================================================
// Worker for Consumer
// example:
// func Do(ctx context.Context, task Task) {
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			log.Println("DoTask Done", task.ID)
// 			return
// 		default:
// 		}
// 		/**
// 		 * TODO
// 		 */
// 	}
// }
type Worker interface {
	// task read-only, should not be changed
	Do(ctx context.Context, task proto.Task)
}

//================================================

type Consumer struct {
	Tasks
	Worker
	Quota int32

	// count int32 // TODO: once
	doing int32
}

func (csm *Consumer) Run(ctx context.Context) {

	var (
		xl        = xlog.FromContextSafe(ctx)
		fullCount = 0
		errCount  = 0
	)

	// GainNewTask
	for {

		time.Sleep(slotDur)

		doingNum := atomic.LoadInt32(&csm.doing)
		if doingNum >= csm.Quota {
			if fullCount%10 == 0 {
				xl.Infof("full, %d / %d", doingNum, csm.Quota)
			}
			fullCount++
			continue
		}

		task, err := csm.Tasks.Gain(ctx)
		if err != nil {
			if errCount%10 == 0 {
				xl.Warnf("Gain err, %+v", err)
			}
			errCount++
			continue
		}

		xl.Infof("DoTask: %s", task.ID)
		subCtx, cancelFunc := context.WithCancel(ctx)
		task.CancelFunc = cancelFunc
		atomic.AddInt32(&csm.doing, 1)
		go csm.DoTask(subCtx, task)

		// reset
		fullCount = 0
		errCount = 0
	}
}

// task read-only, should not be changed
func (csm *Consumer) DoTask(ctx context.Context, task *proto.Task) {

	defer func() {
		log.Println("DoTask Cancel", task.ID)
		task.CancelFunc()
		atomic.AddInt32(&csm.doing, -1)
	}()

	go csm.KeepAlive(ctx, task)

	csm.Worker.Do(ctx, *task)
}

// task read-only, should not be changed
func (csm *Consumer) KeepAlive(ctx context.Context, task *proto.Task) {

	defer func() {
		log.Println("Touch Cancel", task.ID)
		task.CancelFunc()
	}()

	for {
		select {
		case <-ctx.Done():
			log.Println("Touch Done", task.ID)
			return
		default:
		}

		err := csm.Tasks.Touch(ctx, task)
		log.Println("Touch", task.ID, err)
		if err != nil {
			return
		}

		time.Sleep(slotDur)
	}
}
