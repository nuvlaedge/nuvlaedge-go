package nuvlaedge

import (
	"github.com/sirupsen/logrus"
	"nuvlaedge-go/nuvlaedge/orchestrator"
	"os"
	"time"
)

var log = logrus.New()

type NuvlaEdge struct {
	coe           orchestrator.Coe   // coe: Orchestration engine to control deployments
	settings      *NuvlaEdgeSettings // settings:
	agent         *Agent             // Agent: Nuvla-NuvlaEdge interface manager
	systemManager *SystemManager     // systemManager: Manages the local system
	jobProcessor  *JobProcessor      // jobProcessor: Read and execute actions coming from Nuvla
	telemetry     *Telemetry         // telemetry: Reads the local telemetry and exposes it. We provide two options, local NuvlaEdge telemetry or Prometheus exporter.
}

func NewNuvlaEdge(settings *NuvlaEdgeSettings) *NuvlaEdge {
	coeClient, err := orchestrator.NewCoe(orchestrator.DockerType)
	if err != nil {
		log.Errorf("Error creating COE client: %s", err)
	}

	// MetricsMonitor
	log.Infof("Creating MetricsMonitor with period %d", settings.Agent.TelemetryPeriod)

	telemetry := NewTelemetry(coeClient, settings.Agent.TelemetryPeriod)

	// jobChan: Agent -> JobProcessor
	jobChan := make(chan string, 10)

	return &NuvlaEdge{
		coe:           coeClient,
		settings:      settings,
		agent:         NewAgent(&settings.Agent, coeClient, telemetry, jobChan),
		systemManager: NewSystemManager(&settings.SystemManager, coeClient),
		jobProcessor:  NewJobProcessor(jobChan),
		telemetry:     telemetry,
	}
}

// Start starts the NuvlaEdge, initialising all the components by calling their Start() method. Each component is responsible for its own initialisation.
// Requirements check: Checks if the local system meets the requirements to run NuvlaEdge
// Agent: Initialises the agent, reads the local storage for previous installations of NuvlaEdge and acts accordingly
// SystemManager: Initialises the local system based on the settings
// JobProcessor: Reads the local storage for dangling actions/deployments
// Telemetry: Starts the telemetry collection right away
func (ne *NuvlaEdge) Start() error {
	// Run requirements check
	log.Infof("Running requirements check")

	// Run COE telemetry initial sweep
	err := ne.coe.TelemetryStart()
	if err != nil {
		log.Errorf("Error starting COE telemetry: %s, cannot continue", err)
		return err
	}

	// Start Agent
	err = ne.agent.Start()
	if err != nil {
		log.Errorf("Error starting Agent: %s, cannot continue", err)
		return err
	}

	// Start SystemManager
	err = ne.systemManager.Start()
	if err != nil {
		log.Errorf("Error starting SystemManager: %s, cannot continue", err)
		return err
	}

	// Start JobProcessor
	err = ne.jobProcessor.Start()
	if err != nil {
		log.Errorf("Error starting JobProcessor: %s, cannot continue", err)
		return err
	}

	// Start Telemetry
	err = ne.telemetry.Start()
	if err != nil {
		log.Errorf("Error starting Telemetry: %s, cannot continue", err)
		return err
	}

	return nil
}

func (ne *NuvlaEdge) Run() (os.Signal, error) {
	log.Infof("Running NuvlaEdge...")
	_ = ne.agent.Run()
	_ = ne.jobProcessor.Run()
	for {
		time.Sleep(1 * time.Second)
	}
}
