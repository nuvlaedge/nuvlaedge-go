package nuvlaedge

import (
	"context"
	"github.com/nuvla/api-client-go/clients"
	"github.com/nuvla/api-client-go/clients/resources"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/types"
	"nuvlaedge-go/types/settings"
)

type Workers map[types.WorkerType]types.Worker

type NuvlaEdge struct {
	ctx  context.Context
	conf *settings.NuvlaEdgeSettings

	workers Workers

	// Channels
	commissionerChan chan types.CommissionData // Connects Telemetry/EngineMonitor with Commissioner
	jobChan          chan string               // Connects Agent and Telemetry with Job Processor
	deploymentChan   chan string               // Connects Job Processor with Deployment handler

	// Agent types
	heartBeatPeriod int
	client          *clients.NuvlaEdgeClient
	sessionOpts     *clients.NuvlaEdgeSessionFreeze

	currentState resources.NuvlaEdgeState
}

func NewNuvlaEdge(conf *settings.NuvlaEdgeSettings) (*NuvlaEdge, error) {
	ne := &NuvlaEdge{
		conf:             conf,
		workers:          make(Workers),
		commissionerChan: make(chan types.CommissionData),
		jobChan:          make(chan string),
		deploymentChan:   make(chan string),
		currentState:     resources.NuvlaEdgeStateNew,
	}

	// Validate settings
	cli, err := ValidateSettings(conf)
	if err != nil {
		return nil, err
	}

	ne.client = cli

	// Initialise NuvlaEdge:
	// Activate and trigger commission if needed
	// Trigger first telemetry

	// Initialize workers
	if err := ne.InitWorkers(); err != nil {
		return nil, err
	}

	return ne, nil
}

func (ne *NuvlaEdge) InitClient() error {
	return nil
}

func (ne *NuvlaEdge) InitWorkers() error {
	return nil
}

func (ne *NuvlaEdge) Run(ctx context.Context) error {
	if err := ne.InitWorkers(); err != nil {
		log.Errorf("Error initializing workers: %s", err)
		return err
	}

	return nil
}
