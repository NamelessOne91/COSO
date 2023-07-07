package network

import (
	"fmt"
	"os/exec"
	"strings"
)

const (
	chainName = "cosonet"
)

// HostNetWorkManager holds the BridgeManager and VethManager responsible for the correct
// configuration of the host's network devices
type HostNetworkManager struct {
	bridgeManager BridgeManager
	vethManager   VethManager
}

func NewHostNetworkManager(bm BridgeManager, vm VethManager) *HostNetworkManager {
	return &HostNetworkManager{
		bridgeManager: bm,
		vethManager:   vm,
	}
}

// Configure creates new bridge and a veth pair devices in the host namespace.
//
// One of the veth devices is attached to the brige, the other is then moved to the namespace
// identified by the given pid. This allows to route traffic to the new namespace.
func (h *HostNetworkManager) Configure(netConfig NetworkConfig, pid int) error {
	fmt.Println("Started Host's network configuration")
	// create bridge device
	bridge, err := h.bridgeManager.Create(netConfig.BridgeName, netConfig.BridgeIP, netConfig.Subnet)
	if err != nil {
		return err
	}
	fmt.Println("Created bridge device")

	// create veth devices pair
	hostVeth, containerVeth, err := h.vethManager.Create(netConfig.HostVethName, netConfig.ContainerVethName)
	if err != nil {
		return err
	}
	fmt.Println("Created veth devices")

	// attach host veth device to the bridge device
	if err = h.bridgeManager.Attach(bridge, hostVeth); err != nil {
		return err
	}
	fmt.Println("Attached veth to bridge device")

	// move the peer veth device to the new namespace
	if err = h.vethManager.MoveToNetworkNamespace(containerVeth, pid); err != nil {
		return err
	}
	fmt.Println("Moved veth device tp container's namespace")

	// check NAT iptable
	exists, err := natChainExists(chainName)
	if err != nil {
		return err
	}
	if !exists {
		// create NAT chain
		if err = createNATChain(chainName); err != nil {
			return err
		}
		fmt.Println("Created NAT chain")

		// add routing rules
		if err = addNATRules(netConfig.BridgeName, chainName); err != nil {
			return err
		}
		fmt.Println("NAT rules setup complete")
	} else {
		fmt.Println("NAT chain already exists")
	}

	return nil
}

// natChainExists verify if the NAT chain with the given name already exists in the related iptable
func natChainExists(chainName string) (bool, error) {
	listCmd := exec.Command("sudo", "iptables", "-t", "nat", "-L")
	output, err := listCmd.Output()
	if err != nil {
		return false, err
	}
	// Check if the desired chain exists
	exists := strings.Contains(string(output), "Chain "+chainName)
	return exists, nil
}

// createChain creates a new chain with the given name in the NAT table of iptables
func createNATChain(chainName string) error {
	cmd := exec.Command("sudo", "iptables", "-t", "nat", "-N", chainName)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error creating NAT chain - %s\n", err)
		return err
	}
	return nil
}

// addNATRule adds a rule to the specified NAT chain in the NAT table
func addNATRule(chainName, match, target string) error {
	matchArgs := strings.Split(match, " ")
	args := append([]string{"sudo", "iptables", "-t", "nat", "-A", chainName}, matchArgs...)
	args = append(args, "-j", target)

	cmd := exec.Command(args[0], args[1:]...)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error creating NAT rule - chain %s - match %s - target %s || %s\n", chainName, match, target, err)
		return err
	}

	return nil
}

// addNATRules adds the multiple NAT rules needed to route to/from the bridge device
func addNATRules(bridgeName, chainName string) error {
	// Add PREROUTING rule to jump to cosonet chain for packets with destination type LOCAL
	if err := addNATRule("PREROUTING", "-m addrtype --dst-type LOCAL", chainName); err != nil {
		return err
	}

	// Add OUTPUT rule to jump to cosonet chain for packets not destined for 127.0.0.0/8 and have destination type LOCAL
	if err := addNATRule("OUTPUT", "! -d 127.0.0.0/8 -m addrtype --dst-type LOCAL", chainName); err != nil {
		return err
	}

	// Add POSTROUTING rule to perform MASQUERADE on packets with source address in 10.10.10.0/24 and not destined for <bridgeName>
	if err := addNATRule("POSTROUTING", "-s 10.10.10.0/24 ! -o "+bridgeName, "MASQUERADE"); err != nil {
		return err
	}

	// Add RETURN rule to cosonet chain for packets incoming from <bridgeName> interface
	if err := addNATRule(chainName, "-i "+bridgeName, "RETURN"); err != nil {
		return err
	}

	return nil
}
