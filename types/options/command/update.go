package command

type UpdateCmdOptions struct {
	Force bool
	Quiet bool

	// Job tracking
	JobId string `mapstructure:"job-id"`

	// NuvlaEdge Configuration
	Environment []string `mapstructure:"environment"`
	Project     string   `mapstructure:"project"`
	WorkingDir  string   `mapstructure:"working-dir"`

	// Version tracking
	TargetVersion  string `mapstructure:"target-version"`
	CurrentVersion string `mapstructure:"current-version"`

	// Compose Update
	ComposeFiles []string `mapstructure:"compose-files"`

	// Update failure handling
	OnUpdateFailure string `mapstructure:"on-update-failure"`
	GitHubReleases  string `mapstructure:"github-releases"`
	HardReset       bool   `mapstructure:"hard-reset"`
}
