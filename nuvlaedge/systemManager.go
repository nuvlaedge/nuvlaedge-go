package nuvlaedge

import "nuvlaedge-go/nuvlaedge/coe"

type SystemManager struct {
	settings *SystemManagerSettings

	coeClient coe.Coe
}

func NewSystemManager(settings *SystemManagerSettings, coeClient coe.Coe) *SystemManager {
	return &SystemManager{
		settings:  settings,
		coeClient: coeClient,
	}
}
