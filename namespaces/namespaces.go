package namespaces

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/NamelessOne91/coso/command"
	"github.com/NamelessOne91/coso/filesystem"
)

// InitNamespaces performs the set of necessary syscalls allowing to run
// a child process in its own isolated namespace(s)
func InitNamespaces() {
	newrootPath := os.Args[1]

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
