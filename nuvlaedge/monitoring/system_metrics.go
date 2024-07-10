package monitoring

import (
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	log "github.com/sirupsen/logrus"
	"strings"
	"sync"
	"time"
)

type RamMetrics struct {
	Used     uint64 `json:"used"`
	Capacity uint64 `json:"capacity"`
	Topic    string `json:"topic"`
}

func NewRamMetrics() *RamMetrics {
	return &RamMetrics{
		Topic: "ram",
	}
}

func (r *RamMetrics) Update() error {
	// Get RAM usage
	ram, err := mem.VirtualMemory()
	if err != nil {
		return err
	}
	r.Used = ram.Used / 1024 / 1024
	r.Capacity = ram.Total / 1024 / 1024
	return nil
}

type DiskMetrics struct {
	Topic    string `json:"topic"`
	Device   string `json:"device"`
	Used     int32  `json:"used"`
	Capacity int32  `json:"capacity"`
}

func gatherDiskMetrics() ([]DiskMetrics, error) {

	partitions, err := disk.Partitions(true)
	if err != nil {
		return nil, err
	}
	var diskArr []DiskMetrics
	diskMap := make(map[string]DiskMetrics)

	for _, partition := range partitions {
		if !strings.HasPrefix(partition.Device, "/dev/") {
			log.Debugf("Skipping partition %s", partition.Device)
			continue
		}
		if _, ok := diskMap[partition.Device]; ok {
			log.Debugf("Skipping partition %s", partition.Device)
			continue
		}
		log.Debugf("Getting disk metrics for %s", partition.Device)

		itDisk := DiskMetrics{
			Device: partition.Device,
			Topic:  "disk",
		}
		log.Debugf("Getting disk usage for %s", partition.Mountpoint)
		usage, err := disk.Usage(partition.Mountpoint)

		if err != nil {
			return nil, err
		}
		itDisk.Used = int32(usage.Used / 1024 / 1024 / 1024)
		itDisk.Capacity = int32(usage.Total / 1024 / 1024 / 1024)
		if itDisk.Capacity <= 0 || itDisk.Used <= 0 {
			log.Debugf("Skipping disk %s. Total disk space is 0", partition.Device)
			continue
		}

		diskMap[partition.Device] = itDisk
		diskArr = append(diskArr, itDisk)
	}
	return diskArr, nil
}

type ResourceMetrics struct {
	ContainerStats []map[string]any `json:"container-stats,omitempty"`
	Cpu            *CPUMetrics      `json:"cpu,omitempty"`
	Ram            *RamMetrics      `json:"ram,omitempty"`
	Disks          []DiskMetrics    `json:"disks,omitempty"`
	NetStats       []IfaceNetStats  `json:"net-stats,omitempty"`

	networkInfo *NetworkMetricsUpdater
	containers  *ContainerStats
}

func NewResourceMetrics(network *NetworkMetricsUpdater, containers *ContainerStats) *ResourceMetrics {
	cpuMetrics := NewCPUMetrics()
	cpuMetrics.Run()
	return &ResourceMetrics{
		Cpu:         cpuMetrics,
		Ram:         NewRamMetrics(),
		networkInfo: network,
		containers:  containers,
	}
}

func (r *ResourceMetrics) GetMetricsAsMap() (map[string]interface{}, error) {
	jsonMetrics, err := json.Marshal(r)
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return nil, err
	}
	var mapMetrics map[string]interface{}
	err = json.Unmarshal(jsonMetrics, &mapMetrics)
	if err != nil {
		fmt.Println("Error unmarshalling to map:", err)
		return nil, err
	}
	return mapMetrics, nil
}

type ResourceMetricsUpdater struct {
	metrics *ResourceMetrics

	updateLock sync.Mutex
	updateTime time.Time
}

func NewResourceMetricsUpdater(network *NetworkMetricsUpdater, containers *ContainerStats) *ResourceMetricsUpdater {
	return &ResourceMetricsUpdater{
		metrics:    NewResourceMetrics(network, containers),
		updateLock: sync.Mutex{},
	}
}

func (r *ResourceMetricsUpdater) updateMetrics() error {
	err := r.metrics.Cpu.Update()
	if err != nil {
		return fmt.Errorf("error getting CPU metrics: %s", err)
	}

	err = r.metrics.Ram.Update()
	if err != nil {
		return fmt.Errorf("error getting RAM metrics: %s", err)
	}

	diskMetrics, err := gatherDiskMetrics()
	if err != nil {
		return fmt.Errorf("error getting disk metrics: %s", err)
	}
	r.metrics.Disks = diskMetrics

	netStats, err := r.metrics.networkInfo.GetStats()
	if err != nil {
		return fmt.Errorf("error getting network stats: %s", err)
	}
	r.metrics.NetStats = netStats

	containerStats, err := r.metrics.containers.getStats()
	if err != nil {
		log.Warnf("Error retrieving container stats: %s", err)
		return err
	}
	log.Debugf("Last Container Stats update: %v", containerStats)
	r.metrics.ContainerStats = containerStats

	r.updateTime = time.Now()
	return nil
}

func (r *ResourceMetricsUpdater) updateResourceMetricsIfNeeded() error {
	r.updateLock.Lock()
	defer r.updateLock.Unlock()
	if time.Since(r.updateTime) > 10*time.Second {
		return r.updateMetrics()
	}
	return nil
}

func (r *ResourceMetricsUpdater) GetResourceMetrics() (map[string]interface{}, error) {
	err := r.updateResourceMetricsIfNeeded()
	if err != nil {
		return nil, err
	}

	mapMetrics, err := r.metrics.GetMetricsAsMap()
	if err != nil {
		return nil, err
	}

	return mapMetrics, nil
}
