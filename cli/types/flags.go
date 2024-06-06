package types

import "encoding/json"

type InstallFlags struct {
	Version    string
	InstallDir string
	ConfigFile string

	Service    bool
	Process    bool
	Docker     bool
	Kubernetes bool

	// Run flags
	Uuid string
}

func (f *InstallFlags) String() string {
	s, _ := json.MarshalIndent(f, "", "  ")
	return string(s)
}
