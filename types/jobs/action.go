package jobs

import "context"

type JobAction func(ctx context.Context, opts *JobOpts) error
