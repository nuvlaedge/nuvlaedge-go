package orchestrator

import (
	"context"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"nuvlaedge-go/testutils"
	"nuvlaedge-go/types"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var service testutils.SwarmServiceMock

func init() {
	ResetMock()
}

var tempDir string

func createMockFile() error {
	d, err := os.MkdirTemp("/tmp/", "swarm")
	tempDir = d
	if err != nil {
		removeMockFile()
		return err
	}

	content := `version: '3.5'

services:
  socat:
    image: sixsq/socat:latest
    command: "tcp-l:1234,fork,reuseaddr tcp:${DESTINATION}"
    ports:
      - "${EXPOSED_PORT}:1234"
    environment:
      - DESTINATION=${DESTINATION}
      - EXPOSED_PORT=${EXPOSED_PORT}
    deploy:
      restart_policy:
        condition: any
        delay: 5s`

	err = os.WriteFile(tempDir+"/docker-compose.yml", []byte(content), 0644)
	if err != nil {
		removeMockFile()
		return err
	}
	return nil
}

func removeMockFile() {
	if tempDir == "" {
		return
	}
	_ = os.RemoveAll(tempDir)
}

func ResetMock() {
	service = testutils.SwarmServiceMock{}
}

func Test_NewSwarmOrchestrator(t *testing.T) {
	defer ResetMock()
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	assert.Nil(t, err, "Error creating docker client")

	so, err := NewSwarmOrchestrator(dockerClient)
	assert.NoErrorf(t, err, "Error creating swarm orchestrator: %v", err)
	assert.NotNil(t, so, "Swarm orchestrator shouldn't be nil")
}

func Test_SwarmOrchestrator_Start(t *testing.T) {
	defer ResetMock()
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	assert.Nil(t, err, "Error creating docker client")

	so, err := NewSwarmOrchestrator(dockerClient)
	assert.NoErrorf(t, err, "Error creating swarm orchestrator: %v", err)
	assert.NotNil(t, so, "Swarm orchestrator shouldn't be nil")
	mockService := &testutils.SwarmServiceMock{}
	so.swarmService = mockService

	if err := createMockFile(); err != nil {
		t.Fatalf("Error creating mock file: %v", err)
	}
	defer removeMockFile()

	sOpts := &types.StartOpts{
		ProjectName: "test",
		CFiles:      []string{filepath.Join(tempDir, "docker-compose.yml")},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = so.Start(ctx, sOpts)
	assert.NoErrorf(t, err, "No error expected: %v", err)
}
