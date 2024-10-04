package settings

type NuvlaEdgeSettings struct {
	// NuvlaEdge Database Location
	DBPPath    string `toml:"db-path" json:"db-path,omitempty" mapstructure:"db-path"`
	ConfigFile string `toml:"config-file" json:"config-file,omitempty" mapstructure:"config-file"`
	RootFs     string `toml:"rootfs" json:"rootfs,omitempty" mapstructure:"rootfs"`

	// nuvla endpoint definition
	NuvlaEndpoint string `mapstructure:"nuvla-endpoint" toml:"nuvla-endpoint" json:"nuvla-endpoint,omitempty"`
	NuvlaInsecure bool   `mapstructure:"nuvla-insecure" toml:"nuvla-insecure" json:"nuvla-insecure,omitempty"`

	// nuvlaedge resource id and (optional) credentials
	NuvlaEdgeUUID string `mapstructure:"nuvlaedge-uuid" toml:"nuvlaedge-uuid" json:"nuvlaedge-uuid,omitempty"`
	ApiKey        string `mapstructure:"api-key" toml:"api-key" json:"-"`
	ApiSecret     string `mapstructure:"api-secret" toml:"api-secret" json:"-"`

	// NuvlaEdge main jobs periods
	HeartbeatPeriod  int `mapstructure:"heartbeat-period" toml:"heartbeat-period" json:"heartbeat-period,omitempty"`
	TelemetryPeriod  int `mapstructure:"telemetry-period" toml:"telemetry-period" json:"telemetry-period,omitempty"`
	RemoteSyncPeriod int `mapstructure:"remote-sync-period" toml:"remote-sync-period" json:"remote-sync-period,omitempty"`
	CleanUpPeriod    int `mapstructure:"cleanup-period" toml:"cleanup-period" json:"cleanup-period,omitempty"`

	// Resource cleanup
	Resources []string `mapstructure:"resources" toml:"resources" json:"resources,omitempty"`

	// VPN settings
	VpnEnabled     bool   `mapstructure:"vpn-enabled" toml:"vpn-enabled" json:"vpn-enabled,omitempty"`
	VpnExtraConfig string `mapstructure:"vpn-extra-config" toml:"vpn-extra-config" json:"vpn-extra-config,omitempty"`

	// Job Engine
	JobEngineImage         string `mapstructure:"job-engine-image" toml:"job-engine-image" json:"job-engine-image,omitempty"`
	EnableJobLegacySupport bool   `mapstructure:"enable-legacy-job" toml:"enable-legacy-job" json:"enable-legacy-job,omitempty"`

	// Logging
	LogLevel string `mapstructure:"log-level" toml:"log-level" json:"log-level,omitempty"`
	Debug    bool   `mapstructure:"debug" toml:"debug" json:"debug,omitempty"`

	// Irs
	Irs string `mapstructure:"irs" toml:"irs" json:"irs,omitempty"`
}
