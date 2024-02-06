package coe

import (
	"fmt"
	log "github.com/sirupsen/logrus"
)

type CoeType string

const (
	DockerType     CoeType = "docker"
	KubernetesType CoeType = "kubernetes"
)

type Coe interface {
	// Basic orchestration engine management
	GetCoeType() CoeType
	GetCoeVersion() (string, error)
	String() string

	/****************************************
	Container management
	*****************************************/
	RunContainer(image string, configuration map[string]string) (string, error)
	StopContainer(containerId string, force bool) (bool, error)
	RemoveContainer(containerId string, containerName string) (bool, error)
	//getContainerLogs(containerId string) []string
	//collectContainerMetrics() (map[string]any, error)
	//runCommandInContainer(
	//	imageName string,
	//	containerName string,
	//	command string,
	//	args string,
	//	network string,
	//	remove bool)

	// Orchestration Management
	//getNodeLabels() []string
	//readSystemIssues(nodeInfo string) ([]string, error)
	GetClusterData() (*ClusterData, error)
	GetOrchestratorCredentials() (map[string]string, error)
	//getClusterJoinAddress(nodeId string) string
	//isNodeActive(nodeId string) bool
	//getContainerPlugins() []string

	// NuvlaEdge related Functions
	//getInstallationParameters() (map[string]any, error)
	//isVPNClientRunning() bool
	//installSSHKey(sshPubKey string, hostName string) bool
	//isNuvlaJobRunning(jobId string, jobExecutionId string) (bool, error)
	//defineNuvlaInfraService(
	//	endpoint string,
	//	clientCA string,
	//	clientCert string,
	//	clientKey string,
	//) map[string]any
	//getNuvlaEdgeComponents() []string
	//getCoeVersion() string
	TelemetryStart() (bool, error)
	TelemetryStatus() (int, error)
	TelemetryStop() (bool, error)
}

func NewCoe(coeType CoeType) (Coe, error) {
	log.Infof("Creating new %s COE", coeType)
	switch coeType {
	case DockerType:
		return newDockerCoe(), nil
	}
	return nil, nil
}

func newDockerCoe() *DockerCoe {
	coe := DockerCoe{}
	return &coe
}

func newKubernetesCoe() {

}

// DetectCoe Checks the system configuration and returns the COE type if any of the supported COE clients
// is detected, or an error otherwise
func DetectCoe() (string, error) {

	return "", fmt.Errorf("could not find any COE. NuvlaEdge cannot run without it. Installing COE not " +
		"supported ATM")
}
