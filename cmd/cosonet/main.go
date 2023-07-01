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
	defaultVethPrefix       = "coso-veth"
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

	// create bridge device
	bridge := network.NewBridge()
	bridgeIP, bridgeSubnet, err := net.ParseCIDR(bridgeAddress)
	if err != nil {
		fmt.Printf("Error during bridge configuration - %s\n", err)
		os.Exit(1)
	}

	bridgeDevice, err := bridge.Create(bridgeName, bridgeIP, bridgeSubnet)
	if err != nil {
		fmt.Printf("Error creating bridge - %s", err)
		os.Exit(1)
	}

	// create veth pair
	veth := network.Veth{}
	hostVeth, containerVeth, err := veth.Create(vethNamePrefix+defaultVethName, vethNamePrefix+defaultVethPeerName)
	if err != nil {
		fmt.Printf("Error creating veth pair - %s\n", err)
		os.Exit(1)
	}

	// attach host veth to the bridge
	if err = bridge.Attach(bridgeDevice, hostVeth); err != nil {
		fmt.Printf("Error attaching host veth to bridge - %s\n", err)
		os.Exit(1)
	}

	// move veth to new namespace
	if err = veth.MoveToNetworkNamespace(containerVeth, pid); err != nil {
		fmt.Printf("Error moving container veth in namespace - %s\n", err)
		os.Exit(1)
	}

}
