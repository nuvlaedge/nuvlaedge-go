package metrics

type SystemMetrics struct {
	Hostname        string `json:"hostname,omitempty"`
	OperatingSystem string `json:"operating-system,omitempty"`
	Architecture    string `json:"architecture,omitempty"`
	LastBoot        string `json:"last-boot,omitempty"`
}

func (s SystemMetrics) WriteToStatus(status *NuvlaEdgeStatus) error {
	status.HostName = s.Hostname
	status.OperatingSystem = s.OperatingSystem
	status.Architecture = s.Architecture
	status.LastBoot = s.LastBoot
	return nil
}

type NetworkMetrics struct {
	DefaultGw string `json:"default-gw"`

	IPs struct {
		Public string `json:"public"`
		Swarm  string `json:"swarm"`
		Vpn    string `json:"vpn"`
		Local  string `json:"local"`
	} `json:"ips"`

	Interfaces Interfaces `json:"interfaces"`
}

func (n NetworkMetrics) WriteToStatus(status *NuvlaEdgeStatus) error {
	status.Network = n

	var globalIp string

	if status.Network.IPs.Vpn != "" {
		globalIp = status.Network.IPs.Vpn

	} else if status.Network.IPs.Local != "" {
		globalIp = status.Network.IPs.Local

	} else if status.Network.IPs.Public != "" {
		globalIp = status.Network.IPs.Public

	} else if status.Network.IPs.Swarm != "" {
		globalIp = status.Network.IPs.Swarm

	} else {
		globalIp = ""
	}

	status.IpV4Address = globalIp

	return nil
}

type Interfaces []InterfaceInfo

type InterfaceInfo struct {
	Interface string              `json:"interface"`
	Ips       []map[string]string `json:"ips"`
}
