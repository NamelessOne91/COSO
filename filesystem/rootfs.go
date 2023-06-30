package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

const (
	DefaultRootfsPath = "/tmp/coso/rootfs"
)

// the pivot_root syscall has some requirements:
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

func VerifyRootfsExists(rootfsPath string) {
	if _, err := os.Stat(rootfsPath); os.IsNotExist(err) {
		errorMsg := fmt.Sprintf(
			`"%s" does not exist.\nPlease create this directory and unpack a suitable root filesystem inside it.\n
			You can run the following command to set it up with the provided Alpine filesystem for x86_64 architecture:\nmake fssetup`,
			rootfsPath,
		)

		fmt.Println(errorMsg)
		os.Exit(1)
	}
}
