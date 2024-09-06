package jobs

import "context"

type JobExecutor interface {
	Reboot(ctx context.Context) error
	AddSSHKey(ctx context.Context) error
	RevokeSSHKey(ctx context.Context) error
	Update(ctx context.Context) error
	LogFetch(ctx context.Context) error
}

type KubernetesExecutor struct{}

type HostExecutor struct{}
