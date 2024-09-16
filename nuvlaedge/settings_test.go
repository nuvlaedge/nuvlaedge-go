package nuvlaedge

import (
	"github.com/nuvla/api-client-go/clients"
	"github.com/nuvla/api-client-go/types"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"nuvlaedge-go/common/constants"
	"nuvlaedge-go/types/settings"
	"os"
	"path/filepath"
	"testing"
)

var tempDir string
var mockNuvlaEdgeId string
var mockNuvlaEndpoint string

func init() {
	mockNuvlaEdgeId = "nuvlabox/nuvlaedge-uuid"
	mockNuvlaEndpoint = "https://nuvla.io"

	log.SetLevel(log.PanicLevel)
}

func NewTempDir() string {
	d, _ := os.MkdirTemp("/tmp/", "settings_test_")
	tempDir = d
	return d
}

func RemoveTempDir() {
	_ = os.RemoveAll(tempDir)
}

func Test_ValidateSettings(t *testing.T) {
	NewTempDir()
	defer RemoveTempDir()

	sf := &clients.NuvlaEdgeSessionFreeze{
		NuvlaEdgeId: mockNuvlaEdgeId,
		Credentials: &types.ApiKeyLogInParams{
			Key:    "key",
			Secret: "secret",
		},
	}
	err := sf.Save(filepath.Join(tempDir, constants.NuvlaEdgeSessionFile))
	assert.NoError(t, err, "Error saving mock session file")

	set := &settings.NuvlaEdgeSettings{
		NuvlaEdgeUUID: mockNuvlaEdgeId + "_1",
		DBPPath:       tempDir,
	}

	cli, err := ValidateSettings(set)
	assert.ErrorContains(t, err, "NuvlaEndpoint is missing and required")
	assert.Nil(t, cli, "client should be nil when error is returned")

	set.NuvlaEndpoint = mockNuvlaEndpoint
	cli, err = ValidateSettings(set)
	assert.NoError(t, err, "Unexpected error validating settings")
	assert.NotNil(t, cli, "Client is nil")
}

func Test_findOldSession(t *testing.T) {
	NewTempDir()
	defer RemoveTempDir()

	f, ok := findOldSession(tempDir)
	assert.False(t, ok, "Session file should not be found")

	sf := &clients.NuvlaEdgeSessionFreeze{
		NuvlaEdgeId: mockNuvlaEdgeId,
		Credentials: &types.ApiKeyLogInParams{
			Key:    "key",
			Secret: "secret",
		},
	}
	sFile := filepath.Join(tempDir, constants.NuvlaEdgeSessionFile)
	err := sf.Save(sFile)
	assert.NoError(t, err, "Error saving mock session file")

	f, ok = findOldSession(tempDir)
	assert.True(t, ok, "Session file not found")
	assert.NotNil(t, f, "Session file is nil")
	assert.Equal(t, mockNuvlaEdgeId, f.NuvlaEdgeId, "Unexpected NuvlaEdgeId")

	_ = os.WriteFile(sFile, []byte("invalid json"), 0644)
	f, ok = findOldSession(tempDir)
	assert.False(t, ok, "Session file should not be found")
	assert.Nil(t, f, "Session file should be nil")
}

func Test_NewClientFromSettings(t *testing.T) {
	set := &settings.NuvlaEdgeSettings{
		NuvlaEdgeUUID: mockNuvlaEdgeId,
		NuvlaEndpoint: mockNuvlaEndpoint,
	}

	cli := newClientFromSettings(set)
	assert.NotNil(t, cli, "Client is nil")
	assert.Equal(t, mockNuvlaEdgeId, cli.NuvlaEdgeId.String(), "Unexpected NuvlaEdgeId")
	assert.Equal(t, mockNuvlaEndpoint, cli.SessionOpts.Endpoint, "Unexpected NuvlaEndpoint")
}

func Test_MinSettings(t *testing.T) {
	set := &settings.NuvlaEdgeSettings{
		NuvlaEdgeUUID: mockNuvlaEdgeId,
		NuvlaEndpoint: mockNuvlaEndpoint,
	}

	err := minSettings(set)
	assert.NoError(t, err, "Unexpected error validating settings")

	set.NuvlaEndpoint = ""
	err = minSettings(set)
	assert.ErrorContains(t, err, "NuvlaEndpoint is missing and required")

	set.NuvlaEndpoint = mockNuvlaEndpoint
	set.ApiKey = ""
	set.ApiSecret = ""
	set.NuvlaEdgeUUID = ""
	err = minSettings(set)
	assert.ErrorContains(t, err, "missing API KEY and SECRET or NuvlaEdge UUID to start a NuvlaEdge")
}

func Test_SanitiseUUID(t *testing.T) {
	uuid := "nuvlaedge-uuid"

	assert.Empty(t, SanitiseUUID("", "nuvlabox"), "Empty UUID should return empty string")
	assert.Equal(t, mockNuvlaEdgeId, SanitiseUUID(mockNuvlaEdgeId, "nuvlabox"), "Unexpected sanitised UUID")
	assert.Equal(t, mockNuvlaEdgeId, SanitiseUUID(uuid, "nuvlabox"), "Unexpected sanitised UUID")
	assert.Empty(t, SanitiseUUID("session/id", "nuvlabox"), "Invalid UUID should return empty string")
}
