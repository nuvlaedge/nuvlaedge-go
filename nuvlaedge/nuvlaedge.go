package nuvlaedge

import (
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/nuvlaedge/common"
	"nuvlaedge-go/nuvlaedge/orchestrator"
	"os"
	"time"
)

var DataLocation string

func init() {
	DataLocation = common.DefaultDBPath
}

type NuvlaEdge struct {
	coe          orchestrator.Coe // coe: Orchestration engine to control deployments
	settings     *Settings        // settings:
	agent        *Agent           // Agent: Nuvla-NuvlaEdge interface manager
	jobProcessor *JobProcessor    // jobProcessor: Read and execute jobs coming from Nuvla
	telemetry    *Telemetry       // telemetry: Reads the local telemetry and exposes it. We provide two options, local NuvlaEdge telemetry or Prometheus exporter.
}

func NewNuvlaEdge(settings *Settings) *NuvlaEdge {
	// Set global data location
	DataLocation = settings.DataLocation
	log.Infof("Setting global data location to: %s", DataLocation)

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
		coe:      coeClient,
		settings: settings,
		agent:    NewAgent(&settings.Agent, coeClient, telemetry, jobChan),
		// Job engine needs a pointer to Nuvla Client which is created by the agent depending on input s
		// and previous installations so, we defer the creation of the jobProcessor to the start method
		telemetry: telemetry,
	}
}

// Start starts the NuvlaEdge, initialising all the components by calling their Start() method. Each component is responsible for its own initialisation.
// Requirements check: Checks if the local system meets the requirements to run NuvlaEdge
// Agent: Initialises the agent, reads the local storage for previous installations of NuvlaEdge and acts accordingly
// SystemManager: Initialises the local system based on the s
// JobProcessor: Reads the local storage for dangling jobs/deployments
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

	// Start JobProcessor
	ne.jobProcessor = NewJobProcessor(
		ne.agent.jobChan,
		ne.agent.client.GetNuvlaClient(),
		ne.coe,
		ne.settings.Agent.EnableLegacyJobSupport,
		ne.settings.Agent.JobEngineImage)

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
	go ne.agent.Run()
	_ = ne.jobProcessor.Run()
	for {
		time.Sleep(1 * time.Second)
	}
}
