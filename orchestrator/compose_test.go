package orchestrator

import (
	"context"
	"errors"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"nuvlaedge-go/testutils"
	"nuvlaedge-go/types"
	"testing"
)

func Test_NewComposeOrchestrator(t *testing.T) {
	// Test code here
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	assert.Nil(t, err, "Error creating docker client")

	co, err := NewComposeOrchestrator(dockerClient)
	assert.Nil(t, err, "Error creating compose orchestrator")
	assert.NotNil(t, co, "Compose orchestrator shouldn't be nil")
	assert.NotNil(t, co.service, "Compose service shouldn't be nil")
	assert.NotNil(t, co.dCli, "Docker CLI shouldn't be nil")
}

func Test_Compose_Start(t *testing.T) {
	// Test code here
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	assert.Nil(t, err, "Error creating docker client")

	co, err := NewComposeOrchestrator(dockerClient)
	mockService := testutils.ComposeClientMock{}
	co.service = &mockService

	assert.Nil(t, err, "Error creating compose orchestrator")
	ctx := context.Background()
	startOpts := &types.StartOpts{}
	err = co.Start(ctx, startOpts)
	assert.Nil(t, err, "Error should not be nil")

	mockService.PullErr = errors.New("pull error")
	err = co.Start(ctx, startOpts)
	assert.NotNil(t, err, "Error should not be nil")
	assert.Contains(t, err.Error(), "pull error", "Error should contain pull error")

	mockService.PullErr = nil
	mockService.UpErr = errors.New("up error")
	err = co.Start(ctx, startOpts)
	assert.NotNil(t, err, "Error should not be nil")
	assert.Contains(t, err.Error(), "up error", "Error should contain up error")
}

func Test_Compose_Stop(t *testing.T) {
	// Test code here
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	assert.Nil(t, err, "Error creating docker client")

	co, err := NewComposeOrchestrator(dockerClient)
	mockService := testutils.ComposeClientMock{}
	co.service = &mockService

	assert.Nil(t, err, "Error creating compose orchestrator")
	ctx := context.Background()
	stopOpts := &types.StopOpts{}
	err = co.Stop(ctx, stopOpts)
	assert.Nil(t, err, "Error should not be nil")

	mockService.DownErr = errors.New("down error")
	err = co.Stop(ctx, stopOpts)
	assert.NotNil(t, err, "Error should not be nil")
	assert.Contains(t, err.Error(), "down error", "Error should contain down error")
}

func Test_Compose_List(t *testing.T) {
	// Test code here
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	assert.Nil(t, err, "Error creating docker client")

	co, err := NewComposeOrchestrator(dockerClient)
	mockService := testutils.ComposeClientMock{}
	co.service = &mockService

	assert.Nil(t, err, "Error creating compose orchestrator")
	ctx := context.Background()
	_, err = co.List(ctx)
	assert.Nil(t, err, "Error should not be nil")

	mockService.ListErr = errors.New("list error")
	_, err = co.List(ctx)
	assert.NotNil(t, err, "Error should not be nil")
	assert.Contains(t, err.Error(), "list error", "Error should contain list error")
}

func Test_Compose_GetProjectStatus(t *testing.T) {
	// Test code here
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	assert.Nil(t, err, "Error creating docker client")

	co, err := NewComposeOrchestrator(dockerClient)
	mockService := testutils.ComposeClientMock{}
	co.service = &mockService

	assert.Nil(t, err, "Error creating compose orchestrator")
	ctx := context.Background()
	_, err = co.GetProjectStatus(ctx, "project")
	assert.Nil(t, err, "Error should not be nil")

	mockService.PsErr = errors.New("ps error")
	_, err = co.GetProjectStatus(ctx, "project")
	assert.NotNil(t, err, "Error should not be nil")
	assert.Contains(t, err.Error(), "ps error", "Error should contain ps error")
}
