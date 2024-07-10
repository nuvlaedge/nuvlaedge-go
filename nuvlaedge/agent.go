package nuvlaedge

import (
	"context"
	"encoding/json"
	nuvla "github.com/nuvla/api-client-go"
	"github.com/nuvla/api-client-go/clients"
	apiCommon "github.com/nuvla/api-client-go/common"
	"github.com/nuvla/api-client-go/types"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"nuvlaedge-go/nuvlaedge/common/resources"
	"nuvlaedge-go/nuvlaedge/orchestrator"
	"path"
	"path/filepath"
	"time"
)

const (
	DefaultHeartbeatPeriod = 20
	MinHeartbeatPeriod     = 10
	DefaultTelemetryPeriod = 60
	MinTelemetryPeriod     = 30
	NuvlaSessionDataFile   = "nuvlaedge_session.json"
)

type Agent struct {
	settings *AgentSettings
	ctx      context.Context

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
func NewNuvlaEdgeClient(settings *AgentSettings) (*clients.NuvlaEdgeClient, error) {

	sessionFile := path.Join(DataLocation, NuvlaSessionDataFile)
	err := settings.CheckMinimumSettings(sessionFile)
	if err != nil {
		log.Errorf("Error checking minimum settings: %s", err)
		return nil, err
	}
	// Print settings
	log.Infof("Settings: %+v", settings)
	log.Infof("Creating NuvlaEdge with ID %s", settings.NuvlaEdgeUUID)
	// If the freeze file does not exist, create a new client from the settings
	return NewNuvlaEdgeClientFromSettings(settings), nil
}

func NewNuvlaEdgeClientFromSettings(settings *AgentSettings) *clients.NuvlaEdgeClient {
	var credentials *types.ApiKeyLogInParams
	if settings.ApiKey != "" && settings.ApiSecret != "" {
		credentials = types.NewApiKeyLogInParams(settings.ApiKey, settings.ApiSecret)
	}

	client := clients.NewNuvlaEdgeClient(
		settings.NuvlaEdgeUUID,
		credentials,
		nuvla.WithEndpoint(settings.NuvlaEndpoint),
		nuvla.WithInsecureSession(settings.NuvlaInsecure))

	return client
}

func NewAgent(
	ctx context.Context,
	settings *AgentSettings,
	coeClient orchestrator.Coe,
	telemetry *Telemetry,
	jobChan chan string) *Agent {

	// Set default values
	return &Agent{
		ctx:             ctx,
		settings:        settings,
		coeClient:       coeClient,
		jobChan:         jobChan,
		telemetry:       telemetry,
		telemetryPeriod: DefaultTelemetryPeriod,
		heartBeatPeriod: DefaultHeartbeatPeriod,
	}
}

/* ------------------  NuvlaEdge worker Interface implementation ------------------------------- */

// Start initialises the NuvlaEdge.
// The initialisation executes the following steps:
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
func (a *Agent) Start() error {
	c, err := NewNuvlaEdgeClient(a.settings)
	if err != nil {
		log.Errorf("Error creating NuvlaEdge client: %s", err)
		return err
	}
	a.client = c

	// We assume the client is not activated if credentials are not set in the client
	if a.client.Credentials == nil || a.client.Credentials.Key == "" || a.client.Credentials.Secret == "" {
		err := a.client.Activate()
		if err != nil {
			log.Errorf("Error activating client: %s", err)
			return err
		}
		log.Debugf("Client activated... Success.")
	}

	// Log in with the activation credentials
	err = a.client.LogIn()
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
	a.commissioner = NewCommissioner(a.ctx, a.client, a.coeClient)
	// Run first iteration for the commissioner
	a.commissioner.SingleIteration()

	// Run a heart beat
	err = a.sendHeartBeat()
	if err != nil {
		log.Errorf("Error sending heartbeat: %s", err)
		return err
	}

	// Run a telemetry
	err = a.sendTelemetry()
	if err != nil {
		log.Errorf("Error sending telemetry: %s", err)
		return err
	}

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
	status, del := a.telemetry.GetStatusToSend()
	log.Infof("Preparing telemetry... Success.")

	if len(del) == 0 {
		del = nil
	}
	log.Infof("Sending telemetry...")
	log.Debugf("Sending a total of %d metrics", len(status))
	log.Debugf("Deleting a total of %d metrics", len(del))
	res, err := a.client.Telemetry(status, del)
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
		case <-a.ctx.Done():
			log.Infof("Exiting agent...")
			return nil
		}
	}
}

func (a *Agent) IsRunning() bool {
	// Check if the Agent is running
	return true
}
