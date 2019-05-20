package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	xlog "github.com/qiniu/xlog.v1"
	"qiniu.com/vas-app/biz/deeper/proto"
)

const YituSuccessResultCode = "200"

type YituClientConfig struct {
	Host string `json:"host"`

	Username     string `json:"username"`
	Password     string `json:"password"`
	LibIds       string `json:"lib_ids"`
	Threshold    int    `json:"threshold"`
	ExpireSecond int64  `json:"expire_second"`
}

type YituClient struct {
	YituClientConfig

	cookies   []*http.Cookie
	lastLogin int64
	loginLock sync.Mutex
}

func NewYituClient(config YituClientConfig) *YituClient {
	return &YituClient{
		YituClientConfig: config,
	}
}

func (c *YituClient) mustLogin(ctx context.Context, req *http.Request) error {
	xl := xlog.FromContextSafe(ctx)
	if time.Now().Unix()-c.lastLogin >= c.ExpireSecond {
		c.loginLock.Lock()
		defer c.loginLock.Unlock()
		if time.Now().Unix()-c.lastLogin >= c.ExpireSecond {
			xl.Debugf("yitu api trying to login")
			path := c.Host + "/FDS/login?loginName=" + c.Username + "&password=" + c.Password
			// TODO 这个请求的content-type是, 还是说可以不传?
			resp, err := http.Post(path, "", nil)
			if err != nil {
				xl.Errorf("yitu api login error: %+v", err)
				return err
			}
			defer resp.Body.Close()
			ret := proto.YituLoginResp{}
			err = UnmarshalResp(resp, &ret)
			if err != nil {
				xl.Errorf("yitu api login result parse failed: %+v", err)
				return err
			}
			if ret.ResultCode != YituSuccessResultCode {
				err := fmt.Errorf("yitu api login failed: ResultCode is %s", ret.ResultCode)
				xl.Error(err)
				return err
			}

			c.lastLogin = time.Now().Unix()
			c.cookies = resp.Cookies()
			xl.Debugf("yitu api login success")
		}
	}
	for _, cookie := range c.cookies {
		req.AddCookie(cookie)
	}
	return nil
}

func (c *YituClient) ListFaceLib(ctx context.Context, userID string) ([]proto.YituLibItem, error) {
	// NOTE 暂时用不着, 不用实现
	// xl := xlog.FromContextSafe(ctx)
	// path := c.Host + "/FDS/faceLib/list"

	return nil, nil
}

func (c *YituClient) DetectFace(ctx context.Context, imageData []byte) (items []proto.YituFaceRect, err error) {
	xl := xlog.FromContextSafe(ctx)
	path := c.Host + "/FDS/feature/checkFace"

	var body bytes.Buffer
	w := multipart.NewWriter(&body)

	fw, _ := w.CreateFormFile("image", "image.jpg")
	io.Copy(fw, bytes.NewReader(imageData))
	w.Close()

	req, err := http.NewRequest("POST", path, &body)
	if err != nil {
		xl.Errorf("checkFace build request: %+v", err)
		return
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	err = c.mustLogin(ctx, req)
	if err != nil {
		xl.Errorf("checkFace call login error: %+v", err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		xl.Errorf("checkFace error: %+v", err)
		return
	}
	defer resp.Body.Close()
	ret := proto.YituFeatureCheckFaceResp{}
	err = UnmarshalResp(resp, &ret)
	if err != nil {
		xl.Errorf("checkFace result parse failed: %+v", err)
		return
	}
	if ret.ResultCode != YituSuccessResultCode {
		err = fmt.Errorf("checkFace failed: Yitu api ResultCode is %s", ret.ResultCode)
		xl.Error(err)
		return
	}
	items = ret.Data.FaceInfo
	return
}

func (c *YituClient) SearchFace(ctx context.Context, imageData []byte, pts [4][2]int) (items []proto.YituSearchResultItem, err error) {
	xl := xlog.FromContextSafe(ctx)

	path := c.Host + "/FDS/search/searchFace"

	if bytes.HasSuffix(imageData, []byte("data:image/")) {
		// NOTE 依图api只接受没有前缀的base64 string
		imageData = bytes.SplitN(imageData, []byte(";base64,"), 2)[1]
	}

	data := url.Values{
		"yituIds": []string{c.LibIds},
		// "threshold": []string{string(c.Threshold)},
		"imageData": []string{string(imageData)},
	}

	if rect, valid := (proto.YituFaceRect{}.FromPTS(pts)); valid {
		data["faceRect"] = []string{rect.String()}
	}
	req, err := http.NewRequest("POST", path, strings.NewReader(data.Encode()))
	if err != nil {
		xl.Errorf("search face build request: %+v", err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	err = c.mustLogin(ctx, req)
	if err != nil {
		xl.Errorf("search face call login error: %+v", err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		xl.Errorf("search face error: %+v", err)
		return
	}
	defer resp.Body.Close()
	ret := proto.YituSearchResp{}
	err = UnmarshalResp(resp, &ret)
	if err != nil {
		xl.Errorf("search face result parse failed: %+v", err)
		return
	}
	if ret.ResultCode != YituSuccessResultCode {
		err = fmt.Errorf("login failed: Yitu api ResultCode is %s", ret.ResultCode)
		xl.Error(err)
		return
	}
	items = ret.Data.List
	return
}

func UnmarshalResp(resp *http.Response, obj interface{}) error {
	if resp.StatusCode != http.StatusOK {
		return errors.New("bad request")
	}
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bs, obj)
	if err != nil {
		return err
	}
	return nil
}
