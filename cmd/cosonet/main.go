package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/NamelessOne91/coso/network"
)

const (
	defaultBridgeName       = "coso-bridge"
	defaultBridgeAddress    = "10.10.10.1/24"
	defaultVethPrefix       = "coso-veth-"
	defaultVethName         = "host"
	defaultVethPeerName     = "peer"
	defaultContainerAddress = "10.10.10.2/24"
	defaultPid              = 0
)

func main() {
	var bridgeName, bridgeAddress, containerAddress, vethNamePrefix string
	var pid int

	flag.StringVar(&bridgeName, "bridgeName", defaultBridgeName, "Name to assign to bridge device")
	flag.StringVar(&bridgeAddress, "bridgeAddress", defaultBridgeAddress, "Address to assign to bridge device (CIDR notation)")
	flag.StringVar(&vethNamePrefix, "vethNamePrefix", defaultVethName, "Name prefix for veth devices")
	flag.StringVar(&containerAddress, "containerAddress", defaultContainerAddress, "Address to assign to the container (CIDR notation)")
	flag.IntVar(&pid, "pid", defaultPid, "pid of a process in the container's network namespace")
	flag.Parse()

	bridge := network.NewBridge()
	veth := network.NewVeth()

	hostManager := network.NewHostNetworkManager(bridge, veth)
	containerManager := network.NewContainerNetworkManager()

	bridgeIP, bridgeSubnet, err := net.ParseCIDR(bridgeAddress)
	if err != nil {
		fmt.Printf("Error trying to parse bridge CIDR - %s\n", err)
		os.Exit(1)
	}

	containerIP, _, err := net.ParseCIDR(containerAddress)
	if err != nil {
		fmt.Printf("Error trying to parse container CIDR - %s\n", err)
		os.Exit(1)
	}

	config := network.NetworkConfig{
		BridgeName:        bridgeName,
		BridgeIP:          bridgeIP,
		ContainerIP:       containerIP,
		Subnet:            bridgeSubnet,
		HostVethName:      defaultVethPrefix + defaultVethName,
		ContainerVethName: defaultVethPrefix + defaultVethPeerName,
	}

	networkmanager := network.NewNetworkManager(hostManager, containerManager)
	if err := networkmanager.Configure(config, pid); err != nil {
		fmt.Printf("Error trying to configure network devices - %s\n", err)
		os.Exit(1)
	}
}
