package nuvlaedge

import (
	"encoding/json"
	"fmt"
	"nuvlaedge-go/nuvlaedge/coe"
	"nuvlaedge-go/nuvlaedge/common"
	"nuvlaedge-go/nuvlaedge/common/resources"
	"nuvlaedge-go/nuvlaedge/nuvlaClient"
	"os"

	"time"

	log "github.com/sirupsen/logrus"
)

const (
	DefaultHeartbeatPeriod = 20
	DefaultTelemetryPeriod = 60
)

type Agent struct {
	settings *AgentSettings

	client    *nuvlaClient.NuvlaEdgeClient // client: Http client library to interact with Nuvla
	coeClient coe.Coe

	// Features
	heartBeatPeriod int
	telemetryPeriod int

	// Channels
	exitChan       chan bool
	telemetryChan  chan resources.NuvlaEdgeStatus
	hostConfigChan chan map[string]interface{}
	jobChan        chan []string
}

func findUUIDInFile() string {
	file, err := os.Open("config/temp_uuid.json")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return ""
	}
	defer file.Close()

	var data map[string]string
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return ""
	}

	uuid := data["uuid"]
	log.Infof("Found UUID: %s", uuid)

	return uuid
}

func NewAgent(
	settings *AgentSettings,
	coeClient coe.Coe,
	hostConfigChan chan map[string]interface{},
	telemetryChan chan resources.NuvlaEdgeStatus,
	jobChan chan []string,
) *Agent {
	// Set default values
	c := nuvlaClient.NewNuvlaEdgeClient(settings.NuvlaEdgeUUID, settings.NuvlaEndpoint, settings.NuvlaInsecure)
	return &Agent{
		settings:       settings,
		client:         c,
		coeClient:      coeClient,
		telemetryChan:  telemetryChan,
		hostConfigChan: hostConfigChan,
		jobChan:        jobChan,
	}
}

func (a *Agent) Start() {
	log.Infof("Running Agent starting process")
	// Assert local Agent state
	log.Infof("Actiavting my self")
	// Activate
	err := a.client.Activate()
	if err != nil {
		panic("NuvlaEdge not activated properly, exiting")
	}

	err = a.client.GetNuvlaEdgeInformation()
	common.GenericErrorHandler("error retrieving nuvlaedge information", err)
}

func (a *Agent) Stop() {
	log.Infof("Running Agent stopping process")
}

func (a *Agent) activate() {
	log.Infof("Running Agent activation process")
}

func (a *Agent) commission() {
	log.Infof("Running Agent commission process")
}

func (a *Agent) Run() {
	log.Infof("Running Nuvlaedge main loop")

	log.Infof("Starting Agent main loop")
	go a.runHeartBeat()
	go a.runTelemetry()

}

func (a *Agent) runHeartBeat() {
	for {
		startTime := time.Now()

		jobs, err := a.client.HeartBeat()
		if err != nil {
			continue
		}

		if len(jobs) > 0 {
			a.jobChan <- jobs
		}

		err = common.WaitPeriodicAction(startTime, a.heartBeatPeriod, "Heartbeat Loop")
		if err != nil {
			panic(err)
		}
	}
}

func (a *Agent) runTelemetry() {

}

func (a *Agent) Pause() {
	log.Infof("Running Agent pause process")
}

func (a *Agent) Update() {
	log.Infof("Running Agent updating process")
}
