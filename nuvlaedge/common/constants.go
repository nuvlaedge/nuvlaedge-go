package common

import log "github.com/sirupsen/logrus"

// Nuvla Endpoint constant configuration
const (
	NuvlaEndPoint  string = "https://nuvla.io"
	EndPointSecure bool   = true
	DatetimeFormat string = "2006-01-02T15:04:05Z"
)

// SessionTemplate
// NuvlaClient Constant templates
const (
	SessionTemplate = "session-template/api-key"
	SessionEndpoint = "/api/session"
)

const (
	// BasePath NuvlaEdge local configuration path constants
	BasePath string = "/etc/nuvlaedge/"
	// ConfPath Location to NuvlaEdge configuration files locally
	ConfPath = BasePath + "config/"
	// NuvlaEdgeLocalDB Local database path
	NuvlaEdgeLocalDB = BasePath + ".local/"
)

// NuvlaEdgeUserConfig
const (
	NuvlaEdgeConfigFileName string = "settings.toml"
	NuvlaEdgeUserConfig            = ConfPath + NuvlaEdgeConfigFileName
)

// TODO: Release usage of pathlib

// BaseImageName common NuvlaEdge image
// Image Constants
const (
	BaseImageName           string = "alpine:3.18"
	JobEngineContainerImage        = "sixsq/nuvlaedge:latest"
)

// LogLevel Global log level propagator
var LogLevel log.Level = log.InfoLevel

func SetLogLevel(level string) {
	l, err := log.ParseLevel(level)
	if err != nil {
		log.Warnf("Invalid log level: %s", level)
		LogLevel = log.InfoLevel
	} else {
		LogLevel = l
	}
}
