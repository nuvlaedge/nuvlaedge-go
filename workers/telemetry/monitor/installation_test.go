package monitor

import (
	"context"
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"nuvlaedge-go/testutils"
	"nuvlaedge-go/types/metrics"
	"os"
	"testing"
	"time"
)

func init() {
	// Set log level to panic to avoid logs during tests
	log.SetLevel(log.PanicLevel)
}

func Test_InstallationBase_Run_StopsGracefullyOnContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	monitor := NewHostInstallationMonitor(10, make(chan metrics.Metric))
	monitor.period = 1
	go func() {
		time.Sleep(2 * time.Millisecond)
		cancel()
	}()
	err := monitor.Run(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}

func Test_InstallationBase_Run_ReportsInstallationDataPeriodically(t *testing.T) {
	ch := make(chan metrics.Metric, 1)
	monitor := NewHostInstallationMonitor(10, ch)
	monitor.Ticker = time.NewTicker(1 * time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_ = monitor.Run(ctx)
	}()
	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Error("Expected to receive installation data within 2 seconds")
	}
}

func Test_NewInstallationMonitor(t *testing.T) {
	monitor := NewInstallationMonitor(10, nil, make(chan metrics.Metric))
	assert.NotNil(t, monitor)
}

func Test_NewHostInstallationMonitor(t *testing.T) {
	monitor := NewHostInstallationMonitor(10, make(chan metrics.Metric))
	assert.NotNil(t, monitor)
	assert.NotNil(t, monitor.updaterFunc)
}

func TestHostInstallationUpdater(t *testing.T) {
	monitor := NewHostInstallationMonitor(10, make(chan metrics.Metric))
	err := monitor.updaterFunc()
	assert.Nil(t, err)
	assert.Equal(t, monitor.installationData.ProjectName, "nuvlaedge")
}

func Test_NewDockerInstallationMonitor(t *testing.T) {
	monitor := NewDockerInstallationMonitor(10, nil, make(chan metrics.Metric))
	assert.NotNil(t, monitor)
	assert.NotNil(t, monitor.updaterFunc)
}

func TestDockerInstallationUpdater(t *testing.T) {
	dc := testutils.TestDockerMetricsClient{}

	dc.ContainerInspectReturn = types.ContainerJSON{
		Config: &container.Config{
			Labels: map[string]string{
				"com.docker.compose.project.config_files": "docker-compose.yml",
				"com.docker.compose.project":              "nuvlaedge",
				"com.docker.compose.project.working_dir":  "/home/nuvlaedge",
			},
		},
	}
	os.Setenv("COMPOSE_PROJECT_NAME", "nuvlaedge")
	defer os.Unsetenv("COMPOSE_PROJECT_NAME")

	monitor := NewDockerInstallationMonitor(10, &dc, make(chan metrics.Metric))

	err := monitor.updateInstallationData()

	assert.NoError(t, err)
	assert.Equal(t, []string{"docker-compose.yml"}, monitor.installationData.ConfigFiles)
	assert.Equal(t, "nuvlaedge", monitor.installationData.ProjectName)
	assert.Equal(t, "/home/nuvlaedge", monitor.installationData.WorkingDir)
}

func TestDockerInstallationUpdaterWithMissingEnv(t *testing.T) {
	dc := testutils.TestDockerMetricsClient{}

	os.Unsetenv("COMPOSE_PROJECT_NAME")
	monitor := NewDockerInstallationMonitor(10, &dc, make(chan metrics.Metric))

	err := monitor.updateInstallationData()

	assert.Error(t, err)
}

func TestDockerInstallationUpdaterWithMissingContainer(t *testing.T) {
	dc := testutils.TestDockerMetricsClient{}
	dc.ContainerInspectErr = errors.New("container not found")
	t.Setenv("COMPOSE_PROJECT_NAME", "nuvlaedge")
	defer os.Unsetenv("COMPOSE_PROJECT_NAME")

	monitor := NewDockerInstallationMonitor(10, &dc, make(chan metrics.Metric))

	err := monitor.updateInstallationData()

	assert.Error(t, err)
}
