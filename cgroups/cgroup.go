package cgroups

import (
	"os"
)

const (
	cGroupPath   = "/sys/fs/cgroup/cpu/coso_cgroup"
	cpuQuotaPath = "/cpu.cfs_quota_us"
)

// ConfigureCGroup writes Cgroup creates a custom Cgroup, and the necessary files needed
// to limit resource usage, inside the /fs/sys/cgroup directory available at the specified rooth FS path
func ConfigureCgroup(rootPath, cpuQuota string) error {
	// Create a new cgroup
	cgroup := rootPath + cGroupPath
	if err := os.MkdirAll(cgroup, 0755); err != nil && !os.IsExist(err) {
		return err
	}

	if err := setCpuQuota(cgroup, cpuQuota); err != nil {
		return err
	}

	return nil
}

// setCPUQuota writes a file limiting the CPU bandwidth available to a group up to quota microseconds
func setCpuQuota(cGroupPath, quota string) error {
	// Set the cgroup resource limits
	if err := os.WriteFile(cGroupPath+cpuQuotaPath, []byte(quota), 0644); err != nil {
		return err
	}
	return nil
}
