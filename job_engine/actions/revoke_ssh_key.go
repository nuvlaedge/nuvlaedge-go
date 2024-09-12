package actions

import "context"

type RevokeSshKeyAction struct{}

func (r *RevokeSshKeyAction) Execute(context.Context) error {
	return r.RevokeSshKey()
}

func (r *RevokeSshKeyAction) Init(opts *ActionOpts) error {
	//TODO implement me
	panic("implement me")
}

func (r *RevokeSshKeyAction) RevokeSshKey() error {
	return nil
}

func (r *RevokeSshKeyAction) revokeSshKeyHost() error {
	return nil
}

func (r *RevokeSshKeyAction) revokeSshKeyDocker() error {
	return nil
}

func (r *RevokeSshKeyAction) revokeSshKeyHelm() error {
	return nil
}
