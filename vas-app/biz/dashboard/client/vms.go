package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type VmsConfig struct {
	Host     string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type VmsClient struct {
	VmsConfig

	authtoken string
	lastLogin int64
}

type LoginReq struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

type LoginResp struct {
	ErrCode int `json:"code"`
	Data    struct {
		Account string `json:"account"`
		Token   string `json:"token"`
	} `json:"data"`
}

func NewVmsClient(config VmsConfig) *VmsClient {
	return &VmsClient{
		VmsConfig: config,
	}
}

func (client *VmsClient) reqJson(method, path string, body io.ReadCloser) (ret []byte, err error) {
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return
	}
	if client.authtoken == "" || client.lastLogin+86400 < time.Now().Unix() {
		path := fmt.Sprintf("%s/v2/user/token/create", client.Host)
		bs, _ := json.Marshal(LoginReq{
			Account:  client.Username,
			Password: client.Password,
		})

		resp, err := http.DefaultClient.Post(path, "application/json", bytes.NewReader(bs))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		bs, err = ioutil.ReadAll(resp.Body)

		loginResp := LoginResp{}
		json.Unmarshal(bs, &loginResp)
		if loginResp.ErrCode != 0 {
			return nil, errors.New("login failed")
		}

		client.authtoken = loginResp.Data.Token
		client.lastLogin = time.Now().Unix()
	}
	req.SetBasicAuth(client.authtoken, "")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	ret, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	return
}

func (client *VmsClient) GetDevice(deviceID string) (ret []byte, err error) {
	path := fmt.Sprintf("%s/v2/device/%s", client.Host, deviceID)
	return client.reqJson("GET", path, nil)
}

func (client *VmsClient) GetSubDevices(deviceID string, page, perPage int) (ret []byte, err error) {
	path := fmt.Sprintf("%s/v2/sub_device?device_id=%s&page=%d&per_page=%d", client.Host, deviceID, page, perPage)
	return client.reqJson("GET", path, nil)
}

func (client *VmsClient) GetDeviceChannels(deviceID string) (ret []byte, err error) {
	path := fmt.Sprintf("%s/v2/device/%s/available_channels", client.Host, deviceID)
	return client.reqJson("GET", path, nil)
}

func (client *VmsClient) CreateSubDevice(body io.ReadCloser) (ret []byte, err error) {
	path := fmt.Sprintf("%s/v2/sub_device", client.Host)
	return client.reqJson("POST", path, body)
}

func (client *VmsClient) UpdateSubDevice(subDeviceID string, body io.ReadCloser) (ret []byte, err error) {
	path := fmt.Sprintf("%s/v2/sub_device/%s", client.Host, subDeviceID)
	return client.reqJson("PUT", path, body)
}

func (client *VmsClient) RemoveSubDevice(subDeviceID string, body io.ReadCloser) (ret []byte, err error) {
	path := fmt.Sprintf("%s/v2/sub_device/%s", client.Host, subDeviceID)
	return client.reqJson("DELETE", path, body)
}

func (client *VmsClient) RemoveSubDevices(body io.ReadCloser) (ret []byte, err error) {
	path := fmt.Sprintf("%s/v2/sub_device", client.Host)
	return client.reqJson("DELETE", path, body)
}

func (client *VmsClient) GetPlayUrl(subDeviceID string, streamID int, protocol string) (ret []byte, err error) {
	path := fmt.Sprintf("%s/v2/sub_device/%s/play_url?stream_id=%d&type=%s", client.Host, subDeviceID, streamID, protocol)
	return client.reqJson("GET", path, nil)
}

func (client *VmsClient) GetTypeEnums() (ret []byte, err error) {
	path := fmt.Sprintf("%s/v2/sub_device/type", client.Host)
	return client.reqJson("GET", path, nil)
}

func (client *VmsClient) GetVendorEnums() (ret []byte, err error) {
	path := fmt.Sprintf("%s/v2/sub_device/vendor", client.Host)
	return client.reqJson("GET", path, nil)
}

func (client *VmsClient) GetDiscoveryProtocol(typeValue string) (ret []byte, err error) {
	path := fmt.Sprintf("%s/v2/sub_device/discovery_protocol/%s", client.Host, typeValue)
	return client.reqJson("GET", path, nil)
}

func (client *VmsClient) GetChannelTypeEnums() (ret []byte, err error) {
	path := fmt.Sprintf("%s/v2/sub_device/channel_type", client.Host)
	return client.reqJson("GET", path, nil)
}
