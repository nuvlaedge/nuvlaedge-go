package src

import (
	"nuvlaedge-go/src/agent"
	"nuvlaedge-go/src/coe"
)

type NuvlaEdge struct {
	agent             agent.Agent       // Agent: Nuvla-NuvlaEdge interface manager
	coe               *coe.Coe          // coe: Orchestration engine to control deployments
	systemManager     map[string]string // systemManager: Manages the local system
	jobProcessor      map[string]string // jobProcessor: Read and execute jobs coming from Nuvla
	peripheralManager map[string]string // peripheralManager: Scans and reports discovered peripherals
	vpnClient         map[string]string // vpnClient(optional): Allows the setup of the tunnel to the provided Nuvla VPN server. Uses container orchestration Docker/K8s to set up the VPn
	telemetry         map[string]string // telemetry: Reads the local telemetry and exposes it. We provide two options, local NuvlaEdge telemetry or Prometheus exporter.
}

// NewNuvlaEdge creates a new instance of the NuvlaEdge struct, which represents a Nuvla-NuvlaEdge interface manager.
// It takes a nuvlaEdgeConfig map[string]string as a parameter, but it currently does not use it to initialize any fields.
// The function returns a pointer to the created NuvlaEdge object.
func NewNuvlaEdge(nuvlaEdgeConfig map[string]string) *NuvlaEdge {
	return &NuvlaEdge{}
}

func (ne *NuvlaEdge) run() error {
	for {
		select {}
	}
}
