package nuvlaedge

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"nuvlaedge-go/nuvlaedge/common"
	"reflect"
)

var s *Settings

func init() {
	// Create settings and set Viper defaults and env bindings
	s = NewSettings()
	SetDefaults()
}

func NewSettings() *Settings {
	return &Settings{}
}

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

	// NuvlaEdge Telemetry s
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

func SetDefaults() {

	// NuvlaEdge Defaults
	err := viper.BindEnv("agent.nuvlaedge-uuid", "NUVLAEDGE_UUID")
	if err != nil {
		log.Errorf("Error binding env var: %s", err)
	}
	viper.SetDefault("data-location", "/var/lib/nuvlaedge/")
	_ = viper.BindEnv("data-location", "DATABASE_PATH", "DATA_LOCATION")
	viper.SetDefault("config-file", common.DefaultConfigPath+"nuvlaedge.toml")
	_ = viper.BindEnv("config-file", "NUVLAEDGE_SETTINGS")

	viper.SetDefault("agent.job-engine-image", common.JobEngineContainerImage)
	_ = viper.BindEnv("agent.job-engine-image", "NUVLAEDGE_JOB_ENGINE_LITE_IMAGE", "JOB_LEGACY_IMAGE")
	viper.SetDefault("agent.enable-legacy-job-support", false)
	_ = viper.BindEnv("agent.enable-legacy-job-support", "JOB_LEGACY_ENABLE")

	// Bind envs without defaults
	_ = viper.BindEnv("agent.api-key", "NUVLAEDGE_API_KEY")
	_ = viper.BindEnv("agent.api-secret", "NUVLAEDGE_API_SECRET")

	// Agent Defaults
	viper.SetDefault("agent.nuvla-endpoint", "https://nuvla.io")
	_ = viper.BindEnv("agent.nuvla-endpoint", "NUVLA_ENDPOINT")
	viper.SetDefault("agent.nuvla-insecure", false)
	_ = viper.BindEnv("agent.nuvla-insecure", "NUVLA_INSECURE")

	viper.SetDefault("agent.heartbeat-period", 20)
	_ = viper.BindEnv("agent.heartbeat-period", "HEARTBEAT_PERIOD")
	viper.SetDefault("agent.telemetry-period", 60)
	_ = viper.BindEnv("agent.telemetry-period", "TELEMETRY_PERIOD")

	// Logging Defaults
	viper.SetDefault("logging.debug", false)
	_ = viper.BindEnv("logging.debug", "NUVLAEDGE_DEBUG")
	viper.SetDefault("logging.log-level", "info")
	_ = viper.BindEnv("logging.log-level", "LOG_LEVEL", "NUVLAEDGE_LOG_LEVEL")
	viper.SetDefault("logging.log-to-file", false)
	_ = viper.BindEnv("logging.log-to-file", "LOG_TO_FILE")
	viper.SetDefault("logging.log-file", "/var/log/nuvlaedge.log")
	viper.SetDefault("logging.log-path", "/var/log/")
	viper.SetDefault("logging.log-max-size", 10)
	viper.SetDefault("logging.log-max-backups", 5)
}

// SetConfigFile sets the config file for the application. We need to set the config file before getting to this point
// since the config file comes from Viper configuration. It can either be set by environmental variable (NUVLAEDGE_CONFIG)
// or via flag --config-file
func SetConfigFile() error {
	// Sets viper config file
	file := viper.Get("config-file")

	if file == nil || file.(string) == "" {
		// No need to return an error here, we don't want the Config file as a mandatory flag if UUID is provided
		log.Info("No config file provided. Using defaults, envs and flags.")
		return nil
	}
	viper.SetConfigFile(file.(string))
	err := viper.ReadInConfig()
	if err != nil {
		log.Errorf("Error reading config file: %s", err)
		return err
	}
	return nil
}

// UnfoldSettings unfolds the settings from Viper into the s struct. Before that, it checks if the config file is set
// and reads it. If the config file is not set, it will use the defaults and the environmental variables.
func UnfoldSettings() error {
	if err := SetConfigFile(); err != nil {
		// If there is an error reading the config file, we should we try running without it. Just log a warning
		log.Warnf("Error reading config file: %s", err)
	}

	// Unfolds s
	if err := viper.Unmarshal(s); err != nil {
		// If the configuration is not properly unmarshalled into the s struct, we should return an error, we cannot run
		// the nuvlaedge without some configuration
		return err
	}

	return nil
}

func GetSettings() (*Settings, error) {
	if err := UnfoldSettings(); err != nil {
		log.Errorf("Error unfolding s: %s", err)
		return nil, err
	}
	return s, nil
}
