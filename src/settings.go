package src

import (
	"github.com/BurntSushi/toml"
	"github.com/caarlos0/env/v10"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/src/common"
	"os"
)

type NuvlaEdgeSettings struct {
	// Agent settings
	Agent AgentSettings `toml:"agent"`

	// NuvlaEdge System Manager settings
	SystemManager SystemManagerSettings `toml:"system-manager"`

	// Higher logging levels
	Logging LoggingSettings `toml:"logging"`
}

type LoggingSettings struct {
	Debug         bool   `toml:"debug" env:"DEBUG"`
	Level         string `toml:"level" env:"LOG_LEVEL"`
	LogFile       string `toml:"log-file" env:"LOG_FILE"`
	LogPath       string `toml:"log-path" env:"LOG_PATH"`
	LogMaxSize    int    `toml:"log-max-size" env:"LOG_MAX_SIZE"`
	LogMaxBackups int    `toml:"log-max-backups" env:"LOG_MAX_BACKUPS"`
}

// SystemManagerSettings struct holds the configuration settings for the NuvlaEdge System Manager.
// SystemRequirements: Defines the minimum system requirements for the NuvlaEdge System Manager.
// VpnEnabled: Indicates whether the VPN is enabled from the start.
// Mqtt: Holds the configuration settings for the MQTT broker.
type SystemManagerSettings struct {

	// SystemRequirements struct defines the minimum system requirements for the NuvlaEdge.
	// Cores: The minimum number of CPU cores required.
	// Memory: The minimum amount of memory required.
	// Disk: The minimum amount of disk space required.
	// DockerVersion: The required version of Docker.
	// K8sVersion: The required version of Kubernetes.
	SystemRequirements struct {
		// CPU requirements
		Cores int `toml:"cores" env:"CORES"`
		// Memory requirements
		Memory int `toml:"memory" env:"MEMORY"`
		// Disk requirements
		Disk int `toml:"disk" env:"DISK"`
		// COE requirements
		DockerVersion string `toml:"docker-version" env:"DOCKER_VERSION"`
		K8sVersion    string `toml:"k8s-version" env:"K8S_VERSION"`
	} `toml:"system-requirements"`

	// VpnEnabled indicates whether the VPN is enabled from the start.
	VpnEnabled bool `toml:"vpn-enabled" env:"VPN_ENABLED"`

	// Mqtt struct holds the configuration settings for the MQTT broker.
	// Enabled: Indicates whether the MQTT broker is enabled.
	// Host: The host of the MQTT broker.
	// Port: The port of the MQTT broker.
	Mqtt struct {
		Enabled bool   `toml:"enabled" env:"MQTT_ENABLED"`
		Host    string `toml:"host" env:"MQTT_HOST"`
		Port    int    `toml:"port" env:"MQTT_PORT"`
	} `toml:"mqtt"`
}

// AgentSettings struct holds the configuration settings for the NuvlaEdge agent.
// NuvlaEndpoint: The endpoint for the Nuvla service.
// NuvlaInsecure: Indicates whether the Nuvla service should be accessed in insecure mode.
// NuvlaEdgeUUID: The UUID of the NuvlaEdge resource.
// ApiKey: The API key for accessing the Nuvla service.
// ApiSecret: The API secret for accessing the Nuvla service.
// HeartbeatPeriod: The period for the heartbeat action of the NuvlaEdge agent.
// TelemetryPeriod: The period for the telemetry action of the NuvlaEdge agent.
// Commissioner: Holds the settings for the NuvlaEdge commissioner.
// SystemConfiguration: Holds the settings for the host system configuration.
// Telemetry: Holds the settings for the NuvlaEdge telemetry.
// Vpn: Holds the settings for the NuvlaEdge VPN.
type AgentSettings struct {
	// nuvla endpoint definition
	NuvlaEndpoint string `toml:"nuvla-endpoint" env:"NUVLA_ENDPOINT"`
	NuvlaInsecure bool   `toml:"nuvla-insecure" env:"NUVLA_INSECURE"`
	// nuvlaedge resource id and (optional) credentials
	NuvlaEdgeUUID string `toml:"nuvlaedge-uuid" env:"NUVLAEDGE_UUID"`
	ApiKey        string `toml:"api-key" env:"API_KEY"`
	ApiSecret     string `toml:"api-secret" env:"API_SECRET"`
	// NuvlaEdge main actions periods
	HeartbeatPeriod int `toml:"heartbeat-period" env:"HEARTBEAT_PERIOD"`
	TelemetryPeriod int `toml:"telemetry-period" env:"TELEMETRY_PERIOD"`

	// Commissioner settings
	Commissioner struct {
		Period int `toml:"period" env:"COMMISSIONER_PERIOD"`
	} `toml:"commissioner"`

	// NuvlaEdge Telemetry settings
	Telemetry struct {
		Period int `toml:"period" env:"TELEMETRY_PERIOD"`
	} `toml:"telemetry"`

	// HostConfiguration settings
	HostConfiguration struct {
		Period int `toml:"period" env:"SYSTEM_CONFIGURATION_PERIOD"`
	} `toml:"host-configuration"`

	// Vpn settings
	Vpn struct {
		Enabled     bool   `toml:"enabled" env:"VPN_ENABLED"`
		ExtraConfig string `toml:"extra-config" env:"VPN_EXTRA_CONFIG"`
	} `toml:"vpn"`
}

func NewNuvlaEdgeSettings() *NuvlaEdgeSettings {
	path := GetSettingsPath()
	cfg := &NuvlaEdgeSettings{}
	readTomlSettings(cfg, path)
	readEnv(cfg)

	return cfg
}

func readEnv(cfg *NuvlaEdgeSettings) {
	err := env.Parse(cfg)
	if err != nil {
		log.Warnf("Error parsing environment variables: %s", err)
	}
}

func readTomlSettings(cfg *NuvlaEdgeSettings, path string) {
	if !common.FileExistsAndNotEmpty(path) {
		log.Infof("Settings file not found: %s. Will try with environmeltal variables", path)
		return
	}
	log.Debugf("Reading settings from file: %s...", path)
	_, err := toml.DecodeFile(path, cfg)
	if err != nil {
		log.Warnf("Error reading settings file: %s", err)
		return
	}
	log.Debugf("Reading settings from file: %s... Success", path)

}

func readCommandLineSetting(cfg *NuvlaEdgeSettings) {

}

func GetSettingsPath() string {
	envSettingsFile := os.Getenv("NUVLAEDGE_SETTINGS")
	if envSettingsFile != "" {
		return envSettingsFile
	}
	return "/etc/nuvlaedge/settings.toml"

}
