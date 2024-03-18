package agent

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
	nuvlaSessionData  map[string]interface{}
}

func getInstallationParameters(configPath string) (*InstallationParameters, error) {
	return nil, nil
}

func FindPreviousInstallation(configPath string, uuid string) (*InstallationParameters, error) {
	return nil, nil
}
