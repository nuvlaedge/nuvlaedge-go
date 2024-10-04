package flags

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"nuvlaedge-go/common/constants"
	"nuvlaedge-go/types/settings"
)

var OnError = func(err error, msg string) {
	if err != nil {
		log.Warnf("%s: %s", msg, err)
	}
}

// AddRunFlags defines and adds flags related to the NuvlaEdge runtime configuration to a given flag set.
// This function is used to configure the command-line interface of the NuvlaEdge agent, allowing users
// to specify various operational parameters such as database path, NuvlaEdge ID, API keys, and more.
//
// Parameters:
// - flags (*pflag.FlagSet): A pointer to the flag set to which the NuvlaEdge runtime flags will be added.
//
// The function does not return any value.
func AddRunFlags(flags *pflag.FlagSet) {
	// Database location
	flags.String("db-path", "", "NuvlaEdge Database path")
	flags.String("rootfs", constants.DefaultRootFs, "Root filesystem")

	// NuvlaEdge resource
	flags.String("uuid", "", "NuvlaEdge ID")
	flags.String("api-key", "", "NuvlaEdge API key")
	flags.String("api-secret", "", "NuvlaEdge API secret")
	flags.String("irs", "", "NuvlaEdge IRS")

	// Agent configuration
	flags.Int("heartbeat-period", 0, "Heartbeat period")
	flags.Int("telemetry-period", 0, "Telemetry period")
	flags.Int("remote-sync-period", 0, "Remote sync period")

	// Resource cleanup
	flags.Int("cleanup-period", 0, "COE (Docker/K8s) Cleanup period")
	flags.StringSlice("resources", []string{}, "Resources to cleanup")

	// VPN settings
	flags.Bool("vpn-enabled", false, "VPN enabled")
	flags.String("vpn-extra-config", "", "VPN extra configuration")

	// Job Engine
	flags.String("job-image", "", "Job Engine image")
	flags.Bool("enable-legacy-job", false, "Enable legacy job support")

	// Nuvla endpoint definition
	flags.String("nuvla-endpoint", "", "Nuvla endpoint")
	flags.Bool("nuvla-insecure", false, "Insecure connection")

	// NuvlaEdge logging
	flags.String("log-level", "info", "Log level")
	flags.Bool("debug", false, "Debug mode. Will ignore any Level set in log-level")
}

// setDefaultRunFlags sets the default values for the NuvlaEdge runtime configuration flags.
// This function configures default settings for various operational parameters such as the database path,
// NuvlaEdge endpoint, telemetry periods, and VPN settings among others. These defaults are used unless
// overridden by command-line flags or environment variables.
//
// The function leverages Viper's SetDefault method to establish these defaults, ensuring that the application
// has sensible default values even if no specific configurations are provided by the user.
//
// No parameters are required and the function does not return any value.
func setDefaultRunFlags() {
	// Set default run flags
	viper.SetDefault("db-path", constants.DefaultDBPath)
	viper.SetDefault("rootfs", constants.DefaultRootFs)
	viper.SetDefault("nuvla-endpoint", constants.DefaultEndPoint)
	viper.SetDefault("nuvla-insecure", constants.DefaultInsecure)
	viper.SetDefault("heartbeat-period", constants.DefaultHeartbeatPeriod)
	viper.SetDefault("telemetry-period", constants.DefaultTelemetryPeriod)
	viper.SetDefault("remote-sync-period", constants.DefaultRemoteSyncPeriod)
	viper.SetDefault("vpn-enabled", constants.DefaultVPNEnabled)
	viper.SetDefault("job-engine-image", constants.DefaultJobEngineImage)
	viper.SetDefault("enable-legacy-job", constants.DefaultEnableLegacyJob)
	viper.SetDefault("log-level", constants.DefaultLogLevel)
	viper.SetDefault("debug", constants.DefaultDebug)
	viper.SetDefault("cleanup-period", 86400)
	viper.SetDefault("resources", []string{"images"})
}

// bindViperRunFlags binds each command-line flag to its corresponding Viper configuration key.
// This function ensures that the values provided through the command-line flags are accessible
// via Viper's unified configuration system, allowing for easy retrieval of configuration values
// throughout the application. It also sets up error handling for each binding operation to log
// warnings in case of any issues during the binding process.
//
// Parameters:
// - flags (*pflag.FlagSet): A pointer to the flag set from which the flags will be bound to Viper keys.
//
// This function does not return any value. It logs a warning message if an error occurs during the
// binding of a flag to a Viper key, using the provided onError callback function for error handling.
func bindViperRunFlags(flags *pflag.FlagSet) {
	// Bind viper run flags
	errMsg := "Error binding flag to viper var"

	OnError(viper.BindPFlag("db-path", flags.Lookup("db-path")), errMsg)
	OnError(viper.BindPFlag("rootfs", flags.Lookup("rootfs")), errMsg)
	OnError(viper.BindPFlag("nuvla-endpoint", flags.Lookup("nuvla-endpoint")), errMsg)
	OnError(viper.BindPFlag("nuvla-insecure", flags.Lookup("nuvla-insecure")), errMsg)
	OnError(viper.BindPFlag("nuvlaedge-uuid", flags.Lookup("uuid")), errMsg)
	OnError(viper.BindPFlag("api-key", flags.Lookup("api-key")), errMsg)
	OnError(viper.BindPFlag("api-secret", flags.Lookup("api-secret")), errMsg)
	OnError(viper.BindPFlag("heartbeat-period", flags.Lookup("heartbeat-period")), errMsg)
	OnError(viper.BindPFlag("telemetry-period", flags.Lookup("telemetry-period")), errMsg)
	OnError(viper.BindPFlag("remote-sync-period", flags.Lookup("remote-sync-period")), errMsg)
	OnError(viper.BindPFlag("cleanup-period", flags.Lookup("cleanup-period")), errMsg)
	OnError(viper.BindPFlag("resources", flags.Lookup("resources")), errMsg)
	OnError(viper.BindPFlag("vpn-enabled", flags.Lookup("vpn-enabled")), errMsg)
	OnError(viper.BindPFlag("vpn-extra-config", flags.Lookup("vpn-extra-config")), errMsg)
	OnError(viper.BindPFlag("job-engine-image", flags.Lookup("job-image")), errMsg)
	OnError(viper.BindPFlag("enable-legacy-job", flags.Lookup("enable-legacy-job")), errMsg)
	OnError(viper.BindPFlag("log-level", flags.Lookup("log-level")), errMsg)
	OnError(viper.BindPFlag("debug", flags.Lookup("debug")), errMsg)
	OnError(viper.BindPFlag("irs", flags.Lookup("irs")), errMsg)
}

// setEnvBindings binds environment variables to their corresponding Viper configuration keys.
// This function ensures that configuration values can be provided through environment variables,
// offering an alternative to command-line flags. It sets up error handling for each binding operation
// to log warnings in case of any issues during the binding process. This is particularly useful for
// configuring the application in environments where setting environment variables is preferred over
// passing command-line arguments.
//
// No parameters are required and the function does not return any value. It logs a warning message
// if an error occurs during the binding of an environment variable to a Viper key, using the provided
// onError callback function for error handling.
func setEnvBindings() {
	// Env run flags
	errMsg := "Error binding env var to viper var"

	OnError(viper.BindEnv("db-path", "DB_PATH", "DATA_LOCATION"), errMsg)
	OnError(viper.BindEnv("rootfs", "ROOTFS"), errMsg)
	OnError(viper.BindEnv("nuvla-endpoint", "NUVLA_ENDPOINT"), errMsg)
	OnError(viper.BindEnv("nuvla-insecure", "NUVLA_INSECURE"), errMsg)
	OnError(viper.BindEnv("nuvlaedge-uuid", "NUVLAEDGE_UUID"), errMsg)
	OnError(viper.BindEnv("api-key", "NUVLAEDGE_API_KEY"), errMsg)
	OnError(viper.BindEnv("api-secret", "NUVLAEDGE_API_SECRET"), errMsg)
	OnError(viper.BindEnv("irs", "NUVLAEDGE_IRS"), errMsg)
	OnError(viper.BindEnv("heartbeat-period", "HEARTBEAT_PERIOD"), errMsg)
	OnError(viper.BindEnv("telemetry-period", "TELEMETRY_PERIOD"), errMsg)
	OnError(viper.BindEnv("remote-sync-period", "REMOTE_SYNC_PERIOD"), errMsg)
	OnError(viper.BindEnv("cleanup-period", "CLEANUP_PERIOD"), errMsg)
	OnError(viper.BindEnv("resources", "CLEAN_RESOURCES"), errMsg)
	OnError(viper.BindEnv("job-engine-image", "NUVLAEDGE_JOB_ENGINE_LITE_IMAGE", "JOB_LEGACY_IMAGE"), errMsg)
	OnError(viper.BindEnv("enable-legacy-job", "ENABLE_LEGACY_JOB", "JOB_LEGACY_ENABLE"), errMsg)
	OnError(viper.BindEnv("vpn-enabled", "VPN_ENABLED"), errMsg)
	OnError(viper.BindEnv("vpn-extra-config", "VPN_EXTRA_CONFIG"), errMsg)
	OnError(viper.BindEnv("log-level", "NUVLAEDGE_LOG_LEVEL"), errMsg)
	OnError(viper.BindEnv("debug", "DEBUG", "NUVLAEDGE_DEBUG"), errMsg)
}

// ParseSettings initializes and applies configuration settings for the NuvlaEdge application.
// This function binds command-line flags to their corresponding Viper configuration keys,
// sets default values for various operational parameters, binds environment variables to
// their corresponding Viper configuration keys, and finally unmarshals the configuration
// into a NuvlaEdgeSettings struct. It ensures that all configurations are correctly applied
// and accessible throughout the application. This function should be called after defining
// command-line flags using AddRunFlags to ensure all flags are correctly initialized.
//
// Parameters:
// - flags (*pflag.FlagSet): A pointer to the flag set containing the command-line flags.
// - set (*settings.NuvlaEdgeSettings): A pointer to the struct where the unmarshaled configuration will be stored.
//
// Returns:
// - error: Returns an error if unmarshaling the configuration into the NuvlaEdgeSettings struct fails.
func ParseSettings(flags *pflag.FlagSet, opts *settings.NuvlaEdgeSettings) error {
	bindViperRunFlags(flags) // Bind command-line flags to Viper keys.
	setDefaultRunFlags()     // Set default values for the configuration flags.
	setEnvBindings()         // Bind environment variables to Viper keys.
	viper.AutomaticEnv()     // Automatically load environment variables.

	if err := viper.Unmarshal(opts); err != nil {
		log.Infof("Error unmarshaling settings: %s", err)
		return err
	}
	return nil
}
