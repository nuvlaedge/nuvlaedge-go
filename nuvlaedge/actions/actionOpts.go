package actions

import (
	"github.com/nuvla/api-client-go/clients"
	nuvlaTypes "github.com/nuvla/api-client-go/types"
	"nuvlaedge-go/nuvlaedge/orchestrator"
)

type ActionBaseOptsFunc func(*ActionBaseOpts)

type ActionBaseOpts struct {
	NuvlaClient *clients.NuvlaEdgeClient
	CoeClient   orchestrator.Coe
	credentials nuvlaTypes.LogInParams
}

func DefaultActionBaseOpts() *ActionBaseOpts {
	return &ActionBaseOpts{}
}

func WithNuvlaClient(nuvlaClient *clients.NuvlaEdgeClient) ActionBaseOptsFunc {
	return func(opts *ActionBaseOpts) {
		opts.NuvlaClient = nuvlaClient
	}
}

func WithCoeClient(coeClient orchestrator.Coe) ActionBaseOptsFunc {
	return func(opts *ActionBaseOpts) {
		opts.CoeClient = coeClient
	}
}

func WithCredentials(creds nuvlaTypes.LogInParams) ActionBaseOptsFunc {
	return func(opts *ActionBaseOpts) {
		opts.credentials = creds
	}
}
