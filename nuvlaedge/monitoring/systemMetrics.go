package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"nuvlaedge-go/nuvlaedge/common"
	"strings"
	"sync"
	"time"
)

type CpuMetrics struct {
	Load  float64 `json:"load"`
	Load1 float64 `json:"load-1"`
	Load5 float64 `json:"load-5"`
	//SystemCalls        int64   `json:"system-calls"`
	Capacity int `json:"capacity"`
	//Interrupts         int64   `json:"interrupts"`
	Topic string `json:"topic"`
	//SoftwareInterrupts int64   `json:"software-interrupts"`
	//ContextSwitches    int64   `json:"context-switches"`

	cpuUsageAccumulator *common.CircularBuffer
	ctx                 context.Context
	cancel              context.CancelFunc
	once                *sync.Once
}

func NewCpuMetrics() *CpuMetrics {
	ctx, cancel := context.WithCancel(context.Background())
	c := &CpuMetrics{
		Topic:               "cpu",
		cpuUsageAccumulator: common.NewCircularBuffer(15 * 60),
		once:                &sync.Once{},
		ctx:                 ctx,
		cancel:              cancel,
	}
	c.once.Do(c.Run)
	return c
}

// Run starts the CPU metrics gathering for load, load-1 and load-5. Load represent a 15-min average
func (c *CpuMetrics) Run() {
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				return
			default:
				percent, err := cpu.Percent(time.Second, false)
				if err != nil {
					log.Errorf("Error getting CPU percentage: %s", err)
					c.cancel()
					return
				}
				c.cpuUsageAccumulator.Add(percent[0])
			}
		}
	}()
}

// healIfNeeded checks if the CPU metrics gathering is still running and restarts it if needed
func (c *CpuMetrics) healIfNeeded() {
	if c.ctx.Err() != nil {
		log.Info("CPU metrics gathering is not running. Restarting...")
		c.once = &sync.Once{}
		c.ctx, c.cancel = context.WithCancel(context.Background())
		c.once.Do(c.Run)
	}
}

func (c *CpuMetrics) Update() error {
	c.healIfNeeded()

	// Get CPU load
	loads1, err := c.cpuUsageAccumulator.GetLatestAvg(1 * 60)
	if err != nil {
		log.Errorf("Error getting CPU load-1: %s", err)
		//return err
	}
	c.Load1 = loads1

	loads5, err := c.cpuUsageAccumulator.GetLatestAvg(5 * 60)
	if err != nil {
		log.Errorf("Error getting CPU load-5: %s", err)
		//return err
	}
	c.Load5 = loads5

	loads15, err := c.cpuUsageAccumulator.GetLatestAvg(15 * 60)
	if err != nil {
		log.Errorf("Error getting CPU load-15: %s", err)
		//return err
	}
	c.Load = loads15

	// Get CPU count
	cpuCount, err := cpu.Counts(false)
	if err != nil {
		return err
	}
	c.Capacity = cpuCount

	// Get CPU interrupts
	t, _ := cpu.Times(false)
	log.Infof("CPU Times: %s", t[0])
	return nil
}

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
	Used     uint64 `json:"used"`
	Capacity uint64 `json:"capacity"`
}

func gatherDiskMetrics() ([]DiskMetrics, error) {

	partitions, err := disk.Partitions(true)
	if err != nil {
		return nil, err
	}
	var diskArr []DiskMetrics
	for _, partition := range partitions {
		if !strings.HasPrefix(partition.Device, "/dev/") {
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
		itDisk.Used = usage.Used / 1024 / 1024
		itDisk.Capacity = usage.Total / 1024 / 1024
		diskArr = append(diskArr, itDisk)
	}
	return diskArr, nil
}

type ResourceMetrics struct {
	ContainerStats []any           `json:"container-stats,omitempty"`
	Cpu            *CpuMetrics     `json:"cpu,omitempty"`
	Ram            *RamMetrics     `json:"ram,omitempty"`
	Disks          []DiskMetrics   `json:"disks,omitempty"`
	NetStats       []IfaceNetStats `json:"net-stats,omitempty"`

	networkInfo *NetworkMetricsUpdater
}

func NewResourceMetrics(info *NetworkMetricsUpdater) *ResourceMetrics {
	return &ResourceMetrics{
		Cpu:         NewCpuMetrics(),
		Ram:         NewRamMetrics(),
		networkInfo: info,
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

func NewResourceMetricsUpdater(info *NetworkMetricsUpdater) *ResourceMetricsUpdater {
	return &ResourceMetricsUpdater{
		metrics:    NewResourceMetrics(info),
		updateLock: sync.Mutex{},
	}
}

func (r *ResourceMetricsUpdater) updateMetrics() error {
	err := r.metrics.Cpu.Update()
	if err != nil {
		return err
	}

	err = r.metrics.Ram.Update()
	if err != nil {
		return err
	}

	diskMetrics, err := gatherDiskMetrics()
	if err != nil {
		return err
	}
	r.metrics.Disks = diskMetrics

	netStats, err := r.metrics.networkInfo.GetStats()
	if err != nil {
		return err
	}
	r.metrics.NetStats = netStats

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
