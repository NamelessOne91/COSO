package cgroups

import (
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	cosoDir     = "/tmp/coso"
	rootFsPath  = cosoDir + "/rootfs"
	alpineTarGz = "../assets/alpine-minirootfs-3.18.2-x86_64.tar.gz"
)

var _ = Describe("CGroup", func() {

	var (
		cpuQuota string
	)

	BeforeSuite(func() {
		err := extractTarGz(alpineTarGz)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterSuite(func() {
		err := os.RemoveAll(cosoDir)
		Expect(err).NotTo(HaveOccurred())
	})

	BeforeEach(func() {
		cpuQuota = "10000"
	})

	AfterEach(func() {
		Expect(cleanup(rootFsPath + cGroupPath)).To(Succeed())
	})

	Describe("ConfigureCgroup", func() {
		It("Creates a CPU limiting Cgroup", func() {
			err := ConfigureCgroup(rootFsPath, cpuQuota)
			Expect(err).NotTo(HaveOccurred())

			cpuQuotaFilePath := rootFsPath + cGroupPath + cpuQuotaPath
			Expect(cpuQuotaFilePath).To(BeAnExistingFile())

			content, err := os.ReadFile(cpuQuotaFilePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring(cpuQuota))
		})
	})

})

func extractTarGz(tarFilePath string) error {
	if err := os.MkdirAll(rootFsPath, 0755); err != nil {
		return err
	}

	cmd := exec.Command("tar", "-C", rootFsPath, "-xzf", tarFilePath)
	return cmd.Run()
}

func cleanup(path string) error {
	return os.RemoveAll(path)
}
