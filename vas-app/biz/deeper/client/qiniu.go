package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	xlog "github.com/qiniu/xlog.v1"
)

type QiniuClientConfig struct {
	Host      string  `json:"host"`
	Threshold float64 `json:"threshold"`
}

type QiniuClient struct {
	QiniuClientConfig
}

type FaceDetect struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Result  Result `json:"result"`
}

type QScore struct {
	Blur  float64 `json:"blur"`
	Clear float64 `json:"clear"`
	Cover float64 `json:"cover"`
	Neg   float64 `json:"neg"`
	Pose  float64 `json:"pose"`
}

type Detections struct {
	Class       string    `json:"class"`
	Index       int       `json:"index"`
	Orientation string    `json:"orientation"`
	Pts         [4][2]int `json:"pts"`
	QScore      QScore    `json:"q_score"`
	Quality     string    `json:"quality"`
	Score       float64   `json:"score"`
}

type Result struct {
	Detections []Detections `json:"detections"`
}

func NewQiniuClient(config QiniuClientConfig) *QiniuClient {
	return &QiniuClient{
		QiniuClientConfig: config,
	}
}

func (c *QiniuClient) DetectFace(ctx context.Context, data []byte) (faceDetect FaceDetect, err error) {
	xl := xlog.FromContextSafe(ctx)
	const prefix = "data:application/octet-stream;base64,"
	var uri string
	if bytes.HasPrefix(data, []byte("http")) {
		uri = string(data)
	} else {
		uri = prefix + base64.StdEncoding.EncodeToString(data)
	}

	bs, err := json.Marshal(struct {
		Data struct {
			URI string `json:"uri"`
		} `json:"data"`
	}{
		struct {
			URI string `json:"uri"`
		}{
			URI: uri,
		},
	})
	if err != nil {
		xl.Errorf("json.Marshal: %+v", err)
		return
	}
	req, err := http.NewRequest("POST", c.Host+"/v1/eval/facex-detect", bytes.NewBuffer(bs))
	if err != nil {
		xl.Errorf("build req: %+v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "QiniuStub uid=1&ut=0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		xl.Errorf("httpPostError: %+v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("http statuscode is : %d", resp.StatusCode)
		xl.Errorf("httpPostError: %+v", err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		xl.Errorf("ioError: %+v", err)
		return
	}

	err = json.Unmarshal(body, &faceDetect)

	if err != nil {
		xl.Errorf("jsonToResultError: %+v", err)
		return
	}
	return
}
