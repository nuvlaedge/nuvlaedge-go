package actions

import "context"

type RebootAction struct {
}

func (r *RebootAction) Execute(context.Context) error {
	return r.Reboot()
}

func (r *RebootAction) Init(opts *ActionOpts) error {
	//TODO implement me
	panic("implement me")
}

func (r *RebootAction) Reboot() error {
	return nil
}

func (r *RebootAction) rebootHost() error {
	return nil
}

func (r *RebootAction) rebootDocker() error {
	return nil
}

func (r *RebootAction) rebootKubernetes() error {
	return nil
}
