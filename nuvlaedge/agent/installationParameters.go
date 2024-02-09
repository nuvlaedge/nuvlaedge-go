package agent

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/nuvlaedge/common"
	"strings"
)

const (
	DefaultConfigPath = "/etc/nuvlaedge/"
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

}

func FindPreviousInstallation(configPath string, uuid string) (*InstallationParameters, error) {
	var foundUuid string

	// 1. Try finding installation in config Path
	if common.FileExists(configPath) {
		// Read installation data
		params, err := getInstallationParameters(configPath)
		if err != nil {
			return nil, err
		}
		foundUuid = params.commissioningData["uuid"].(string)
		log.Infof("Found previous installation with uuid: %s", foundUuid)
		return strings.Compare(foundUuid, uuid) == 0, nil
	} else {
		log.Infof("No previous installation found in %s", configPath)
	}

	// 2. Try finding installation in default path
	if common.FileExists(DefaultConfigPath) {
		// Read installation data
		params, err := getInstallationParameters(DefaultConfigPath)
		if err != nil {
			return nil, err
		}
		foundUuid = params.commissioningData["uuid"].(string)
		log.Infof("Found previous installation with uuid: %s", foundUuid)
		return strings.Compare(foundUuid, uuid) == 0, nil
	} else {
		log.Infof("No previous installation found in %s", DefaultConfigPath)
	}

	// 3. Try finding installation in user home
	if common.FileExists(HomeConfigPath) {
		// Read installation data
		params, err := getInstallationParameters(HomeConfigPath)
		if err != nil {
			return nil, err
		}
		foundUuid = params.commissioningData["uuid"].(string)
		log.Infof("Found previous installation with uuid: %s", foundUuid)
		return strings.Compare(foundUuid, uuid) == 0, nil
	} else {
		log.Infof("No previous installation found in %s", HomeConfigPath)
	}

	return nil, fmt.Errorf("no previous installation found in any path")
}
