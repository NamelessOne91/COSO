package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

const (
	DefaultRootfsPath = "/tmp/coso/rootfs"
)

// PivotRoot executes a pivot_root syscall, setting a new root filesystem for the calling process
//
// A pivot_root syscall has some requirements:
//  1. They must both be directories
//  2. They must not be on the same filesystem as the current root
//  3. putold must be underneath newroot
//  4. No other filesystem may be mounted on putold
func PivotRoot(newroot string) error {
	putold := filepath.Join(newroot, "/.pivot_root")

	// bind mount newroot to itself - this is a slight hack
	// needed to work around a pivot_root requirement
	err := syscall.Mount(
		newroot,
		newroot,
		"",
		syscall.MS_BIND|syscall.MS_REC,
		"",
	)
	if err != nil {
		return err
	}

	// create putold directory
	if err := os.MkdirAll(putold, 0700); err != nil {
		return err
	}

	// call pivot_root
	if err := syscall.PivotRoot(newroot, putold); err != nil {
		return err
	}

	// ensure current working directory is set to new root
	if err := os.Chdir("/"); err != nil {
		return err
	}

	// umount putold, which now lives at /.pivot_root
	putold = "/.pivot_root"
	if err := syscall.Unmount(
		putold,
		syscall.MNT_DETACH,
	); err != nil {
		return err
	}

	// remove putold
	if err := os.RemoveAll(putold); err != nil {
		return err
	}

	return nil
}

// MountProc executes a mount syscallmount, mounting the proc filesystem at mountpoint (default is /proc).
//
// This is useful when creating a new PID namespace.
// MountProc should not be executed inside the default/global namespace since the /proc mount would
// otherwise mess up existing programs on the system.
func MountProc(newroot string) error {
	source := "proc"
	target := filepath.Join(newroot, "/proc")
	fstype := "proc"
	flags := 0
	data := ""

	os.MkdirAll(target, 0755)
	err := syscall.Mount(
		source,
		target,
		fstype,
		uintptr(flags),
		data,
	)
	return err
}

// VerifyRootfsExists checks a valid root filesystem to use as lower layer has been provided
func VerifyRootfsExists(rootfsPath string) {
	if _, err := os.Stat(rootfsPath); os.IsNotExist(err) {
		sb := strings.Builder{}
		sb.WriteString(fmt.Sprintf("\n'%s' does not exist.\n", rootfsPath))
		sb.WriteString("Please create this directory and unpack a suitable root filesystem inside it.\n")
		sb.WriteString("You can run the following command to set it up with the provided Alpine filesystem for x86_64 architecture:")
		sb.WriteString("\n\nmake fs-setup\n")

		fmt.Println(sb.String())
		os.Exit(1)
	}
}
