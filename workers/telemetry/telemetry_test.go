package telemetry

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"nuvlaedge-go/testutils"
	"nuvlaedge-go/types"
	"nuvlaedge-go/types/metrics"
	"nuvlaedge-go/types/worker"
	"nuvlaedge-go/workers/telemetry/monitor"
	"strings"
	"testing"
	"time"
)

var commissionerChan = make(chan types.CommissionData, 10)
var jobChan = make(chan string, 10)

func init() {
	log.SetLevel(log.PanicLevel)
}

func newTelemetry(
	period int,
	neClient types.TelemetryClientInterface,
	dockerClient types.DockerMetricsClient,
	commissionerChan chan types.CommissionData,
	jobChan chan string) *Telemetry {

	t := &Telemetry{
		TimedWorker: worker.NewTimedWorker(period, worker.Telemetry),
		nuvla:       neClient,
		metricsChan: make(chan metrics.Metric, 10), // Buffer size 10 to allow all different metric types to be sent without blocking
		jobChan:     jobChan,
	}

	t.monitors = map[string]monitor.NuvlaEdgeMonitor{
		"engine":       monitor.NewDockerMonitor(dockerClient, t.GetPeriod(), t.metricsChan, neClient.GetEndpoint(), commissionerChan),
		"system":       monitor.NewSystemMonitor(t.GetPeriod(), t.metricsChan),
		"resources":    monitor.NewResourceMonitor(t.GetPeriod(), t.metricsChan),
		"installation": monitor.NewInstallationMonitor(t.GetPeriod(), dockerClient, t.metricsChan),
	}

	return t
}

func Test_StartMonitors_WithValidContext_StartsAllMonitorsWithoutError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockTelemetryClient := testutils.MockTelemetryClient{
		GetEndpointResponse: "https://nuvla.io",
	}
	mockDockerMetricsClient := testutils.TestDockerMetricsClient{}
	telemetry := newTelemetry(10, &mockTelemetryClient, &mockDockerMetricsClient, commissionerChan, jobChan)

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
	telemetry := newTelemetry(10, mockTelemetryClient, mockDockerMetricsClient, commissionerChan, jobChan)
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
	telemetry := newTelemetry(10, mockTelemetryClient, mockDockerMetricsClient, commissionerChan, jobChan)
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
	telemetry := newTelemetry(10, mockTelemetryClient, mockDockerMetricsClient, commissionerChan, jobChan)
	telemetry.localStatus = metrics.NuvlaEdgeStatus{Status: "New"}
	telemetry.lastStatus = metrics.NuvlaEdgeStatus{Status: "New"}

	err := telemetry.sendTelemetry()
	assert.NoError(t, err)
}

func Test_Telemetry_MonitorStatus(t *testing.T) {
	mockTelemetryClient := &testutils.MockTelemetryClient{
		GetEndpointResponse: "https://nuvla.io",
	}
	mockDockerMetricsClient := &testutils.TestDockerMetricsClient{}
	telemetry := newTelemetry(10, mockTelemetryClient, mockDockerMetricsClient, commissionerChan, jobChan)

	telemetry.monitors = make(map[string]monitor.NuvlaEdgeMonitor)
	telemetry.monitors["test"] = testutils.NewMonitorMock()

	assert.False(t, telemetry.monitors["test"].Running())

	telemetry.monitorStatus(context.Background())
	time.Sleep(100 * time.Millisecond)
	assert.True(t, telemetry.monitors["test"].Running())
}
