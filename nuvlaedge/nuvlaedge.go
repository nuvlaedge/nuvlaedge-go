package nuvlaedge

import (
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/nuvlaedge/coe"
	"os"
)

type NuvlaEdge struct {
	coe           coe.Coe            // coe: Orchestration engine to control deployments
	settings      *NuvlaEdgeSettings // settings:
	agent         *Agent             // Agent: Nuvla-NuvlaEdge interface manager
	systemManager *SystemManager     // systemManager: Manages the local system
	jobProcessor  *JobProcessor      // jobProcessor: Read and execute jobs coming from Nuvla
	vpnHandler    map[string]string  // vpnHandler(optional): Allows the setup of the tunnel to the provided Nuvla VPN server. Uses container orchestration Docker/K8s to set up the VPn
	telemetry     map[string]string  // telemetry: Reads the local telemetry and exposes it. We provide two options, local NuvlaEdge telemetry or Prometheus exporter.
}

func NewNuvlaEdge(settings *NuvlaEdgeSettings) *NuvlaEdge {
	coeClient, err := coe.NewCoe(coe.DockerType)
	if err != nil {
		log.Errorf("Error creating COE client: %s", err)
	}

	// Create channels
	// hostConfigChan: SystemManager -> Agent & VPNHandler -> Agent
	hostConfigChan := make(chan map[string]interface{}, 3)
	// telemetryChan: Telemetry -> Agent
	telemetryChan := make(chan map[string]interface{}, 3)

	// jobChan: Agent -> JobProcessor
	jobChan := make(chan []string, 10)

	return &NuvlaEdge{
		agent:         NewAgent(&settings.Agent, coeClient, hostConfigChan, telemetryChan, jobChan),
		coe:           coeClient,
		systemManager: NewSystemManager(&settings.SystemManager, coeClient),
		jobProcessor:  NewJobProcessor(jobChan),
		vpnHandler:    make(map[string]string),
		telemetry:     make(map[string]string),
	}
}

func (ne *NuvlaEdge) Start() error {
	return nil
}

func (ne *NuvlaEdge) Run() (os.Signal, error) {
	for {
		select {}
	}
}
