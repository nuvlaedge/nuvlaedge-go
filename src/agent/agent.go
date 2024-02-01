package agent

import (
	"encoding/json"
	"fmt"
	"nuvlaedge-go/src/agent/telemetry"
	"nuvlaedge-go/src/coe"
	"nuvlaedge-go/src/common"
	"nuvlaedge-go/src/nuvlaClient"
	"os"

	"time"

	log "github.com/sirupsen/logrus"
)

const (
	DefaultHeartbeatPeriod = 20
	DefaultTelemetryPeriod = 60
)

type Config struct {
	Uuid          string `json:"nuvlabox-uuid"`
	NuvlaEndpoint string `json:"nuvla-endpoint"`

	// Possible pre-initialised credentials
	ApiKey    string `json:"api-key"`
	SecretKey string `json:"secret-key"`
}

type Agent struct {
	client    *nuvlaClient.NuvlaEdgeClient // client: Http client library to interact with Nuvla
	telemetry *telemetry.Telemetry
	coeClient coe.Coe

	// Features
	heartBeatPeriod int
	telemetryPeriod int

	// Channels
	exitChan      chan bool
	telemetryChan chan map[string]interface{}
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

func NewAgent(configFile Config, coeClient coe.Coe) *Agent {
	reportChan := make(chan map[string]interface{})
	exitChan := make(chan bool)
	return &Agent{
		client:        nuvlaClient.NewNuvlaEdgeClient(findUUIDInFile(), "https://nuvla.io", false),
		coeClient:     coeClient,
		telemetry:     telemetry.NewTelemetry(coeClient, reportChan, exitChan, DefaultTelemetryPeriod),
		telemetryChan: reportChan,
		exitChan:      exitChan,
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

func (a *Agent) Run() {
	log.Infof("Running Nuvlaedge main loop")
	// SetUp connectivity between agent
	jobsCh := make(chan []string)

	go a.telemetry.Run()
	go a.runHeartBeat(20, jobsCh)
	go a.runTelemetry(60, jobsCh)

	log.Infof("Starting Agent main loop")
	for {
		time.Sleep(time.Second)
		jobs := <-jobsCh
		log.Infof("Received jobs %s", jobs)
	}
}

func (a *Agent) runHeartBeat(period int, jobsCh chan []string) {
	for {
		startTime := time.Now()

		jobs, err := a.client.HeartBeat()
		if err != nil {
			continue
		}

		if len(jobs) > 0 {
			jobsCh <- jobs
		}

		err = common.WaitPeriodicAction(startTime, period, "Heartbeat Loop")
		if err != nil {
			panic(err)
		}
	}
}

func (a *Agent) runTelemetry(period int, jobsCh chan []string) {

	for {
		//startTime := time.Now()

		//jobs, err := a.client.Telemetry(&statusCopy)
		//common.GenericErrorHandler("error sending telemetry", err)

		//if len(jobs) > 0 {
		//	log.Infof("Jobs received in telemetry: %d", len(jobs))
		//	jobsCh <- jobs
		//}
		//err = common.WaitPeriodicAction(startTime, period, "Telemetry Loop")
		//if err != nil {
		//	panic(err)
		//}
	}
}

func (a *Agent) Pause() {
	log.Infof("Running Agent pause process")
}

func (a *Agent) Update() {
	log.Infof("Running Agent updating process")
}
