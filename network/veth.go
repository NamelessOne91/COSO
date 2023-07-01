package network

import (
	"net"

	"github.com/vishvananda/netlink"
)

// Veth represents a virtual Ethernet device pair. Veth devices are always created in interconnected pairs.
//
// Veth devices can act as a tunnel between network namespaces to create a bridge to a
// physical network device in another namespace. That's because packets transmitted on one device in the pair
// are immediately received on the other device
type Veth struct{}

func NewVeth() *Veth {
	return &Veth{}
}

// Create creates a new pair of veth devices with the given names.
// The devices are immediatly set 'up' once created (aka, are active)
func (v *Veth) Create(hostVethName, containerVethName string) (*net.Interface, *net.Interface, error) {
	// check if veth devices already exist
	if interfaceExists(hostVethName) {
		return vethInterfacesByName(hostVethName, containerVethName)
	}

	// configure and create veth pair
	vethLinkAttrs := netlink.NewLinkAttrs()
	vethLinkAttrs.Name = hostVethName
	veth := &netlink.Veth{
		LinkAttrs: vethLinkAttrs,
		PeerName:  containerVethName,
	}
	if err := netlink.LinkAdd(veth); err != nil {
		return nil, nil, err
	}

	// set device in 'up' state
	if err := netlink.LinkSetUp(veth); err != nil {
		return nil, nil, err
	}

	return vethInterfacesByName(hostVethName, containerVethName)
}

// MoveToNetworkNamespace moves the given veth device into another namespace, corresponding to trhe given pid
func (v *Veth) MoveToNetworkNamespace(containerVeth *net.Interface, pid int) error {
	// find veth deviced
	containerVethLink, err := netlink.LinkByName(containerVeth.Name)
	if err != nil {
		return err
	}

	// move veth to container namespace
	return netlink.LinkSetNsPid(containerVethLink, pid)
}

// vethInterfacesByName retrieves and returns the pair of veth devices with the given names
func vethInterfacesByName(hostVethName, containerVethName string) (*net.Interface, *net.Interface, error) {
	hostVeth, err := net.InterfaceByName(hostVethName)
	if err != nil {
		return nil, nil, err
	}

	containerVeth, err := net.InterfaceByName(containerVethName)
	if err != nil {
		return nil, nil, err
	}

	return hostVeth, containerVeth, nil
}
