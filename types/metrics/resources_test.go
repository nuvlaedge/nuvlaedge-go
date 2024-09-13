package metrics

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_CPUMetrics_WriteToStatus(t *testing.T) {
	c := CPUMetrics{
		Load:     0.1,
		Load1:    0.2,
		Load5:    0.3,
		Capacity: 4,
	}

	status := &NuvlaEdgeStatus{}
	err := c.WriteToStatus(status)
	assert.NoErrorf(t, err, "error writing cpu metrics to status")
	assert.Equal(t, "cpu", status.Resources.CPUMetrics.Topic, "topic not set correctly")
	c.Topic = "cpu"
	b, _ := json.Marshal(c)
	assert.Equal(t, string(b), status.Resources.CPUMetrics.Raw, "raw sample not set correctly")
}

func Test_RamMetrics_WriteToStatus(t *testing.T) {
	r := RamMetrics{
		Used:     1,
		Capacity: 2,
	}

	status := &NuvlaEdgeStatus{}
	err := r.WriteToStatus(status)
	assert.NoErrorf(t, err, "error writing ram metrics to status")
	assert.Equal(t, "ram", status.Resources.RamMetrics.Topic, "topic not set correctly")
	r.Topic = "ram"
	b, _ := json.Marshal(r)
	assert.Equal(t, string(b), status.Resources.RamMetrics.Raw, "raw sample not set correctly")
}

func Test_DiskMetrics_WriteToStatus(t *testing.T) {
	dm := DiskMetrics{
		{
			Device:   "device",
			Used:     1,
			Capacity: 2,
		},
	}

	status := &NuvlaEdgeStatus{}
	err := dm.WriteToStatus(status)
	assert.NoErrorf(t, err, "error writing disk metrics to status")
	for i := range dm {
		assert.Equal(t, "disk", status.Resources.DiskMetrics[i].Topic, "topic not set correctly")
		dm[i].Topic = "disk"
		assert.Equal(t, dm[i].Raw, status.Resources.DiskMetrics[i].Raw, "raw sample not set correctly")
	}
}

func Test_IfacesMetrics_WriteToStatus(t *testing.T) {
	n := IfacesMetrics{
		{
			Interface:        "name",
			BytesReceived:    1,
			BytesTransmitted: 2,
		},
	}

	status := &NuvlaEdgeStatus{}
	err := n.WriteToStatus(status)
	assert.NoErrorf(t, err, "error writing network metrics to status")
	assert.Equal(t, n, status.Resources.NetStats, "network metrics not written to status")
}

func Test_GetRaw(t *testing.T) {
	i := struct {
		A string `json:"a"`
	}{
		A: "a",
	}
	b, err := json.Marshal(i)
	assert.NoError(t, err)
	assert.Equal(t, string(b), GetRaw(i))

	x := GetRaw(make(chan struct{}))
	assert.Equal(t, "", x)
}
