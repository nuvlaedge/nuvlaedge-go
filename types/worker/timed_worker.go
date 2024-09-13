package worker

import (
	"sync"
	"time"
)

type TimedWorker struct {
	WorkerBase

	period     int
	BaseTicker *time.Ticker
	mu         sync.Mutex
}

func NewTimedWorker(period int, wType WorkerType) TimedWorker {
	return TimedWorker{
		WorkerBase: NewWorkerBase(wType),
		period:     period,
		BaseTicker: time.NewTicker(time.Duration(period) * time.Second),
		mu:         sync.Mutex{},
	}
}

func (w *TimedWorker) GetPeriod() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.period
}

func (w *TimedWorker) SetPeriod(period int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.period = period
	w.BaseTicker.Reset(time.Duration(w.period) * time.Second)
}
