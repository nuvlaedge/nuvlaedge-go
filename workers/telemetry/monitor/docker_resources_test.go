package monitor

import (
	"github.com/stretchr/testify/assert"
	"nuvlaedge-go/testutils"
	"slices"
	"testing"
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
