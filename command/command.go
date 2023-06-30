// Package command is a Docker's reexec inspired package allowing to bypass Go's forking limitations
// and run arbitrary code after new namespaces have been created but a process still has to be executed
package command

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/unix"
)

const (
	// flags is the bit pattern passed to the clone syscall to create new namespaces
	flags uintptr = syscall.CLONE_NEWNS | syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC |
		syscall.CLONE_NEWPID | syscall.CLONE_NEWNET | syscall.CLONE_NEWUSER
	// self is the path to the current process' binary.
	self = "/proc/self/exe"
)

// registeredInitializers is a map of custom function mapped to an argument
var registeredInitializers = make(map[string]func())

// Register adds an initialization func under the specified name
func Register(name string, initializer func()) {
	if _, exists := registeredInitializers[name]; exists {
		panic(fmt.Sprintf("reexec func already registered under name %q", name))
	}

	registeredInitializers[name] = initializer
}

// Init is called as the first part of the exec process and returns true if an
// initialization function was called.
func Init() bool {
	initializer, exists := registeredInitializers[os.Args[0]]
	if exists {
		initializer()

		return true
	}
	return false
}

// NewReexecCommand return a pointer to an exec.Cmd which will be executed inside new namespaces
func NewReexecCommand(args ...string) *exec.Cmd {
	cmd := &exec.Cmd{
		Path: self,
		Args: args,
		SysProcAttr: &syscall.SysProcAttr{
			Pdeathsig: unix.SIGTERM,
		},
	}
	setupNewNamespaces(cmd)
	SetupProcessEnv(cmd)
	return cmd
}

// SetupProcessEnv pipes stdin/stdout/err from the calling process and sets
// the default interaction prompt (PS1) env variable
func SetupProcessEnv(cmd *exec.Cmd) {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = []string{"PS1=-[coso]- # "}
}

// setupNewNamespaces set the system flags needed to run the process inside new namespaces
func setupNewNamespaces(cmd *exec.Cmd) {
	// define clone flags, aka namespaces
	cmd.SysProcAttr.Cloneflags = flags
	// map root (ID 0) in the new User namespace
	// to the user and group IDs who invoked coso
	cmd.SysProcAttr.UidMappings = []syscall.SysProcIDMap{
		{
			ContainerID: 0,
			HostID:      os.Getuid(),
			Size:        1,
		},
	}
	cmd.SysProcAttr.GidMappings = []syscall.SysProcIDMap{
		{
			ContainerID: 0,
			HostID:      os.Getgid(),
			Size:        1,
		},
	}
}
