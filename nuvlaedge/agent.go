package nuvlaedge

import (
	"encoding/json"
	nuvla "github.com/nuvla/api-client-go"
	"github.com/nuvla/api-client-go/clients"
	apiCommon "github.com/nuvla/api-client-go/common"
	"github.com/nuvla/api-client-go/types"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"nuvlaedge-go/nuvlaedge/common"
	"nuvlaedge-go/nuvlaedge/common/resources"
	"nuvlaedge-go/nuvlaedge/orchestrator"
	neTypes "nuvlaedge-go/nuvlaedge/types"
	"path/filepath"
	"time"
)

const (
	DefaultHeartbeatPeriod = 20
	MinHeartbeatPeriod     = 10
	DefaultTelemetryPeriod = 60
	MinTelemetryPeriod     = 30
	NuvlaSessionDataFile   = "nuvla-session.json"
	LegacySessionDataFile  = "nuvlaedge_session.json"
)

type Agent struct {
	settings *AgentSettings

	client       *clients.NuvlaEdgeClient // client: Http client library to interact with Nuvla
	coeClient    orchestrator.Coe
	telemetry    *Telemetry
	commissioner *Commissioner

	// Features
	heartBeatPeriod     int
	telemetryPeriod     int
	sentNuvlaEdgeStatus *resources.NuvlaEdgeStatus

	// Channels
	exitChan chan bool
	jobChan  chan string
}

// NewNuvlaEdgeClient tries to create a new Nuvla client first from the local files if available, else from the s
func NewNuvlaEdgeClient(settings *AgentSettings) *clients.NuvlaEdgeClient {

	clientFile := filepath.Join(DataLocation, NuvlaSessionDataFile)
	legacyFile := filepath.Join(DataLocation, LegacySessionDataFile)

	var client *clients.NuvlaEdgeClient

	// Try to load the client from the freeze file of Golang version
	if common.FileExists(clientFile) {
		log.Infof("Loading NuvlaEdge client from freeze file: %s", clientFile)
		client = NewNuvlaEdgeClientFromSessionFile(clientFile)
		// Try to load the client from the freeze file of Python version
	} else if common.FileExists(legacyFile) {
		log.Infof("Loading NuvlaEdge client from legacy freeze file: %s", legacyFile)
		client = NewNuvlaEdgeClientFromLegacySession(legacyFile)
	}

	if client != nil {
		log.Infof("Successfully created NuvlaEdge client from freeze file")
		return client
	}
	// If the freeze file does not exist, create a new client from the settings
	return NewNuvlaEdgeClientFromSettings(settings)
}

// NewNuvlaEdgeClientFromSessionFile creates a new Nuvla client from a freeze file.
func NewNuvlaEdgeClientFromSessionFile(file string) *clients.NuvlaEdgeClient {
	// Check if the file exists
	if !common.FileExists(file) {
		log.Infof("Freeze file does not exist: %s", file)
		return nil
	}
	log.Infof("Restoring NuvlaEdge client from file: %s", file)

	f := &clients.NuvlaEdgeSessionFreeze{}
	err := f.Load(file)
	if err != nil {
		log.Warnf("Error loading NuvlaEdge session freeze file: %s", err)
		return nil
	}
	return clients.NewNuvlaEdgeClientFromSessionFreeze(f)
}

// NewNuvlaEdgeClientFromLegacySession creates a new Nuvla client from a legacy freeze file.
// It first converts the legacy freeze file to the new freeze file format and then creates the client.
func NewNuvlaEdgeClientFromLegacySession(file string) *clients.NuvlaEdgeClient {
	if file == "" {
		file = filepath.Join(DataLocation, LegacySessionDataFile)
	}

	if !common.FileExists(file) {
		log.Infof("Legacy freeze file does not exist: %s", file)
		return nil
	}

	l := &neTypes.LegacySession{}
	if err := l.Load(file); err != nil {
		log.Errorf("Error loading legacy session freeze file: %s", err)
		return nil
	}

	f := l.ConvertToNuvlaSession()
	sessionFile := filepath.Join(DataLocation, NuvlaSessionDataFile)
	if err := f.Save(sessionFile); err != nil {
		log.Errorf("Error saving Nuvla session freeze file: %s", err)
		return nil
	}
	return NewNuvlaEdgeClientFromSessionFile(sessionFile)
}

func NewNuvlaEdgeClientFromSettings(settings *AgentSettings) *clients.NuvlaEdgeClient {
	var credentials *types.ApiKeyLogInParams
	if settings.ApiKey != "" && settings.ApiSecret != "" {
		credentials = types.NewApiKeyLogInParams(settings.ApiKey, settings.ApiSecret)
	}

	log.Infof("Creating NuvlaEdge client with options: %v", settings)
	client := clients.NewNuvlaEdgeClient(
		settings.NuvlaEdgeUUID,
		credentials,
		nuvla.WithEndpoint(settings.NuvlaEndpoint),
		nuvla.WithInsecureSession(settings.NuvlaInsecure))

	return client
}

func NewAgent(
	settings *AgentSettings,
	coeClient orchestrator.Coe,
	telemetry *Telemetry,
	jobChan chan string) *Agent {

	// Set default values
	return &Agent{
		settings:        settings,
		coeClient:       coeClient,
		jobChan:         jobChan,
		telemetry:       telemetry,
		telemetryPeriod: DefaultTelemetryPeriod,
		heartBeatPeriod: DefaultHeartbeatPeriod,
	}
}

/* ------------------  NuvlaEdge worker Interface implementation ------------------------------- */

func (a *Agent) Start() error {
	// Start the Agent
	// Find
	// TODO: Write a default function to generate Client opts from NuvlaEdge s
	a.client = NewNuvlaEdgeClient(a.settings)

	// We assume the client is not activated if credentials are not set in the client
	if a.client.Credentials == nil {
		err := a.client.Activate()
		if err != nil {
			log.Errorf("Error activating client: %s", err)
			return err
		}
		log.Debugf("Client activated... Success.")
	}

	// Log in with the activation credentials
	err := a.client.LogIn()
	if err != nil {
		log.Panicf("Error logging in with activation credentials: %s", err)
	}

	err = a.client.UpdateResource()

	if err != nil {
		log.Errorf("Error getting NuvlaEdge resource: %s", err)
		return err
	}

	// Freeze the client here
	freezeFile := filepath.Join(DataLocation, NuvlaSessionDataFile)
	err = a.client.Freeze(freezeFile)

	// Create commissioner
	a.commissioner = NewCommissioner(a.client, a.coeClient)

	return nil
}

func (a *Agent) updateRefreshPeriods(tickers map[string]*time.Ticker) error {
	err := a.client.UpdateResourceSelect([]string{"refresh-interval", "heartbeat-interval"})
	if err != nil {
		log.Errorf("Error retrieving intervals: %s", err)
		return err
	}
	res := a.client.GetNuvlaEdgeResource()
	// Extract the refresh and heartbeat intervals
	refreshInterval := res.RefreshInterval
	heartbeatInterval := res.HeartbeatInterval

	// Update the telemetry period and reset the ticker if necessary
	if refreshInterval != a.telemetryPeriod && refreshInterval > MinTelemetryPeriod {
		log.Infof("Updating telemetry period to %d", refreshInterval)
		a.telemetryPeriod = refreshInterval
		tickers["telemetry"].Reset(time.Duration(a.telemetryPeriod) * time.Second)
	}

	// Update the heartbeat period and reset the ticker if necessary
	if heartbeatInterval != a.heartBeatPeriod && heartbeatInterval > MinHeartbeatPeriod {
		log.Infof("Updating heartbeat period to %d", heartbeatInterval)
		a.heartBeatPeriod = heartbeatInterval
		tickers["heartbeat"].Reset(time.Duration(a.heartBeatPeriod) * time.Second)
	}

	return nil
}

func (a *Agent) sendHeartBeat() error {
	// Send heartbeat to Nuvla
	log.Infof("Sending heartbeat...")
	res, err := a.client.Heartbeat()
	if err != nil {
		log.Infof("Error sending heartbeat: %s", err)
		return err
	}

	err = a.processResponseWithJobs(res, "heartbeat")
	if err != nil {
		log.Errorf("Error processing heartbeat response: %s", err)
		return nil
	}

	return nil
}

func (a *Agent) sendTelemetry() error {
	// Run the Agent
	log.Infof("Preparing telemetry...")
	status, err := a.telemetry.GetStatusToSend()
	if err != nil {
		log.Errorf("Error getting status to send: %s", err)
		return err
	}
	log.Infof("Preparing telemetry... Success.")

	log.Infof("Sending telemetry...")
	common.CleanMap(status)
	res, err := a.client.Telemetry(status, nil)
	if err != nil {
		log.Errorf("Error sending telemetry: %s", err)
		return err
	}
	log.Infof("Sending telemetry... Success.")
	err = a.processResponseWithJobs(res, "telemetry")
	if err != nil {
		log.Errorf("Error processing telemetry response: %s", err)
	}
	return nil
}

func (a *Agent) processResponseWithJobs(res *http.Response, action string) error {
	log.Infof("Processing response with jobs...")
	body, err := io.ReadAll(res.Body)
	defer apiCommon.CloseGenericResponseWithLog(res, err)
	if err != nil {
		log.Errorf("Error reading response body: %s", err)
		return err
	}

	var sample struct {
		Message string   `json:"message"`
		Jobs    []string `json:"jobs"`
	}
	err = json.Unmarshal(body, &sample)
	if err != nil {
		log.Errorf("Error unmarshaling response body: %s", err)
		return err
	}

	if sample.Jobs != nil && len(sample.Jobs) > 0 {
		log.Infof("Jobs received: %v", sample.Jobs)
		for _, job := range sample.Jobs {
			a.jobChan <- job
		}
	}

	if sample.Message != "" {
		log.Infof("Message received: %s", sample.Message)
	}

	return nil
}

func (a *Agent) Stop() error {
	// Stop the Agent
	log.Warnf("Stopping agent...")
	a.exitChan <- true
	return nil
}

func (a *Agent) Run() error {
	// Start workers
	go a.commissioner.Run()

	// Create ticker for sendHeartBeat function
	heartbeatTicker := time.NewTicker(time.Second * time.Duration(a.heartBeatPeriod))
	defer heartbeatTicker.Stop()

	// Create ticker for sendTelemetry function
	telemetryTicker := time.NewTicker(time.Second * time.Duration(a.telemetryPeriod))
	defer telemetryTicker.Stop()
	tickers := map[string]*time.Ticker{
		"heartbeat": heartbeatTicker,
		"telemetry": telemetryTicker,
	}
	// Updater ticker
	updaterTicker := time.NewTicker(time.Second * 60)
	defer updaterTicker.Stop()
	for {
		select {
		case <-heartbeatTicker.C:
			err := a.sendHeartBeat()
			if err != nil {
				log.Errorf("Error sending heartbeat: %s", err)
			}
		case <-telemetryTicker.C:
			err := a.sendTelemetry()
			if err != nil {
				log.Errorf("Error sending telemetry: %s", err)
			}
		case <-updaterTicker.C:
			err := a.updateRefreshPeriods(tickers)
			if err != nil {
				log.Errorf("Error updating refresh periods: %s", err)
			}
		case <-a.exitChan:
			log.Infof("Exiting agent...")
			return nil
		}
	}
}

func (a *Agent) IsRunning() bool {
	// Check if the Agent is running
	return true
}
