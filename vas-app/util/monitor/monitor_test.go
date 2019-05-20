package monitor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMonitor(t *testing.T) {
	assertion := assert.New(t)

	assertion.Nil(ResponseTime("api", 0))
	ResponseTimeFrom("api", 0, time.Now())
	assertion.Nil(RequestsCounter("api", 0))
	RequestsCounterInc("api", 0)

	assertion.Nil(RequestsParallel("api"))
	RequestsParallelInc("api")
	RequestsParallelDec("api")

	Init("project", "app")
	assertion.NotNil(ResponseTime("api", 0))
	ResponseTimeFrom("api", 0, time.Now())
	assertion.NotNil(RequestsCounter("api", 0))
	RequestsCounterInc("api", 0)

	assertion.NotNil(RequestsParallel("api"))
	RequestsParallelInc("api")
	RequestsParallelDec("api")
}
