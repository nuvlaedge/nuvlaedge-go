package telemetry

import (
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/v3/host"
	log "github.com/sirupsen/logrus"
	"native-nuvlaedge/src/coe"
	"native-nuvlaedge/src/common"
	"native-nuvlaedge/src/common/resources"
	"os"
	"reflect"
	"runtime"
	"sync"
	"time"
)

type UpdaterFunction func(chan<- error)

type UpdaterError struct {
	Err         error
	UpdaterName string
}

type Telemetry struct {
	// Base telemetry data
	nuvlaEdgeStatus *resources.NuvlaEdgeStatus

	// Helper map to gather and run all the update functions
	updateFuncs map[string]UpdaterFunction

	// Metric reading updaters
	coeClient               coe.Coe // Interfaces need no pointer
	resourcesMetricsUpdater *ResourceMetricsUpdater
	networkMetricsUpdater   *NetworkMetricsUpdater

	// Report and exit channels
	reportChan chan map[string]interface{}
	exitChan   chan bool

	// Telemetry report update period
	updatePeriod int
}

func NewTelemetry(
	coeClient coe.Coe,
	reportChan chan map[string]interface{},
	exitChan chan bool,
	updatePeriod int) *Telemetry {

	networkUpdater := NewNetworkMetricsUpdater()
	t := &Telemetry{
		nuvlaEdgeStatus:         &resources.NuvlaEdgeStatus{},
		updateFuncs:             make(map[string]UpdaterFunction),
		coeClient:               coeClient,
		resourcesMetricsUpdater: NewResourceMetricsUpdater(networkUpdater),
		networkMetricsUpdater:   networkUpdater,
		reportChan:              reportChan,
		exitChan:                exitChan,
		updatePeriod:            updatePeriod,
	}
	val := reflect.ValueOf(*t.nuvlaEdgeStatus)
	log.Infof("Number of fields: %d", val.NumField())
	for i := 0; i < val.NumField(); i++ {
		fieldName := val.Type().Field(i).Name
		updaterName := "Updater" + fieldName
		updaterFunc := reflect.ValueOf(t).MethodByName(updaterName)
		if updaterFunc.IsValid() {
			t.updateFuncs[fieldName] = updaterFunc.Interface().(func(chan<- error))
		}
	}
	return t
}

func (t *Telemetry) GetStatus() string {
	j, _ := json.MarshalIndent(t.nuvlaEdgeStatus, "", "    ")
	return string(j)
}

func (t *Telemetry) update() error {
	log.Infof("Updating telemetry data")
	defer common.ExecutionTime(time.Now(), "Updating telemetry data")
	var wg sync.WaitGroup
	errChan := make(chan error)

	// Prepare the wait group
	wg.Add(len(t.updateFuncs))

	// Run all the update functions in parallel
	for _, updateFunc := range t.updateFuncs {
		log.Debugf("Starting update function %s", updateFunc)
		go func(updateFunc UpdaterFunction) {
			defer wg.Done()
			updateFunc(errChan)
		}(updateFunc)
	}

	// Wait for all the update functions to finish
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Process all the possible errors
	for err := range errChan {
		if err != nil {
			log.Warnf("Error updating telemetry data: %s", err)
		}
	}

	return nil
}

func (t *Telemetry) Run() {
	log.Info("Starting telemetry update")
	for {
		startTime := time.Now()
		err := t.update()

		if err != nil {
			log.Errorf("error %s updating telemetry data", err)
		} else {
			log.Infof("Status: \n%s", t.GetStatus())
			err = common.WaitPeriodicAction(startTime, t.updatePeriod, "Telemetry Update")
			if err != nil {
				panic(err)
			}
		}
	}
}

func (t *Telemetry) UpdaterOrchestrator(errChan chan<- error) {
	log.Infof("Updating orchestrator")
	t.nuvlaEdgeStatus.Orchestrator = string(t.coeClient.GetCoeType())

	errChan <- nil

}

func (t *Telemetry) UpdaterNodeId(errChan chan<- error) {
	log.Infof("Updating node id")
	clusterData, err := t.coeClient.GetClusterData()
	if err != nil {
		errChan <- err
		return
	}
	log.Debugf("Node id: %s", clusterData.NodeId)
	t.nuvlaEdgeStatus.NodeId = clusterData.NodeId
}

func (t *Telemetry) UpdaterClusterId(errChan chan<- error) {
	log.Infof("Updating cluster id")
	clusterData, err := t.coeClient.GetClusterData()
	if err != nil {
		errChan <- err
		return
	}
	log.Debugf("Cluster id: %s", clusterData.ClusterId)
	t.nuvlaEdgeStatus.ClusterId = clusterData.ClusterId
}

func (t *Telemetry) UpdaterClusterManagers(errChan chan<- error) {
	log.Infof("Updating cluster managers")
	clusterData, err := t.coeClient.GetClusterData()
	if err != nil {
		errChan <- err
		return
	}
	log.Debugf("Cluster managers: %s", clusterData.ClusterManagers)
	t.nuvlaEdgeStatus.ClusterManagers = clusterData.ClusterManagers
}

func (t *Telemetry) UpdaterClusterNodes(errChan chan<- error) {
	clusterData, err := t.coeClient.GetClusterData()
	if err != nil {
		errChan <- err
		return
	}
	log.Debugf("Cluster nodes: %s", clusterData.ClusterNodes)
	t.nuvlaEdgeStatus.ClusterNodes = clusterData.ClusterNodes
}

func (t *Telemetry) UpdaterClusterNodeLabels(errChan chan<- error) {
	clusterData, err := t.coeClient.GetClusterData()
	if err != nil {
		errChan <- err
		return
	}
	log.Debugf("Cluster node labels: %s", clusterData.ClusterNodeLabels)
	t.nuvlaEdgeStatus.ClusterNodeLabels = clusterData.ClusterNodeLabels
}

func (t *Telemetry) UpdaterClusterNodeRole(errChan chan<- error) {
	clusterData, err := t.coeClient.GetClusterData()
	if err != nil {
		errChan <- err
		return
	}
	log.Debugf("Cluster node role: %s", clusterData.NodeRole)
	t.nuvlaEdgeStatus.ClusterNodeRole = clusterData.NodeRole
}

func (t *Telemetry) UpdaterClusterJoinAddress(errChan chan<- error) {
	clusterData, err := t.coeClient.GetClusterData()
	if err != nil {
		errChan <- err
		return
	}
	log.Debugf("Cluster join address: %s", clusterData.ClusterJoinAddress)
	t.nuvlaEdgeStatus.ClusterJoinAddress = clusterData.ClusterJoinAddress
}

func (t *Telemetry) UpdaterSwarmNodeCertExpiryDate(errChan chan<- error) {
	clusterData, err := t.coeClient.GetClusterData()
	if err != nil {
		errChan <- err
		return
	}
	log.Debugf("Cluster join address: %s", clusterData.SwarmNodeCertExpiryDate)
	t.nuvlaEdgeStatus.SwarmNodeCertExpiryDate = clusterData.SwarmNodeCertExpiryDate
}

func (t *Telemetry) UpdaterArchitecture(errChan chan<- error) {
	t.nuvlaEdgeStatus.Architecture = runtime.GOARCH
	if t.nuvlaEdgeStatus.Architecture == "" {
		errChan <- fmt.Errorf("error retrieving architecture from runtime package")
		return
	}
	log.Debugf("Architecture: %s", t.nuvlaEdgeStatus.Architecture)
}

func (t *Telemetry) UpdaterOperatingSystem(errChan chan<- error) {
	t.nuvlaEdgeStatus.OperatingSystem = runtime.GOOS
	if t.nuvlaEdgeStatus.OperatingSystem == "" {
		errChan <- fmt.Errorf("error retrieving OperatingSystem from runtime package")
		return
	}
	log.Debugf("Operating system: %s", t.nuvlaEdgeStatus.OperatingSystem)
}

func (t *Telemetry) UpdaterIpV4Address(errChan chan<- error) {
	ip, err := t.networkMetricsUpdater.GetIPV4()
	if err != nil {
		errChan <- err
		return
	}
	t.nuvlaEdgeStatus.IpV4Address = ip
	errChan <- nil
}

func (t *Telemetry) UpdaterLastBoot(errChan chan<- error) {
	epochTime, err := host.BootTime()
	if err != nil {
		errChan <- err
		return
	}

	tTime := time.Unix(int64(epochTime), 0)
	t.nuvlaEdgeStatus.LastBoot = tTime.Format(common.DatetimeFormat)
	log.Debugf("Last boot time found %s", t.nuvlaEdgeStatus.LastBoot)
}

func (t *Telemetry) UpdaterHostName(errChan chan<- error) {
	hostname, err := os.Hostname()
	if err != nil {
		errChan <- err
		return
	}
	t.nuvlaEdgeStatus.HostName = hostname

}

func (t *Telemetry) UpdaterDockerServerVersion(errChan chan<- error) {
	dockerVer, err := t.coeClient.GetCoeVersion()
	if err != nil {
		errChan <- err
		return
	}
	t.nuvlaEdgeStatus.DockerServerVersion = dockerVer
}

func (t *Telemetry) UpdaterContainerPlugins(errChan chan<- error) {
	clusterData, err := t.coeClient.GetClusterData()
	if err != nil {
		errChan <- err
		return
	}
	log.Debugf("Container plugins: %s", clusterData.ContainerPlugins)
	t.nuvlaEdgeStatus.ContainerPlugins = clusterData.ContainerPlugins
}

func (t *Telemetry) UpdaterCurrentTime(errChan chan<- error) {
	t.nuvlaEdgeStatus.CurrentTime = time.Now().Format(common.DatetimeFormat)
	errChan <- nil
}

func (t *Telemetry) UpdaterNuvlaEdgeEngineVersion(errChan chan<- error) {
	errChan <- nil
}

func (t *Telemetry) UpdaterHostUserHome(errChan chan<- error) {
	t.nuvlaEdgeStatus.HostUserHome = os.Getenv("HOME")
	errChan <- nil
}

func (t *Telemetry) UpdaterInstallationParameters(errChan chan<- error) {
	errChan <- nil
}

func (t *Telemetry) UpdaterComponents(errChan chan<- error) {
	errChan <- nil
}

func (t *Telemetry) UpdaterStatus(errChan chan<- error) {
	errChan <- nil
}

func (t *Telemetry) UpdaterStatusNotes(errChan chan<- error) {
	errChan <- nil
}

func (t *Telemetry) UpdaterNetwork(errChan chan<- error) {
	networkInfo, err := t.networkMetricsUpdater.GetNetworkInfoMap()
	if err != nil {
		errChan <- err
		return
	}
	log.Debugf("Network info: %s", networkInfo)
	t.nuvlaEdgeStatus.Network = networkInfo
	errChan <- nil
}

func (t *Telemetry) UpdaterResources(errChan chan<- error) {
	metrics, err := t.resourcesMetricsUpdater.GetResourceMetrics()
	if err != nil {
		log.Warnf("Error retrieving resource metrics: %s", err)
		errChan <- err
		return
	}
	log.Infof("Metrics: %s", metrics)
	t.nuvlaEdgeStatus.Resources = metrics
	errChan <- nil
}
