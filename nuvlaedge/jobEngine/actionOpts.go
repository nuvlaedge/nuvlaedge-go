package jobEngine

import (
	nuvla "github.com/nuvla/api-client-go"
	"github.com/nuvla/api-client-go/clients"
	"nuvlaedge-go/nuvlaedge/orchestrator"
)

type ActionBaseOptsFunc func(*ActionBaseOpts)

type ActionBaseOpts struct {
	NuvlaClient *nuvla.NuvlaClient
	CoeClient   orchestrator.Coe
	jobResource *clients.JobResource
}

func DefaultActionBaseOpts() *ActionBaseOpts {
	return &ActionBaseOpts{}
}

func WithNuvlaClient(nuvlaClient *nuvla.NuvlaClient) ActionBaseOptsFunc {
	return func(opts *ActionBaseOpts) {
		opts.NuvlaClient = nuvlaClient
	}
}

func WithCoeClient(coeClient orchestrator.Coe) ActionBaseOptsFunc {
	return func(opts *ActionBaseOpts) {
		opts.CoeClient = coeClient
	}
}

func WithJobResource(jobResource *clients.JobResource) ActionBaseOptsFunc {
	return func(opts *ActionBaseOpts) {
		opts.jobResource = jobResource
	}
}
