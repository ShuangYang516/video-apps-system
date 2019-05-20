package main

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"qbox.us/cc/config"

	restrpc "github.com/qiniu/http/restrpc.v1"
	servestk "github.com/qiniu/http/servestk.v1"
	xlog "github.com/qiniu/xlog.v1"
	jsonlog "qbox.us/http/audit/jsonlog.v3"

	"qiniu.com/vas-app/biz/deeper"
	"qiniu.com/vas-app/biz/deeper/client"
	"qiniu.com/vas-app/biz/deeper/dao"
	"qiniu.com/vas-app/biz/deeper/handler"
	"qiniu.com/vas-app/util"
)

type Config struct {
	AuditLog jsonlog.Config `json:"audit_log"`
	LogLevel int            `json:"log_level"`

	JobDao dao.JobDaoConfig `json:"mq"`
	Client struct {
		Yitu client.YituClientConfig `json:"yitu"`
	} `json:"client"`
	Handler struct {
		FaceSearchHandler *handler.FaceSearchHandlerConfig `json:"face_search"`
	} `json:"handler"`
	Worker deeper.WorkerConfig `json:"worker"`
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	var (
		xl = xlog.NewWith("main")
	)

	var cfg Config
	config.Init("f", "deeper", "deeper.conf")
	_ = config.Load(&cfg)
	xlog.SetOutputLevel(cfg.LogLevel)

	//================================
	xl.Info("Init AuditLog...")
	al, logf, err := jsonlog.Open("DEEP_ANALYZER", &cfg.AuditLog, nil)
	if err != nil {
		xl.Fatalf("Failed to open jsonlog, err: %v", err)
	}
	defer logf.Close()
	mux := servestk.New(restrpc.NewServeMux(), al.Handler)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok deeper"))
	})
	util.HandlePprof(mux)

	//================================
	xl.Info("Init Service...")
	srv, worker, err := initDeeper(cfg)
	if err != nil {
		xl.Fatalf("Failed to init deeper service & worker, err: %v", err)
	}
	worker.Start()

	//================================
	xl.Info("Register Router...")
	{
		router := restrpc.Router{
			PatternPrefix: "v1",
			Factory:       restrpc.Factory,
			Mux:           mux,
		}
		router.Register(srv)
	}

	//================================
	envPort := os.Getenv("PORT_HTTP")
	if envPort == "" {
		envPort = "80"
	}

	//================================
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		xl.Info("Shutting down server...")
		time.AfterFunc(20*time.Second, func() {
			os.Exit(0)
		})
		worker.Stop()
		os.Exit(0)
	}()

	if err := http.ListenAndServe(
		net.JoinHostPort("0.0.0.0", envPort),
		mux); err != nil {
		xl.Fatal("http.ListenAndServe", err)
	}
}

func initDeeper(cfg Config) (srv *deeper.DeeperService, worker *deeper.Worker, err error) {
	// client
	yituClient := client.NewYituClient(cfg.Client.Yitu)

	// dao
	dao, err := dao.NewJobDao(cfg.JobDao)
	if err != nil {
		return
	}

	// worker
	worker = deeper.NewWorker(cfg.Worker, dao)
	if cfg.Handler.FaceSearchHandler != nil {
		faceSearchHandler := handler.NewFaceSearchHandler(*cfg.Handler.FaceSearchHandler, nil, yituClient)
		worker.RegisterHandler(faceSearchHandler)
	}

	//  service
	srv = deeper.NewDeeperService()

	return
}
