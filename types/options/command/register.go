package command

type RegisterCmdOptions struct {
	// Nuvla USER API keys
	Key    string
	Secret string

	// Nuvla Endpoint conf
	Endpoint string
	Insecure bool

	// (Optional) NuvlaEdge UUID
	NuvlaEdgeUUID string `mapstructure:"uuid"`

	// NuvlaEdge configuration
	NamePrefix string `mapstructure:"name-prefix"` // Used to auto-generate the name.
	Name       string `json:"name"`

	// (Optional) NuvlaEdge configuration
	Description       string   `json:"description,omitempty"`
	VPNEnabled        bool     `mapstructure:"vpn"`
	VPNServerId       string   `json:"vpn-server-id,omitempty"`
	RefreshInterval   int      `mapstructure:"refresh-interval" json:"refresh-interval,omitempty"`
	HeartbeatInterval int      `mapstructure:"heartbeat-interval" json:"heartbeat-interval"`
	Tags              []string `json:"tags,omitempty"`
}
