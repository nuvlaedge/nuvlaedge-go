package main

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"nuvlaedge-go/nuvlaedge"
	"nuvlaedge-go/nuvlaedge/common"
)

func SetupRootCommand(rootCmd *cobra.Command) {
	rootCmd.SetVersionTemplate("Docker version {{.Version}}\n")

	rootCmd.PersistentFlags().Bool("help", false, "Print usage")
	rootCmd.PersistentFlags().Bool("version", false, "Print version")

	addCmdFlags(rootCmd)
	bindCmdFlagsToViper(rootCmd)
}

func addCmdFlags(cmd *cobra.Command) {
	// Basic configuration
	cmd.Flags().String("uuid", "", "NuvlaEdge ID")
	cmd.Flags().String("config-file", "", "Configuration file")
	cmd.Flags().String("data-location", "", "Data location")
	// Agent configuration
	cmd.Flags().Int("heartbeat-period", 20, "Heartbeat period")
	cmd.Flags().Int("telemetry-period", 60, "Telemetry period")
	cmd.Flags().Bool("vpn-enabled", false, "VPN enabled")
	// Endpoint
	cmd.Flags().String("endpoint", "https://nuvla.io", "Nuvla endpoint")
	cmd.Flags().Bool("insecure", false, "Insecure connection")
	cmd.MarkFlagsRequiredTogether("endpoint", "insecure")
	cmd.Flags().String("api-key", "", "NuvlaEdge API key")
	cmd.Flags().String("api-secret", "", "NuvlaEdge API secret")
	cmd.MarkFlagsRequiredTogether("api-key", "api-secret")
	// Logging
	cmd.Flags().Bool("debug", false, "Debug mode")
	cmd.Flags().String("log-level", "info", "Log level")
	cmd.Flags().Bool("log-to-file", false, "Log to file")
	cmd.Flags().String("log-file", "", "Log file")
}

func bindCmdFlagsToViper(runCmd *cobra.Command) {
	onError := func(err error) {
		if err != nil {
			log.Warn("Error binding flag to viper var: %s", err)
		}
	}
	onError(viper.BindPFlag("agent.nuvlaedge-uuid", runCmd.Flags().Lookup("uuid")))
	onError(viper.BindPFlag("config-file", runCmd.Flags().Lookup("config-file")))
	onError(viper.BindPFlag("data-location", runCmd.Flags().Lookup("data-location")))
	onError(viper.BindPFlag("agent.heartbeat-period", runCmd.Flags().Lookup("heartbeat-period")))
	onError(viper.BindPFlag("agent.telemetry-period", runCmd.Flags().Lookup("telemetry-period")))
	onError(viper.BindPFlag("agent.vpn.enabled", runCmd.Flags().Lookup("vpn-enabled")))
	onError(viper.BindPFlag("agent.nuvla-endpoint", runCmd.Flags().Lookup("endpoint")))
	onError(viper.BindPFlag("agent.nuvla-insecure", runCmd.Flags().Lookup("insecure")))
	onError(viper.BindPFlag("agent.api-key", runCmd.Flags().Lookup("api-key")))
	onError(viper.BindPFlag("agent.api-secret", runCmd.Flags().Lookup("api-secret")))
	onError(viper.BindPFlag("logging.debug", runCmd.Flags().Lookup("debug")))
	onError(viper.BindPFlag("logging.log-level", runCmd.Flags().Lookup("log-level")))
	onError(viper.BindPFlag("logging.log-to-file", runCmd.Flags().Lookup("log-to-file")))
	onError(viper.BindPFlag("logging.log-file", runCmd.Flags().Lookup("log-file")))
}

func setSettingsDefaults() {
	// NuvlaEdge Defaults
	viper.SetDefault("data-location", "/var/lib/nuvlaedge/")
	viper.SetDefault("config-file", common.DefaultConfigPath+"nuvlaedge.toml")
	viper.SetDefault("agent.job-engine-image", common.JobEngineContainerImage)
	viper.SetDefault("agent.enable-legacy-job-support", true)
	// Agent Defaults
	viper.SetDefault("agent.nuvla-endpoint", "https://nuvla.io")
	viper.SetDefault("agent.nuvla-insecure", false)
	viper.SetDefault("agent.heartbeat-period", 20)
	viper.SetDefault("agent.telemetry-period", 60)
	// Logging Defaults
	viper.SetDefault("logging.debug", false)
	viper.SetDefault("logging.log-level", "info")
	viper.SetDefault("logging.log-to-file", false)
	viper.SetDefault("logging.log-file", "/var/log/nuvlaedge.log")
	viper.SetDefault("logging.log-path", "/var/log/")
	viper.SetDefault("logging.log-max-size", 10)
	viper.SetDefault("logging.log-max-backups", 5)
}

func bindEnvs() {
	onError := func(err error) {
		if err != nil {
			log.Warn("Error binding env var: %s", err)
		}
	}
	// NuvlaEdge Defaults
	onError(viper.BindEnv("agent.nuvlaedge-uuid", "NUVLAEDGE_UUID"))
	onError(viper.BindEnv("data-location", "DATABASE_PATH", "DATA_LOCATION"))
	onError(viper.BindEnv("config-file", "NUVLAEDGE_SETTINGS"))
	onError(viper.BindEnv("agent.job-engine-image", "NUVLAEDGE_JOB_ENGINE_LITE_IMAGE", "JOB_LEGACY_IMAGE"))
	onError(viper.BindEnv("agent.enable-legacy-job-support", "JOB_LEGACY_ENABLE"))
	// Agent Defaults
	onError(viper.BindEnv("agent.nuvla-endpoint", "NUVLA_ENDPOINT"))
	onError(viper.BindEnv("agent.nuvla-insecure", "NUVLA_INSECURE"))
	onError(viper.BindEnv("agent.heartbeat-period", "HEARTBEAT_PERIOD"))
	onError(viper.BindEnv("agent.telemetry-period", "TELEMETRY_PERIOD"))
	// Logging Defaults
	onError(viper.BindEnv("logging.debug", "NUVLAEDGE_DEBUG"))
	onError(viper.BindEnv("logging.log-level", "LOG_LEVEL", "NUVLAEDGE_LOG_LEVEL"))
	onError(viper.BindEnv("logging.log-to-file", "LOG_TO_FILE"))
}

func setConfigFile() error {
	file := viper.Get("config-file")

	if file == nil || file.(string) == "" || !common.FileExists(file.(string)) {
		// No need to return an error here, we don't want the Config file as a mandatory flag if UUID is provided
		log.Info("No config file provided. Using defaults, envs and flags.")
		return errors.New("no config file provided")
	}

	viper.SetConfigFile(file.(string))
	err := viper.ReadInConfig()
	if err != nil {
		log.Errorf("Error reading config file: %s", err)
		return err
	}
	return nil
}

func GetSettings() *nuvlaedge.Settings {
	setSettingsDefaults()
	bindEnvs()

	if err := setConfigFile(); err != nil {
		log.Warnf("Error setting config file: %s", err)
	}

	settings := &nuvlaedge.Settings{}
	if err := viper.Unmarshal(settings); err != nil {
		log.Errorf("Error unmarshalling settings: %s", err)
		return nil
	}
	return settings

}
