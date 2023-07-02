package network

import (
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	// default path to the cosonet binary after 'make net-setup'
	DefaultCosonetPath = "/usr/local/bin/cosonet"
)

// NetworkConfig hold the needed information to configure  bridge and veth pair devices
// in order to redirect traffic from the default namespace to the new one
type NetworkConfig struct {
	BridgeName        string
	BridgeIP          net.IP
	ContainerIP       net.IP
	Subnet            *net.IPNet
	HostVethName      string
	ContainerVethName string
}

// Manager provides methods to apply a network configuration
type Manager interface {
	Configure(config NetworkConfig, pid int) error
}

// NetworkManager holds the actual Manager struct responsible for
// the nwtwork configuratios of the host and container namespaces
type NetworkManager struct {
	hostManager      Manager
	containerManager Manager
}

func NewNetworkManager(hostManager, containerManager Manager) *NetworkManager {
	return &NetworkManager{
		hostManager:      hostManager,
		containerManager: containerManager,
	}
}

// Configure sets up the needed network devices for the host and container
func (nm *NetworkManager) Configure(netConfig NetworkConfig, pid int) error {
	if err := nm.configureHost(netConfig, pid); err != nil {
		fmt.Printf("Error configuring host network - %s\n", err)
		os.Exit(1)
	}

	if err := nm.configureContainer(netConfig, pid); err != nil {
		fmt.Printf("Error configuring container network - %s\n", err)
		os.Exit(1)
	}

	return nil
}

// configureHost sets up the needed host's networks devices according to the given configuration
func (nm *NetworkManager) configureHost(netConfig NetworkConfig, pid int) error {
	return nm.hostManager.Configure(netConfig, pid)
}

// configureContainer sets up the needed devices for the new network namespace, according to the given configuration
func (nm *NetworkManager) configureContainer(netConfig NetworkConfig, pid int) error {
	return nm.containerManager.Configure(netConfig, pid)
}

// VerifyNetworkManagerExists checks whether the  binary responsible for the correct creation of network devices
// is present at the given path
func VerifyNetworkManagerExists(executablePath string) {
	if _, err := os.Stat(executablePath); os.IsNotExist(err) {
		sb := strings.Builder{}
		sb.WriteString(fmt.Sprintf("Unable to find the executable at '%s'.\n", executablePath))
		sb.WriteString("An external binary used to configure networking is needed.\n")
		sb.WriteString("You must build it, chown it to the root user and apply the setuid bit.\n")
		sb.WriteString("This can be done for the provided cosonet tool as follows:\n\nmake net-setup\n")

		fmt.Println(sb.String())
		os.Exit(1)
	}
}

// interfaceExists check whether a network device with the given name exists
func interfaceExists(name string) bool {
	_, err := net.InterfaceByName(name)
	return err == nil
}
