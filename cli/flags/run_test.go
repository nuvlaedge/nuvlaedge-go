package flags

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"nuvlaedge-go/common/constants"
	"nuvlaedge-go/types/settings"
	"os"
	"testing"
)

func init() {
	// Set log level to panic to avoid logs during tests
	log.SetLevel(log.PanicLevel)
}

func TestAddRunFlags(t *testing.T) {
	flags := &pflag.FlagSet{}
	AddRunFlags(flags)

	err := flags.Parse([]string{
		"--db-path", "test",
		"--nuvla-endpoint", "test",
		"--nuvla-insecure",
		"--uuid", "test_uuid",
		"--api-key", "test",
		"--api-secret", "test",
		"--heartbeat-period", "1",
		"--telemetry-period", "1",
		"--remote-sync-period", "1",
		"--vpn-enabled",
		"--vpn-extra-config", "test",
		"--job-image", "test",
		"--enable-legacy-job",
		"--log-level", "debug",
		"--debug",
	})

	assert.NoError(t, err)
	assert.Equal(t, "test", flags.Lookup("db-path").Value.String())
	assert.Equal(t, "test", flags.Lookup("nuvla-endpoint").Value.String())
	assert.Equal(t, "true", flags.Lookup("nuvla-insecure").Value.String())
	assert.Equal(t, "test_uuid", flags.Lookup("uuid").Value.String())
	assert.Equal(t, "test", flags.Lookup("api-key").Value.String())
	assert.Equal(t, "test", flags.Lookup("api-secret").Value.String())
	assert.Equal(t, "1", flags.Lookup("heartbeat-period").Value.String())
	assert.Equal(t, "1", flags.Lookup("telemetry-period").Value.String())
	assert.Equal(t, "1", flags.Lookup("remote-sync-period").Value.String())
	assert.Equal(t, "true", flags.Lookup("vpn-enabled").Value.String())
	assert.Equal(t, "test", flags.Lookup("job-image").Value.String())
	assert.Equal(t, "true", flags.Lookup("enable-legacy-job").Value.String())
	assert.Equal(t, "debug", flags.Lookup("log-level").Value.String())
	assert.Equal(t, "true", flags.Lookup("debug").Value.String())
}

func TestSetDefaultRunFlags(t *testing.T) {
	setDefaultRunFlags()
	defer viper.Reset()
	assert.Equal(t, constants.DefaultDBPath, viper.GetString("db-path"))
	assert.Equal(t, constants.DefaultEndPoint, viper.GetString("nuvla-endpoint"))
	assert.Equal(t, constants.DefaultInsecure, viper.GetBool("nuvla-insecure"))
	assert.Equal(t, constants.DefaultHeartbeatPeriod, viper.GetInt("heartbeat-period"))
	assert.Equal(t, constants.DefaultTelemetryPeriod, viper.GetInt("telemetry-period"))
	assert.Equal(t, constants.DefaultRemoteSyncPeriod, viper.GetInt("remote-sync-period"))
	assert.Equal(t, constants.DefaultVPNEnabled, viper.GetBool("vpn-enabled"))
	assert.Equal(t, constants.DefaultJobEngineImage, viper.GetString("job-engine-image"))
	assert.Equal(t, constants.DefaultEnableLegacyJob, viper.GetBool("enable-legacy-job"))
	assert.Equal(t, constants.DefaultLogLevel, viper.GetString("log-level"))
	assert.Equal(t, constants.DefaultDebug, viper.GetBool("debug"))
}

func TestBindViperRunFlags(t *testing.T) {
	flags := &pflag.FlagSet{}
	AddRunFlags(flags)
	setDefaultRunFlags()
	bindViperRunFlags(flags)
	defer viper.Reset()
	err := flags.Parse([]string{
		"--db-path", "test",
		"--nuvla-endpoint", "test",
		"--nuvla-insecure",
		"--uuid", "test_uuid",
		"--api-key", "test",
		"--api-secret", "test",
		"--heartbeat-period", "1",
		"--telemetry-period", "1",
		"--remote-sync-period", "1",
		"--vpn-enabled",
		"--vpn-extra-config", "test",
		"--job-image", "test",
		"--enable-legacy-job",
		"--log-level", "warn",
		"--debug",
	})

	assert.NoError(t, err)
	assert.Equal(t, "test", viper.GetString("db-path"))
	assert.Equal(t, "test", viper.GetString("nuvla-endpoint"))
	assert.Equal(t, true, viper.GetBool("nuvla-insecure"))
	assert.Equal(t, "test_uuid", viper.GetString("nuvlaedge-uuid"))
	assert.Equal(t, "test", viper.GetString("api-key"))
	assert.Equal(t, "test", viper.GetString("api-secret"))
	assert.Equal(t, 1, viper.GetInt("heartbeat-period"))
	assert.Equal(t, 1, viper.GetInt("telemetry-period"))
	assert.Equal(t, 1, viper.GetInt("remote-sync-period"))
	assert.Equal(t, true, viper.GetBool("vpn-enabled"))
	assert.Equal(t, "test", viper.GetString("job-engine-image"))
	assert.Equal(t, true, viper.GetBool("enable-legacy-job"))
	assert.Equal(t, "warn", viper.GetString("log-level"))
	assert.Equal(t, true, viper.GetBool("debug"))
}

var envs = map[string]string{
	"DB_PATH":              "test",
	"NUVLA_ENDPOINT":       "test",
	"NUVLA_INSECURE":       "true",
	"NUVLAEDGE_UUID":       "test_uuid",
	"NUVLAEDGE_API_KEY":    "test",
	"NUVLAEDGE_API_SECRET": "test",
	"HEARTBEAT_PERIOD":     "1",
	"TELEMETRY_PERIOD":     "1",
	"REMOTE_SYNC_PERIOD":   "1",
	"VPN_ENABLED":          "true",
	"VPN_EXTRA_CONFIG":     "test",
	"JOB_LEGACY_IMAGE":     "test",
	"ENABLE_LEGACY_JOB":    "true",
	"NUVLAEDGE_LOG_LEVEL":  "error",
	"DEBUG":                "true",
}

func setEnvs(e map[string]string) {
	for k, v := range e {
		_ = os.Setenv(k, v)
	}
}

func unSetEnvs(e map[string]string) {
	for k := range e {
		_ = os.Unsetenv(k)
	}
}

func TestSetEnvBindings(t *testing.T) {
	setEnvBindings()
	// Export Envs
	setEnvs(envs)
	defer unSetEnvs(envs)
	defer viper.Reset()
	viper.AutomaticEnv()

	assert.Equal(t, "test", viper.GetString("db-path"))
	assert.Equal(t, "test", viper.GetString("nuvla-endpoint"))
	assert.Equal(t, true, viper.GetBool("nuvla-insecure"))
	assert.Equal(t, "test_uuid", viper.GetString("nuvlaedge-uuid"))
	assert.Equal(t, "test", viper.GetString("api-key"))
	assert.Equal(t, "test", viper.GetString("api-secret"))
	assert.Equal(t, 1, viper.GetInt("heartbeat-period"))
	assert.Equal(t, 1, viper.GetInt("telemetry-period"))
	assert.Equal(t, 1, viper.GetInt("remote-sync-period"))
	assert.Equal(t, true, viper.GetBool("vpn-enabled"))
	assert.Equal(t, "test", viper.GetString("vpn-extra-config"))
	assert.Equal(t, "test", viper.GetString("job-engine-image"))
	assert.Equal(t, true, viper.GetBool("enable-legacy-job"))
	assert.Equal(t, "error", viper.GetString("log-level"))
	assert.Equal(t, true, viper.GetBool("debug"))
}

func TestParseSettings(t *testing.T) {
	// Export Envs
	setEnvs(envs)
	defer unSetEnvs(envs)
	defer viper.Reset()
	flags := &pflag.FlagSet{}
	AddRunFlags(flags)

	set := &settings.NuvlaEdgeSettings{}
	err := ParseSettings(flags, set)
	assert.NoError(t, err)
	assert.Equal(t, "test", set.DBPPath)
	assert.Equal(t, "test", set.NuvlaEndpoint)
	assert.Equal(t, true, set.NuvlaInsecure)
	assert.Equal(t, "test_uuid", set.NuvlaEdgeUUID)
	assert.Equal(t, "test", set.ApiKey)
	assert.Equal(t, "test", set.ApiSecret)
	assert.Equal(t, 1, set.HeartbeatPeriod)
	assert.Equal(t, 1, set.TelemetryPeriod)
	assert.Equal(t, 1, set.RemoteSyncPeriod)
	assert.Equal(t, true, set.VpnEnabled)
	assert.Equal(t, "test", set.VpnExtraConfig)
	assert.Equal(t, "test", set.JobEngineImage)
	assert.Equal(t, true, set.EnableJobLegacySupport)
	assert.Equal(t, "error", set.LogLevel)
	assert.Equal(t, true, set.Debug)

	err = ParseSettings(flags, nil)
	assert.NotNil(t, err, "Expected error")
}
