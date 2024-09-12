package engine

import (
	"io"
	"nuvlaedge-go/types/jobs"
	"time"
)

type CoeType string

const (
	DockerType     CoeType = "swarm"
	KubernetesType CoeType = "kubernetes"
)

type Coe interface {
	RunContainer(image string, configuration map[string]string) (string, error)
	RunJobEngineContainer(conf *jobs.LegacyJobConf) (string, error)
	StopContainer(containerId string, force bool) (bool, error)
	RemoveContainer(containerId string, containerName string) (bool, error)
	GetContainerLogs(containerId string, since string) (io.ReadCloser, error)
	WaitContainerFinish(containerId string, timeout time.Duration, printLogs bool) (int64, error)
}
