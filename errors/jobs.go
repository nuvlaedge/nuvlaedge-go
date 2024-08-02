package errors

import "nuvlaedge-go/types"

// NotImplementedJobError is returned when an action is not implemented and should be used to assert
// whether to run the action from the python container
type NotImplementedJobError struct {
	actionName types.JobType
}

func (na NotImplementedJobError) Error() string {
	return "job " + string(na.actionName) + " is not implemented"
}

func NewNotImplementedActionError(actionName string) NotImplementedJobError {
	return NotImplementedJobError{actionName: types.JobType(actionName)}
}
