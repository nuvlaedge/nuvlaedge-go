package nuvlaedge

import (
	"errors"
	"fmt"
	nuvlaApi "github.com/nuvla/api-client-go"
	"github.com/nuvla/api-client-go/clients"
	"github.com/nuvla/api-client-go/common"
	nuvlaTypes "github.com/nuvla/api-client-go/types"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/common/constants"
	"nuvlaedge-go/types/settings"
	"path/filepath"
	"strings"
)

// ValidateSettings validates the settings and returns a NuvlaEdge client
func ValidateSettings(settings *settings.NuvlaEdgeSettings) (*clients.NuvlaEdgeClient, error) {
	oldSession, ok := findOldSession(settings.DBPPath)
	if ok {
		log.Infof("Found stored NuvlaEdge session")
		mergeSessionIntoSettings(settings, oldSession)
	}

	if err := minSettings(settings); err != nil {
		return nil, err
	}
	return newClientFromSettings(settings), nil
}

// findOldSession finds the old session
func findOldSession(dbpPath string) (*clients.NuvlaEdgeSessionFreeze, bool) {
	sessionFile := filepath.Join(dbpPath, constants.NuvlaEdgeSessionFile)
	if !common.FileExists(sessionFile) {
		return nil, false
	}

	f := &clients.NuvlaEdgeSessionFreeze{}
	if err := f.Load(sessionFile); err != nil {
		log.Errorf("Error loading session file: %s", err)
		return nil, false
	}

	return f, true
}

func mergeSessionIntoSettings(settings *settings.NuvlaEdgeSettings, session *clients.NuvlaEdgeSessionFreeze) {
	sessionId := SanitiseUUID(session.NuvlaEdgeId, "nuvlabox")
	settId := SanitiseUUID(settings.NuvlaEdgeUUID, "nuvlabox")
	if settId != sessionId {
		log.Warnf("NuvlaEdge UUID in settings (%s) is different from stored session (%s)", settId, sessionId)
		log.Warnf("NuvlaEdge will try to use the stored session, if you are trying to start a new NuvlaEdge, " +
			"please remove the stored session")
	}

	settings.NuvlaEdgeUUID = sessionId

	if session.Credentials != nil {
		settings.ApiKey = session.Credentials.Key
		settings.ApiSecret = session.Credentials.Secret
	}
}

// newClientFromSettings creates a new NuvlaEdge client from the settings. Settings must be validated before calling this function
func newClientFromSettings(settings *settings.NuvlaEdgeSettings) *clients.NuvlaEdgeClient {
	var creds *nuvlaTypes.ApiKeyLogInParams

	if isRestoreNuvlaEdge(settings) {
		creds = &nuvlaTypes.ApiKeyLogInParams{
			Key:    settings.ApiKey,
			Secret: settings.ApiSecret,
		}
	} else {
		creds = nil
	}

	cli := clients.NewNuvlaEdgeClient(
		settings.NuvlaEdgeUUID,
		creds,
		nuvlaApi.WithEndpoint(settings.NuvlaEndpoint),
		nuvlaApi.WithInsecureSession(settings.NuvlaInsecure),
		nuvlaApi.WithoutPersistCookie,
		nuvlaApi.ReAuthenticateSession)

	return cli
}

func isRestoreNuvlaEdge(settings *settings.NuvlaEdgeSettings) bool {
	return settings.ApiKey != "" && settings.ApiSecret != ""
}

func minSettings(settings *settings.NuvlaEdgeSettings) error {
	if settings.NuvlaEndpoint == "" {
		return errors.New("NuvlaEndpoint is missing and required")
	}

	if (settings.ApiKey == "" || settings.ApiSecret == "") && settings.NuvlaEdgeUUID == "" {
		return errors.New("missing API KEY and SECRET or NuvlaEdge UUID to start a NuvlaEdge")
	}

	if settings.ApiKey != "" && settings.ApiSecret != "" && settings.NuvlaEdgeUUID == "" {
		remoteId, err := getNuvlaEdgeIdFromApiKeys(settings)
		if err != nil {
			return err
		}
		settings.NuvlaEdgeUUID = remoteId
	}

	if settings.NuvlaEdgeUUID == "" {
		return errors.New("missing NuvlaEdge UUID, cannot start NuvlaEdge")
	}

	settings.NuvlaEdgeUUID = SanitiseUUID(settings.NuvlaEdgeUUID, "nuvlabox")
	return nil
}

// SanitiseUUID returns the resource ID. If UUID starts with the resource name, means we already have the full ID.
// Else, we need to add the resource name to the UUID.
func SanitiseUUID(uuid, resourceName string) string {
	if uuid == "" {
		log.Debugf("UUID for resource %s is empty", resourceName)
		return uuid
	}

	if strings.HasPrefix(uuid, resourceName) {
		return uuid
	}

	if strings.Contains(uuid, "/") {
		s := strings.Split(uuid, "/")
		if len(s) == 2 {
			log.Infof("UUID (%s) belongs to resource %s, not %s", s[0], s[1], resourceName)
		}
		return ""
	}

	return fmt.Sprintf("%s/%s", resourceName, uuid)
}

func getNuvlaEdgeIdFromApiKeys(settings *settings.NuvlaEdgeSettings) (string, error) {
	sOpts := nuvlaApi.DefaultSessionOpts()
	sOpts.Endpoint = settings.NuvlaEndpoint
	sOpts.Insecure = settings.NuvlaInsecure

	cli := nuvlaApi.NewNuvlaClient(
		&nuvlaTypes.ApiKeyLogInParams{
			Key:    settings.ApiKey,
			Secret: settings.ApiSecret,
		},
		sOpts)

	// Get the NuvlaEdge ID
	col, err := cli.Search("session", nil)
	if err != nil {
		return "", err
	}
	ErrNotFoundSession := errors.New("no session found")
	if len(col.Resources) == 0 {
		return "", ErrNotFoundSession
	}

	s, ok := col.Resources[0]["id"]
	if !ok {
		return "", ErrNotFoundSession
	}

	res, err := cli.Get(s.(string), nil)
	if err != nil {
		return "", ErrNotFoundSession
	}

	return res.Id, nil
}
