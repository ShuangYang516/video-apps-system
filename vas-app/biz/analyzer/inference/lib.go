package inference

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
#include <limits.h>
#include <stdlib.h>
#include <stdint.h>
#include <stdio.h>

typedef void (*INIT_FUNC) ( \
	void*, const int, \
	int*, char**);
typedef void* (*CREATE_FUNC)( \
	void*, const int, \
	int*, char**);
typedef void (*PREPROCESS_FUNC)( \
	const void*, void*, const int, \
	int*, char**, void*, int*);
typedef void (*INFERENCE_FUNC)(\
	const void*, void*, const int, \
	int*, char**, void*, int*);
typedef void (*STREAMREQUEST_FUNC)(\
	const void*, void*,\
	const int,int*,\
	char**);
typedef void (*FRAMEINFERENCE_FUNC)(\
	const void*, void*,\
	int*, int*, char**);
typedef void (*NETRELEASE_FUNC)(\
	const void *);
typedef void (*RESETSTREAM_FUNC)(\
	const void *,const char*);

static uintptr_t pluginOpen(const char* path, char** err) {
	void* h = dlopen(path, RTLD_NOW|RTLD_GLOBAL);
	if (h == NULL) {
		*err = (char*)dlerror();
	}
	return (uintptr_t)h;
}

static INIT_FUNC lookupInitFunc(uintptr_t h, char** err) {
	void* r = dlsym((void*)h, "initEnv");
	if (r == NULL) {
		*err = (char*)dlerror();
	}
	return (INIT_FUNC)r;
}
void callInitFunc(INIT_FUNC f,
	void* args, const int args_size,
	int* code, char** err) {
	f(args, args_size, code, err);
	return;
}

static CREATE_FUNC lookupCreateFunc(uintptr_t h, char** err) {
	void* r = dlsym((void*)h, "createNet");
	if (r == NULL) {
		*err = (char*)dlerror();
	}
	return (CREATE_FUNC)r;
}
void* callCreateFunc(CREATE_FUNC f,
	void* args, const int args_size,
	int* code, char** err) {
	return f(args, args_size, code, err);
}

static PREPROCESS_FUNC lookupPreprocessFunc(uintptr_t h, char** err) {
	void* r = dlsym((void*)h, "netPreprocess");
	if (r == NULL) {
		*err = (char*)dlerror();
	}
	return (PREPROCESS_FUNC)r;
}
void callPreprocessFunc(PREPROCESS_FUNC f,
	const void* net, void* args, const int args_size,
	int* code, char** err, void* ret, int* ret_size) {
	return f(net, args, args_size, code, err, ret, ret_size);
}

static INFERENCE_FUNC lookupInferenceFunc(uintptr_t h, char** err) {
	void* r = dlsym((void*)h, "netInference");
	if (r == NULL) {
		*err = (char*)dlerror();
	}
	return (INFERENCE_FUNC)r;
}
void callInferenceFunc(INFERENCE_FUNC f,
	const void* net, void* args, int args_size,
	int* code, char** err, void* ret, int* ret_size) {
	return f(net, args, args_size, code, err, ret, ret_size);
}

static STREAMREQUEST_FUNC lookupStreamrequestFunc(uintptr_t h, char** err) {
	void* r = dlsym((void*)h, "streamRequest");
	if (r == NULL) {
		*err = (char*)dlerror();
	}
	return (STREAMREQUEST_FUNC)r;
}
void callStreamrequestFunc(STREAMREQUEST_FUNC f,
    const void *ctx, void *request, const int request_size,
    int *code, char **err){
	return f(ctx, request, request_size, code, err);
}

static FRAMEINFERENCE_FUNC lookupFrameinferenceFunc(uintptr_t h, char** err) {
	void* r = dlsym((void*)h, "frameInference");
	if (r == NULL) {
		*err = (char*)dlerror();
	}
	return (FRAMEINFERENCE_FUNC)r;
}
void callFrameinferenceFunc(FRAMEINFERENCE_FUNC f,
    const void *ctx, void *ret, int *ret_size, int *code, char **err){
	return f(ctx, ret, ret_size, code, err);
}

static NETRELEASE_FUNC lookupNetreleaseFunc(uintptr_t h, char** err) {
	void* r = dlsym((void*)h, "netRelease");
	if (r == NULL) {
		*err = (char*)dlerror();
	}
	return (NETRELEASE_FUNC)r;
}
void callNetreleaseFunc(NETRELEASE_FUNC  f,const void *ctx){
	return f(ctx);
}

static RESETSTREAM_FUNC lookupResetStreamFunc(uintptr_t h, char** err) {
	void* r = dlsym((void*)h, "resetStream");
	if (r == NULL) {
		*err = (char*)dlerror();
	}
	return (RESETSTREAM_FUNC)r;
}

void callResetStreamFunc(RESETSTREAM_FUNC  f,const void *ctx, const char* url){
	return f(ctx,url);
}

*/
import "C"
import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"unsafe"

	log "github.com/qiniu/log.v1"

	"github.com/gogo/protobuf/proto"
	lib "qiniu.com/vas-app/biz/analyzer/inference/lib"

	httputil "github.com/qiniu/http/httputil.v1"

	xlog "github.com/qiniu/xlog.v1"
)

var _ Creator = &Lib{}

var libMux = sync.Mutex{}

type Lib struct {
	Workspace string

	Path   string
	plugin C.uintptr_t

	//initFUNC       C.INIT_FUNC
	createFUNC C.CREATE_FUNC
	//preprocessFUNC C.PREPROCESS_FUNC
	//inferenceFUNC  C.INFERENCE_FUNC
	streamRequestFunc  C.STREAMREQUEST_FUNC
	frameInferenceFunc C.FRAMEINFERENCE_FUNC
	netReleaseFunc     C.NETRELEASE_FUNC
	resetStreamFunc    C.RESETSTREAM_FUNC

	initialized bool
	*sync.Mutex
}

func NewLib(ctx context.Context, workspace, path string) (*Lib, error) {
	var (
		xl = xlog.FromContextSafe(ctx)

		cPath    = make([]byte, C.PATH_MAX+1)
		cRelName = make([]byte, len(path)+1)
	)

	copy(cRelName, path)
	if C.realpath(
		(*C.char)(unsafe.Pointer(&cRelName[0])),
		(*C.char)(unsafe.Pointer(&cPath[0])),
	) == nil {
		err := errors.New(`plugin.Open("` + path + `"): realpath failed`)
		xl.Errorf("%v", err)
		return nil, err
	}

	var (
		cErr *C.char
	)
	h := C.pluginOpen((*C.char)(unsafe.Pointer(&cPath[0])), &cErr)
	if h == 0 {
		err := errors.New(`plugin.Open("` + path + `"): ` + C.GoString(cErr))
		xl.Errorf("%v", err)
		return nil, err
	}

	lib := &Lib{Workspace: workspace, Path: path, plugin: h, Mutex: new(sync.Mutex)}
	/*
		lib.initFUNC = C.lookupInitFunc(lib.plugin, &cErr)
		if lib.initFUNC == nil {
			err := fmt.Errorf("function init: %s", C.GoString(cErr))
			xl.Errorf("%v", err)
			return lib, err
		}
	*/
	lib.createFUNC = C.lookupCreateFunc(lib.plugin, &cErr)
	if lib.createFUNC == nil {
		err := fmt.Errorf("function create: %s", C.GoString(cErr))
		xl.Errorf("%v", err)
		return lib, err
	}

	lib.streamRequestFunc = C.lookupStreamrequestFunc(lib.plugin, &cErr)
	if lib.streamRequestFunc == nil {
		err := fmt.Errorf("function streamRequest: %s", C.GoString(cErr))
		xl.Errorf("%v", err)
		return lib, err
	}
	// 这个接口设计不完整，先全部禁止
	// lib.preprocessFUNC = C.lookupPreprocessFunc(lib.plugin, &cErr)
	// if lib.preprocessFUNC == nil {
	// 	err := fmt.Errorf("function preprocess: %s", C.GoString(cErr))
	// 	xl.Errorf("%v", err)
	// 	// return lib, err
	// }
	lib.frameInferenceFunc = C.lookupFrameinferenceFunc(lib.plugin, &cErr)
	if lib.frameInferenceFunc == nil {
		err := fmt.Errorf("function inference: %s", C.GoString(cErr))
		xl.Errorf("%v", err)
		return lib, err
	}

	lib.netReleaseFunc = C.lookupNetreleaseFunc(lib.plugin, &cErr)
	if lib.netReleaseFunc == nil {
		err := fmt.Errorf("function netReleaseFunc: %s", C.GoString(cErr))
		xl.Errorf("%v", err)
		return lib, err
	}

	lib.resetStreamFunc = C.lookupResetStreamFunc(lib.plugin, &cErr)
	if lib.resetStreamFunc == nil {
		err := fmt.Errorf("function resetStreamFunc: %s", C.GoString(cErr))
		xl.Errorf("%v", err)
		return lib, err
	}
	return lib, nil
}

func (_lib *Lib) Create(ctx context.Context, params *CreateParams) (Instance, error) {
	libMux.Lock()
	defer libMux.Unlock()

	var (
		// xl = xlog.FromContextSafe(ctx)

		_params = &lib.CreateParams{
			UseDevice: proto.String(params.UseDevice),
			BatchSize: proto.Int32(int32(params.BatchSize)),
			Env: &lib.CreateParams_Env{
				App:       proto.String(params.App),
				Workspace: proto.String(params.Workspace),
			},
		}
	)
	{
		for name, file := range params.ModelFiles {
			bs, _ := ioutil.ReadFile(file)
			_params.ModelFiles = append(
				_params.ModelFiles,
				&lib.CreateParams_File{
					Name: proto.String(name),
					Body: bs,
				})
		}
	}
	if params.ModelParams != nil {
		bs, _ := json.Marshal(params.ModelParams)
		_params.ModelParams = proto.String(string(bs))
	}
	{
		for name, file := range params.CustomFiles {
			bs, _ := ioutil.ReadFile(file)
			_params.CustomFiles = append(
				_params.CustomFiles,
				&lib.CreateParams_File{
					Name: proto.String(name),
					Body: bs,
				})
		}
	}
	if params.CustomParams != nil {
		bs, _ := json.Marshal(params.CustomParams)
		_params.CustomParams = proto.String(string(bs))
	}

	return newLibInstance(_lib, _params)
}

var _ Instance = &LibInstance{}

type LibInstance struct {
	*Lib
	ctx      unsafe.Pointer
	readBody bool
	ch       chan struct {
		done chan bool
		f    func()
	}
}

func newLibInstance(_lib *Lib, params *lib.CreateParams) (*LibInstance, error) {

	i := &LibInstance{
		readBody: false,
		Lib:      _lib,
		ch: make(chan struct {
			done chan bool
			f    func()
		}),
	}

	var (
		xl   = xlog.NewDummy()
		wait sync.WaitGroup

		createParams, _ = proto.Marshal(params)
		cCode           C.int
		cErr            *C.char
	)

	wait.Add(1)

	go func(ctx context.Context) {
		runtime.LockOSThread()

		_CCreateParams := C.CBytes(createParams)
		defer C.free(_CCreateParams)
		i.ctx = C.callCreateFunc(_lib.createFUNC,
			_CCreateParams, C.int(len(createParams)),
			&cCode, &cErr)
		C.fflush(C.stdout)
		C.fflush(C.stderr)

		wait.Done()

		for r := range i.ch {
			func() {
				defer func() {
					if err := recover(); err != nil {
						// TODO
						log.Println(err)
					}
					r.done <- true
				}()
				r.f()
			}()
		}
	}(xlog.NewContext(
		context.Background(),
		xlog.NewWith("instance."+xlog.GenReqId()),
	))

	wait.Wait()
	if int(cCode) != 0 && int(cCode) != 200 {
		xl.Errorf("%d %s", int(cCode), C.GoString(cErr))
		return nil, httputil.NewError(int(cCode), C.GoString(cErr))
	}

	return i, nil
}

func (i *LibInstance) invoke(f func()) {
	done := make(chan bool)
	i.ch <- struct {
		done chan bool
		f    func()
	}{
		done: done,
		f:    f,
	}
	<-done
}

func (i *LibInstance) StreamRequest(
	ctx context.Context, request *lib.InferenceRequest) error {
	return i.streamRequest(ctx, request)
}

func (i *LibInstance) streamRequest(
	ctx context.Context, request *lib.InferenceRequest) error {
	var (
		xl          = xlog.FromContextSafe(ctx)
		_request, _ = proto.Marshal(request)

		cCode C.int
		cErr  *C.char
	)

	i.invoke(func() {
		_CRequests := C.CBytes(_request)
		defer C.free(_CRequests)
		C.callStreamrequestFunc(
			i.streamRequestFunc, i.ctx,
			_CRequests, C.int(len(_request)),
			&cCode, &cErr,
		)
		C.fflush(C.stdout)
		C.fflush(C.stderr)
	})
	if int(cCode) != 0 && int(cCode) != 200 {
		xl.Errorf("%d %s", int(cCode), C.GoString(cErr))
		code, message := foramtCodeMessage(int(cCode), C.GoString(cErr))
		return httputil.NewError(code, message)
	}
	return nil
}

func (i *LibInstance) NetRelease(
	ctx context.Context) {

	i.invoke(func() {
		C.callNetreleaseFunc(
			i.netReleaseFunc, i.ctx,
		)
		C.fflush(C.stdout)
		C.fflush(C.stderr)
	})
	close(i.ch)
	return
}

func (i *LibInstance) ResetStream(ctx context.Context, streamURL string) {
	i.invoke(func() {
		C.callResetStreamFunc(
			i.resetStreamFunc, i.ctx, C.CString(streamURL),
		)
		C.fflush(C.stdout)
		C.fflush(C.stderr)
	})
	return
}

func (i *LibInstance) FrameInference(ctx context.Context) (*lib.InferenceResponse, error) {

	var (
		xl = xlog.FromContextSafe(ctx)

		retSize C.int
		ret     = C.malloc(1024 * 1024 * 32)
		cCode   C.int
		cErr    *C.char
	)
	defer C.free(ret)

	i.invoke(func() {
		C.callFrameinferenceFunc(
			i.frameInferenceFunc, i.ctx,
			ret, &retSize,
			&cCode, &cErr,
		)
		C.fflush(C.stdout)
		C.fflush(C.stderr)
	})

	if int(cCode) != 0 && int(cCode) != 200 {
		xl.Errorf("%d %s", int(cCode), C.GoString(cErr))
		code, message := foramtCodeMessage(int(cCode), C.GoString(cErr))
		return nil, httputil.NewError(code, message)
	}

	var (
		bs       = C.GoBytes(ret, retSize)
		response = &lib.InferenceResponse{}
	)
	if err := proto.Unmarshal(bs, response); err != nil {
		xl.Errorf("parse inference response failed. %v", err)
		return nil, err
	}

	return response, nil
}

func (i *LibInstance) newFilename(uri string) string {
	sum := sha1.Sum([]byte(strings.Join([]string{xlog.GenReqId(), uri}, "_")))
	return filepath.Join(i.Workspace, hex.EncodeToString(sum[:]))
}
