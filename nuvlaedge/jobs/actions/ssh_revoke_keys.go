package actions

import (
	nuvla "github.com/nuvla/api-client-go"
	"github.com/nuvla/api-client-go/clients/resources"
)

type SSHRevokeKeys struct {
}

func (a *SSHRevokeKeys) Init(jobResource *resources.JobResource, client *nuvla.NuvlaClient) error {
	return nil
}
