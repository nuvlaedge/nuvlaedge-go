package common

import (
	"context"
	"sync"
)

type RunFunction func()

type NuvlaEdgeRoutine struct {
	running bool

	runFunction RunFunction

	mutex  *sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
	once   *sync.Once
}

func NewNuvlaEdgeRoutine(runFunction RunFunction) *NuvlaEdgeRoutine {
	ctx, cancel := context.WithCancel(context.Background())
	return &NuvlaEdgeRoutine{
		runFunction: runFunction,
		mutex:       &sync.Mutex{},
		cancel:      cancel,
		once:        &sync.Once{},
		ctx:         ctx,
	}
}

func (r *NuvlaEdgeRoutine) Start() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.running {
		return
	}
	r.once.Do(r.runFunction)
}
