package util

import (
	"net/http"
	"net/http/pprof"

	servestk "github.com/qiniu/http/servestk.v1"
)

func HandlePprof(mux *servestk.ServeStack) {
	mux2 := http.NewServeMux()
	mux.SetDefault(mux2)
	mux2.HandleFunc("/debug/pprof/", pprof.Index)
	mux2.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux2.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux2.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux2.HandleFunc("/debug/pprof/trace", pprof.Trace)
}
