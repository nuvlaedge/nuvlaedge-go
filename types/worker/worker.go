package worker

import (
	"context"
)

type WorkerType string

const (
	Telemetry       WorkerType = "telemetry"
	Commissioner    WorkerType = "commissioner"
	Deployments     WorkerType = "deployments"
	JobProcessor    WorkerType = "job-processor"
	Heartbeat       WorkerType = "heartbeat"
	ConfUpdater     WorkerType = "conf-updater"
	ResourceCleaner WorkerType = "resource-cleaner"
)

type Worker interface {
	Init(opts *WorkerOpts, conf *WorkerConfig) error
	Start(ctx context.Context) error
	Reconfigure(conf *WorkerConfig) error
	Run(ctx context.Context) error
	Stop(ctx context.Context) error
	GetName() string
	GetConfChannel() chan *WorkerConfig
}

type WorkerBase struct {
	ConfChan   chan *WorkerConfig
	workerType WorkerType
}

func (wb *WorkerBase) GetName() string {
	return string(wb.workerType)
}

func (wb *WorkerBase) GetConfChannel() chan *WorkerConfig {
	return wb.ConfChan
}

func NewWorkerBase(wType WorkerType) WorkerBase {
	return WorkerBase{
		workerType: wType,
		ConfChan:   make(chan *WorkerConfig),
	}
}
