package nuvlaedge

import "nuvlaedge-go/nuvlaedge/orchestrator"

type SystemManager struct {
	settings *SystemManagerSettings

	coeClient orchestrator.Coe

	nuvlaEdgeComponents []map[string]interface{}
}

func NewSystemManager(settings *SystemManagerSettings, coeClient orchestrator.Coe) *SystemManager {
	return &SystemManager{
		settings:  settings,
		coeClient: coeClient,
	}
}

func (s *SystemManager) startVPNClient() error {
	log.Infof("Starting VPN client")
	return nil
}

func (s *SystemManager) startMQTTBroker() error {
	log.Infof("Starting MQTT broker")
	return nil
}

func (s *SystemManager) Start() error {
	// Start VPN client based on configuration
	if s.settings.VpnEnabled {
		err := s.startVPNClient()
		if err != nil {
			log.Warnf("Error starting VPN client: %s", err)
			return err
		}
	}

	// Start mqtt broker based on configuration
	err := s.startMQTTBroker()
	if err != nil {
		log.Warnf("Error starting MQTT broker: %s", err)
		return err
	}

	// Start local storage based on configuration

	return nil
}
