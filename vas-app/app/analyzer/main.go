package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	mgoutil "github.com/qiniu/db/mgoutil.v3"
	restrpc "github.com/qiniu/http/restrpc.v1"
	servestk "github.com/qiniu/http/servestk.v1"
	xlog "github.com/qiniu/xlog.v1"
	"qbox.us/cc/config"
	"qiniu.com/vas-app/biz/analyzer"
	"qiniu.com/vas-app/biz/analyzer/dao"
	"qiniu.com/vas-app/biz/analyzer/tasks"
	"qiniu.com/vas-app/util"
	"qiniu.com/vas-app/util/monitor"
)

type Config struct {
	Mgo          *mgoutil.Config  `json:"mgo"`
	WorkerConfig *analyzer.Config `json:"worker_config"`
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	var (
		xl      = xlog.NewWith("main")
		taskTag string
		cfg     Config
	)

	if hostname, err := os.Hostname(); err != nil {
		xl.Warnf("get hostname failed, use ip as tag, err: %v", err)
		taskTag = fmt.Sprintf("%s_%d", util.GetLocalIP(), os.Getpid())
	} else {
		taskTag = hostname
	}

	config.Init("f", "analyzer", "analyzer.conf")
	_ = config.Load(&cfg)
	xl.Infof("config load, %s", util.JsonStr(cfg))

	//================================
	var (
		mux   *servestk.ServeStack
		colls struct {
			VasTasksColl mgoutil.Collection `coll:"vas_tasks"`
			VasEvent     mgoutil.Collection `coll:"vas_event"`
		}
	)

	//================================
	mux = servestk.New(restrpc.NewServeMux())
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	mux.Handle("GET /metrics", promhttp.Handler())
	// pprof
	util.HandlePprof(mux)

	xl.Info("Init mongo...")

	sess, err := mgoutil.Open(&colls, cfg.Mgo)
	if err != nil {
		xl.Errorf("open mongo failed: %+v", err)
		return
	}
	sess.SetPoolLimit(100)
	defer sess.Close()

	//================================

	//================================
	monitor.Init("vas_app", "vas_app")

	var (
		vasTasks    = tasks.NewTasks(&colls.VasTasksColl, time.Second*30, taskTag)
		eventDao    = dao.NewEventDao(&colls.VasEvent)
		vasWoker    = analyzer.NewVasWoker(cfg.WorkerConfig, eventDao)
		vasConsumer = tasks.Consumer{
			Tasks:  vasTasks,
			Worker: vasWoker,
			Quota:  1,
		}
	)

	//================================
	xl.Info("VasConsumer.Run...")

	vasConsumer.Run(context.Background())

}
