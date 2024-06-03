package common

// Path constants
const (
	// BinarySudoPath default path for the binary when running with sudo
	BinarySudoPath = "/usr/local/bin/"
	// ConfigSudoPath default path for the configuration when running with sudo
	ConfigSudoPath = "/etc/nuvlaedge/"
	// DatabaseSudoPath default path for the database when running with sudo
	DatabaseSudoPath = "/var/lib/nuvlaedge/"
	// LogsSudoPath default path for the logs when running with sudo
	LogsSudoPath = "/var/log/nuvlaedge/"

	// BinaryPath default path for the binary
	BinaryPath = ".nuvlaedge/bin/"
	// ConfigPath default path for the configuration
	ConfigPath = ".nuvlaedge/config/"
	// DatabasePath default path for the database
	DatabasePath = ".nuvlaedge/db/"
	// LogsPath default path for the logs
	LogsPath = ".nuvlaedge/logs/"

	// TempDir temporal directory for downloads
	TempDir = "/tmp/nuvlaedge/"

	// DefaultServicePath default path for the service file
	DefaultServicePath = "/etc/systemd/system/"
)
