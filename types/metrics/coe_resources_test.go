package metrics

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCoeResources_WriteToStatus(t *testing.T) {
	coeResources := CoeResources{
		DockerResources: DockerResources{
			Containers: []map[string]interface{}{
				{"container1": "value1"},
			},
		},
	}

	status := NuvlaEdgeStatus{}
	err := coeResources.WriteToStatus(&status)
	assert.Nil(t, err)
	assert.Equal(t, status.CoeResources.DockerResources.Containers, coeResources.DockerResources.Containers)
}
