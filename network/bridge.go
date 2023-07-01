package network

import (
	"net"

	"github.com/vishvananda/netlink"
)

type Bridge struct{}

func NewBridge() *Bridge {
	return &Bridge{}
}

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
