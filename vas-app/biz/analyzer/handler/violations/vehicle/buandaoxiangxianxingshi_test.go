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

type VehicleBuandaoxiangxianxingshiTestSuite struct {
	suite.Suite
	handler *BuandaoxiangxianxingshiViolation
}

func (suite *VehicleBuandaoxiangxianxingshiTestSuite) SetupTest() {
	suite.handler = NewBuandaoxiangxianxingshiViolation(
		context.TODO(),
		&BuandaoxiangxianxingshiViolationConfig{
			Timeout: 60,
		},
	)
}

func (suite *VehicleBuandaoxiangxianxingshiTestSuite) TearDownTest() {
	if suite.handler != nil {
		suite.handler.Release()
	}
}

func (suite *VehicleBuandaoxiangxianxingshiTestSuite) TestCase1() {
	var (
		err     error
		imgBody = &proto.ImageBody{}
		data    proto.VehicleModelData
		event   *vio.ViolationEvent
	)

	jsonData := []byte(`
	{
		"boxes":
			[
				{
					"score":1,
					"violation_type":0,
					"id":-1,"cross_line_id":0,
					"violation_frame_idx":0,
					"frame_count":1029,
					"plate_content":"皖K94C15",
					"pts":[[1762,1143],[2461,1759]]
				},
				{
					"score":1,
					"violation_type":2106,
					"id":34,
					"cross_line_id":0,
					"violation_frame_idx":2,
					"frame_count":1148,
					"plate_content":"皖K94C15",
					"pts":[[1611,600],[2007,1141]]
				},
				{
					"score":1,
					"violation_type":2106,
					"id":34,
					"cross_line_id":0,
					"violation_frame_idx":3,
					"frame_count":1149,
					"plate_content":"皖K94C15",
					"pts":[[1620,595],[2009,1137]]
				}
			]
	}
	`)

	assert.NoError(suite.T(), json.Unmarshal(jsonData, &data), "parse request data failed")
	event, err = suite.handler.Handle(&data, imgBody)
	assert.NoError(suite.T(), err, "handler failed, err: %v", err)
	assert.NotNil(suite.T(), event, "shold have event")
}

func TestVehicleBuandaoxiangxianxingshiTestSuite(t *testing.T) {
	suite.Run(t, new(VehicleBuandaoxiangxianxingshiTestSuite))
}
