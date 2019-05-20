package client

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"qiniu.com/vas-app/util"
)

type Client struct {
	endpoint string
}

func New(endpoint string) (*Client, error) {
	client := &Client{
		endpoint: endpoint,
	}
	return client, nil
}

func (c *Client) GetEvents(req *GetEventReq) (resp GetEventResp, err error) {
	url := c.endpoint + "/v1/events/export/private"
	url += fmt.Sprintf("?start=%d&end=%d", req.Start, req.End)
	if req.CameraId != "" {
		url += fmt.Sprintf("&cameraIds=%s", req.CameraId)
	}
	if req.Class != 0 {
		url += fmt.Sprintf("&classes=%d", req.Class)
	}
	if req.Type != "" {
		url += "&type=" + req.Type
	}
	if req.EventType != 0 {
		url += fmt.Sprintf("&eventTypes=%d", req.EventType)
	}
	if req.Limit > 0 {
		url += "&limit=" + strconv.Itoa(req.Limit)
	}

	if req.Marking != "" {
		url += "&marking=" + req.Marking
	}

	if req.HasLabel != 0 {
		url += fmt.Sprintf("&hasLabel=%d", req.HasLabel)
	}
	if req.LabelScore > 0 {
		url += fmt.Sprintf("&labelScore=%f", req.LabelScore)
	}

	if req.HasFace != 0 {
		url += fmt.Sprintf("&hasFace=%d", req.HasFace)
	}
	if req.Similarity > 0 {
		url += fmt.Sprintf("&similarity=%f", req.Similarity)
	}

	body, err := util.Get(url)
	if err != nil {
		log.Printf("util.Get(%s): %v\n", url, err)
		return
	}

	err = json.Unmarshal(body, &resp)

	if err != nil {
		log.Printf("json.Unmarshal error: %+v\n", err)
		return
	}

	return
}
