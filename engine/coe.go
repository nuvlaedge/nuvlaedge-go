package engine

import (
	"context"
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
	RunContainer(ctx context.Context, image string, configuration map[string]string) (string, error)
	RunJobEngineContainer(ctx context.Context, conf *jobs.LegacyJobConf) (string, error)
	StopContainer(ctx context.Context, containerId string, force bool) (bool, error)
	RemoveContainer(ctx context.Context, containerId string, containerName string) (bool, error)
	GetContainerLogs(ctx context.Context, containerId string, since string) (io.ReadCloser, error)
	WaitContainerFinish(ctx context.Context, containerId string, timeout time.Duration, printLogs bool) (int64, error)
}
