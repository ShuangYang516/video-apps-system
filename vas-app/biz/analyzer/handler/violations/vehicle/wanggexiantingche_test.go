package vehicle

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	vio "qiniu.com/vas-app/biz/analyzer/handler/violations"
	"qiniu.com/vas-app/biz/proto"
)

type WanggexiantingcheViolationTestSuite struct {
	suite.Suite
	handler *WanggexiantingcheViolation
}

func (suite *WanggexiantingcheViolationTestSuite) SetupTest() {
	suite.handler = NewWanggexiantingcheViolation(
		context.TODO(),
		&WanggexiantingcheViolationConfig{
			Timeout:        60,
			ParkingSec:     5,
			MaxMovePercent: 0.05,
		},
	)
}

func (suite *WanggexiantingcheViolationTestSuite) TearDownTest() {
	if suite.handler != nil {
		suite.handler.Release()
	}
}

func (suite *WanggexiantingcheViolationTestSuite) TestCase1() {
	var (
		err     error
		imgBody = &proto.ImageBody{}
		data    proto.VehicleModelData
		event   *vio.ViolationEvent
	)

	jsonData := []byte(`
	{
		"boxes":
			[{
				"score":1,
				"violation_type":0,
				"id":-1,
				"cross_line_id":0,
				"violation_frame_idx":0,
				"frame_count":0,
				"plate_content":"沪A1F703",
				"pts":[[2289,575],[2540,751]]
			},{
				"score":1,
				"violation_type":0,
				"id":-1,
				"cross_line_id":0,
				"violation_frame_idx":0,
				"frame_count":0,
				"plate_content":"沪F86482",
				"pts":[[1986,688],[2223,907]]
			}]
	}
	`)

	assert.NoError(suite.T(), json.Unmarshal(jsonData, &data), "parse request data failed")
	event, err = suite.handler.Handle(&data, imgBody)
	assert.NoError(suite.T(), err, "handler failed, err: %v", err)
	assert.Nil(suite.T(), event, "shold no event")
}

func (suite *WanggexiantingcheViolationTestSuite) TestCase2() {
	var (
		err     error
		imgBody = &proto.ImageBody{}
		data    proto.VehicleModelData
		event   *vio.ViolationEvent
	)

	// 获取到车牌
	// 发现实线变道
	jsonData := []byte(`
	{
		"boxes":[
			{
				"score":1,
				"violation_type":0,
				"id":-1,
				"cross_line_id":0,
				"violation_frame_idx":0,
				"frame_count":0,
				"plate_content":"沪A060S8",
				"pts":[[1871,1092],[2281,1366]]
			}, {
				"score":1,
				"violation_type":2108,
				"id":0,
				"cross_line_id":0,
				"violation_frame_idx":2,
				"frame_count":41,
				"plate_content":"沪A060S8",
				"pts":[[1799,762],[2130,967]]
			}
		]
	}
	`)

	assert.NoError(suite.T(), json.Unmarshal(jsonData, &data), "parse request data failed")
	event, err = suite.handler.Handle(&data, imgBody)
	assert.NoError(suite.T(), err, "handler failed, err: %v", err)
	assert.Nil(suite.T(), event, "should no event")
}

func (suite *WanggexiantingcheViolationTestSuite) TestCase3() {
	var (
		err     error
		imgBody = &proto.ImageBody{}
		data    proto.VehicleModelData
		event   *vio.ViolationEvent
	)

	// 完整的大弯小转违法过程
	jsonData := []byte(`
	{
		"boxes":[
			{
				"score":1,
				"violation_type":0,
				"id":-1,
				"cross_line_id":0,
				"violation_frame_idx":0,
				"frame_count":0,
				"plate_content":"沪A060S8",
				"pts":[[1871,1092],[2281,1366]]
			}, {
				"score":1,
				"violation_type":2108,
				"id":0,
				"cross_line_id":0,
				"violation_frame_idx":2,
				"frame_count":41,
				"plate_content":"沪A060S8",
				"pts":[[1799,762],[2130,967]]
			}, {
				"score":1,
				"violation_type":2108,
				"id":0,
				"cross_line_id":0,
				"violation_frame_idx":3,
				"frame_count":226,
				"plate_content":"",
				"pts":[[1000,433],[1232,545]]
			}
		]
	}
	`)

	assert.NoError(suite.T(), json.Unmarshal(jsonData, &data), "parse request data failed")
	event, err = suite.handler.Handle(&data, imgBody)
	assert.NoError(suite.T(), err, "handler failed, err: %v", err)
	assert.Nil(suite.T(), event, "should have event")
}

func (suite *WanggexiantingcheViolationTestSuite) TestCase4() {
	var (
		err     error
		imgBody = &proto.ImageBody{}
		data    proto.VehicleModelData
		event   *vio.ViolationEvent
		// exist   bool
	)

	// 完整的实线变道违法过程
	jsonData := []byte(`
	{
		"boxes":[
			{
				"score":1,
				"violation_type":0,
				"id":-1,
				"cross_line_id":0,
				"violation_frame_idx":0,
				"frame_count":0,
				"plate_content":"沪A060S8",
				"pts":[[1871,1092],[2281,1366]]
			}, {
				"score":1,
				"violation_type":2108,
				"id":0,
				"cross_line_id":0,
				"violation_frame_idx":2,
				"frame_count":41,
				"plate_content":"沪A060S8",
				"pts":[[1799,762],[2130,967]]
			}, {
				"score":1,
				"violation_type":2108,
				"id":0,
				"cross_line_id":0,
				"violation_frame_idx":2,
				"frame_count":41,
				"plate_content":"沪A060S8",
				"pts":[[1799,762],[2130,967]]
			}, {
				"score":1,
				"violation_type":2108,
				"id":0,
				"cross_line_id":0,
				"violation_frame_idx":2,
				"frame_count":41,
				"plate_content":"沪A060S8",
				"pts":[[1799,762],[2130,967]]
			}, {
				"score":1,
				"violation_type":2108,
				"id":0,
				"cross_line_id":0,
				"violation_frame_idx":3,
				"frame_count":226,
				"plate_content":"",
				"pts":[[1000,433],[1232,545]]
			}
		]
	}
	`)

	assert.NoError(suite.T(), json.Unmarshal(jsonData, &data), "parse request data failed")
	tmp := data
	now := time.Now()
	tmp.Boxes = data.Boxes[:2]
	event, err = suite.handler.Handle(&tmp, imgBody)
	assert.NoError(suite.T(), err, "handler failed, err: %v", err)
	assert.Nil(suite.T(), event, "should have event")

	now = now.Add(time.Second * 2)
	suite.handler.now = &now
	tmp.Boxes = data.Boxes[2:3]
	event, err = suite.handler.Handle(&tmp, imgBody)
	assert.NoError(suite.T(), err, "handler failed, err: %v", err)
	assert.Nil(suite.T(), event, "should have event")

	now = now.Add(time.Second * 4)
	suite.handler.now = &now
	tmp.Boxes = data.Boxes[3:]
	event, err = suite.handler.Handle(&tmp, imgBody)
	assert.NoError(suite.T(), err, "handler failed, err: %v", err)
	assert.NotNil(suite.T(), event, "should have event")
}

func TestWangGexianTingCheTestSuite(t *testing.T) {
	suite.Run(t, new(WanggexiantingcheViolationTestSuite))
}
func (suite *WanggexiantingcheViolationTestSuite) TestCase5() {
	var (
		err     error
		imgBody = &proto.ImageBody{}
		data    proto.VehicleModelData
		event   *vio.ViolationEvent
	)

	// 完整的违法过程，但是不是实线变道类型
	jsonData := []byte(`
	{
		"boxes":[
			{
				"score":1,
				"violation_type":0,
				"id":-1,
				"cross_line_id":0,
				"violation_frame_idx":0,
				"frame_count":0,
				"plate_content":"沪A060S8",
				"pts":[[1871,1092],[2281,1366]]
			}, {
				"score":1,
				"violation_type":2108,
				"id":0,
				"cross_line_id":0,
				"violation_frame_idx":2,
				"frame_count":41,
				"plate_content":"沪A060S8",
				"pts":[[1799,762],[2130,967]]
			}, {
				"score":1,
				"violation_type":2108,
				"id":0,
				"cross_line_id":0,
				"violation_frame_idx":3,
				"frame_count":226,
				"plate_content":"",
				"pts":[[1000,433],[1232,545]]
			}
		]
	}
	`)

	assert.NoError(suite.T(), json.Unmarshal(jsonData, &data), "parse request data failed")
	event, err = suite.handler.Handle(&data, imgBody)
	assert.NoError(suite.T(), err, "handler failed, err: %v", err)
	assert.Nil(suite.T(), event, "should have event")
}
