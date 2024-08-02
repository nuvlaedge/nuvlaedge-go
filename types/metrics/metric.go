package metrics

type Metric interface {
	WriteToStatus(status *NuvlaEdgeStatus) error
}
