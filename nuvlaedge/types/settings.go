package types

import "encoding/json"

type RunFlags struct {
	// Required
	Uuid string `json:"uuid,omitempty"`

	// Optional
	// NuvlaEdge configuration
	ConfigFile      string `json:"config-file,omitempty"`
	DataLocation    string `json:"data-location,omitempty"`
	HeartbeatPeriod int    `json:"heartbeat-period,omitempty"`
	TelemetryPeriod int    `json:"telemetry-period,omitempty"`
	VPNEnabled      bool   `json:"vpn-enabled,omitempty"`

	// Endpoint credentials and configuration
	NuvlaEndpoint string `json:"nuvla-endpoint,omitempty"`
	NuvlaInsecure bool   `json:"nuvla-insecure,omitempty"`
	ApiKey        string `json:"api-key,omitempty"`
	ApiSecret     string `json:"api-secret,omitempty"`

	// Logging Configuration
	Debug     bool   `json:"debug,omitempty"`
	LogLevel  string `json:"log-level,omitempty"`
	LogToFile bool   `json:"log-to-file,omitempty"`
	LogFile   string `json:"log-file,omitempty"`
}

func (f *RunFlags) String() string {
	s, _ := json.MarshalIndent(f, "", "  ")
	return string(s)
}
