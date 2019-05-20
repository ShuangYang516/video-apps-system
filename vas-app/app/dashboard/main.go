package main

import (
	"encoding/json"
	"net/http"
	"runtime"

	"github.com/imdario/mergo"
	mgoutil "github.com/qiniu/db/mgoutil.v3"
	restrpc "github.com/qiniu/http/restrpc.v1"
	servestk "github.com/qiniu/http/servestk.v1"
	log "github.com/qiniu/log.v1"
	"qiniu.com/ava/com/version"
	"qiniu.com/vas-app/biz/dashboard/dao/db"
	"qiniu.com/vas-app/biz/dashboard/service"
	config "qiniupkg.com/x/config.v7"
)

type Config struct {
	MaxProcs   int             `json:"max_procs"`
	BindHost   string          `json:"bind_host"`
	DebugLevel int             `json:"debug_level"`
	DbConfig   *mgoutil.Config `json:"db_config"`
	Service    *service.Config `json:"service"`
}

const appName = "dashboard"

func main() {
	log.Println("version:", version.Version())
	// load default config
	var conf Config
	config.Init("f", "qiniu", appName+".conf")

	var fileConf Config
	if e := config.Load(&fileConf); e != nil {
		log.Fatal("config.Load failed:", e)
	}
	mergo.MergeWithOverwrite(&conf, fileConf)
	buf, _ := json.MarshalIndent(conf, "", "    ")
	log.Printf("loaded conf \n%s", string(buf))

	runtime.GOMAXPROCS(conf.MaxProcs)
	log.SetOutputLevel(conf.DebugLevel)

	err := db.Init(conf.DbConfig, 0)
	if err != nil {
		log.Fatal(err)
	}

	// new Service
	svc, e := service.NewDashboardService(conf.Service)
	if e != nil {
		log.Fatal("failed to create service instance:", e)
	}

	// run Service
	mux := servestk.New(restrpc.DefaultServeMux)
	router := restrpc.Router{
		PatternPrefix: "v1",
		Factory:       restrpc.Factory,
		Mux:           mux,
	}

	router.Register(svc)

	log.Infof("Starting %s..., listen on %s", appName, conf.BindHost)
	log.Fatal("http.ListenAndServe:", http.ListenAndServe(conf.BindHost, mux))
	log.Info(appName + " stopped, process exit")
}
