package worker

import (
	"github.com/docker/docker/client"
	"github.com/nuvla/api-client-go/clients"
	"github.com/nuvla/api-client-go/clients/resources"
	"nuvlaedge-go/common/constants"
	"nuvlaedge-go/types"
	"nuvlaedge-go/types/jobs"
)

type WorkerConfig struct {
	TelemetryPeriod int
	HeartBeatPeriod int

	// Resource cleaner
	CleanUpPeriod int
	RemoveObjects []string

	CommissionPeriod int

	// Job Processor
	EnableJobLegacy bool
	LegacyJobImage  string
}

func NewDefaultWorkersConfig() *WorkerConfig {
	return &WorkerConfig{
		TelemetryPeriod:  constants.DefaultTelemetryPeriod,
		HeartBeatPeriod:  constants.DefaultHeartbeatPeriod,
		CleanUpPeriod:    constants.DefaultCleanUpPeriod,
		CommissionPeriod: constants.MinCommissioningPeriod,
		EnableJobLegacy:  false,
	}
}

func (wc *WorkerConfig) UpdateFromResource(res *resources.NuvlaEdgeResource) {
	wc.TelemetryPeriod = res.RefreshInterval
	wc.HeartBeatPeriod = res.HeartbeatInterval
	//wc.CleanUpPeriod = res.CleanUpPeriod
	//wc.CommissionPeriod = res.CommissionPeriod

	// TODO: Update when nuvla has the new fields
	wc.CleanUpPeriod = constants.DefaultCleanUpPeriod
	wc.RemoveObjects = []string{"images"}
}

type WorkerOpts struct {
	NuvlaClient  *clients.NuvlaEdgeClient
	DockerClient client.APIClient

	CommissionCh     chan types.CommissionData
	JobCh            chan string
	DeploymentCh     chan jobs.Job
	ConfLastUpdateCh chan string
	ConfigChannels   []chan *WorkerConfig

	// Thread safe job registry. Shared between JobProcessor and DeploymentHandler
	Jobs *jobs.JobRegistry
}
