package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/NamelessOne91/coso/command"
	"github.com/NamelessOne91/coso/filesystem"
	"github.com/NamelessOne91/coso/namespaces"
	"github.com/NamelessOne91/coso/network"
)

func init() {
	command.Register("nsInit", namespaces.InitNamespaces)
	if command.Init() {
		// avoid infinite loops of the program rexec-uting itself
		os.Exit(0)
	}
}

func main() {
	var rootfsPath, cosonetPath string
	flag.StringVar(&rootfsPath, "rootfs", filesystem.DefaultRootfsPath, "Path to the root filesystem to use")
	flag.StringVar(&cosonetPath, "cosonet", network.DefaultCosonetPath, "Path to the cosonet binary")
	flag.Parse()

	filesystem.VerifyRootfsExists(rootfsPath)
	network.VerifyCosonetExists(cosonetPath)

	// rexec is used to bypass forking limitations of Go
	// allowing to run code after the namespace creation but before the process starts
	cmd := command.NewReexecCommand("nsInit", rootfsPath)

	// syscalls here
	// 1) clone: creates process
	// 2) setns: allows the calling process to join an existing namespace
	// 3) unshare: moves the calling process to a new namespace

	// not blocking
	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting the reexec.Command - %s\n", err)
		os.Exit(1)
	}
	// child process PID
	pid := fmt.Sprintf("%d", cmd.Process.Pid)

	// executed in the host namespace
	cosoNetCmd := exec.Command(cosonetPath, "-pid", pid)
	if err := cosoNetCmd.Run(); err != nil {
		fmt.Printf("Error running cosonet - %s\n", err)
		os.Exit(1)
	}

	if err := cmd.Wait(); err != nil {
		fmt.Printf("Error waiting for reexec.Command - %s\n", err)
	}
}
