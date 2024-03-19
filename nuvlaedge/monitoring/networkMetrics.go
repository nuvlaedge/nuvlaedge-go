package monitoring

import (
	"encoding/json"
	"fmt"
	"github.com/jackpal/gateway"
	psNet "github.com/shirou/gopsutil/v3/net"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

// IfaceNetStats represents the network IO statistics of an interface.
type IfaceNetStats struct {
	Interface        string `json:"interface"`
	BytesTransmitted uint64 `json:"bytes-transmitted"`
	BytesReceived    uint64 `json:"bytes-received"`
}

// InterfaceInfo represents the IP addresses of an interface.
type InterfaceInfo struct {
	Interface string              `json:"interface"`
	Ips       []map[string]string `json:"ips"`
}

// NetworkInfo is the main wrapper for the network information.
type NetworkInfo struct {
	DefaultGw string `json:"default-gw"`

	IPs struct {
		Public string `json:"public"`
		Swarm  string `json:"swarm"`
		Vpn    string `json:"vpn"`
		Local  string `json:"local"`
	} `json:"ips"`

	Interfaces []*InterfaceInfo `json:"interfaces"`
}

// NetworkMetricsUpdater is responsible for updating the network metrics. It offers helper methods to retrieve the
// information in different formats
type NetworkMetricsUpdater struct {
	NetworkInfo *NetworkInfo
	IfaceStats  []IfaceNetStats

	updateLock sync.Mutex
	updateTime time.Time
}

// NewNetworkMetricsUpdater creates a new NetworkMetricsUpdater.
func NewNetworkMetricsUpdater() *NetworkMetricsUpdater {
	return &NetworkMetricsUpdater{
		NetworkInfo: &NetworkInfo{},
	}
}

// getGateway retrieves the default gateway IP address.
func getGateway() (string, error) {
	g, err := gateway.DiscoverInterface()
	return g.String(), err
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

// updateNetworkInfo retrieves all the network information and updates the NetworkInfo struct.
func (n *NetworkMetricsUpdater) updateNetworkInfo() error {

	// Retrieve public IP address
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := http.Get("https://api.ipify.org")
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
		n.NetworkInfo.IPs.Public = string(ip)
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
	gw, err := getGateway()

	// Reset network info
	n.IfaceStats = make([]IfaceNetStats, 0)
	n.NetworkInfo.Interfaces = make([]*InterfaceInfo, 0)

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
					n.NetworkInfo.DefaultGw = iface.Name
					n.NetworkInfo.IPs.Local = ip.String()
				}

				// Retrieve the IO data for the interface
				tx, rx, err := getIODataFromInterface(iface.Name, ioCounters)
				n.IfaceStats = append(n.IfaceStats, IfaceNetStats{
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
				n.NetworkInfo.Interfaces = append(
					n.NetworkInfo.Interfaces,
					&InterfaceInfo{
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

func (n *NetworkMetricsUpdater) updateNetworkInfoIfNeeded() error {
	n.updateLock.Lock()
	defer n.updateLock.Unlock()
	if time.Since(n.updateTime) > 10*time.Second {
		return n.updateNetworkInfo()
	}
	return nil
}

func (n *NetworkMetricsUpdater) GetStats() ([]IfaceNetStats, error) {
	err := n.updateNetworkInfoIfNeeded()
	if err != nil {
		return nil, err
	}

	return n.IfaceStats, nil

}

func (n *NetworkMetricsUpdater) GetIPV4() (string, error) {
	err := n.updateNetworkInfoIfNeeded()
	if err != nil {
		return "", err
	}
	if n.NetworkInfo.IPs.Vpn != "" {
		return n.NetworkInfo.IPs.Vpn, nil
	}
	if n.NetworkInfo.IPs.Local != "" {
		return n.NetworkInfo.IPs.Local, nil
	}
	if n.NetworkInfo.IPs.Public != "" {
		return n.NetworkInfo.IPs.Public, nil
	}
	if n.NetworkInfo.IPs.Swarm != "" {
		return n.NetworkInfo.IPs.Swarm, nil
	}
	return "", fmt.Errorf("no IP found")
}

func (n *NetworkMetricsUpdater) GetNetworkInfoMap() (map[string]interface{}, error) {
	err := n.updateNetworkInfoIfNeeded()
	if err != nil {
		return nil, err
	}

	netJson, err := json.Marshal(n.NetworkInfo)
	if err != nil {
		return nil, err
	}

	var netInfoMap map[string]interface{}
	err = json.Unmarshal(netJson, &netInfoMap)
	if err != nil {
		return nil, err
	}
	return netInfoMap, nil
}
