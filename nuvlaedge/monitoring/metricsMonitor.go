package monitoring

import (
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/v3/host"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/nuvlaedge/common"
	"nuvlaedge-go/nuvlaedge/common/resources"
	"nuvlaedge-go/nuvlaedge/orchestrator"
	"nuvlaedge-go/nuvlaedge/version"
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

type MetricsMonitor struct {
	// Base monitoring data
	nuvlaEdgeStatus *resources.NuvlaEdgeStatus
	updateMutex     *sync.Mutex

	// Helper map to gather and run all the update functions
	updateFuncs map[string]UpdaterFunction

	// Metric reading updaters
	coeClient               orchestrator.Coe // Interfaces need no pointer
	resourcesMetricsUpdater *ResourceMetricsUpdater
	networkMetricsUpdater   *NetworkMetricsUpdater

	// Report and exit channels
	reportChan chan resources.NuvlaEdgeStatus
	exitChan   chan bool

	// MetricsMonitor time in seconds to refresh the NuvlaEdgeStatus resource attribute
	refreshRate int
}

func NewMetricsMonitor(
	coeClient orchestrator.Coe,
	refreshRate int) *MetricsMonitor {

	networkUpdater := NewNetworkMetricsUpdater()
	containerUpdater := NewContainerStats(&coeClient, refreshRate)

	t := &MetricsMonitor{
		nuvlaEdgeStatus:         &resources.NuvlaEdgeStatus{},
		updateFuncs:             make(map[string]UpdaterFunction),
		coeClient:               coeClient,
		resourcesMetricsUpdater: NewResourceMetricsUpdater(networkUpdater, containerUpdater),
		networkMetricsUpdater:   networkUpdater,
		refreshRate:             refreshRate,
		updateMutex:             &sync.Mutex{},
	}
	val := reflect.ValueOf(*t.nuvlaEdgeStatus)
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

func (t *MetricsMonitor) GetStatus() string {
	j, _ := json.MarshalIndent(t.nuvlaEdgeStatus, "", "    ")
	return string(j)
}

func (t *MetricsMonitor) GetNewFullStatus(status *resources.NuvlaEdgeStatus) error {
	// For debugging purposes, lets time the execution of this function
	defer common.ExecutionTime(time.Now(), "GetNewStatus")
	t.updateMutex.Lock()
	defer t.updateMutex.Unlock()

	b, err := json.Marshal(t.nuvlaEdgeStatus)
	if err != nil {
		log.Warnf("Error marshalling nuvlaEdgeStatus: %s", err)
		log.Errorf("Cannot update status monitoring")
		return err
	}

	err = json.Unmarshal(b, status)
	if err != nil {
		log.Warnf("Error unmarshalling nuvlaEdgeStatus: %s", err)
		log.Errorf("Cannot update status monitoring")
		return err
	}
	return nil
}

func (t *MetricsMonitor) update() error {
	log.Debug("Updating monitoring data")
	defer common.ExecutionTime(time.Now(), "Updating monitoring data")
	var wg sync.WaitGroup
	errChan := make(chan error)

	// Prepare the wait group
	wg.Add(len(t.updateFuncs))

	// Run all the update functions concurrently
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
			log.Warnf("Error updating monitoring data: %s", err)
		}
	}

	return nil
}

func (t *MetricsMonitor) Run() {
	log.Info("Starting monitoring update")
	for {
		startTime := time.Now()
		t.updateMutex.Lock()
		t.nuvlaEdgeStatus = &resources.NuvlaEdgeStatus{}
		// Try updating all the monitoring data
		if err := t.update(); err != nil {
			log.Errorf("error %s updating monitoring data", err)
		}

		// Release the mutex for the nuvlaEdgeStatus
		t.updateMutex.Unlock()
		if err := common.WaitPeriodicAction(startTime, t.refreshRate, "MetricsMonitor Update"); err != nil {
			log.Errorf("Error waiting for periodic action: %s", err)
		}
	}
}

func (t *MetricsMonitor) SetRefreshRate(newRate int) {
	if newRate < 10 {
		log.Infof("Cannot set %vs as refresh rate. Minimum rate is 10s", newRate)
		return
	}
	t.refreshRate = newRate
}

func (t *MetricsMonitor) UpdaterOrchestrator(errChan chan<- error) {
	log.Debug("Updating orchestrator")
	t.nuvlaEdgeStatus.Orchestrator = string(t.coeClient.GetCoeType())

	errChan <- nil

}

func (t *MetricsMonitor) UpdaterNodeId(errChan chan<- error) {
	log.Debugln("Updating node id")
	clusterData, err := t.coeClient.GetClusterData()
	if err != nil {
		errChan <- err
		return
	}
	log.Debugf("Node id: %s", clusterData.NodeId)
	t.nuvlaEdgeStatus.NodeId = clusterData.NodeId
}

func (t *MetricsMonitor) UpdaterClusterId(errChan chan<- error) {
	log.Debugln("Updating cluster id")
	clusterData, err := t.coeClient.GetClusterData()
	if err != nil {
		errChan <- err
		return
	}
	log.Debugf("Cluster id: %s", clusterData.ClusterId)
	t.nuvlaEdgeStatus.ClusterId = clusterData.ClusterId
}

func (t *MetricsMonitor) UpdaterClusterManagers(errChan chan<- error) {
	log.Debugln("Updating cluster managers")
	clusterData, err := t.coeClient.GetClusterData()
	if err != nil {
		errChan <- err
		return
	}
	log.Debugf("Cluster managers: %s", clusterData.ClusterManagers)
	t.nuvlaEdgeStatus.ClusterManagers = clusterData.ClusterManagers
}

func (t *MetricsMonitor) UpdaterClusterNodes(errChan chan<- error) {
	clusterData, err := t.coeClient.GetClusterData()
	if err != nil {
		errChan <- err
		return
	}
	log.Debugf("Cluster nodes: %s", clusterData.ClusterNodes)
	t.nuvlaEdgeStatus.ClusterNodes = clusterData.ClusterNodes
}

func (t *MetricsMonitor) UpdaterClusterNodeLabels(errChan chan<- error) {
	clusterData, err := t.coeClient.GetClusterData()
	if err != nil {
		errChan <- err
		return
	}
	log.Debugf("Cluster node labels: %s", clusterData.ClusterNodeLabels)
	t.nuvlaEdgeStatus.ClusterNodeLabels = clusterData.ClusterNodeLabels
}

func (t *MetricsMonitor) UpdaterClusterNodeRole(errChan chan<- error) {
	clusterData, err := t.coeClient.GetClusterData()
	if err != nil {
		errChan <- err
		return
	}
	log.Debugf("Cluster node role: %s", clusterData.NodeRole)
	t.nuvlaEdgeStatus.ClusterNodeRole = clusterData.NodeRole
}

func (t *MetricsMonitor) UpdaterClusterJoinAddress(errChan chan<- error) {
	clusterData, err := t.coeClient.GetClusterData()
	if err != nil {
		errChan <- err
		return
	}
	log.Debugf("Cluster join address: %s", clusterData.ClusterJoinAddress)
	t.nuvlaEdgeStatus.ClusterJoinAddress = clusterData.ClusterJoinAddress
}

func (t *MetricsMonitor) UpdaterSwarmNodeCertExpiryDate(errChan chan<- error) {
	clusterData, err := t.coeClient.GetClusterData()
	if err != nil {
		errChan <- err
		return
	}
	log.Debugf("Cluster join address: %s", clusterData.SwarmNodeCertExpiryDate)
	t.nuvlaEdgeStatus.SwarmNodeCertExpiryDate = clusterData.SwarmNodeCertExpiryDate
}

func (t *MetricsMonitor) UpdaterArchitecture(errChan chan<- error) {
	t.nuvlaEdgeStatus.Architecture = runtime.GOARCH
	if t.nuvlaEdgeStatus.Architecture == "" {
		errChan <- fmt.Errorf("error retrieving architecture from runtime package")
		return
	}
	log.Debugf("Architecture: %s", t.nuvlaEdgeStatus.Architecture)
}

func (t *MetricsMonitor) UpdaterOperatingSystem(errChan chan<- error) {
	t.nuvlaEdgeStatus.OperatingSystem = runtime.GOOS
	if t.nuvlaEdgeStatus.OperatingSystem == "" {
		errChan <- fmt.Errorf("error retrieving OperatingSystem from runtime package")
		return
	}
	log.Debugf("Operating system: %s", t.nuvlaEdgeStatus.OperatingSystem)
}

func (t *MetricsMonitor) UpdaterIpV4Address(errChan chan<- error) {
	ip, err := t.networkMetricsUpdater.GetIPV4()
	if err != nil {
		errChan <- err
		return
	}
	t.nuvlaEdgeStatus.IpV4Address = ip
	errChan <- nil
}

func (t *MetricsMonitor) UpdaterLastBoot(errChan chan<- error) {
	epochTime, err := host.BootTime()
	if err != nil {
		errChan <- err
		return
	}

	tTime := time.Unix(int64(epochTime), 0)
	t.nuvlaEdgeStatus.LastBoot = tTime.Format(common.DatetimeFormat)
	log.Debugf("Last boot time found %s", t.nuvlaEdgeStatus.LastBoot)
}

func (t *MetricsMonitor) UpdaterHostName(errChan chan<- error) {
	hostname, err := os.Hostname()
	if err != nil {
		errChan <- err
		return
	}
	t.nuvlaEdgeStatus.HostName = hostname

}

func (t *MetricsMonitor) UpdaterDockerServerVersion(errChan chan<- error) {
	dockerVer, err := t.coeClient.GetCoeVersion()
	if err != nil {
		errChan <- err
		return
	}
	t.nuvlaEdgeStatus.DockerServerVersion = dockerVer
}

func (t *MetricsMonitor) UpdaterContainerPlugins(errChan chan<- error) {
	clusterData, err := t.coeClient.GetClusterData()
	if err != nil {
		errChan <- err
		return
	}
	log.Debugf("Container plugins: %s", clusterData.ContainerPlugins)
	t.nuvlaEdgeStatus.ContainerPlugins = clusterData.ContainerPlugins
}

func (t *MetricsMonitor) UpdaterCurrentTime(errChan chan<- error) {
	t.nuvlaEdgeStatus.CurrentTime = time.Now().Format(common.DatetimeFormat)
	errChan <- nil
}

func (t *MetricsMonitor) UpdaterNuvlaEdgeEngineVersion(errChan chan<- error) {
	log.Debugf("NuvlaEdgeEngineVersion: %s-go", version.GetVersion())
	t.nuvlaEdgeStatus.NuvlaEdgeEngineVersion = fmt.Sprintf("%s-go", version.GetVersion())
	errChan <- nil
}

func (t *MetricsMonitor) UpdaterHostUserHome(errChan chan<- error) {
	t.nuvlaEdgeStatus.HostUserHome = os.Getenv("HOME")
	errChan <- nil
}

func (t *MetricsMonitor) UpdaterInstallationParameters(errChan chan<- error) {
	log.Info("Updating installation parameters")
	if t.nuvlaEdgeStatus.InstallationParameters == nil {
		t.nuvlaEdgeStatus.InstallationParameters = &resources.InstallationParameters{}
	}
	err := t.coeClient.GetInstallationParameters(t.nuvlaEdgeStatus.InstallationParameters)
	if err != nil {
		errChan <- err
		return
	}
	errChan <- nil
}

func (t *MetricsMonitor) UpdaterComponents(errChan chan<- error) {
	errChan <- nil
}

func (t *MetricsMonitor) UpdaterStatus(errChan chan<- error) {
	t.nuvlaEdgeStatus.Status = "OPERATIONAL"
	errChan <- nil
}

func (t *MetricsMonitor) UpdaterStatusNotes(errChan chan<- error) {
	errChan <- nil
}

func (t *MetricsMonitor) UpdaterNetwork(errChan chan<- error) {
	networkInfo, err := t.networkMetricsUpdater.GetNetworkInfoMap()
	if err != nil {
		errChan <- err
		return
	}
	log.Debugf("Network info: %s", networkInfo)
	t.nuvlaEdgeStatus.Network = networkInfo
	errChan <- nil
}

func (t *MetricsMonitor) UpdaterResources(errChan chan<- error) {
	metrics, err := t.resourcesMetricsUpdater.GetResourceMetrics()
	if err != nil {
		log.Warnf("Error retrieving resource metrics: %s", err)
		errChan <- err
		return
	}
	pMetrics, _ := json.MarshalIndent(metrics, "", "    ")
	log.Debugf("Last metrics update: %s", string(pMetrics))
	t.nuvlaEdgeStatus.Resources = metrics
	errChan <- nil
}
