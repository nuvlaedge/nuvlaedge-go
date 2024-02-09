package nuvlaedge

import (
	"nuvlaedge-go/nuvlaedge/agent"
	"nuvlaedge-go/nuvlaedge/coe"
	"nuvlaedge-go/nuvlaedge/common"
	"nuvlaedge-go/nuvlaedge/common/resources"
	"nuvlaedge-go/nuvlaedge/nuvlaClient"
	"reflect"

	"time"
)

const (
	DefaultHeartbeatPeriod = 20
	DefaultTelemetryPeriod = 60
)

type Agent struct {
	settings *AgentSettings

	client       *nuvlaClient.NuvlaEdgeClient // client: Http client library to interact with Nuvla
	coeClient    coe.Coe
	telemetry    *Telemetry
	commissioner *agent.Commissioner

	// Features
	heartBeatPeriod     int
	telemetryPeriod     int
	sentNuvlaEdgeStatus *resources.NuvlaEdgeStatus

	// Channels
	exitChan chan bool
	jobChan  chan []string
}

func NewAgent(
	settings *AgentSettings,
	coeClient coe.Coe,
	telemetry *Telemetry,
	jobChan chan []string) *Agent {

	// Set default values
	return &Agent{
		settings:  settings,
		coeClient: coeClient,
		jobChan:   jobChan,
		telemetry: telemetry,
	}
}

/* ------------------  NuvlaEdge worker Interface implementation ------------------------------- */

// Start runs the initial setup of the agent.
// Check if there is a nuvlaedge already installed in the system
// Synchronizes resources from Nuvla: nuvlabox, nuvlabox-status, vpn-credential, etc
// It also runs the commissioner start process which will retrieve the information of the local system beforehand
func (a *Agent) Start() error {
	log.Infof("Running Agent starting process")

	// 1. Check Previous installation
	params, err := agent.FindPreviousInstallation(a.settings.ConfPath, a.settings.NuvlaEdgeUUID)
	if err != nil {
		log.Infof("Error finding previous installation: %s", err)
	}

	if params == nil {
		log.Infof("No previous installation found")
	} else {
		log.Infof("Found previous installation with uuid: %s", a.settings.NuvlaEdgeUUID)
	}
	a.client = nuvlaClient.NewNuvlaEdgeClient(a.settings.NuvlaEdgeUUID, a.settings.NuvlaEndpoint, a.settings.NuvlaInsecure)
	// 2. Try retrieving NuvlaEdge status from Nuvla
	// 3. Base on the state execute the activation or commission process
	return nil
}

func (a *Agent) Stop() error {
	return nil
}

func (a *Agent) Run() error {
	log.Infof("Starting Agent main loop")
	go a.runHeartBeat()
	go a.runTelemetry()
	return nil
}

func (a *Agent) activate() {
	log.Infof("Running Agent activation process")
}

func (a *Agent) commission() {
	log.Infof("Running Agent commission process")
}

func (a *Agent) runHeartBeat() {
	for {
		startTime := time.Now()

		jobs, err := a.client.HeartBeat()
		if err != nil {
			log.Warnf("Error sending heartbeat: %s", err)
			// Wait for the next heartbeat
			err = common.WaitPeriodicAction(startTime, a.heartBeatPeriod, "Heartbeat Loop")
			if err != nil {
				log.Errorf("Error waiting for heartbeat: %s", err)
			}
			continue
		}

		if len(jobs) > 0 {
			a.jobChan <- jobs
		}
		// Wait for the next heartbeat
		err = common.WaitPeriodicAction(startTime, a.heartBeatPeriod, "Heartbeat Loop")
		if err != nil {
			log.Errorf("Error waiting for heartbeat: %s", err)
		}
	}
}

// Function that takes a pointer to NuvlaEdge status and compares it with the local NuvlaEdgeStatus.
// It returns two slices, one contains the fields that have changes from the local status and the other and
// the fields that are no longer present in the parsed status.
func (a *Agent) compareNuvlaEdgeStatus(newNuvlaEdgeStatus *resources.NuvlaEdgeStatus) ([]string, []string) {
	v1 := reflect.ValueOf(a.sentNuvlaEdgeStatus).Elem()
	v2 := reflect.ValueOf(newNuvlaEdgeStatus).Elem()
	changedFields := make([]string, 0)
	removedFields := make([]string, 0)

	for i := 0; i < v1.NumField(); i++ {
		fieldName := v1.Type().Field(i).Name
		value1 := v1.Field(i).Interface()
		value2 := v2.FieldByName(fieldName).Interface()

		if value2 != value1 {
			changedFields = append(changedFields, fieldName)
		}
	}

	for i := 0; i < v1.NumField(); i++ {
		fieldName := v1.Type().Field(i).Name
		_, ok := v2.Type().FieldByName(fieldName)

		if !ok {
			removedFields = append(removedFields, fieldName)
		}
	}
	return changedFields, removedFields
}

func (a *Agent) runTelemetry() {
	for {
		startTime := time.Now()

		status, toDelete := a.telemetry.GetStatusToSend()
		if status == nil {
			return
		}
		jobs, err := a.client.Telemetry(status, toDelete)
		if len(jobs) > 0 {
			a.jobChan <- jobs
		}

		err = common.WaitPeriodicAction(startTime, a.telemetryPeriod, "MetricsMonitor Loop")
		if err != nil {
			panic(err)
		}
	}
}
