package command

import (
	"os"
	"os/exec"
	"syscall"
)

// Flags is the bit pattern passed to the clone syscall to create new namespaces
var flags uintptr = syscall.CLONE_NEWNS | syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC |
	syscall.CLONE_NEWPID | syscall.CLONE_NEWNET | syscall.CLONE_NEWUSER

// SetupProcessEnv pipes stdin/stdout/err from the calling process and sets
// the default interaction prompt (PS1) env variable
func SetupProcessEnv(cmd *exec.Cmd) {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = []string{"PS1=-[coso]- # "}
}

// SetupNewNamespaces set the system flags needed to run the process inside new namespaces
func SetupNewNamespaces(cmd *exec.Cmd) {
	// define clone flags, aka namespaces
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: flags,
		// map root (ID 0) in the new User namespace
		// to the user and group IDs who invoked coso
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getuid(),
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getgid(),
				Size:        1,
			},
		},
	}
}
