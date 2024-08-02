package jobs

import "nuvlaedge-go/jobs/actions"

type NativeJob interface {
	actions.JobBase
	Execute() error
}
