package nuvlaedge

import (
	"fmt"
	"nuvlaedge-go/nuvlaedge/common"
	"reflect"
)

type Settings struct {
	// Agent s
	Agent AgentSettings `toml:"agent" mapstructure:"agent" json:"agent,omitempty"`

	// Higher logging levels
	Logging common.LoggingSettings `toml:"logging" mapstructure:"logging" json:"logging,omitempty"`

	// NuvlaEdge Database Location
	DataLocation string `toml:"data-location" json:"data-location,omitempty" mapstructure:"data-location"`
	ConfigFile   string `toml:"config-file" json:"config-file,omitempty" mapstructure:"config-file"`
}

// AgentSettings struct holds the configuration s for the NuvlaEdge agent.
// NuvlaEndpoint: The endpoint for the Nuvla service.
// NuvlaInsecure: Indicates whether the Nuvla service should be accessed in insecure mode.
// NuvlaEdgeUUID: The UUID of the NuvlaEdge resource.
// ApiKey: The API key for accessing the Nuvla service.
// ApiSecret: The API secret for accessing the Nuvla service.
// HeartbeatPeriod: The period for the heartbeat action of the NuvlaEdge agent.
// TelemetryPeriod: The period for the monitoring action of the NuvlaEdge agent.
// Commissioner: Holds the s for the NuvlaEdge commissioner.
// Telemetry: Holds the s for the NuvlaEdge monitoring.
// Vpn: Holds the s for the NuvlaEdge VPN.
type AgentSettings struct {
	// nuvla endpoint definition
	NuvlaEndpoint string `mapstructure:"nuvla-endpoint" toml:"nuvla-endpoint" json:"nuvla-endpoint,omitempty"`
	NuvlaInsecure bool   `mapstructure:"nuvla-insecure" toml:"nuvla-insecure" json:"nuvla-insecure,omitempty"`

	// nuvlaedge resource id and (optional) credentials
	NuvlaEdgeUUID string `mapstructure:"nuvlaedge-uuid" toml:"nuvlaedge-uuid" json:"nuvlaedge-uuid,omitempty"`
	ApiKey        string `mapstructure:"api-key" toml:"api-key" json:"api-key,omitempty"`
	ApiSecret     string `mapstructure:"api-secret" toml:"api-secret" json:"api-secret,omitempty"`

	// NuvlaEdge main jobs periods
	HeartbeatPeriod int `mapstructure:"heartbeat-period" toml:"heartbeat-period" json:"heartbeat-period,omitempty"`
	TelemetryPeriod int `mapstructure:"telemetry-period" toml:"telemetry-period" json:"telemetry-period,omitempty"`

	// Commissioner s
	Commissioner struct {
		Period int `mapstructure:"period" toml:"period" json:"period,omitempty"`
	} `mapstructure:"commissioner" toml:"commissioner" json:"commissioner,omitempty"`

	// NuvlaEdge Telemetry
	Telemetry struct {
		Period int `mapstructure:"period" toml:"period" json:"period,omitempty"`
	} `mapstructure:"telemetry" toml:"telemetry" json:"telemetry,omitempty"`

	// HostConfiguration s
	HostConfiguration struct {
		Period int `mapstructure:"period" toml:"period" json:"period,omitempty"`
	} `mapstructure:"host-configuration" toml:"host-configuration" json:"host-configuration,omitempty"`

	// Vpn s
	Vpn struct {
		Enabled     bool   `mapstructure:"enabled" toml:"enabled" json:"enabled,omitempty"`
		ExtraConfig string `mapstructure:"extra-config" toml:"extra-config" json:"extra-config,omitempty"`
	} `mapstructure:"vpn" toml:"vpn" json:"vpn,omitempty"`

	// Job Engine Configuration
	JobEngineImage         string `mapstructure:"job-engine-image" toml:"job-engine-image" json:"job-engine-image,omitempty"`
	EnableLegacyJobSupport bool   `mapstructure:"enable-legacy-job-support" toml:"enable-legacy-job-support" json:"enable-legacy-job-support,omitempty"`
}

func (a *AgentSettings) String() string {

	s := "\n"
	v := reflect.ValueOf(a).Elem()
	for i := 0; i < v.NumField(); i++ {
		s += fmt.Sprintf("%s: %v\n", v.Type().Field(i).Name, v.Field(i))
	}
	return s
}
