package nuvlaedge

import (
	"errors"
	"fmt"
	"github.com/nuvla/api-client-go/clients"
	log "github.com/sirupsen/logrus"
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

// CheckMinimumSettings checks if the minimum required settings are present in the AgentSettings struct.
// Takes sessionFile as input and loads the session into a file
func (a *AgentSettings) CheckMinimumSettings(sessionFile string) error {
	var f *clients.NuvlaEdgeSessionFreeze
	log.Info("Checking NuvlaEdge minimum settings...")

	if common.FileExists(sessionFile) {
		log.Info("Session file exists, loading it...")
		// Load session file
		f = &clients.NuvlaEdgeSessionFreeze{}
		if err := f.Load(sessionFile); err != nil {
			log.Infof("Error loading NuvlaEdge session freeze file: %s", err)
			// If there is an error, just reset the session freeze
			f = nil
		}
	}

	if f == nil {
		a.NuvlaEdgeUUID = common.SanitiseUUID(a.NuvlaEdgeUUID, "nuvlabox")
		// If the session file does not exist, just return and assume settings
		log.Infof("Session file does not exist, assuming settings. Probably first run")
		return nil
	}

	if f.NuvlaEdgeId != a.NuvlaEdgeUUID {
		log.Warnf("NuvlaEdge UUID in session file does not match the one in the settings, ingonring settings")
		a.NuvlaEdgeUUID = f.NuvlaEdgeId
	}
	a.NuvlaEdgeUUID = common.SanitiseUUID(a.NuvlaEdgeUUID, "nuvlabox")

	if f.Credentials.Key != a.ApiKey {
		log.Warnf("API Key in session file does not match the one in the settings, ignoring settings")
		a.ApiKey = f.Credentials.Key
	}

	if f.Credentials.Secret != a.ApiSecret {
		log.Warnf("API Secret in session file does not match the one in the settings, ignoring settings")
		a.ApiSecret = f.Credentials.Secret
	}

	if f.Endpoint != a.NuvlaEndpoint {
		log.Warnf("NuvlaEndpoint in session file does not match the one in the settings, ignoring settings")
		a.NuvlaEndpoint = f.Endpoint
	}

	if f.Insecure != a.NuvlaInsecure {
		log.Warnf("NuvlaInsecure in session file does not match the one in the settings, ignoring settings")
		a.NuvlaInsecure = f.Insecure
	}

	if a.NuvlaEdgeUUID == "" {
		return errors.New("NuvlaEdge UUID is missing and required")
	}

	return nil
}
