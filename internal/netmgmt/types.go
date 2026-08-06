// Package netmgmt provides network management commands for the server CLI.
// It supports listing, inspecting, adding, and removing network interfaces,
// VLANs, and bonds using interactive bubbletea forms.
package netmgmt

import (
	"fmt"
	"net"
	"os"
	"strings"

	gopsnet "github.com/shirou/gopsutil/v4/net"
)

// InterfaceType categorises a network interface.
type InterfaceType string

const (
	TypePhysical InterfaceType = "physical"
	TypeVLAN     InterfaceType = "vlan"
	TypeBond     InterfaceType = "bond"
	TypeBridge   InterfaceType = "bridge"
	TypeLoopback InterfaceType = "loopback"
	TypeVirtual  InterfaceType = "virtual"
	TypeUnknown  InterfaceType = "unknown"
)

// Interface describes a network interface with its addresses and stats.
type Interface struct {
	Name        string        `json:"name"`
	Type        InterfaceType `json:"type"`
	MTU         int           `json:"mtu"`
	MAC         string        `json:"mac"`
	Up          bool          `json:"up"`
	IPAddresses []string      `json:"ip_addresses"`
	RxBytes     uint64        `json:"rx_bytes"`
	TxBytes     uint64        `json:"tx_bytes"`
	RxPackets   uint64        `json:"rx_packets"`
	TxPackets   uint64        `json:"tx_packets"`
	RxErrors    uint64        `json:"rx_errors"`
	TxErrors    uint64        `json:"tx_errors"`
	Speed       string        `json:"speed"`
	Duplex      string        `json:"duplex"`
}

// BondInfo holds bond-specific details if the interface is a bond.
type BondInfo struct {
	Mode    string
	Slaves  []string
	Active  string
	Primary string
}

// VLANInfo holds VLAN-specific details if the interface is a VLAN.
type VLANInfo struct {
	ParentIF string
	VLANID   int
}

// GatherInterfaces collects all network interfaces and their statistics.
func GatherInterfaces() ([]Interface, error) {
	gopsIfaces, err := gopsnet.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("listing interfaces: %w", err)
	}

	counters, _ := gopsnet.IOCounters(true)
	counterMap := make(map[string]gopsnet.IOCountersStat)
	for _, c := range counters {
		counterMap[c.Name] = c
	}

	var result []Interface
	for _, gopsIF := range gopsIfaces {
		netIF, err := net.InterfaceByName(gopsIF.Name)
		if err != nil {
			continue
		}

		var ips []string
		addrs, _ := netIF.Addrs()
		for _, addr := range addrs {
			ips = append(ips, addr.String())
		}

		mac := netIF.HardwareAddr.String()
		if mac == "" {
			mac = gopsIF.HardwareAddr
		}

		i := Interface{
			Name:        gopsIF.Name,
			Type:        classifyInterface(gopsIF.Name, netIF),
			MTU:         gopsIF.MTU,
			MAC:         mac,
			Up:          netIF.Flags&net.FlagUp != 0,
			IPAddresses: ips,
		}

		if c, ok := counterMap[gopsIF.Name]; ok {
			i.RxBytes = c.BytesRecv
			i.TxBytes = c.BytesSent
			i.RxPackets = c.PacketsRecv
			i.TxPackets = c.PacketsSent
			i.RxErrors = c.Errin
			i.TxErrors = c.Errout
		}

		i.Speed, i.Duplex = readEthtool(gopsIF.Name)

		result = append(result, i)
	}

	return result, nil
}

// classifyInterface determines the type of a network interface.
func classifyInterface(name string, iface *net.Interface) InterfaceType {
	if name == "lo" {
		return TypeLoopback
	}

	// Check for VLAN (e.g. eth0.100)
	if strings.Contains(name, ".") {
		return TypeVLAN
	}

	// Check for bond
	if strings.HasPrefix(name, "bond") {
		return TypeBond
	}

	// Check for bridge
	if strings.HasPrefix(name, "br") || strings.HasPrefix(name, "docker") {
		return TypeBridge
	}

	// Check for virtual/veth
	if strings.HasPrefix(name, "veth") || strings.HasPrefix(name, "virbr") {
		return TypeVirtual
	}

	// Check /sys/class/net for interface type
	sysPath := "/sys/class/net/" + name
	if _, err := os.Stat(sysPath + "/bonding"); err == nil {
		return TypeBond
	}
	if _, err := os.Stat(sysPath + "/bridge"); err == nil {
		return TypeBridge
	}
	if _, err := os.Stat(sysPath + "/vlan_id"); err == nil {
		return TypeVLAN
	}

	// If it has a device symlink, it's likely physical
	if _, err := os.Stat(sysPath + "/device"); err == nil {
		return TypePhysical
	}

	return TypeUnknown
}

// readEthtool attempts to read speed and duplex from /sys/class/net.
func readEthtool(name string) (speed, duplex string) {
	base := "/sys/class/net/" + name
	if b, err := os.ReadFile(base + "/speed"); err == nil {
		speed = strings.TrimSpace(string(b))
		if speed == "-1" {
			speed = "unknown"
		}
	}
	if b, err := os.ReadFile(base + "/duplex"); err == nil {
		duplex = strings.TrimSpace(string(b))
	}
	return
}

// GatherBondInfo reads bond details from /proc/net/bonding.
func GatherBondInfo(name string) (*BondInfo, error) {
	data, err := os.ReadFile("/proc/net/bonding/" + name)
	if err != nil {
		return nil, err
	}

	info := &BondInfo{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "Bonding Mode:"):
			info.Mode = strings.TrimSpace(strings.TrimPrefix(line, "Bonding Mode:"))
		case strings.HasPrefix(line, "Currently Active Slave:"):
			info.Active = strings.TrimSpace(strings.TrimPrefix(line, "Currently Active Slave:"))
		case strings.HasPrefix(line, "Primary Slave:"):
			info.Primary = strings.TrimSpace(strings.TrimPrefix(line, "Primary Slave:"))
		case strings.HasPrefix(line, "Slave Interface:"):
			info.Slaves = append(info.Slaves, strings.TrimSpace(strings.TrimPrefix(line, "Slave Interface:")))
		}
	}

	return info, nil
}

// GatherVLANInfo reads VLAN ID from /sys/class/net.
func GatherVLANInfo(name string) (*VLANInfo, error) {
	data, err := os.ReadFile("/sys/class/net/" + name + "/vlan_id")
	if err != nil {
		return nil, err
	}
	var id int
	fmt.Sscanf(strings.TrimSpace(string(data)), "%d", &id)

	// Parent is the part before the dot
	parent := name
	if idx := strings.Index(name, "."); idx > 0 {
		parent = name[:idx]
	}

	return &VLANInfo{
		ParentIF: parent,
		VLANID:   id,
	}, nil
}