package namespaces

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/NamelessOne91/coso/cgroups"
	"github.com/NamelessOne91/coso/command"
	"github.com/NamelessOne91/coso/filesystem"
)

// InitNamespaces performs the set of necessary syscalls allowing to run
// a child process in its own isolated namespace(s)
func InitNamespaces() {
	newrootPath := os.Args[1]

	if err := cgroups.ConfigureCgroup(newrootPath, "10000"); err != nil {
		fmt.Printf("Error creating Cgroups - %s\n", err)
		os.Exit(1)
	}

	if err := filesystem.MountProc(newrootPath); err != nil {
		fmt.Printf("Error mounting /proc - %s\n", err)
		os.Exit(1)
	}

	// the pivot_root syscall must happen inside the new mount namespace
	// otherwise, you'll end up changing the host's /
	if err := filesystem.PivotRoot(newrootPath); err != nil {
		fmt.Printf("Error running pivot_root - %s\n", err)
		os.Exit(1)
	}

	if err := syscall.Sethostname([]byte("coso")); err != nil {
		fmt.Printf("Error setting hostname - %s\n", err)
		os.Exit(1)
	}

	// hold launching /bin/sh untill the network is ready
	if err := waitForNetwork(); err != nil {
		fmt.Printf("Error waiting for network - %s\n", err)
		os.Exit(1)
	}

	nsRun()
}

// nsRun starts the system shell inside the namespace
func nsRun() {
	cmd := exec.Command("/bin/sh")

	command.SetupProcessEnv(cmd)

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running the /bin/sh command - %s\n", err)
		os.Exit(1)
	}
}

// waitForNetwork checks for up to 3 seconds if a new network interface has been created
//
// After the namespaces have been created, a veth interface should appear
func waitForNetwork() error {
	maxWait := time.Second * 3
	checkInterval := time.Second
	timeStarted := time.Now()

	for {
		interfaces, err := net.Interfaces()
		if err != nil {
			return err
		}

		// pretty basic check ...
		// > 1 as a lo device will already exist
		if len(interfaces) > 1 {
			return nil
		}

		if time.Since(timeStarted) > maxWait {
			return fmt.Errorf("timeout after %s waiting for network", maxWait)
		}

		time.Sleep(checkInterval)
	}
}
