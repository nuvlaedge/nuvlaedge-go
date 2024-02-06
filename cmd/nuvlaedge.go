package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/nuvlaedge"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	// Initialise settings
	nuvlaEdgeSettings := getNuvlaEdgeSettings()

	// Initialise logging
	initializeLogging(&nuvlaEdgeSettings.Logging)

	log.Debugf("Checking NuvlAedge minimum settings...")
	missingSettings, err := areMinimumSettingsPresent(nuvlaEdgeSettings)
	if err != nil {
		log.Errorf("Error checking minimum settings: %s", err)
		log.Errorf("Missing settings: %v", missingSettings)
	}
	log.Debugf("Checking NuvlaEdge minimum settings... Success.")
	log.Debugf("Starting NuvlaEdge with settings: %v", nuvlaEdgeSettings)

	// Create NuvlaEdge structure
	nuvlaEdge := nuvlaedge.NewNuvlaEdge(nuvlaEdgeSettings)

	// Initialize NuvlaEdge
	log.Debug("Initializing NuvlaEdge...")
	err = nuvlaEdge.Start()
	if err != nil {
		log.Error("Initializing NuvlaEdge... Failed. Exiting.")
		log.Panic(err)
		return
	}
	log.Debug("Initializing NuvlaEdge... Success.")

	// Handle signaling to graciously exit NuvlaEdge
	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)

	// Run NuvlaEdge loop
	log.Debug("Running NuvlaEdge...")
	_, err = nuvlaEdge.Run()
	// This point should only be reached if a signal is received
	if err != nil {
		log.Errorf("Error running NuvlaEdge: %s", err)
		log.Panic(err)
		// Here we should probably handle and identify signal. And Idea is to send an emergency message to Nuvla with
		// the status of the NuvlaEdge and the error message.
	}
}

func initializeLogging(loggingSettings *nuvlaedge.LoggingSettings) {
	if loggingSettings.Debug {
		log.SetLevel(log.DebugLevel)
	}
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
}

// getNuvlaEdgeSettings returns new NuvlaEdgeSettings.
// Moved to a function for testing purposes and possible future error handling.
func getNuvlaEdgeSettings() *nuvlaedge.NuvlaEdgeSettings {
	return nuvlaedge.NewNuvlaEdgeSettings()
}

// areMinimumSettingsPresent checks if the minimum settings are present.
// Moved to a function for testing purposes and possible future error handling.
func areMinimumSettingsPresent(nuvlaEdgeSettings *nuvlaedge.NuvlaEdgeSettings) ([]string, error) {
	if nuvlaEdgeSettings.Agent.NuvlaEdgeUUID == "" {
		return []string{"Agent.NuvlaEdgeUUID"}, fmt.Errorf("Agent.NuvlaEdgeUUID is missing")
	}

	return nil, nil
}
