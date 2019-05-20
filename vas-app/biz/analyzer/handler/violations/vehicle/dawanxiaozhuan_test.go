package vehicle

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	vio "qiniu.com/vas-app/biz/analyzer/handler/violations"
	"qiniu.com/vas-app/biz/proto"
)

type VehicleDawanxiaozhuanTestSuite struct {
	suite.Suite
	handler *DawanxiaozhuanViolation
}

func (suite *VehicleDawanxiaozhuanTestSuite) SetupTest() {
	suite.handler = NewDawanxiaozhuanViolation(
		context.TODO(),
		&DawanxiaozhuanViolationConfig{
			Timeout: 60,
		},
	)
}

func (suite *VehicleDawanxiaozhuanTestSuite) TearDownTest() {
	if suite.handler != nil {
		suite.handler.Release()
	}
}

func (suite *VehicleDawanxiaozhuanTestSuite) TestVehicleCase1() {
	var (
		err     error
		imgBody = &proto.ImageBody{}
		data    proto.VehicleModelData
		exist   bool
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
	_, exist = suite.handler.labelSnap["沪A1F703"]
	assert.True(suite.T(), exist, "沪A1F703 should in labelSnap")
	_, exist = suite.handler.labelSnap["沪F86482"]
	assert.True(suite.T(), exist, "沪F86482 should in labelSnap")
}

func (suite *VehicleDawanxiaozhuanTestSuite) TestVehicleCase2() {
	var (
		err     error
		imgBody = &proto.ImageBody{}
		data    proto.VehicleModelData
		exist   bool
		event   *vio.ViolationEvent
	)

	// 获取到车牌
	// 发现压线
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
				"violation_type":2101,
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
	_, exist = suite.handler.labelSnap["沪A060S8"]
	assert.True(suite.T(), exist, "沪A060S8 should in labelSnap(label <-> snap)")
	_, exist = suite.handler.objVioEvent[0]
	assert.True(suite.T(), exist, "沪A060S8 (id:0) should in m(id <-> vehicle info)")
	assert.NotNil(suite.T(), suite.handler.objVioEvent[0].Snapshots[1], "沪A060S8 (id:0) info should not be nil")
}

func (suite *VehicleDawanxiaozhuanTestSuite) TestVehicleCase3() {
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
				"violation_type":2101,
				"id":0,
				"cross_line_id":0,
				"violation_frame_idx":2,
				"frame_count":41,
				"plate_content":"沪A060S8",
				"pts":[[1799,762],[2130,967]]
			}, {
				"score":1,
				"violation_type":2101,
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
	assert.NotNil(suite.T(), event, "should have event")
	// 发现了大弯小转违法行为，buffer 会被清理
	assert.Equal(suite.T(), 0, len(suite.handler.labelSnap), "labelSnap should be empty")
	for id, ev := range suite.handler.objVioEvent {
		suite.T().Logf("id: %d, event: %+v", id, ev)
	}
	assert.Equal(suite.T(), 0, len(suite.handler.objVioEvent), "objVioEvent should be empty")
}

func (suite *VehicleDawanxiaozhuanTestSuite) TestVehicleCase4() {
	var (
		err     error
		imgBody = &proto.ImageBody{}
		data    proto.VehicleModelData
		event   *vio.ViolationEvent
		// exist   bool
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
				"violation_type":2101,
				"id":0,
				"cross_line_id":0,
				"violation_frame_idx":2,
				"frame_count":41,
				"plate_content":"沪A060S8",
				"pts":[[1799,762],[2130,967]]
			}, {
				"score":1,
				"violation_type":2101,
				"id":0,
				"cross_line_id":0,
				"violation_frame_idx":2,
				"frame_count":41,
				"plate_content":"沪A060S8",
				"pts":[[1799,762],[2130,967]]
			}, {
				"score":1,
				"violation_type":2101,
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
	assert.NotNil(suite.T(), event, "should have event")
	// 发现了大弯小转违法行为，buffer 会被清理
	assert.Equal(suite.T(), 0, len(suite.handler.labelSnap), "labelSnap should be empty")
	// for id, info := range suite.handler.m {
	// 	suite.T().Logf("id: %d, info: %v", id, info)
	// }
	assert.Equal(suite.T(), 0, len(suite.handler.objVioEvent), "objVioEvent should be empty")
}

func TestVehicleDawanxiaozhuanTestSuite(t *testing.T) {
	suite.Run(t, new(VehicleDawanxiaozhuanTestSuite))
}
func (suite *VehicleDawanxiaozhuanTestSuite) TestVehicleCase5() {
	var (
		err     error
		imgBody = &proto.ImageBody{}
		data    proto.VehicleModelData
		event   *vio.ViolationEvent
	)

	// 完整的违法过程，但是不是大弯小转类型
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
				"violation_type":2102,
				"id":0,
				"cross_line_id":0,
				"violation_frame_idx":2,
				"frame_count":41,
				"plate_content":"沪A060S8",
				"pts":[[1799,762],[2130,967]]
			}, {
				"score":1,
				"violation_type":2102,
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
