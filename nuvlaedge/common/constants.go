package common

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

// BasePath NuvlaEdge local configuration path constants
// ConfPath
// NuvlaEdgeLocalDB
const (
	BasePath         string = "/etc/nuvlaedge/"
	ConfPath                = BasePath + "config/"
	NuvlaEdgeLocalDB        = BasePath + ".local/"
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
	BaseImageName string = "alpine:3.18"
)
