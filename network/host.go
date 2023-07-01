package network

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
	// create bridge device
	bridge, err := h.bridgeManager.Create(netConfig.BridgeName, netConfig.BridgeIP, netConfig.Subnet)
	if err != nil {
		return err
	}

	// create veth devices pair
	hostVeth, containerVeth, err := h.vethManager.Create(netConfig.HostVethName, netConfig.ContainerVethName)
	if err != nil {
		return err
	}

	// attach host veth device to the bridge device
	err = h.bridgeManager.Attach(bridge, hostVeth)
	if err != nil {
		return err
	}

	// move the peer veth device to the new namespace
	err = h.vethManager.MoveToNetworkNamespace(containerVeth, pid)
	if err != nil {
		return err
	}

	return nil
}
