package workers

import (
	"context"
	"errors"
	nuvlaTypes "github.com/nuvla/api-client-go/types"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"nuvlaedge-go/testutils"
	"nuvlaedge-go/types"
	"nuvlaedge-go/types/metrics"
	"nuvlaedge-go/types/worker"
	"testing"
	"time"
)

func init() {
	// Set log level to panic to avoid logs during tests
	log.SetLevel(log.PanicLevel)
}

func newCommissioner(period int, client types.CommissionClientInterface, ch chan types.CommissionData) *Commissioner {
	c := Commissioner{}
	_ = c.Init(&worker.WorkerOpts{CommissionCh: ch}, &worker.WorkerConfig{CommissionPeriod: period})
	c.client = client
	return &c
}

func Test_Commissioner_GetNodeIdFromStatus(t *testing.T) {
	client := &testutils.CommissionerClientMock{}
	ch := make(chan types.CommissionData)
	c := newCommissioner(1, client, ch)

	client.GetStatusIdReturn = ""
	assert.Equal(t, "", c.getNodeIdFromStatus())

	client.GetStatusIdReturn = "status-id"
	client.GetError = errors.New("get error")
	assert.Equal(t, "", c.getNodeIdFromStatus())
	client.GetError = nil

	client.GetReturn = nuvlaTypes.NuvlaResource{
		Data: map[string]interface{}{"node-id": "node-id"},
	}
	assert.Equal(t, "node-id", c.getNodeIdFromStatus())

	client.GetReturn = nuvlaTypes.NuvlaResource{
		Data: make(map[string]interface{}),
	}
	assert.Equal(t, "", c.getNodeIdFromStatus())
}

func Test_Commissioner_Run_ProperContextCancel(t *testing.T) {
	bCtx := context.Background()
	client := &testutils.CommissionerClientMock{}
	ch := make(chan types.CommissionData)
	c := newCommissioner(1, client, ch)
	ctx, cancel := context.WithCancel(bCtx)

	go func() {
		err := c.Run(ctx)
		assert.Equal(t, context.Canceled, err)
	}()
	time.Sleep(100 * time.Millisecond)
	cancel()
}

func Test_Commissioner_Run_CommissionSuccess(t *testing.T) {
	bCtx := context.Background()
	client := &testutils.CommissionerClientMock{}
	ch := make(chan types.CommissionData)
	c := newCommissioner(1, client, ch)

	sample := metrics.SwarmData{SwarmEndPoint: "mock_endpoint"}
	ctx2, cancel2 := context.WithTimeout(bCtx, 2*time.Second)

	go func() {
		time.Sleep(100 * time.Millisecond)
		ch <- sample
	}()
	err := c.Run(ctx2)
	assert.Equal(t, context.DeadlineExceeded, err)

	cancel2()
	assert.Equal(t, c.currentData.SwarmEndPoint, sample.SwarmEndPoint)
	assert.Equal(t, c.lastCommission.SwarmEndPoint, sample.SwarmEndPoint)
}

func Test_Commissioner_Run_CommissionNotNeeded(t *testing.T) {
	bCtx := context.Background()
	client := &testutils.CommissionerClientMock{}
	ch := make(chan types.CommissionData)
	c := newCommissioner(1, client, ch)
	sample := metrics.SwarmData{SwarmEndPoint: "mock_endpoint"}

	ctx, cancel := context.WithTimeout(bCtx, 1*time.Second)
	defer cancel()
	client.CommissionErr = errors.New("commission error")

	defer cancel()
	go func() {
		time.Sleep(100 * time.Millisecond)
		sample.SwarmEndPoint = "mock_ca"
		ch <- sample
	}()

	err := c.Run(ctx)
	assert.Equal(t, context.DeadlineExceeded, err)
	assert.Equal(t, c.currentData.SwarmEndPoint, sample.SwarmEndPoint)
	assert.NotEqualf(t, c.lastCommission.SwarmEndPoint, sample.SwarmEndPoint, "SwarmEndPoint should not be updated")
}

func Test_Commissioner_NeedsCommission(t *testing.T) {
	client := &testutils.CommissionerClientMock{}
	ch := make(chan types.CommissionData)
	c := newCommissioner(1, client, ch)
	d, o := c.needsCommissioning()
	assert.NotNil(t, d)
	assert.True(t, o)
	c.lastCommission = c.currentData

	d, o = c.needsCommissioning()
	assert.False(t, o)
	assert.Nil(t, d)

	c.lastCommission.SwarmEndPoint = "value"
	c.currentData.SwarmEndPoint = ""
	d, o = c.needsCommissioning()
	assert.True(t, o)
	_, ok := d["removed"]
	assert.True(t, ok)
}
