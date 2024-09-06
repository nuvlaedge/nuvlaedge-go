package flags

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"nuvlaedge-go/common/constants"
	"nuvlaedge-go/types/options/command"
)

func AddRegisterFlags(cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.SetInterspersed(false)

	// Register flags
	flags.String("key", "", "Nuvla User API key")
	_ = cmd.MarkFlagRequired("key")
	flags.String("secret", "", "Nuvla User API secret")
	_ = cmd.MarkFlagRequired("secret")
	cmd.MarkFlagsRequiredTogether("key", "secret")

	flags.String("endpoint", constants.DefaultEndPoint, "Nuvla API endpoint")
	flags.Bool("insecure", constants.DefaultInsecure, "Skip SSL certificate verification")

	// Resource configuration
	flags.String("uuid", "", "NuvlaEdge UUID")
	flags.Bool("vpn", constants.DefaultVPNEnabled, "Enable VPN")
	flags.String("name", "", "Resource name")
	flags.String("description", "", "Resource description")
	flags.String("name-prefix", "", "Resource name prefix")
	cmd.MarkFlagsMutuallyExclusive("name", "name-prefix")

	flags.Int("refresh-interval", constants.DefaultTelemetryPeriod, "Refresh interval in seconds")
	flags.Int("heartbeat-interval", constants.DefaultHeartbeatPeriod, "Heartbeat interval in seconds")
	flags.StringArray("tags", []string{}, "Resource tags")
}

func setDefaultRegisterFlags() {
	// Set default register flags
	viper.SetDefault("endpoint", constants.DefaultEndPoint)
	viper.SetDefault("insecure", constants.DefaultInsecure)
	viper.SetDefault("refresh-interval", constants.DefaultTelemetryPeriod)
	viper.SetDefault("heartbeat-interval", constants.DefaultHeartbeatPeriod)
	viper.SetDefault("vpn", constants.DefaultVPNEnabled)
}

func bindViperRegisterFlags(flags *pflag.FlagSet) {
	errMsg := "Failed to bind register cmd flag to viper"
	OnError(viper.BindPFlag("key", flags.Lookup("key")), errMsg)
	OnError(viper.BindPFlag("secret", flags.Lookup("secret")), errMsg)

	OnError(viper.BindPFlag("endpoint", flags.Lookup("endpoint")), errMsg)
	OnError(viper.BindPFlag("insecure", flags.Lookup("insecure")), errMsg)

	OnError(viper.BindPFlag("uuid", flags.Lookup("uuid")), errMsg)
	OnError(viper.BindPFlag("vpn", flags.Lookup("vpn")), errMsg)
	OnError(viper.BindPFlag("name", flags.Lookup("name")), errMsg)
	OnError(viper.BindPFlag("description", flags.Lookup("description")), errMsg)
	OnError(viper.BindPFlag("name-prefix", flags.Lookup("name-prefix")), errMsg)

	OnError(viper.BindPFlag("refresh-interval", flags.Lookup("refresh-interval")), errMsg)
	OnError(viper.BindPFlag("heartbeat-interval", flags.Lookup("heartbeat-interval")), errMsg)
	OnError(viper.BindPFlag("tags", flags.Lookup("tags")), errMsg)
}

func setRegisterEnvBindings() {
	errMsg := "Failed to set register cmd flag env bindings"
	OnError(viper.BindEnv("key", "NUVLA_API_KEY"), errMsg)
	OnError(viper.BindEnv("secret", "NUVLA_API_SECRET"), errMsg)

	OnError(viper.BindEnv("endpoint", "NUVLA_ENDPOINT"), errMsg)
	OnError(viper.BindEnv("insecure", "NUVLA_INSECURE"), errMsg)

	OnError(viper.BindEnv("uuid", "NUVLAEDGE_UUID"), errMsg)
	OnError(viper.BindEnv("vpn", "NUVLAEDGE_VPN"), errMsg)
	OnError(viper.BindEnv("name", "NUVLAEDGE_NAME"), errMsg)
	OnError(viper.BindEnv("description", "NUVLAEDGE_DESCRIPTION"), errMsg)
	OnError(viper.BindEnv("name-prefix", "NUVLAEDGE_NAME_PREFIX"), errMsg)

	OnError(viper.BindEnv("refresh-interval", "NUVLAEDGE_REFRESH_INTERVAL"), errMsg)
	OnError(viper.BindEnv("heartbeat-interval", "NUVLAEDGE_HEARTBEAT_INTERVAL"), errMsg)
	OnError(viper.BindEnv("tags", "NUVLAEDGE_TAGS"), errMsg)
}

func ParseRegisterFlags(cmd *cobra.Command, opts *command.RegisterCmdOptions) error {
	flags := cmd.Flags()
	bindViperRegisterFlags(flags)
	setRegisterEnvBindings()
	setDefaultRegisterFlags()
	viper.AutomaticEnv()

	if err := viper.Unmarshal(opts); err != nil {
		log.Info("Failed to unmarshal register cmd flags")
		return err
	}

	return nil
}
