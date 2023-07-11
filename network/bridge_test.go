package network

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"testing"

	"github.com/vishvananda/netlink"
)

const (
	bridgeName = "test-bridge"
	bridgeCIDR = "10.10.10.1/24"
)

func cleanup(name string) error {
	if _, err := net.InterfaceByName(name); err == nil {
		link, err := netlink.LinkByName(name)
		if err != nil {
			return err
		}
		return netlink.LinkDel(link)
	}
	return nil
}

func TestBridgeCreate(t *testing.T) {
	bridge := NewBridge()

	bridgeIP, bridgeSubnet, err := net.ParseCIDR(bridgeCIDR)
	if err != nil {
		t.Errorf("Failed to parse bridge CIDR with error: %s", err)
	}

	firstBI, err := bridge.Create(bridgeName, bridgeIP, bridgeSubnet)
	if err != nil {
		t.Errorf("Failed to create bridge with error: %s", err)
	}
	defer func() {
		err := cleanup(bridgeName)
		if err != nil {
			t.Errorf("Failed to cleanup bridge device with error: %s", err)
		}
	}()

	// should return the same device if a bridge with the given name already exists
	bridgeInterface, err := bridge.Create(bridgeName, bridgeIP, bridgeSubnet)
	if err != nil {
		t.Errorf("Expected no error when bridge with the same name exists - got: %s", err)
	}
	if firstBI.Name != bridgeInterface.Name {
		t.Errorf("Expected bridge interface with the same name - got %s and %s", firstBI.Name, bridgeInterface.Name)
	}

	// correct name
	if bridgeInterface.Name != bridgeName {
		t.Errorf("Expected bridge interface name to be %s - got %s", bridgeName, bridgeInterface.Name)
	}

	// device is in 'UP' state
	var stdout bytes.Buffer
	cmd := exec.Command("sh", "-c", fmt.Sprintf("ip link list %s", bridgeName))
	cmd.Stdout = &stdout
	err = cmd.Run()
	if err != nil {
		t.Errorf("Error executing bash command to check bridge's state: %s", err)
	}
	if bytes.Contains(stdout.Bytes(), []byte("state DOWN")) {
		t.Error("Bridge device is in 'DOWN' state")
	}

	// check bridge address
	bridgeAddresses, err := bridgeInterface.Addrs()
	if err != nil {
		t.Errorf("Error checking bridge device address: %s", err)
	}
	if len(bridgeAddresses) != 2 {
		t.Errorf("Expected len(bridgeAddreses) to be 2 - got %d", len(bridgeAddresses))
	}
	if bridgeAddresses[0].String() != bridgeCIDR {
		t.Errorf("Expected bridge address to be %s - got %s", bridgeCIDR, bridgeAddresses[0].String())
	}
}

func TestBridgeAttach(t *testing.T) {
	bridge := NewBridge()
	veth := NewVeth()

	bridgeIP, bridgeSubnet, err := net.ParseCIDR(bridgeCIDR)
	if err != nil {
		t.Errorf("Failed to parse bridge CIDR with error: %s", err)
	}

	bridgeI, err := bridge.Create(bridgeName, bridgeIP, bridgeSubnet)
	if err != nil {
		t.Errorf("Failed to create bridge with error: %s", err)
	}
	defer func() {
		err := cleanup(bridgeName)
		if err != nil {
			t.Errorf("Failed to cleanup bridge device with error: %s", err)
		}
	}()

	hostVethI, _, err := veth.Create(testHostVeth, testPeerVeth)
	if err != nil {
		t.Errorf("Failed to create veth pair with error: %s", err)
	}
	defer func() {
		err := cleanup(testHostVeth)
		if err != nil {
			t.Errorf("Failed to cleanup veth devices with error: %s", err)
		}
	}()

	err = bridge.Attach(bridgeI, hostVethI)
	if err != nil {
		t.Errorf("Failed to attach veth to bridge with error: %s", err)
	}

	// this should now exists
	confiFilePath := fmt.Sprintf("/sys/class/net/%s/master", testHostVeth)
	_, err = os.Stat(confiFilePath)
	if err != nil {
		t.Errorf("Failed to retrieve %s with error: %s", confiFilePath, err)
	}
}
