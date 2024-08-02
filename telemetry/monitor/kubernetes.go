//go:build !coverage

package monitor

import (
	"context"
)

type KubernetesMonitor struct {
	BaseMonitor
}

func (km *KubernetesMonitor) Run(ctx context.Context) error {
	return nil
}
