package version

var Version string

const (
	DevVersion = "dev"
)

// GetVersion returns the version of the NuvlaEdge
func GetVersion() string {
	if Version == "" {
		return DevVersion
	}
	return Version
}
