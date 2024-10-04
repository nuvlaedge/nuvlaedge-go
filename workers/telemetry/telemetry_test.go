package telemetry

import (
	"context"
	"errors"
	nuvla "github.com/nuvla/api-client-go"
	"github.com/nuvla/api-client-go/clients"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"nuvlaedge-go/common/constants"
	"strings"

	//	"io"
	//	"net/http"
	"nuvlaedge-go/testutils"
	"nuvlaedge-go/types"
	"nuvlaedge-go/types/metrics"
	"nuvlaedge-go/types/worker"
	"nuvlaedge-go/workers/telemetry/monitor"
	//	"strings"
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

func TestTelemetry_Init(t *testing.T) {
	comChan := make(chan types.CommissionData, 10)
	jChan := make(chan string, 10)

	wOpts := &worker.WorkerOpts{
		NuvlaClient:  &clients.NuvlaEdgeClient{NuvlaClient: &nuvla.NuvlaClient{SessionOpts: nuvla.SessionOptions{Endpoint: "https://mock_nuvla.io"}}},
		CommissionCh: comChan,
		JobCh:        jChan,
		DockerClient: nil,
	}
	wConf := &worker.WorkerConfig{
		TelemetryPeriod: 10,
	}

	telemetryTest := &Telemetry{}
	err := telemetryTest.Init(wOpts, wConf)

	assert.NoError(t, err)
	assert.Equal(t, 10, telemetryTest.GetPeriod())
	assert.NotNil(t, telemetryTest.nuvla)
	assert.Equal(t, jChan, telemetryTest.jobChan)

	// Assert telemetryTest.monitors
	assert.NotNil(t, telemetryTest.monitors)
	assert.Len(t, telemetryTest.monitors, 4)
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

func TestTelemetry_Start(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockTelemetryClient := &testutils.MockTelemetryClient{
		GetEndpointResponse: "https://nuvla.io",
	}
	mockDockerMetricsClient := &testutils.TestDockerMetricsClient{}
	telemetry := newTelemetry(10, mockTelemetryClient, mockDockerMetricsClient, commissionerChan, jobChan)

	go func() {
		time.Sleep(1 * time.Second)
		cancel()
	}()
	err := telemetry.Start(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 10, telemetry.GetPeriod())
	assert.Equal(t, "OPERATIONAL", telemetry.localStatus.Status)
	assert.Equal(t, 2, telemetry.localStatus.Version)
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

func TestTelemetry_GetTelemetryDiff(t *testing.T) {
	telemetry := newTelemetry(10, &testutils.MockTelemetryClient{}, &testutils.TestDockerMetricsClient{}, commissionerChan, jobChan)
	telemetry.localStatus = metrics.NuvlaEdgeStatus{
		Status:      "OPERATIONAL",
		Version:     2,
		CurrentTime: time.Now().Format(constants.DatetimeFormat),
	}
	telemetry.lastStatus = metrics.NuvlaEdgeStatus{
		Status:  "OPERATIONAL",
		Version: 1,
	}

	patch, data, attrsToDelete := telemetry.getTelemetryDiff()

	assert.NotNil(t, patch)
	assert.NotNil(t, data)
	assert.Empty(t, attrsToDelete)
}

func Test_Telemetry_SendTelemetry_WithNilClient_ReturnError(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	telemetry := newTelemetry(10, &testutils.MockTelemetryClient{}, &testutils.TestDockerMetricsClient{}, commissionerChan, jobChan)
	telemetry.nuvla = nil
	err := telemetry.sendTelemetry(ctx, nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "telemetry client not initialized, cannot send telemetry")
}

func Test_Telemetry_SendTelemetry_WithNilData_ReturnNil(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	telemetry := newTelemetry(10, &testutils.MockTelemetryClient{}, &testutils.TestDockerMetricsClient{}, commissionerChan, jobChan)
	err := telemetry.sendTelemetry(ctx, nil, nil)
	assert.NoError(t, err)
	assert.Nil(t, err)
	err = nil
}

func Test_Telemetry_SendTelemetry_WithData_ResponseError(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	telemetry := newTelemetry(10, &testutils.MockTelemetryClient{TelemetryResponse: nil, TelemetryErr: assert.AnError}, &testutils.TestDockerMetricsClient{}, commissionerChan, jobChan)
	err := telemetry.sendTelemetry(ctx, "data", nil)
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}

func Test_Telemetry_SendTelemetry_WithData_ResponseOK(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	res := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(`{"key": "value"}`)),
	}
	telemetry := newTelemetry(10, &testutils.MockTelemetryClient{TelemetryResponse: res}, &testutils.TestDockerMetricsClient{}, commissionerChan, jobChan)
	telemetry.localStatus = metrics.NuvlaEdgeStatus{
		Status: "OPERATIONAL",
	}
	err := telemetry.sendTelemetry(ctx, "data", nil)
	assert.NoError(t, err)
	assert.Equal(t, "OPERATIONAL", telemetry.lastStatus.Status)

	res.StatusCode = 400
	err = telemetry.sendTelemetry(ctx, "data", nil)
	assert.Error(t, err)
	assert.Equal(t, "telemetry failed with status code: 400", err.Error())
}

func Test_Telemetry_Reconfigure_UpdatesPeriod_WhenPeriodChanges(t *testing.T) {
	t.Parallel()
	telemetry := newTelemetry(10, &testutils.MockTelemetryClient{}, &testutils.TestDockerMetricsClient{}, commissionerChan, jobChan)
	conf := &worker.WorkerConfig{TelemetryPeriod: 20}

	err := telemetry.Reconfigure(conf)

	assert.NoError(t, err)
	assert.Equal(t, 20, telemetry.GetPeriod())
}

func Test_Telemetry_Reconfigure_DoesNotUpdatePeriod_WhenPeriodUnchanged(t *testing.T) {
	t.Parallel()
	telemetry := newTelemetry(10, &testutils.MockTelemetryClient{}, &testutils.TestDockerMetricsClient{}, commissionerChan, jobChan)
	conf := &worker.WorkerConfig{TelemetryPeriod: 10}

	err := telemetry.Reconfigure(conf)

	assert.NoError(t, err)
	assert.Equal(t, 10, telemetry.GetPeriod())
}

func Test_Telemetry_Reconfigure_HandlesNilConfig(t *testing.T) {
	t.Parallel()
	telemetry := newTelemetry(10, &testutils.MockTelemetryClient{}, &testutils.TestDockerMetricsClient{}, commissionerChan, jobChan)

	err := telemetry.Reconfigure(nil)

	// Assert error contains 'nil configuration received'
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil configuration received")
	assert.Equal(t, 10, telemetry.GetPeriod())
}

func Test_Telemetry_Stop(t *testing.T) {
	t.Parallel()
	telemetry := newTelemetry(10, &testutils.MockTelemetryClient{}, &testutils.TestDockerMetricsClient{}, commissionerChan, jobChan)
	m1 := testutils.NewMonitorMock()
	m2 := testutils.NewMonitorMock()
	telemetry.monitors = map[string]monitor.NuvlaEdgeMonitor{
		"test1": m1,
		"test2": m2,
	}

	err := telemetry.Stop(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, m1.CloseCnt)
	assert.Equal(t, 1, m2.CloseCnt)

	m1.CloseErr = assert.AnError

	err = telemetry.Stop(context.Background())
	assert.Error(t, err)
	// Assert err is assert.AnError
	assert.Equal(t, errors.Join(assert.AnError), err)
	assert.Equal(t, 2, m1.CloseCnt)
	assert.Equal(t, 2, m2.CloseCnt)
}

func Test_Telemetry_Run_ProcessesMetricsAndStopsOnContextCancel(t *testing.T) {
	t.Parallel()
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
}

func Test_Telemetry_Run_SendsTelemetryOnTicker(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockTelemetryClient := &testutils.MockTelemetryClient{
		GetEndpointResponse: "https://nuvla.io",
		TelemetryResponse: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(`{"key": "value"}`))},
	}
	mockDockerMetricsClient := &testutils.TestDockerMetricsClient{}
	telemetry := newTelemetry(1, mockTelemetryClient, mockDockerMetricsClient, commissionerChan, jobChan)
	telemetry.BaseTicker = time.NewTicker(100 * time.Millisecond)

	telemetry.localStatus = metrics.NuvlaEdgeStatus{
		Status: "OPERATIONAL",
	}
	telemetry.lastStatus = metrics.NuvlaEdgeStatus{
		Status: "NON-OPERATIONAL",
	}

	go func() {
		time.Sleep(1 * time.Second)
		cancel()
	}()
	err := telemetry.Run(ctx)
	assert.Equal(t, context.Canceled, err)
	assert.Equal(t, "OPERATIONAL", telemetry.lastStatus.Status)
}

func Test_Telemetry_Run_ReconfiguresOnConfigChange(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockTelemetryClient := &testutils.MockTelemetryClient{
		GetEndpointResponse: "https://nuvla.io",
	}
	mockDockerMetricsClient := &testutils.TestDockerMetricsClient{}
	telemetry := newTelemetry(10, mockTelemetryClient, mockDockerMetricsClient, commissionerChan, jobChan)
	newConf := &worker.WorkerConfig{TelemetryPeriod: 20}
	go func() {
		telemetry.ConfChan <- newConf
		time.Sleep(1 * time.Second)
		cancel()
	}()
	err := telemetry.Run(ctx)
	assert.Equal(t, context.Canceled, err)
	assert.Equal(t, 20, telemetry.GetPeriod())
}
