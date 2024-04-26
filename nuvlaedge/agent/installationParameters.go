package agent

import "github.com/nuvla/api-client-go/clients"

const (
	DefaultConfigPath = "/var/lib/nuvlaedge/"
	HomeConfigPath    = "~/.nuvlaedge/"
)

const (
	CommissioningDataFile = "commission.json"
	VpnDataFile           = "vpnHandler.json"
	NuvlaSessionDataFile  = "nuvla-session.json"
)

type InstallationParameters struct {
	commissioningData map[string]interface{}
	vpnData           map[string]interface{}
	NuvlaSessionData  *clients.NuvlaEdgeSessionFreeze `json:"nuvla-session"`
}

func getInstallationParameters(configPath string) (*InstallationParameters, error) {
	return nil, nil
}

func FindPreviousInstallation(configPath string, uuid string) (*InstallationParameters, error) {
	return nil, nil
}
