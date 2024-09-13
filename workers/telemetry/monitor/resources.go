package monitor

import (
	"context"
	"github.com/jackpal/gateway"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	psNet "github.com/shirou/gopsutil/v3/net"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"nuvlaedge-go/types/metrics"
	"strings"
	"sync"
	"time"
)

type ResourceMonitor struct {
	BaseMonitor

	ramData     metrics.RamMetrics
	cpuData     metrics.CPUMetrics
	disksData   metrics.DiskMetrics
	ifaceData   metrics.IfacesMetrics
	networkData metrics.NetworkMetrics
}

func NewResourceMonitor(period int, ch chan metrics.Metric) *ResourceMonitor {
	return &ResourceMonitor{
		BaseMonitor: NewBaseMonitor(period, ch),
	}
}

func (rm *ResourceMonitor) Run(ctx context.Context) error {
	rm.SetRunning()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-rm.Ticker.C:
			// Send metric to channel
			if err := rm.updateMetrics(); err != nil {
				log.Errorf("Error updating resource metrics: %s", err)
			}

			rm.sendMetrics()
		}
	}

}

func (rm *ResourceMonitor) sendMetrics() {
	rm.reportChan <- rm.ramData
	rm.reportChan <- rm.cpuData
	rm.reportChan <- rm.disksData
	rm.reportChan <- rm.ifaceData
	rm.reportChan <- rm.networkData
}

func (rm *ResourceMonitor) updateMetrics() error {
	var errs []error

	if err := rm.updateRam(); err != nil {
		errs = append(errs, err)
	}

	if err := rm.updateCPU(); err != nil {
		errs = append(errs, err)
	}

	if err := rm.updateDisks(); err != nil {
		errs = append(errs, err)
	}

	if err := rm.updateIface(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func (rm *ResourceMonitor) updateRam() error {
	ram, err := mem.VirtualMemory()
	if err != nil {
		return err
	}
	rm.ramData.Used = ram.Used / 1024 / 1024
	rm.ramData.Capacity = ram.Total / 1024 / 1024
	return nil
}

func (rm *ResourceMonitor) updateDisks() error {
	rm.disksData = make(metrics.DiskMetrics, 0)

	partitions, err := disk.Partitions(true)
	if err != nil {
		return err
	}
	diskMap := make(map[string]metrics.DiskInfo)

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

		itDisk := metrics.DiskInfo{Device: partition.Device}

		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			log.Infof("Error getting disk usage for %s: %s", partition.Mountpoint, err)
			continue
		}

		itDisk.Used = int32(usage.Used / 1024 / 1024 / 1024)
		itDisk.Capacity = int32(usage.Total / 1024 / 1024 / 1024)
		if itDisk.Capacity <= 0 || itDisk.Used <= 0 {
			log.Debugf("Skipping disk %s. Total disk space is 0", partition.Device)
			continue
		}

		diskMap[partition.Device] = itDisk
		rm.disksData = append(rm.disksData, itDisk)
	}
	return nil
}

func (rm *ResourceMonitor) updateIface() error {
	// Retrieve public IP address
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		r, err := http.NewRequestWithContext(ctx, "GET", "https://api.ipify.org", nil)
		if err != nil {
			log.Errorf("Error creating request: %v", err)
			return
		}

		resp, err := http.DefaultClient.Do(r)
		if err != nil {
			log.Errorf("Error getting public IP: %v", err)
			return
		}
		defer resp.Body.Close()

		ip, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Errorf("Error reading public IP: %v", err)
			return
		}
		rm.networkData.IPs.Public = string(ip)
	}()

	// Gather all interfaces information from the system
	ifaces, err := psNet.Interfaces()
	if err != nil {
		return err
	}

	// Retrieve TX/RX data for all interfaces
	ioCounters, err := psNet.IOCounters(true)
	if err != nil {
		return err
	}

	// Retrieve default gateway IP address to find the default interface in the loop
	gw, _ := getGateway()

	// Reset network info
	rm.ifaceData = make([]metrics.IfaceMetrics, 0)
	rm.networkData.Interfaces = make([]metrics.InterfaceInfo, 0)

	// Loop through all interfaces
	for _, iface := range ifaces {
		// Loop through all addresses in the interface
		for _, addr := range iface.Addrs {
			// Parse the IP address as CIDR
			ip, _, err := net.ParseCIDR(addr.Addr)
			if err != nil {
				log.Errorf("Error parsing IP: %v", err)
				continue
			}

			// Check if the IP is private and global unicast. It should mean is a valid IP address
			if ip.IsPrivate() && ip.IsGlobalUnicast() {

				// Check if the IP is the default gateway
				if ip.String() == gw {
					log.Debugf("Found gateway in interface %s", iface.Name)
					rm.networkData.DefaultGw = iface.Name
					rm.networkData.IPs.Local = ip.String()
				}

				// Retrieve the IO data for the interface
				tx, rx, err := getIODataFromInterface(iface.Name, ioCounters)
				rm.ifaceData = append(rm.ifaceData, metrics.IfaceMetrics{
					Interface:        iface.Name,
					BytesTransmitted: tx,
					BytesReceived:    rx,
				})
				if err != nil {
					log.Errorf("Error getting IO data from interface %s: %v", iface.Name, err)
					continue
				}

				// Add the interface to the list of interfaces
				ipAddr := make(map[string]string)
				ipAddr["address"] = ip.String()
				var ips []map[string]string
				ips = append(ips, ipAddr)
				rm.networkData.Interfaces = append(
					rm.networkData.Interfaces,
					metrics.InterfaceInfo{
						Interface: iface.Name,
						Ips:       ips,
					},
				)

			}

		}
	}

	wg.Wait()
	return nil
}

// getIODataFromInterface retrieves the IO data of the interface.
func getIODataFromInterface(ifaceName string, counters []psNet.IOCountersStat) (uint64, uint64, error) {

	for _, ioCounter := range counters {
		if ioCounter.Name == ifaceName {
			return ioCounter.BytesSent, ioCounter.BytesRecv, nil
		}
	}

	return 0, 0, nil
}

// getGateway retrieves the default gateway IP address.
func getGateway() (string, error) {
	g, err := gateway.DiscoverInterface()
	if err != nil {
		return "", err
	}
	return g.String(), err
}
