package flags

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"nuvlaedge-go/types/options/command"
)

func AddUpdateFlags(cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.SetInterspersed(false)

	// Update flags
	flags.Bool("force", false, "Force update")
	flags.Bool("quiet", false, "Quiet mode")

	flags.String("job-id", "", "Job ID")

	flags.StringSlice("environment", []string{}, "Comma separated list of environments KEY=VALUE")
	flags.String("project", "", "Project")
	flags.String("working-dir", "", "Working directory")

	flags.String("target-version", "", "Target version")
	flags.String("current-version", "", "Current version")

	flags.StringSlice("compose-files", []string{}, "Compose files")

	flags.String("on-update-failure", "", "On update failure")
	flags.String("github-releases", "", "GitHub releases")
	flags.Bool("hard-reset", false, "Hard reset")
}

func setDefaultUpdateFlags() {
	// Set default update flags
	viper.SetDefault("github-releases", "")
	viper.SetDefault("on-update-failure", "")
}

func bindViperUpdateFlags(flags *pflag.FlagSet) {
	errMsg := "Failed to bind update cmd flag to viper"
	OnError(viper.BindPFlag("force", flags.Lookup("force")), errMsg)
	OnError(viper.BindPFlag("quiet", flags.Lookup("quiet")), errMsg)

	OnError(viper.BindPFlag("job-id", flags.Lookup("job-id")), errMsg)

	OnError(viper.BindPFlag("environment", flags.Lookup("environment")), errMsg)
	OnError(viper.BindPFlag("project", flags.Lookup("project")), errMsg)
	OnError(viper.BindPFlag("working-dir", flags.Lookup("working-dir")), errMsg)

	OnError(viper.BindPFlag("target-version", flags.Lookup("target-version")), errMsg)
	OnError(viper.BindPFlag("current-version", flags.Lookup("current-version")), errMsg)

	OnError(viper.BindPFlag("compose-files", flags.Lookup("compose-files")), errMsg)

	OnError(viper.BindPFlag("on-update-failure", flags.Lookup("on-update-failure")), errMsg)
	OnError(viper.BindPFlag("github-releases", flags.Lookup("github-releases")), errMsg)
	OnError(viper.BindPFlag("hard-reset", flags.Lookup("hard-reset")), errMsg)
}

func setUpdateEnvBindings() {
	errMsg := "Failed to set update env binding"
	OnError(viper.BindEnv("force", "FORCE_UPDATE"), errMsg)
	OnError(viper.BindEnv("quiet", "QUIET"), errMsg)

	OnError(viper.BindEnv("job-id", "JOB_ID"), errMsg)

	OnError(viper.BindEnv("environment", "ENVIRONMENT"), errMsg)
	OnError(viper.BindEnv("project", "PROJECT"), errMsg)
	OnError(viper.BindEnv("working-dir", "WORKING_DIR"), errMsg)

	OnError(viper.BindEnv("target-version", "TARGET_VERSION"), errMsg)
	OnError(viper.BindEnv("current-version", "CURRENT_VERSION"), errMsg)

	OnError(viper.BindEnv("compose-files", "COMPOSE_FILES"), errMsg)

	OnError(viper.BindEnv("on-update-failure", "ON_UPDATE_FAILURE"), errMsg)
	OnError(viper.BindEnv("github-releases", "GITHUB_RELEASES"), errMsg)
	OnError(viper.BindEnv("hard-reset", "HARD_RESET"), errMsg)
}

func ParseUpdateFlags(flags *pflag.FlagSet, opts *command.UpdateCmdOptions) error {
	bindViperUpdateFlags(flags)
	setDefaultUpdateFlags()
	setUpdateEnvBindings()
	setEnvBindings()
	viper.AutomaticEnv()

	if err := viper.Unmarshal(opts); err != nil {
		log.Info("Failed to unmarshal update cmd flags")
		return err
	}

	return nil
}
