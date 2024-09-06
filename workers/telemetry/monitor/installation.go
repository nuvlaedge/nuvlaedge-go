package monitor

import (
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/common"
	"nuvlaedge-go/types"
	"nuvlaedge-go/types/metrics"
	"os"
	"strings"
)

type InstallationBase struct {
	BaseMonitor
	installationData metrics.InstallationParameters
	updaterFunc      func() error
}

func (ib *InstallationBase) Run(ctx context.Context) error {
	ib.SetRunning()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ib.Ticker.C:
			// Send metric to channel
			if err := ib.updaterFunc(); err != nil {
				log.Errorf("Error updating installation metrics: %s", err)
			}
			ib.reportChan <- ib.installationData
		}
	}
}

type HostInstallationMonitor struct {
	InstallationBase
}

func NewInstallationMonitor(period int, dclient types.InstallationParametersClient, reportChan chan metrics.Metric) NuvlaEdgeMonitor {
	switch common.WhereAmI() {
	case common.Host:
		return NewHostInstallationMonitor(period, reportChan)
	case common.Docker:
		return NewDockerInstallationMonitor(period, dclient, reportChan)
	default:
		return nil
	}
}

func NewHostInstallationMonitor(period int, ch chan metrics.Metric) *HostInstallationMonitor {
	i := &HostInstallationMonitor{
		InstallationBase: InstallationBase{
			BaseMonitor: NewBaseMonitor(period, ch),
		},
	}
	i.updaterFunc = i.updateInstallationData
	return i
}

func (im *HostInstallationMonitor) updateInstallationData() error {
	im.installationData.ConfigFiles = []string{"/bin/nuvlaedge"}
	im.installationData.ProjectName = "nuvlaedge"
	im.installationData.Environment = os.Environ()
	dir, err := os.Getwd()
	if err == nil {
		im.installationData.WorkingDir = dir
	}

	return nil
}

type DockerInstallationMonitor struct {
	InstallationBase
	client types.InstallationParametersClient
}

func NewDockerInstallationMonitor(period int, dclient types.InstallationParametersClient, ch chan metrics.Metric) *DockerInstallationMonitor {
	i := &DockerInstallationMonitor{
		InstallationBase: InstallationBase{
			BaseMonitor: NewBaseMonitor(period, ch),
		},
		client: dclient,
	}
	i.updaterFunc = i.updateInstallationData
	return i
}

func (dim *DockerInstallationMonitor) updateInstallationData() error {
	dim.installationData.Environment = os.Environ()
	pName := os.Getenv("COMPOSE_PROJECT_NAME")
	if pName == "" {
		return errors.New("COMPOSE_PROJECT_NAME not set")
	}
	containerName := fmt.Sprintf("%s-agent-go", pName)
	inspect, err := dim.client.ContainerInspect(context.Background(), containerName)
	if err != nil {
		return err
	}
	dim.installationData.ConfigFiles = strings.Split(inspect.Config.Labels["com.docker.compose.project.config_files"], ",")
	dim.installationData.ProjectName = inspect.Config.Labels["com.docker.compose.project"]
	dim.installationData.WorkingDir = inspect.Config.Labels["com.docker.compose.project.working_dir"]
	return nil
}
