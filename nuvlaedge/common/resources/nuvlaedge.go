package resources

type NuvlaEdge struct {
	// Required
	State             string `json:"state,omitempty"`
	RefreshInterval   int    `json:"refresh-interval,omitempty"`
	HeartbeatInterval int    `json:"heartbeat-interval,omitempty"`
	// Opt
	Location               []float32 `json:"location,omitempty"`
	Capabilities           []string  `json:"capabilities,omitempty"`
	NuvlaEdgeEngineVersion string    `json:"nuvlaedge-engine-version,omitempty"`
	Online                 bool      `json:"online,omitempty"`
	VpnServerId            string    `json:"vpn-server-id,omitempty"`
	NuvlaEdgeStatusId      string    `json:"nuvlabox-status,omitempty"`
}
