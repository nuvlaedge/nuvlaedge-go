package common

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var mockDockerFile = "/tmp/.dockerenv"

// Mock environment setup for Docker
func mockDockerEnv() func() {
	os.Setenv("KUBERNETES_SERVICE_HOST", "")
	os.Setenv("KUBERNETES_SERVICE_PORT", "")
	_, err := os.Create(mockDockerFile)
	if err != nil {
		panic(err)
	}
	return func() {
		os.Remove(mockDockerFile)
	}
}

// Mock environment setup for Kubernetes
func mockKubernetesEnv() func() {
	os.Setenv("KUBERNETES_SERVICE_HOST", "127.0.0.1")
	os.Setenv("KUBERNETES_SERVICE_PORT", "8080")
	return func() {
		os.Unsetenv("KUBERNETES_SERVICE_HOST")
		os.Unsetenv("KUBERNETES_SERVICE_PORT")
	}
}

// RunningOnHostWhenNoIndicatorsPresent asserts that the default running mode is Host
func TestRunningOnHostWhenNoIndicatorsPresent(t *testing.T) {
	assert.Equal(t, Host, WhereAmI(mockDockerFile))
}

// RunningInDockerWhenDockerEnvFileExists asserts that the running mode is Docker when /.dockerenv exists
func TestRunningInDockerWhenDockerEnvFileExists(t *testing.T) {
	cleanup := mockDockerEnv()
	defer cleanup()

	assert.Equal(t, Docker, WhereAmI(mockDockerFile))
}

// RunningInKubernetesWhenEnvVarsSet asserts that the running mode is Kubernetes when KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT are set
func TestRunningInKubernetesWhenEnvVarsSet(t *testing.T) {
	cleanup := mockKubernetesEnv()
	defer cleanup()

	assert.Equal(t, Kubernetes, WhereAmI(mockDockerFile))
}

func TestIsRunningInDocker(t *testing.T) {
	cleanup := mockDockerEnv()
	defer cleanup()

	assert.True(t, IsRunningInDocker(mockDockerFile))
}

func TestIsRunningInKubernetes(t *testing.T) {
	cleanup := mockKubernetesEnv()
	defer cleanup()

	assert.True(t, IsRunningInKubernetes())
}

func TestIsRunningOnHost(t *testing.T) {
	assert.True(t, IsRunningOnHost())
}
