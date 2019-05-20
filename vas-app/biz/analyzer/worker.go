package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"qiniu.com/vas-app/biz/analyzer/dao"
	"qiniu.com/vas-app/biz/analyzer/handler"
	ilib "qiniu.com/vas-app/biz/analyzer/inference/lib"
	"qiniu.com/vas-app/biz/proto"

	log "github.com/qiniu/log.v1"
	"qiniu.com/vas-app/biz/analyzer/inference"
)

type Config struct {
	VmsHost     string `json:"vms_host"`
	DataDir     string `json:"data_dir"`
	TrackSOPath string `json:"track_so_path"`

	GpuNum       int `json:"gpu_num"`
	WorkerPerGpu int `json:"worker_per_gpu"`

	StreamInferenceRetryLimit    int `json:"stream_inference_retry_limit"`    //流推理出错重试限制
	StreamInferenceRetryInterval int `json:"stream_inference_retry_interval"` //流推理出错重试间隔
	StreamRetryLimit             int `json:"stream_retry_limit"`              //流中断结束重试限制
	StreamRestartInterval        int `json:"stream_restart_interval"`         //流中断结束重试间隔

	Fileserver string `json:"fileserver"` //文件服务器
}

type VasWorker struct {
	eventDao   dao.EventDao
	config     *Config
	handlerMap map[string]handler.Handler
}

func NewVasWoker(config *Config, eventDao dao.EventDao) *VasWorker {

	vw := VasWorker{
		eventDao:   eventDao,
		config:     config,
		handlerMap: make(map[string]handler.Handler),
	}
	return &vw
}

func (w *VasWorker) getStreamAddress(cameraId string) (string, error) {

	strs := strings.Split(cameraId, "_")
	if len(strs) < 2 {
		return "", fmt.Errorf("GetStreamAddress err, cameraId invalid, %s", cameraId)
	}

	url := fmt.Sprintf("%s/live?device_id=%s&channel_id=%s&stream_id=0&type=rtmp", w.config.VmsHost, strs[0], strs[1])
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		err := fmt.Errorf("getStreamAddress err, %d", resp.StatusCode)
		return "", err
	}

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var addrInfo struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			PlayUrl string `json:"play_url"`
		} `json:"data"`
	}
	err = json.Unmarshal(bs, &addrInfo)
	if err != nil {
		return "", err
	}

	if addrInfo.Data.PlayUrl == "" {
		return "", fmt.Errorf("GetStreamAddress err, addr empty, %s", cameraId)
	}

	return addrInfo.Data.PlayUrl, nil
}

// Worker主入口
func (w *VasWorker) Do(ctx context.Context, task proto.Task) {
	xlog := log.New(os.Stderr, fmt.Sprintf("[%s] ", task.ID), log.Ldefault)
	ctx = context.WithValue(ctx, "xlog", xlog)

	// init lib
	lib, err := inference.NewLib(ctx, w.config.DataDir, w.config.TrackSOPath)
	if err != nil {
		xlog.Errorf("NewLib err, %+v", err)
		return
	}

	xlog.Infof("LibCreate...")
	instance, err := lib.Create(ctx, &inference.CreateParams{
		CustomParams: task.Config,
	})
	if err != nil {
		xlog.Errorf("Create instance err, %+v", err)
		return
	}
	xlog.Infof("LibCreate End")

	defer func() {
		instance.NetRelease(ctx)
		xlog.Infof("NetRelease, %+v", task)
	}()

	// get addr by cameraId
	if task.StreamAddr == "" {
		addr, err := w.getStreamAddress(task.CameraID)
		if err != nil {
			xlog.Errorf("get stream addr err, %+v", err)
			return
		}
		task.StreamAddr = addr
	}
	xlog.Infof("task: %+v", task)

	xlog.Infof("StreamRequest...")
	err = instance.StreamRequest(ctx, &ilib.InferenceRequest{
		Data: &ilib.InferenceRequest_RequestData{
			Uri: &task.StreamAddr,
		},
	})
	if err != nil {
		xlog.Errorf("StreamRequest err, %+v", err)
		return
	}
	xlog.Infof("StreamRequest End")

	w.initHandler(ctx, task)
	defer w.releaseHandler(task)

	inferenceRetryLimit := w.config.StreamInferenceRetryLimit

	for {
		select {
		case <-ctx.Done():
			xlog.Infof("task canceled, %+v", task)
			return
		default:
		}

		// Do Eval
		startTime := time.Now()
		xlog.Infof("FrameInference...")
		resp, err := instance.FrameInference(ctx)
		inferenceDuration := time.Since(startTime).String()
		startTime2 := time.Now()

		if err != nil {
			xlog.Errorf("FrameInference err, %+v", err)
			if inferenceRetryLimit <= 0 {
				return
			}
			instance.ResetStream(ctx, task.StreamAddr)
			xlog.Infof("sleep for %d seconds ... %+v", w.config.StreamInferenceRetryInterval, task)
			inferenceRetryLimit--
			time.Sleep(time.Second * time.Duration(w.config.StreamInferenceRetryInterval))
			continue
		}

		inferenceRetryLimit = w.config.StreamInferenceRetryLimit
		xlog.Infof("FrameInference result, %+v", *resp.Result)

		var respMap map[string]interface{}
		err = json.Unmarshal([]byte(*resp.Result), &respMap)
		if err != nil {
			xlog.Errorf("Unmarshal err, %+v", err)
			continue
		}
		if len(respMap) == 0 {
			xlog.Infof("result is empty, skip")
			continue
		}

		var imageBody *proto.ImageBody

		if len(resp.Body) > 0 && resp.Message != nil {
			var msg proto.InferenceMessageData
			err = json.Unmarshal([]byte(*resp.Message), &msg)
			if err != nil {
				xlog.Error(err)
			} else {
				imageBody = &proto.ImageBody{
					Body:   resp.Body,
					Height: msg.Height,
					Width:  msg.Width,
				}
			}
		}

		for k, v := range respMap {
			if v == nil {
				continue
			}
			xlog.Infof("resp value, %+v", v)
			if handler, ok := w.handlerMap[k]; ok && handler != nil {
				err = handler.Handle(v, imageBody)
				if err != nil {
					xlog.Errorf("%s handle msg error: %+v", k, err)
				}
			} else {
				xlog.Infof("Unable to find %s handler ", k)
			}
		}
		handlerDuation := time.Since(startTime2).String()
		totalDuration := time.Since(startTime).String()

		xlog.Infof("total frame duration:%s , frame duration:%s , handler duration:%s ", totalDuration, inferenceDuration, handlerDuation)
	}
}

func (w *VasWorker) releaseHandler(task proto.Task) {
	for _, h := range w.handlerMap {
		err := h.Release()
		if err != nil {
			log.Errorf("Release handler error: %+v ", err)
		}
	}
	w.handlerMap = make(map[string]handler.Handler)
}

func (w *VasWorker) initHandler(ctx context.Context, task proto.Task) (err error) {
	var mc proto.ModelConfig
	mcData, err := json.Marshal(task.Config)
	if err != nil {
		log.Println(err)
		return nil
	}
	err = json.Unmarshal(mcData, &mc)
	if err != nil {
		log.Println(err)
		return nil
	}

	if mc.NonMotorOn {
		nonMotorHandler, err := handler.NewNonMotorHandler(ctx, w.eventDao, &task, w.config.Fileserver)
		if err != nil {
			log.Println(err)
			return err
		}
		w.handlerMap[handler.NonMotorHandlerModelKey] = nonMotorHandler
	}

	if mc.VehicleOn {
		vehicleHandler, err := handler.NewVehicleHandler(ctx, w.eventDao, &task, w.config.Fileserver)
		if err != nil {
			log.Println(err)
			return err
		}
		w.handlerMap[handler.VehicleHandlerModelKey] = vehicleHandler
	}

	if mc.WaimaiOn {
		config := handler.WaimaiHandlerConfig{
			Fileserver: w.config.Fileserver,
			Timeout:    7,
			// DetectZone: w.limitedModelConfig.WaimaiDetectZone,
		}
		waimaiHandler, err := handler.NewWaimaiHandler(ctx, w.eventDao, &task, &config)
		if err != nil {
			log.Println(err)
			return err
		}
		w.handlerMap[handler.WaimaiHandlerModelKey] = waimaiHandler
	}

	if mc.ConstructionOn {
		config := handler.ConstructionHandlerConfig{
			Fileserver: w.config.Fileserver,
		}
		constructionHandler, err := handler.NewConstructionHandler(ctx, w.eventDao, &task, &config)
		if err != nil {
			log.Println(err)
			return err
		}
		w.handlerMap[handler.ConstructionHandlerModelKey] = constructionHandler
	}
	return nil
}
