package executors

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
)

const (
	// DefaultTemporaryDirectory is used by compose files
	DefaultTemporaryDirectory = "/tmp"
)

func ExecuteCommand(command string, args ...string) (string, error) {
	log.Infof("Executing command: %s with arguments %v", command, args)

	cmd := exec.Command(command, args...)
	output, err := cmd.Output()

	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			log.Errorf("Program %s exited with error code: %d", command, exitError.ExitCode())
			return string(exitError.Stderr), err
		}
		log.Errorf("Error executing command: %s", err)
		return string(output), err
	}

	return string(output), nil

}

// IsSuperUser Asserts whether the program is running as root or with superuser privileges
func IsSuperUser() bool {
	if output, err := ExecuteCommand("id", "-u"); err != nil {
		return false
	} else {
		return output == "0"
	}
}

type RunningMode string

const (
	HostMode       RunningMode = "host"
	DockerMode     RunningMode = "docker"
	KubernetesMode RunningMode = "kubernetes"
)

// WhereAmI Asserts whether the program is running in a Docker container, Kubernetes pod, or on the host
func WhereAmI() RunningMode {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		// The /.dockerenv file exists, so we're probably running inside a Docker container
		return DockerMode
	}

	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" && os.Getenv("KUBERNETES_SERVICE_PORT") != "" {
		// The KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT environment variables are set,
		// so we're probably running inside a Kubernetes pod
		return KubernetesMode
	}

	// If neither of the above checks were true, assume we're running on the host
	return HostMode
}

// IsRunningOnHost Asserts whether the program is running on the host
func IsRunningOnHost() bool {
	return WhereAmI() == HostMode
}

// IsRunningInDocker Asserts whether the program is running in a Docker container
func IsRunningInDocker() bool {
	return WhereAmI() == DockerMode
}

// IsRunningInKubernetes Asserts whether the program is running in a Kubernetes pod
func IsRunningInKubernetes() bool {
	return WhereAmI() == KubernetesMode
}

func GetProjectNameFromDeploymentId(deploymentId string) string {
	return strings.Replace(deploymentId, "/", "-", -1)
}
