package errors

// ActionRequiresSudoError is returned when an action requires sudo and can't be executed.
type ActionRequiresSudoError struct {
	actionName string
}

func NewActionRequiresSudoError(actionName string) ActionRequiresSudoError {
	return ActionRequiresSudoError{actionName: actionName}
}

func (as ActionRequiresSudoError) Error() string {
	return "action " + as.actionName + " requires sudo"
}

// NotImplementedActionError is returned when an action is not implemented and should be used to assert
// whether to run the action from the python container
type NotImplementedActionError struct {
	actionName string
}

func (na NotImplementedActionError) Error() string {
	return "action " + na.actionName + " is not implemented"
}

func NewNotImplementedActionError(actionName string) NotImplementedActionError {
	return NotImplementedActionError{actionName: actionName}
}
