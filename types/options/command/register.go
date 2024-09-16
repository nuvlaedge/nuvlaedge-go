package command

type RegisterCmdOptions struct {
	// Nuvla USER API keys
	Key    string `json:"-"`
	Secret string `json:"-"`

	// Nuvla Endpoint conf
	Endpoint string `json:"-"`
	Insecure bool   `json:"-"`

	// (Optional) NuvlaEdge UUID
	NuvlaEdgeUUID string `mapstructure:"uuid" json:"-"`

	// NuvlaEdge configuration
	NamePrefix string `mapstructure:"name-prefix" json:"-"` // Used to auto-generate the name.
	Name       string `json:"name"`

	// (Optional) NuvlaEdge configuration
	Description       string   `json:"description,omitempty"`
	VPNEnabled        bool     `mapstructure:"vpn" json:"-"`
	VPNServerId       string   `json:"vpn-server-id,omitempty"`
	RefreshInterval   int      `mapstructure:"refresh-interval" json:"refresh-interval,omitempty"`
	HeartbeatInterval int      `mapstructure:"heartbeat-interval" json:"heartbeat-interval"`
	Tags              []string `json:"tags,omitempty"`
}
