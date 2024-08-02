package types

import (
	"context"
	"time"
)

type WorkerType string

const (
	Telemetry    WorkerType = "telemetry"
	Commissioner WorkerType = "commissioner"
	Deployments  WorkerType = "deployments"
	JobProcessor WorkerType = "job-processor"
	Orchestrator WorkerType = "orchestrator"
	Engine       WorkerType = "engine"
	Agent        WorkerType = "agent"
)

type Worker interface {
	Run(ctx context.Context) error
	GetPeriod() int
	SetPeriod(period int)
	Stop()
	GetStatus() WorkerStatusReport
	GetType() WorkerType
}

type WorkerBase struct {
	wType      WorkerType
	Period     int
	BaseTicker *time.Ticker
	Status     WorkerStatus
}

func NewWorkerBase(period int, workerType WorkerType) WorkerBase {
	return WorkerBase{
		wType:      workerType,
		Period:     period,
		BaseTicker: time.NewTicker(time.Duration(period) * time.Second),
		Status:     NEW,
	}
}

func (w *WorkerBase) GetPeriod() int {
	return w.Period
}

func (w *WorkerBase) SetPeriod(period int) {
	w.Period = period
	w.BaseTicker.Reset(time.Duration(w.Period) * time.Second)
}

func (w *WorkerBase) GetStatus() WorkerStatusReport {
	return WorkerStatusReport{
		Type:   w.wType,
		Status: w.Status,
	}
}

func (w *WorkerBase) GetType() WorkerType {
	return w.wType
}

func (w *WorkerBase) Stop() {
	w.Status = STOPPED
}

type WorkerStatusReport struct {
	Type   WorkerType
	Status WorkerStatus
	Errors []error
}

type WorkerStatus string

const (
	NEW      WorkerStatus = "new"
	STARTING WorkerStatus = "starting"
	RUNNING  WorkerStatus = "running"
	STOPPED  WorkerStatus = "stopped"
	FAILING  WorkerStatus = "failing"
	FAILED   WorkerStatus = "failed"
)
