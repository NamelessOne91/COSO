package network

import (
	"fmt"
	"net"
	"os"
	"path"
	"runtime"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

const (
	dnsNameserver = "nameserver 8.8.8.8" // Google's public IPV4 DNS
)

// ContainerNetworkManager gives access to methods to configure a container's network namespace devices
type ContainerNetworkManager struct{}

func NewContainerNetworkManager() *ContainerNetworkManager {
	return &ContainerNetworkManager{}
}

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
		fmt.Printf("Well, fuck.\nCritical error: failed to switch back to the host's network namespace - %s", err)
		panic(err)
	}

	// setup DNS nameserver
	if err := addDNSNameserver(pid); err != nil {
		fmt.Printf("Error configuring DNS name server - %s\n", err)
		os.Exit(1)
	}
	fmt.Println("DNS nameserver setup complete")

	return err
}

// containerNetworkSetup takes care of configuring the container's veth device and adding a new
// route to it, with the host's bridge as gateway
func containerNetworkSetup(vethName string, containerIP, bridgeIP net.IP, mask net.IPMask) error {
	// find veth
	veth, err := netlink.LinkByName(vethName)
	if err != nil {
		return fmt.Errorf("container veth '%s' not found", vethName)
	}

	// add IP to veth
	addr := &netlink.Addr{IPNet: &net.IPNet{IP: containerIP, Mask: mask}}
	err = netlink.AddrAdd(veth, addr)
	if err != nil {
		return fmt.Errorf("unable to assign IP address '%s' to %s", containerIP, vethName)
	}
	fmt.Println("IP assigned to veth device")

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
	if err = netlink.RouteAdd(route); err != nil {
		return err
	}

	fmt.Println("Route created")
	return nil
}

// addDNSNameserver writes a DNS resolver configuration file for the namespace associated with the given PID
func addDNSNameserver(pid int) error {
	dnsPath := fmt.Sprintf("/etc/netns/%d/resolv.conf", pid)

	// Create the directory if it doesn't exist
	err := os.MkdirAll(path.Dir(dnsPath), 0755)
	if err != nil {
		return err
	}

	// Open the file in append mode with write-only permissions
	file, err := os.OpenFile(dnsPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the content to the file
	_, err = file.WriteString(dnsNameserver)
	if err != nil {
		return err
	}

	return nil
}
