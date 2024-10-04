package monitor

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/volume"
	"github.com/stretchr/testify/assert"
	"nuvlaedge-go/testutils"
	"slices"
	"testing"
	"time"
)

func Test_DockerMonitor_GetGatherers_WithOutSwarm(t *testing.T) {
	cli := &testutils.TestDockerMetricsClient{InspectErr: assert.AnError}
	dm := &DockerMonitor{
		client: cli,
	}

	gatherers := []string{
		"images",
		"containers",
		"volumes",
		"networks",
	}
	g := dm.getGatherers()
	assert.Len(t, g, 4, "Expected 4 gatherers without swarm")
	for _, gatherer := range g {
		assert.False(t, gatherer.needSwarm, "Expected gatherer to not need swarm")
		assert.Contains(t, gatherers, gatherer.resourceName, "Expected gatherer to be in list")
	}

	swarmGatherers := slices.Concat(gatherers, []string{"services", "tasks", "configs", "secrets"})
	cli.InspectErr = nil

	g = dm.getGatherers()
	assert.Len(t, g, 8, "Expected 8 gatherers with swarm")
	for _, gatherer := range g {
		assert.Contains(t, swarmGatherers, gatherer.resourceName, "Expected gatherer to be in list")
	}
}

func Test_sortImages_SortedByCreatedAndID(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		ImageListReturn: []image.Summary{
			{ID: "3", Created: 3},
			{ID: "1", Created: 1},
			{ID: "2", Created: 2},
			{ID: "4", Created: 3},
		},
	}

	expected := []extendedImage{
		{Summary: image.Summary{ID: "1", Created: 1}},
		{Summary: image.Summary{ID: "2", Created: 2}},
		{Summary: image.Summary{ID: "3", Created: 3}},
		{Summary: image.Summary{ID: "4", Created: 3}},
	}

	result, err := sortImages(ctx, dCli)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_sortImages_ErrorCLi(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		ImageListErr: assert.AnError,
	}

	result, err := sortImages(ctx, dCli)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, assert.AnError)
}

func Test_Containers_SortedByCreatedAndID(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		ContainerListReturn: []types.Container{
			{ID: "3", Created: 3},
			{ID: "1", Created: 1},
			{ID: "2", Created: 2},
			{ID: "4", Created: 3},
		},
	}

	expected := []extendedContainer{
		{Container: types.Container{ID: "1", Created: 1}},
		{Container: types.Container{ID: "2", Created: 2}},
		{Container: types.Container{ID: "3", Created: 3}},
		{Container: types.Container{ID: "4", Created: 3}},
	}

	result, err := sortContainers(ctx, dCli)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_Containers_EmptySlice(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		ContainerListReturn: []types.Container{},
	}

	expected := make([]extendedContainer, 0)

	result, err := sortContainers(ctx, dCli)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_Containers_ErrorCLi(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		ContainerListErr: assert.AnError,
	}

	result, err := sortContainers(ctx, dCli)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, assert.AnError)
}

func Test_sortVolumes_SortedByCreatedAndName(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		VolumeListReturn: volume.ListResponse{
			Volumes: []*volume.Volume{
				{Name: "vol3", CreatedAt: "2023-01-03T00:00:00Z"},
				{Name: "vol1", CreatedAt: "2023-01-01T00:00:00Z"},
				{Name: "vol2", CreatedAt: "2023-01-02T00:00:00Z"},
				{Name: "vol4", CreatedAt: "2023-01-03T00:00:00Z"},
			},
		},
	}

	expected := []*volume.Volume{
		{Name: "vol1", CreatedAt: "2023-01-01T00:00:00Z"},
		{Name: "vol2", CreatedAt: "2023-01-02T00:00:00Z"},
		{Name: "vol3", CreatedAt: "2023-01-03T00:00:00Z"},
		{Name: "vol4", CreatedAt: "2023-01-03T00:00:00Z"},
	}

	result, err := sortVolumes(ctx, dCli)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_sortVolumes_EmptySlice(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		VolumeListReturn: volume.ListResponse{
			Volumes: []*volume.Volume{},
		},
	}

	expected := make([]*volume.Volume, 0)

	result, err := sortVolumes(ctx, dCli)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_sortVolumes_SingleElement(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		VolumeListReturn: volume.ListResponse{
			Volumes: []*volume.Volume{
				{Name: "vol1", CreatedAt: "2023-01-01T00:00:00Z"},
			},
		},
	}

	expected := []*volume.Volume{
		{Name: "vol1", CreatedAt: "2023-01-01T00:00:00Z"},
	}

	result, err := sortVolumes(ctx, dCli)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_sortVolumes_SameCreatedDifferentName(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		VolumeListReturn: volume.ListResponse{
			Volumes: []*volume.Volume{
				{Name: "vol2", CreatedAt: "2023-01-01T00:00:00Z"},
				{Name: "vol1", CreatedAt: "2023-01-01T00:00:00Z"},
			},
		},
	}

	expected := []*volume.Volume{
		{Name: "vol1", CreatedAt: "2023-01-01T00:00:00Z"},
		{Name: "vol2", CreatedAt: "2023-01-01T00:00:00Z"},
	}

	result, err := sortVolumes(ctx, dCli)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_sortVolumes_ErrorCLi(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		VolumeListErr: assert.AnError,
	}

	result, err := sortVolumes(ctx, dCli)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, assert.AnError)
}

func Test_sortNetworks_SortedByCreatedAndID(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		NetworkListReturn: []network.Inspect{
			{ID: "3", Created: time.Unix(3, 0)},
			{ID: "1", Created: time.Unix(1, 0)},
			{ID: "2", Created: time.Unix(2, 0)},
			{ID: "4", Created: time.Unix(3, 0)},
		},
	}

	expected := []network.Inspect{
		{ID: "3", Created: time.Unix(3, 0)},
		{ID: "4", Created: time.Unix(3, 0)},
		{ID: "2", Created: time.Unix(2, 0)},
		{ID: "1", Created: time.Unix(1, 0)},
	}

	result, err := sortNetworks(ctx, dCli)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_sortNetworks_EmptySlice(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		NetworkListReturn: []network.Inspect{},
	}

	expected := make([]network.Inspect, 0)

	result, err := sortNetworks(ctx, dCli)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_sortNetworks_SingleElement(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		NetworkListReturn: []network.Inspect{
			{ID: "1", Created: time.Unix(1, 0)},
		},
	}

	expected := []network.Inspect{
		{ID: "1", Created: time.Unix(1, 0)},
	}

	result, err := sortNetworks(ctx, dCli)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_sortNetworks_SameCreatedDifferentID(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		NetworkListReturn: []network.Inspect{
			{ID: "2", Created: time.Unix(1, 0)},
			{ID: "1", Created: time.Unix(1, 0)},
		},
	}

	expected := []network.Inspect{
		{ID: "1", Created: time.Unix(1, 0)},
		{ID: "2", Created: time.Unix(1, 0)},
	}

	result, err := sortNetworks(ctx, dCli)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_sortNetworks_ErrorCLi(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		NetworkListErr: assert.AnError,
	}

	result, err := sortNetworks(ctx, dCli)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, assert.AnError)
}

func Test_sortServices_SortedByCreatedAndID(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		ServiceListReturn: []swarm.Service{
			{ID: "3", Meta: swarm.Meta{CreatedAt: time.Unix(3, 0)}},
			{ID: "1", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
			{ID: "2", Meta: swarm.Meta{CreatedAt: time.Unix(2, 0)}},
			{ID: "4", Meta: swarm.Meta{CreatedAt: time.Unix(3, 0)}},
		},
	}

	expected := []swarm.Service{
		{ID: "3", Meta: swarm.Meta{CreatedAt: time.Unix(3, 0)}},
		{ID: "4", Meta: swarm.Meta{CreatedAt: time.Unix(3, 0)}},
		{ID: "2", Meta: swarm.Meta{CreatedAt: time.Unix(2, 0)}},
		{ID: "1", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
	}

	result, err := sortServices(ctx, dCli)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_sortServices_EmptySlice(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		ServiceListReturn: []swarm.Service{},
	}

	expected := make([]swarm.Service, 0)

	result, err := sortServices(ctx, dCli)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_sortServices_SingleElement(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		ServiceListReturn: []swarm.Service{
			{ID: "1", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
		},
	}

	expected := []swarm.Service{
		{ID: "1", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
	}

	result, err := sortServices(ctx, dCli)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_sortServices_SameCreatedDifferentID(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		ServiceListReturn: []swarm.Service{
			{ID: "2", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
			{ID: "1", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
		},
	}

	expected := []swarm.Service{
		{ID: "1", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
		{ID: "2", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
	}

	result, err := sortServices(ctx, dCli)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_sortServices_ErrorCLi(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		ServiceListErr: assert.AnError,
	}

	result, err := sortServices(ctx, dCli)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, assert.AnError)
}

func Test_sortTasks_SortedByCreatedAndID(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		TaskListReturn: []swarm.Task{
			{ID: "3", Meta: swarm.Meta{CreatedAt: time.Unix(3, 0)}},
			{ID: "1", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
			{ID: "2", Meta: swarm.Meta{CreatedAt: time.Unix(2, 0)}},
			{ID: "4", Meta: swarm.Meta{CreatedAt: time.Unix(3, 0)}},
		},
	}

	expected := []swarm.Task{
		{ID: "3", Meta: swarm.Meta{CreatedAt: time.Unix(3, 0)}},
		{ID: "4", Meta: swarm.Meta{CreatedAt: time.Unix(3, 0)}},
		{ID: "2", Meta: swarm.Meta{CreatedAt: time.Unix(2, 0)}},
		{ID: "1", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
	}

	result, err := sortTasks(ctx, dCli)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_sortTasks_EmptySlice(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		TaskListReturn: []swarm.Task{},
	}

	expected := make([]swarm.Task, 0)

	result, err := sortTasks(ctx, dCli)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_sortTasks_SingleElement(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		TaskListReturn: []swarm.Task{
			{ID: "1", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
		},
	}

	expected := []swarm.Task{
		{ID: "1", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
	}

	result, err := sortTasks(ctx, dCli)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_sortTasks_SameCreatedDifferentID(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		TaskListReturn: []swarm.Task{
			{ID: "2", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
			{ID: "1", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
		},
	}

	expected := []swarm.Task{
		{ID: "1", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
		{ID: "2", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
	}

	result, err := sortTasks(ctx, dCli)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_sortTasks_ErrorCLi(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		TaskListErr: assert.AnError,
	}

	result, err := sortTasks(ctx, dCli)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, assert.AnError)
}

func Test_sortConfigs_SortedByCreatedAndID(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		ConfigListReturn: []swarm.Config{
			{ID: "3", Meta: swarm.Meta{CreatedAt: time.Unix(3, 0)}},
			{ID: "1", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
			{ID: "2", Meta: swarm.Meta{CreatedAt: time.Unix(2, 0)}},
			{ID: "4", Meta: swarm.Meta{CreatedAt: time.Unix(3, 0)}},
		},
	}

	expected := []swarm.Config{
		{ID: "3", Meta: swarm.Meta{CreatedAt: time.Unix(3, 0)}},
		{ID: "4", Meta: swarm.Meta{CreatedAt: time.Unix(3, 0)}},
		{ID: "2", Meta: swarm.Meta{CreatedAt: time.Unix(2, 0)}},
		{ID: "1", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
	}

	result, err := sortConfigs(ctx, dCli)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_sortConfigs_EmptySlice(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		ConfigListReturn: []swarm.Config{},
	}

	expected := make([]swarm.Config, 0)

	result, err := sortConfigs(ctx, dCli)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_sortConfigs_SingleElement(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		ConfigListReturn: []swarm.Config{
			{ID: "1", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
		},
	}

	expected := []swarm.Config{
		{ID: "1", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
	}

	result, err := sortConfigs(ctx, dCli)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_sortConfigs_SameCreatedDifferentID(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		ConfigListReturn: []swarm.Config{
			{ID: "2", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
			{ID: "1", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
		},
	}

	expected := []swarm.Config{
		{ID: "1", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
		{ID: "2", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
	}

	result, err := sortConfigs(ctx, dCli)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_sortConfigs_ErrorCLi(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		ConfigListErr: assert.AnError,
	}

	result, err := sortConfigs(ctx, dCli)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, assert.AnError)
}

func Test_sortSecretsSortedByCreatedAndID(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		SecretListReturn: []swarm.Secret{
			{ID: "3", Meta: swarm.Meta{CreatedAt: time.Unix(3, 0)}},
			{ID: "1", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
			{ID: "2", Meta: swarm.Meta{CreatedAt: time.Unix(2, 0)}},
			{ID: "4", Meta: swarm.Meta{CreatedAt: time.Unix(3, 0)}},
		},
	}

	expected := []swarm.Secret{
		{ID: "3", Meta: swarm.Meta{CreatedAt: time.Unix(3, 0)}},
		{ID: "4", Meta: swarm.Meta{CreatedAt: time.Unix(3, 0)}},
		{ID: "2", Meta: swarm.Meta{CreatedAt: time.Unix(2, 0)}},
		{ID: "1", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
	}

	result, err := sortSecrets(ctx, dCli)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_sortSecretsEmptySlice(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		SecretListReturn: []swarm.Secret{},
	}

	expected := make([]swarm.Secret, 0)

	result, err := sortSecrets(ctx, dCli)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_sortSecretsSingleElement(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		SecretListReturn: []swarm.Secret{
			{ID: "1", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
		},
	}

	expected := []swarm.Secret{
		{ID: "1", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
	}

	result, err := sortSecrets(ctx, dCli)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_sortSecretsSameCreatedDifferentID(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		SecretListReturn: []swarm.Secret{
			{ID: "2", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
			{ID: "1", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
		},
	}

	expected := []swarm.Secret{
		{ID: "1", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
		{ID: "2", Meta: swarm.Meta{CreatedAt: time.Unix(1, 0)}},
	}

	result, err := sortSecrets(ctx, dCli)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_sortSecretsErrorCLi(t *testing.T) {
	ctx := context.Background()
	dCli := &testutils.TestDockerMetricsClient{
		SecretListErr: assert.AnError,
	}

	result, err := sortSecrets(ctx, dCli)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, assert.AnError)
}
