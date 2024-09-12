package actions

import "context"

type AddSshKeyAction struct{}

func (a *AddSshKeyAction) Execute(context.Context) error {
	return a.AddSshKey()
}

func (a *AddSshKeyAction) Init(opts *ActionOpts) error {
	//TODO implement me
	panic("implement me")
}

func (a *AddSshKeyAction) AddSshKey() error {
	return nil
}

func (a *AddSshKeyAction) addSshKeyHost() error {
	return nil
}

func (a *AddSshKeyAction) addSshKeyDocker() error {
	return nil
}

func (a *AddSshKeyAction) addSshKeyHelm() error {
	return nil
}
