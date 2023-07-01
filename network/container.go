package network

import (
	"fmt"
	"net"
	"os"
	"runtime"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

type ContainerNetworkManager struct{}

// Configure configures the container's network devices in a thread safe way.
//
// A network namespace switch is operated and a route to the veth device, which has previously been moved to the
// container's namespace, is created. The host's network namespace is then restored.
func (c *ContainerNetworkManager) Configure(config NetworkConfig, pid int) error {
	// open symbolic link to the container's network namespace
	netnsFile, err := os.Open(fmt.Sprintf("/proc/%d/ns/net", pid))
	if err != nil {
		return fmt.Errorf("unable to find network namespace for process with pid '%d'", pid)
	}
	defer netnsFile.Close()

	fileDescriptor := int(netnsFile.Fd())

	// we have to ensure that the current goroutine remains on the same operating system thread throughout the execution of the function
	// this is necessary because the network namespace switching should occur on the same thread
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// container's network namespace handle
	newns := netns.NsHandle(fileDescriptor)
	// host's network namespace handle
	hostNs, _ := netns.Get()
	defer hostNs.Close()

	// network namespace switch
	if err := netns.Set(newns); err != nil {
		return fmt.Errorf("error while trying to switch to the container's network namespace: %s", err)
	}
	// configure container devices
	err = containerNetworkSetup(config.ContainerVethName, config.ContainerIP, config.BridgeIP, config.Subnet.Mask)

	// back to the host's network namespace
	if err := netns.Set(hostNs); err != nil {
		fmt.Printf("Well, fuck.\nError while trying to switch back to the host's network namespace - %s", err)
		panic(err)
	}

	return err
}

// containerNetworkSetup takes care of configuring the container's veth device and adding a new
// route to it, with the host's bridge as gateway
func containerNetworkSetup(vethName string, ip, bridgeIP net.IP, mask net.IPMask) error {
	// find veth
	veth, err := netlink.LinkByName(vethName)
	if err != nil {
		return fmt.Errorf("container veth '%s' not found", vethName)
	}

	// add IP to veth
	addr := &netlink.Addr{IPNet: &net.IPNet{IP: ip, Mask: mask}}
	err = netlink.AddrAdd(veth, addr)
	if err != nil {
		return fmt.Errorf("unable to assign IP address '%s' to %s", ip, vethName)
	}

	// set veth to 'up'
	if err := netlink.LinkSetUp(veth); err != nil {
		return err
	}

	// create new route
	route := &netlink.Route{
		Scope:     netlink.SCOPE_UNIVERSE, // global route accessible to all network interfaces.
		LinkIndex: veth.Attrs().Index,     // associate with veth device
		Gw:        bridgeIP,               // gateway IP --> Bridge IP
	}
	return netlink.RouteAdd(route)
}
