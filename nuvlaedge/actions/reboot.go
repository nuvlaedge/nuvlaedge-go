package actions

type RebootAction struct {
}

func (ra *RebootAction) Execute() error {
	return nil
}

func (ra *RebootAction) GetActionType() ActionType {
	return RebootActionType
}

func (ra *RebootAction) Init(opts ActionBaseOpts) error {
	return nil
}
