package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/NamelessOne91/coso/command"
	"github.com/NamelessOne91/coso/filesystem"
	"github.com/NamelessOne91/coso/namespaces"
	"github.com/docker/docker/pkg/reexec"
)

func init() {
	reexec.Register("nsInit", namespaces.InitNamespaces)
	if reexec.Init() {
		// avoid infinite loops of the program rexec-uting itself
		os.Exit(0)
	}
}

func main() {
	var rootfsPath string
	flag.StringVar(&rootfsPath, "rootfs", filesystem.DefaultRootfsPath, "Path to the root filesystem to use")
	flag.Parse()

	filesystem.VerifyRootfsExists(rootfsPath)

	// rexec is used to bypass forking limitations of Go
	// allowing to run code after the namespace creation but before the process starts
	cmd := reexec.Command("nsInit", rootfsPath)
	command.SetupProcessEnv(cmd)
	command.SetupNewNamespaces(cmd)

	// syscalls here
	// 1) clone: creates process
	// 2) setns: allows the calling process to join an existing namespace
	// 3) unshare: moves the calling process to a new namespace
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running the /bin/sh command - %s\n", err)
		os.Exit(1)
	}
}
