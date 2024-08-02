package metrics

import (
	"encoding/json"
)

type Resources struct {
	ContainerStats ContainerStats `json:"container-stats,omitempty"`
	CPUMetrics     *CPUMetrics    `json:"cpu,omitempty"`
	RamMetrics     *RamMetrics    `json:"ram,omitempty"`
	DiskMetrics    DiskMetrics    `json:"disks,omitempty"`
	NetStats       IfacesMetrics  `json:"net-stats,omitempty"`
}

type BaseResource struct {
	Topic string `json:"topic,omitempty"`
	Raw   string `json:"raw-sample,omitempty"`
}

type RamMetrics struct {
	BaseResource
	Used     uint64 `json:"used"`
	Capacity uint64 `json:"capacity"`
}

func (r RamMetrics) WriteToStatus(status *NuvlaEdgeStatus) error {
	r.Topic = "ram"
	r.Raw = GetRaw(r)
	status.Resources.RamMetrics = &r
	return nil
}

type DiskMetrics []DiskInfo

func (dm DiskMetrics) WriteToStatus(status *NuvlaEdgeStatus) error {
	for i := range dm {
		dm[i].Topic = "disk"
		dm[i].Raw = GetRaw(dm[i])
	}
	status.Resources.DiskMetrics = dm
	return nil
}

type DiskInfo struct {
	BaseResource
	Device   string `json:"device,omitempty"`
	Used     int32  `json:"used"`
	Capacity int32  `json:"capacity"`
}

type CPUMetrics struct {
	BaseResource
	Load     float64 `json:"load"`
	Load1    float64 `json:"load-1,omitempty"`
	Load5    float64 `json:"load-5,omitempty"`
	Capacity int     `json:"capacity"`
}

func (c CPUMetrics) WriteToStatus(status *NuvlaEdgeStatus) error {
	c.Topic = "cpu"
	c.Raw = GetRaw(c)

	status.Resources.CPUMetrics = &c
	return nil
}

type IfacesMetrics []IfaceMetrics

func (im IfacesMetrics) WriteToStatus(status *NuvlaEdgeStatus) error {
	status.Resources.NetStats = im
	return nil
}

type IfaceMetrics struct {
	BaseResource
	Interface        string `json:"interface,omitempty"`
	BytesTransmitted uint64 `json:"bytes-transmitted"`
	BytesReceived    uint64 `json:"bytes-received"`
}

func GetRaw(i interface{}) string {
	b, err := json.Marshal(i)
	if err != nil {
		return ""
	}
	return string(b)
}
