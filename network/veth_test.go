package network

import (
	"bytes"
	"fmt"
	"os/exec"
	"testing"
)

const (
	testHostVeth = "test-veth"
	testPeerVeth = "test-peer-veth"
)

func TestVethCreate(t *testing.T) {
	veth := NewVeth()

	firstHostVI, firstContainerVI, err := veth.Create(testHostVeth, testPeerVeth)
	if err != nil {
		t.Errorf("Failed to create veth pair with error: %s", err)
	}
	defer func() {
		err := cleanup(testHostVeth)
		if err != nil {
			t.Errorf("Failed to cleanup veth devices with error: %s", err)
		}
	}()

	// should return the same device if a veth with the given name already exists
	hostVeth, containerVeth, err := veth.Create(testHostVeth, testPeerVeth)
	if err != nil {
		t.Errorf("Expected no error when veth with the same name exists - got: %s", err)
	}
	if firstHostVI.Name != hostVeth.Name || firstContainerVI.Name != containerVeth.Name {
		t.Errorf("Expected veth interfaces with the same name - got host %s and %s - container %s and %s", firstHostVI.Name, hostVeth.Name, firstContainerVI.Name, containerVeth.Name)
	}

	// correct veth names
	if hostVeth.Name != testHostVeth || containerVeth.Name != testPeerVeth {
		t.Errorf("Expected veth devices to have name %s and %s - got %s and %s", testHostVeth, testPeerVeth, hostVeth.Name, containerVeth.Name)
	}

	// device is in 'UP' state
	var stdout bytes.Buffer
	cmd := exec.Command("sh", "-c", fmt.Sprintf("ip link list %s", testHostVeth))
	cmd.Stdout = &stdout
	err = cmd.Run()
	if err != nil {
		t.Errorf("Error executing bash command to check veth's state: %s", err)
	}
	if bytes.Contains(stdout.Bytes(), []byte("state DOWN")) {
		t.Error("Veth device is in 'DOWN' state")
	}
}
