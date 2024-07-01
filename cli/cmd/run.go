package cmd

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"nuvlaedge-go/nuvlaedge"
	"nuvlaedge-go/nuvlaedge/common"
	"nuvlaedge-go/nuvlaedge/types"
)

// After adding Viper, runFlags are no longer needed. Keep them for now...
var runFlags types.RunFlags

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs NuvlaEdge on the local machine",
	Run: func(cmd *cobra.Command, args []string) {
		s, err := nuvlaedge.GetSettings()
		if err != nil {
			log.Fatalf("Error producing NuvlaEdge settings: %s", err)
		}

		// Initialises the global logging system. ATM only using default logger
		common.InitLogging(s.Logging)

		// Main loop function, This should never return unless an error is encountered
		err = runNuvlaEdge(s)
		if err != nil {
			log.Fatalf("Error running NuvlaEdge: %s", err)
		}
	},
}

func init() {
	log.Info("Initializing run command")
	rootCmd.AddCommand(runCmd)
	runFlags = types.RunFlags{}

	runCmd.Flags().StringVar(&runFlags.Uuid, "uuid", "", "NuvlaEdge ID")
	_ = viper.BindPFlag("agent.nuvlaedge-uuid", runCmd.Flags().Lookup("uuid"))

	runCmd.Flags().StringVar(&runFlags.ConfigFile, "config-file", "", "Configuration file")
	_ = viper.BindPFlag("config-file", runCmd.Flags().Lookup("config-file"))

	//runCmd.MarkFlagsOneRequired("uuid", "config-file")

	runCmd.Flags().StringVar(&runFlags.DataLocation, "data-location", "", "Data location")
	runCmd.Flags().IntVar(&runFlags.HeartbeatPeriod, "heartbeat-period", 20, "Heartbeat period")
	runCmd.Flags().IntVar(&runFlags.TelemetryPeriod, "telemetry-period", 60, "Telemetry period")
	runCmd.Flags().BoolVar(&runFlags.VPNEnabled, "vpn-enabled", false, "VPN enabled")
	_ = viper.BindPFlag("data-location", runCmd.Flags().Lookup("data-location"))
	_ = viper.BindPFlag("agent.heartbeat-period", runCmd.Flags().Lookup("heartbeat-period"))
	_ = viper.BindPFlag("agent.telemetry-period", runCmd.Flags().Lookup("telemetry-period"))
	_ = viper.BindPFlag("agent.vpn.enabled", runCmd.Flags().Lookup("vpn-enabled"))

	runCmd.Flags().StringVar(&runFlags.NuvlaEndpoint, "endpoint", "https://nuvla.io", "Nuvla endpoint")
	runCmd.Flags().BoolVar(&runFlags.NuvlaInsecure, "insecure", false, "Insecure connection")
	runCmd.MarkFlagsRequiredTogether("endpoint", "insecure")
	_ = viper.BindPFlag("agent.nuvla-endpoint", runCmd.Flags().Lookup("endpoint"))
	_ = viper.BindPFlag("agent.nuvla-insecure", runCmd.Flags().Lookup("insecure"))

	runCmd.Flags().StringVar(&runFlags.ApiKey, "api-key", "", "NuvlaEdge API key")
	runCmd.Flags().StringVar(&runFlags.ApiSecret, "api-secret", "", "NuvlaEdge API secret")
	runCmd.MarkFlagsRequiredTogether("api-key", "api-secret")
	_ = viper.BindPFlag("agent.api-key", runCmd.Flags().Lookup("api-key"))
	_ = viper.BindPFlag("agent.api-secret", runCmd.Flags().Lookup("api-secret"))

	runCmd.Flags().BoolVar(&runFlags.Debug, "debug", false, "Debug mode")
	runCmd.Flags().StringVar(&runFlags.LogLevel, "log-level", "info", "Log level")
	runCmd.Flags().BoolVar(&runFlags.LogToFile, "log-to-file", false, "Log to file")
	runCmd.Flags().StringVar(&runFlags.LogFile, "log-file", "", "Log file")
	_ = viper.BindPFlag("logging.debug", runCmd.Flags().Lookup("debug"))
	_ = viper.BindPFlag("logging.log-level", runCmd.Flags().Lookup("log-level"))
	_ = viper.BindPFlag("logging.log-to-file", runCmd.Flags().Lookup("log-to-file"))
	_ = viper.BindPFlag("logging.log-file", runCmd.Flags().Lookup("log-file"))
}

func runNuvlaEdge(settings *nuvlaedge.Settings) error {
	ne := nuvlaedge.NewNuvlaEdge(settings)

	log.Infof("Initializing NuvlaEdge...")
	if err := ne.Start(); err != nil {
		return err
	}
	log.Infof("NuvlaEdge started successfully")

	err := ne.Run()
	if err != nil {
		log.Errorf("Error running NuvlaEdge: %s", err)
		return err
	}

	return errors.New("NuvlaEdge run exited unexpectedly")
}
