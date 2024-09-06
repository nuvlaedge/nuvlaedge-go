package flags

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"nuvlaedge-go/types/options/command"
	"testing"
)

func init() {
	log.SetLevel(log.PanicLevel)
}

func TestAddRegisterFlags(t *testing.T) {
	cmd := cobra.Command{}
	flags := cmd.Flags()
	AddRegisterFlags(&cmd)

	err := flags.Parse([]string{
		"--endpoint", "test",
		"--insecure",
		"--key", "test-key",
		"--secret", "test-secret",
		"--uuid", "test-uuid",
		"--vpn",
		"--name", "test-name",
		"--description", "test-description",
		"--refresh-interval", "59",
		"--heartbeat-interval", "13",
		"--tags", "tag1",
		"--tags", "tag2",
	})

	assert.NoError(t, err)
	assert.Equal(t, "test", flags.Lookup("endpoint").Value.String())
	assert.Equal(t, "true", flags.Lookup("insecure").Value.String())
	assert.Equal(t, "test-key", flags.Lookup("key").Value.String())
	assert.Equal(t, "test-secret", flags.Lookup("secret").Value.String())
	assert.Equal(t, "test-uuid", flags.Lookup("uuid").Value.String())
	assert.Equal(t, "true", flags.Lookup("vpn").Value.String())
	assert.Equal(t, "test-name", flags.Lookup("name").Value.String())
	assert.Equal(t, "test-description", flags.Lookup("description").Value.String())
	assert.Equal(t, "59", flags.Lookup("refresh-interval").Value.String())
	assert.Equal(t, "13", flags.Lookup("heartbeat-interval").Value.String())
}

func TestSetDefaultRegisterFlags(t *testing.T) {
	setDefaultRegisterFlags()
	assert.Equal(t, "https://nuvla.io", viper.GetString("endpoint"))
	assert.Equal(t, false, viper.GetBool("insecure"))
	assert.Equal(t, 60, viper.GetInt("refresh-interval"))
	assert.Equal(t, 20, viper.GetInt("heartbeat-interval"))
	assert.Equal(t, false, viper.GetBool("vpn"))
}

func TestBindViperRegisterFlags(t *testing.T) {
	cmd := cobra.Command{}
	flags := cmd.Flags()
	AddRegisterFlags(&cmd)
	setDefaultRegisterFlags()
	bindViperRegisterFlags(flags)
	defer viper.Reset()

	err := flags.Parse([]string{
		"--endpoint", "test",
		"--insecure",
		"--key", "test-key",
		"--secret", "test-secret",
		"--uuid", "test-uuid",
		"--vpn",
		"--name", "test-name",
		"--description", "test-description",
		"--refresh-interval", "59",
		"--heartbeat-interval", "13",
		"--tags=tag1,tag2",
	})
	fmt.Printf("Tags, %v\n", flags.Lookup("tags").Value.String())
	assert.NoError(t, err)
	assert.Equal(t, "test", viper.GetString("endpoint"))
	assert.Equal(t, true, viper.GetBool("insecure"))
	assert.Equal(t, "test-key", viper.GetString("key"))
	assert.Equal(t, "test-secret", viper.GetString("secret"))
	assert.Equal(t, "test-uuid", viper.GetString("uuid"))
	assert.Equal(t, true, viper.GetBool("vpn"))
	assert.Equal(t, "test-name", viper.GetString("name"))
	assert.Equal(t, "test-description", viper.GetString("description"))
	assert.Equal(t, 59, viper.GetInt("refresh-interval"))
	assert.Equal(t, 13, viper.GetInt("heartbeat-interval"))
	assert.Equal(t, []string{"tag1,tag2"}, viper.GetStringSlice("tags"))

}

var regEnvs = map[string]string{
	"NUVLA_ENDPOINT":               "test",
	"NUVLA_INSECURE":               "true",
	"NUVLA_API_KEY":                "test-key",
	"NUVLA_API_SECRET":             "test-secret",
	"NUVLAEDGE_UUID":               "test-uuid",
	"NUVLAEDGE_VPN":                "true",
	"NUVLAEDGE_NAME":               "test-name",
	"NUVLAEDGE_DESCRIPTION":        "test-description",
	"NUVLAEDGE_NAME_PREFIX":        "test-name-prefix",
	"NUVLAEDGE_REFRESH_INTERVAL":   "59",
	"NUVLAEDGE_HEARTBEAT_INTERVAL": "13",
	"NUVLAEDGE_TAGS":               "tag1,tag2",
}

func TestSetRegisterEnvBindings(t *testing.T) {
	setRegisterEnvBindings()
	setEnvs(regEnvs)
	defer unSetEnvs(regEnvs)

	assert.Equal(t, "test", viper.GetString("endpoint"))
	assert.Equal(t, true, viper.GetBool("insecure"))
	assert.Equal(t, "test-key", viper.GetString("key"))
	assert.Equal(t, "test-secret", viper.GetString("secret"))
	assert.Equal(t, "test-uuid", viper.GetString("uuid"))
	assert.Equal(t, true, viper.GetBool("vpn"))
	assert.Equal(t, "test-name", viper.GetString("name"))
	assert.Equal(t, "test-description", viper.GetString("description"))
	assert.Equal(t, "test-name-prefix", viper.GetString("name-prefix"))
	assert.Equal(t, 59, viper.GetInt("refresh-interval"))
	assert.Equal(t, 13, viper.GetInt("heartbeat-interval"))
	assert.Equal(t, "tag1,tag2", viper.GetString("tags"))
}

func TestParseRegisterFlags(t *testing.T) {
	cmd := cobra.Command{}
	flags := cmd.Flags()
	AddRegisterFlags(&cmd)
	opts := &command.RegisterCmdOptions{}

	err := flags.Parse([]string{
		"--endpoint", "test",
		"--insecure",
		"--uuid", "test-uuid",
		"--vpn",
		"--name", "test-name",
		"--description", "test-description",
		"--refresh-interval", "59",
		"--heartbeat-interval", "13",
		"--tags", "tag1",
		"--tags", "tag2",
	})
	assert.NoError(t, err)

	err = ParseRegisterFlags(&cmd, opts)
	assert.NoError(t, err)
	assert.Equal(t, "test", opts.Endpoint)
	assert.Equal(t, true, opts.Insecure)
	assert.Equal(t, "test-uuid", opts.NuvlaEdgeUUID)
	assert.Equal(t, true, opts.VPNEnabled)
	assert.Equal(t, "test-name", opts.Name)
	assert.Equal(t, "test-description", opts.Description)
	assert.Equal(t, 59, opts.RefreshInterval)
	assert.Equal(t, 13, opts.HeartbeatInterval)
	assert.Equal(t, []string{"tag1", "tag2"}, opts.Tags)
}
