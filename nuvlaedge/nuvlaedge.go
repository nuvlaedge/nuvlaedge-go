package nuvlaedge

import (
	"context"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/nuvlaedge/common"
	"nuvlaedge-go/nuvlaedge/orchestrator"
	"time"
)

var DataLocation string

func init() {
	DataLocation = common.DefaultDBPath
}

type NuvlaEdge struct {
	ctx          context.Context  // ctx: Context
	coe          orchestrator.Coe // coe: Orchestration engine to control deployments
	settings     *Settings        // settings:
	agent        *Agent           // Agent: Nuvla-NuvlaEdge interface manager
	jobProcessor *JobProcessor    // jobProcessor: Read and execute jobs coming from Nuvla
	telemetry    *Telemetry       // telemetry: Reads the local telemetry and exposes it. We provide two options, local NuvlaEdge telemetry or Prometheus exporter.
}

func NewNuvlaEdge(ctx context.Context, settings *Settings) *NuvlaEdge {
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
		ctx:      ctx,
		coe:      coeClient,
		settings: settings,
		agent:    NewAgent(ctx, &settings.Agent, coeClient, telemetry, jobChan),
		// Job engine needs a pointer to Nuvla Client which is created by the agent depending on input s
		// and previous installations so, we defer the creation of the jobProcessor to the start method
		telemetry: telemetry,
	}
}

// Initialise initialises the NuvlaEdge.
// The initialisation process consists in the creation of all the components and the execution
// of the first required operations as follows:
// - Check minimum settings
//   - This should consider the local nuvlaedge-session file
//
// - Agent: Initialises the agent
//   - Activate if required
//   - Commission if not done already
//   - Send first heartbeat
//
// - Run first telemetry sweep
// - Send first telemetry
// None of these steps are executed in a routine, they should be blocking operations to ensure the correct
// initialisation of the system.
func (ne *NuvlaEdge) Initialise() error {
	return nil
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

func (ne *NuvlaEdge) Run(errChan chan error) {
	log.Infof("Running NuvlaEdge...")
	go func() {
		waitTime := 5
		for {
			err := ne.agent.Run()
			if err == nil {
				log.Warn("Agent has been stopped, exiting")
				return
			}
			// Wait 5 seconds before restarting the agent
			log.Warnf("Agent has been stopped due to an error, restarting in %d seconds", waitTime)
			errChan <- err
			time.Sleep(time.Duration(waitTime) * time.Second)
			waitTime += waitTime
		}
	}()

	_ = ne.jobProcessor.Run()

	select {
	case <-ne.ctx.Done():
		log.Info("NuvlaEdge has been stopped")
	}
}

func (ne *NuvlaEdge) CheckMinimumSettings() error {
	log.Info("Checking NuvlaEdge minimum settings...")
	// Check if session file exists, if so, load it
	if common.FileExists(DataLocation + "nuvlaedge-session.json") {
		log.Info("Session file exists, loading it...")

	}

	return nil
}
