package common

import "os"

type RunningMode int

const (
	Host       RunningMode = 0
	Docker     RunningMode = 1
	Kubernetes RunningMode = 2
)

func WhereAmI(dFile ...string) RunningMode {
	f := "/.dockerenv"
	if len(dFile) >= 1 {
		f = dFile[0]
	}
	if _, err := os.Stat(f); err == nil {
		return Docker
	}

	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" && os.Getenv("KUBERNETES_SERVICE_PORT") != "" {
		// The KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT environment variables are set,
		// so we're probably running inside a Kubernetes pod
		return Kubernetes
	}

	// If neither of the above checks were true, assume we're running on the host
	return Host
}

// IsRunningOnHost Asserts whether the program is running on the host
func IsRunningOnHost() bool {
	return WhereAmI() == Host
}

// IsRunningInDocker Asserts whether the program is running in a Docker container
func IsRunningInDocker(file ...string) bool {
	return WhereAmI(file...) == Docker
}

// IsRunningInKubernetes Asserts whether the program is running in a Kubernetes pod
func IsRunningInKubernetes() bool {
	return WhereAmI() == Kubernetes
}
