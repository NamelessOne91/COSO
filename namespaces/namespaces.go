package namespaces

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/NamelessOne91/coso/filesystem"
)

const (
	hostname = "coso"
)

// Flags is the bit pattern passed to the clone syscall to create new namespaces
var Flags uintptr = syscall.CLONE_NEWNS | syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC |
	syscall.CLONE_NEWPID | syscall.CLONE_NEWNET | syscall.CLONE_NEWUSER

func InitNamespaces() {
	newrootPath := os.Args[1]

	if err := filesystem.MountProc(newrootPath); err != nil {
		fmt.Printf("Error mounting /proc - %s\n", err)
		os.Exit(1)
	}

	if err := filesystem.PivotRoot(newrootPath); err != nil {
		fmt.Printf("Error running pivot_root - %s\n", err)
		os.Exit(1)
	}

	if err := syscall.Sethostname([]byte(hostname)); err != nil {
		fmt.Printf("Error setting hostname - %s\n", err)
		os.Exit(1)
	}

	nsRun()
}

func nsRun() {
	cmd := exec.Command("/bin/sh")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = []string{fmt.Sprintf("PS1=-[%s]- # ", hostname)}

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running the /bin/sh command - %s\n", err)
		os.Exit(1)
	}
}
