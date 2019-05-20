package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	client      *VmsClient
	deviceId    = "QN001158b2c2df1f785eb"
	subDeviceId = "01000000E4277149C5628B15A201000000000000"
	ts          *httptest.Server
)

type Resp struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

func init() {
	ts = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		res := []byte(`{"code": 0, "msg": ""}`)
		if req.Method == "GET" && req.URL.Path == "v2/user/token/create" {
			rw.Write(res)
		} else if matched, _ := regexp.MatchString("v2/sub_device", req.URL.Path); matched && req.Method == "GET" {
			rw.Write(res)
		} else {
			rw.Write(res)
		}

		// TODO
	}))

	client = NewVmsClient(VmsConfig{
		Host:     ts.URL,
		Username: "qiniu",
		Password: "qiniu",
	})

}

func TestGetDevice(t *testing.T) {
	assertion := assert.New(t)

	fmt.Println("getDevice start")

	bs, err := client.GetDevice(deviceId)
	if !assertion.NoError(err) {
		return
	}

	var resp Resp
	err = json.Unmarshal(bs, &resp)
	if !assertion.NoError(err) {
		return
	}
	fmt.Println("getDevice resp:", resp)
	assertion.Equal(0, resp.Code)
}

func TestGetSubDevices(t *testing.T) {
	assertion := assert.New(t)

	bs, err := client.GetSubDevices(deviceId, 1, 5)
	if !assertion.NoError(err) {
		return
	}

	var resp Resp
	err = json.Unmarshal(bs, &resp)
	if !assertion.NoError(err) {
		return
	}
	assertion.Equal(0, resp.Code)
}

func TestGetDeviceChannels(t *testing.T) {
	assertion := assert.New(t)

	bs, err := client.GetDeviceChannels(deviceId)
	if !assertion.NoError(err) {
		return
	}

	var resp Resp
	err = json.Unmarshal(bs, &resp)
	if !assertion.NoError(err) {
		return
	}
	assertion.Equal(0, resp.Code)
}

func TestGetPlayUrl(t *testing.T) {
	assertion := assert.New(t)

	bs, err := client.GetPlayUrl(subDeviceId, 0, "hflv")
	if !assertion.NoError(err) {
		return
	}

	var resp Resp
	err = json.Unmarshal(bs, &resp)
	if !assertion.NoError(err) {
		return
	}
	assertion.Equal(0, resp.Code)
}

func TestGetTypeEnums(t *testing.T) {
	assertion := assert.New(t)

	bs, err := client.GetTypeEnums()
	if !assertion.NoError(err) {
		return
	}

	var resp Resp
	err = json.Unmarshal(bs, &resp)
	if !assertion.NoError(err) {
		return
	}
	assertion.Equal(0, resp.Code)
}

func TestGetVendorEnums(t *testing.T) {
	assertion := assert.New(t)

	bs, err := client.GetVendorEnums()
	if !assertion.NoError(err) {
		return
	}

	var resp Resp
	err = json.Unmarshal(bs, &resp)
	if !assertion.NoError(err) {
		return
	}
	assertion.Equal(0, resp.Code)
}

func TestGetDiscoveryProtocol(t *testing.T) {
	assertion := assert.New(t)

	bs, err := client.GetDiscoveryProtocol("1")
	if !assertion.NoError(err) {
		return
	}

	var resp Resp
	err = json.Unmarshal(bs, &resp)
	if !assertion.NoError(err) {
		return
	}
	assertion.Equal(0, resp.Code)
}

func TestGetChannelTypeEnums(t *testing.T) {
	assertion := assert.New(t)

	bs, err := client.GetChannelTypeEnums()
	if !assertion.NoError(err) {
		return
	}

	var resp Resp
	err = json.Unmarshal(bs, &resp)
	if !assertion.NoError(err) {
		return
	}
	assertion.Equal(0, resp.Code)
}
