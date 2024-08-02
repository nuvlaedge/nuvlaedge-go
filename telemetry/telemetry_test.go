package telemetry

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"nuvlaedge-go/telemetry/monitor"
	"nuvlaedge-go/testutils"
	"nuvlaedge-go/types"
	"nuvlaedge-go/types/metrics"
	"strings"
	"testing"
	"time"
)

var commissionerChan = make(chan types.CommissionData, 10)
var jobChan = make(chan string, 10)

func init() {
	log.SetLevel(log.PanicLevel)
}

func TestNewTelemetry_WithPositivePeriod_InitializesCorrectly(t *testing.T) {
	mockTelemetryClient := new(testutils.MockTelemetryClient)
	mockDockerMetricsClient := new(testutils.TestDockerMetricsClient)
	telemetry := NewTelemetry(10, mockTelemetryClient, mockDockerMetricsClient, commissionerChan, jobChan)

	assert.NotNil(t, telemetry)
	assert.Equal(t, 10, telemetry.period)
	assert.NotNil(t, telemetry.metricsChan)
	assert.Equal(t, 4, len(telemetry.monitors))
}

func Test_StartMonitors_WithValidContext_StartsAllMonitorsWithoutError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockTelemetryClient := testutils.MockTelemetryClient{
		GetEndpointResponse: "https://nuvla.io",
	}
	mockDockerMetricsClient := testutils.TestDockerMetricsClient{}
	telemetry := NewTelemetry(10, &mockTelemetryClient, &mockDockerMetricsClient, commissionerChan, jobChan)

	err := telemetry.StartMonitors(ctx)
	assert.NoError(t, err)
	time.Sleep(200 * time.Millisecond)
	for k, m := range telemetry.monitors {
		assert.True(t, m.Running(), "Monitor %s is not running", k)
	}
}

func Test_Telemetry_Run(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockTelemetryClient := &testutils.MockTelemetryClient{
		GetEndpointResponse: "https://nuvla.io",
	}
	mockDockerMetricsClient := &testutils.TestDockerMetricsClient{}
	telemetry := NewTelemetry(10, mockTelemetryClient, mockDockerMetricsClient, commissionerChan, jobChan)
	m := &testutils.MetricMock{}
	go func() {
		telemetry.metricsChan <- m
		time.Sleep(1 * time.Second)
		cancel()
	}()
	err := telemetry.Run(ctx)
	assert.Equal(t, m.IncCnt, 1)
	assert.Equal(t, context.Canceled, err)
	// Additional checks can be added to verify error logging behavior
}

func Test_SendTelemetry_WithUninitializedClient_ReturnsError(t *testing.T) {
	mockTelemetryClient := &testutils.MockTelemetryClient{
		GetEndpointResponse: "https://nuvla.io",
	}
	mockDockerMetricsClient := &testutils.TestDockerMetricsClient{}
	telemetry := NewTelemetry(10, mockTelemetryClient, mockDockerMetricsClient, commissionerChan, jobChan)
	telemetry.localStatus = metrics.NuvlaEdgeStatus{Status: "New"}
	mockTelemetryClient.TelemetryResponse = http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("")),
	}

	err := telemetry.sendTelemetry()
	assert.NoError(t, err)

	mockTelemetryClient.TelemetryResponse = http.Response{
		StatusCode: 400,
		Body:       io.NopCloser(strings.NewReader("")),
	}

	telemetry.localStatus = metrics.NuvlaEdgeStatus{Status: "New2"}
	err = telemetry.sendTelemetry()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "telemetry failed with status")

	telemetry.nuvla = nil
	err = telemetry.sendTelemetry()
	assert.Error(t, err)
}

func Test_SendTelemetry_WhenPayloadsAreEqual(t *testing.T) {
	mockTelemetryClient := &testutils.MockTelemetryClient{
		GetEndpointResponse: "https://nuvla.io",
	}
	mockDockerMetricsClient := &testutils.TestDockerMetricsClient{}
	telemetry := NewTelemetry(10, mockTelemetryClient, mockDockerMetricsClient, commissionerChan, jobChan)
	telemetry.localStatus = metrics.NuvlaEdgeStatus{Status: "New"}
	telemetry.lastStatus = metrics.NuvlaEdgeStatus{Status: "New"}

	err := telemetry.sendTelemetry()
	assert.NoError(t, err)
}

func Test_Telemetry_GetType(t *testing.T) {
	mockTelemetryClient := &testutils.MockTelemetryClient{
		GetEndpointResponse: "https://nuvla.io",
	}
	mockDockerMetricsClient := &testutils.TestDockerMetricsClient{}
	telemetry := NewTelemetry(10, mockTelemetryClient, mockDockerMetricsClient, commissionerChan, jobChan)

	assert.Equal(t, types.Telemetry, telemetry.GetType())
}

func Test_Telemetry_MonitorStatus(t *testing.T) {
	mockTelemetryClient := &testutils.MockTelemetryClient{
		GetEndpointResponse: "https://nuvla.io",
	}
	mockDockerMetricsClient := &testutils.TestDockerMetricsClient{}
	telemetry := NewTelemetry(10, mockTelemetryClient, mockDockerMetricsClient, commissionerChan, jobChan)

	telemetry.monitors = make(map[string]monitor.NuvlaEdgeMonitor)
	telemetry.monitors["test"] = testutils.NewMonitorMock()

	assert.False(t, telemetry.monitors["test"].Running())

	telemetry.monitorStatus(context.Background())
	time.Sleep(100 * time.Millisecond)
	assert.True(t, telemetry.monitors["test"].Running())
}
