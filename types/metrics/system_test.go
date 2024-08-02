package metrics

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_SystemMetrics_WriteToStatus(t *testing.T) {
	s := SystemMetrics{
		Hostname:        "hostname",
		OperatingSystem: "operating-system",
		Architecture:    "architecture",
		LastBoot:        "last-boot",
	}

	status := &NuvlaEdgeStatus{}
	err := s.WriteToStatus(status)
	assert.NoErrorf(t, err, "error writing system metrics to status")
	assert.Equal(t, s.Hostname, status.HostName, "hostname not set correctly")
	assert.Equal(t, s.OperatingSystem, status.OperatingSystem, "operating system not set correctly")
	assert.Equal(t, s.Architecture, status.Architecture, "architecture not set correctly")
	assert.Equal(t, s.LastBoot, status.LastBoot, "last boot not set correctly")
}

func Test_NetworkMetrics_WriteToStatus(t *testing.T) {
	n := NetworkMetrics{
		DefaultGw: "default-gw",
		IPs: struct {
			Public string `json:"public"`
			Swarm  string `json:"swarm"`
			Vpn    string `json:"vpn"`
			Local  string `json:"local"`
		}{
			Public: "public",
			Swarm:  "swarm",
			Vpn:    "vpn",
			Local:  "local",
		},
		Interfaces: Interfaces{
			{
				Interface: "interface",
				Ips:       []map[string]string{{"key1": "value1"}, {"key2": "value2"}},
			},
		},
	}

	status := &NuvlaEdgeStatus{}
	err := n.WriteToStatus(status)
	assert.NoErrorf(t, err, "error writing network metrics to status")
	assert.Equal(t, n.DefaultGw, status.Network.DefaultGw, "default gw not set correctly")
	assert.Equal(t, n.IPs.Public, status.Network.IPs.Public, "public IP not set correctly")
	assert.Equal(t, n.IPs.Swarm, status.Network.IPs.Swarm, "swarm IP not set correctly")
	assert.Equal(t, n.IPs.Vpn, status.Network.IPs.Vpn, "vpn IP not set correctly")
	assert.Equal(t, n.IPs.Local, status.Network.IPs.Local, "local IP not set correctly")
	assert.Equal(t, n.Interfaces, status.Network.Interfaces, "interfaces not set correctly")
}
