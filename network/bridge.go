package network

import (
	"net"

	"github.com/vishvananda/netlink"
)

// BridgeManager provides the necessary methods to create, destroy and attach a veth interface to a Linux bridge device
type BridgeManager interface {
	Create(name string, ip net.IP, subnet *net.IPNet) (*net.Interface, error)
	Attach(bridge, hostVeth *net.Interface) error
	Remove()
}

// Bridge represents a Linux bridge network device manager
type Bridge struct{}

func NewBridge() *Bridge {
	return &Bridge{}
}

// Create creates a new bridge network device and returns it.
// The bridge is immediatly set to 'up' once created (aka, is active)
func (b *Bridge) Create(name string, ip net.IP, subnet *net.IPNet) (*net.Interface, error) {
	// check if the bridge already exists
	if interfaceExists(name) {
		return net.InterfaceByName(name)
	}

	// define and create device
	linkAttrs := netlink.LinkAttrs{Name: name}
	link := &netlink.Bridge{
		LinkAttrs: linkAttrs,
	}
	if err := netlink.LinkAdd(link); err != nil {
		return nil, err
	}

	// add IP address to the device
	address := &netlink.Addr{IPNet: &net.IPNet{IP: ip, Mask: subnet.Mask}}
	if err := netlink.AddrAdd(link, address); err != nil {
		return nil, err
	}

	// set device in 'up' state
	if err := netlink.LinkSetUp(link); err != nil {
		return nil, err
	}

	return net.InterfaceByName(name)
}

// Attach attaches the given veth device to the given bridge device
func (b *Bridge) Attach(bridge, hostVeth *net.Interface) error {
	// find bridge device
	bridgeLink, err := netlink.LinkByName(bridge.Name)
	if err != nil {
		return err
	}
	//find host veth device
	hostVethLink, err := netlink.LinkByName(hostVeth.Name)
	if err != nil {
		return err
	}
	// attach veth to bridge
	return netlink.LinkSetMaster(hostVethLink, bridgeLink.(*netlink.Bridge))
}

// Remove deletes the bridge device with the given name
func (b *Bridge) Remove(bridgeName string) error {
	bridge, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}

	if err := netlink.LinkDel(bridge); err != nil {
		return err
	}

	return nil
}
