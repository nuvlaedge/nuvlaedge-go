package constants

const (
	// Default nuvla values
	DefaultEndPoint = "https://nuvla.io"
	DefaultInsecure = false

	// Default NuvlaEdge configuration
	DefaultDBPath           = "/var/lib/nuvlaedge/"
	DefaultHeartbeatPeriod  = 20
	DefaultTelemetryPeriod  = 60
	DefaultRemoteSyncPeriod = 60
	DefaultVPNEnabled       = false

	// Default Job Engine configuration
	DefaultJobEngineImage  = "sixsq/nuvlaedge:latest"
	DefaultEnableLegacyJob = false

	// Logging
	DefaultLogLevel = "info"
	DefaultDebug    = false
)
