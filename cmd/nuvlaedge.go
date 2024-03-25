package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	_ "net/http/pprof"
	"nuvlaedge-go/nuvlaedge"
	"os"
	"runtime"
	"strings"
)

var version = "development"

func main() {
	log.Infof("Starting NuvlaEdge version %s", version)

	// Initialise settings
	nuvlaEdgeSettings := getNuvlaEdgeSettings()

	// Initialise logging
	initializeLogging(&nuvlaEdgeSettings.Logging)

	log.Infof("Agent settings: %s", nuvlaEdgeSettings.Agent.String())
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
	log.Info("Initializing NuvlaEdge...")
	err = nuvlaEdge.Start()
	if err != nil {
		log.Error("Initializing NuvlaEdge... Failed. Exiting.")
		log.Panic(err)
		return
	}
	log.Info("Initializing NuvlaEdge... Success.")

	// Handle signaling to graciously exit NuvlaEdge
	//exitSignal := make(chan os.Signal, 1)
	//signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)

	// Run NuvlaEdge loop
	log.Info("Running NuvlaEdge...")
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
	log.SetReportCaller(true)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:          true,
		TimestampFormat:        "02-01-2006 15:04:05",
		DisableLevelTruncation: true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			return "", fmt.Sprintf("%s:%d", formatFilePath(f.File), f.Line)
		},
	})
	log.SetOutput(os.Stdout)
}

func formatFilePath(filePath string) string {
	arr := strings.Split(filePath, "/")
	return arr[len(arr)-1]
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
