package types

type NuvlaEdgeWorker interface {
	Start() error
	Stop() error
	Run() error
	//IsRunning() bool
}
