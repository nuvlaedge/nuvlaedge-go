package types

import (
	"github.com/nuvla/api-client-go/clients"
	"github.com/nuvla/api-client-go/common"
	nuvlaTypes "github.com/nuvla/api-client-go/types"
)

type LegacySession struct {
	Endpoint    string `json:"endpoint,omitempty"`
	Insecure    bool   `json:"insecure,omitempty"`
	Credentials struct {
		Key    string `json:"key"`
		Secret string `json:"secret"`
		Href   string `json:"href"`
	} `json:"credentials,omitempty"`

	NuvlaEdgeUUID       string `json:"nuvlaedge-uuid,omitempty"`
	NuvlaEdgeStatusUUID string `json:"nuvlaedge-status-uuid,omitempty"`
}

func (r *LegacySession) ConvertToNuvlaSession() *clients.NuvlaEdgeSessionFreeze {
	f := &clients.NuvlaEdgeSessionFreeze{}
	f.Credentials = &nuvlaTypes.ApiKeyLogInParams{
		Key:    r.Credentials.Key,
		Secret: r.Credentials.Secret,
	}
	f.NuvlaEdgeId = r.NuvlaEdgeUUID
	f.NuvlaEdgeStatusId = r.NuvlaEdgeStatusUUID
	f.Endpoint = r.Endpoint
	f.Insecure = r.Insecure
	return f
}

func (r *LegacySession) Load(file string) error {
	if err := common.ReadJSONFromFile(file, r); err != nil {
		return err
	}
	return nil
}
