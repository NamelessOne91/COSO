package network

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/vishvananda/netlink"
)

const (
	testHostVeth     = "test-veth"
	testPeerVeth     = "test-peer-veth"
	netNamespaceName = "testNetNs"
)

var _ = Describe("Veth", func() {
	var (
		veth *Veth
	)

	BeforeEach(func() {
		veth = NewVeth()
	})

	AfterEach(func() {
		Expect(cleanup(testHostVeth)).To(Succeed())
	})

	Describe("Create", func() {
		It("creates a veth pair using the provided names", func() {
			hostVeth, containerVeth, err := veth.Create(testHostVeth, testPeerVeth)
			Expect(err).NotTo(HaveOccurred())

			Expect(hostVeth.Name).To(Equal(testHostVeth))
			Expect(containerVeth.Name).To(Equal(testPeerVeth))
		})

		It("brings the veth link up", func() {
			_, _, err := veth.Create(testHostVeth, testPeerVeth)
			Expect(err).NotTo(HaveOccurred())

			stdout := gbytes.NewBuffer()
			cmd := exec.Command("sh", "-c", "ip link show "+testHostVeth)
			_, err = gexec.Start(cmd, stdout, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Consistently(stdout).ShouldNot(gbytes.Say(",DOWN"))
		})

		Context("when a veth pair using the provided name prefix already exists", func() {
			BeforeEach(func() {
				_, _, err := veth.Create(testHostVeth, testPeerVeth)
				Expect(err).NotTo(HaveOccurred())
			})

			It("doesn't error", func() {
				_, _, err := veth.Create(testHostVeth, testPeerVeth)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the host and container veths", func() {
				hostVeth, containerVeth, err := veth.Create(testHostVeth, testPeerVeth)
				Expect(err).NotTo(HaveOccurred())

				Expect(hostVeth.Name).To(Equal(testHostVeth))
				Expect(containerVeth.Name).To(Equal(testPeerVeth))
			})

			Context("and the link is already up", func() {
				BeforeEach(func() {
					link, err := netlink.LinkByName(testHostVeth)
					Expect(err).NotTo(HaveOccurred())
					Expect(netlink.LinkSetUp(link)).To(Succeed())
				})

				It("doesn't error", func() {
					_, _, err := veth.Create(testHostVeth, testPeerVeth)
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns the host and container veths", func() {
					hostVeth, containerVeth, err := veth.Create(testHostVeth, testPeerVeth)
					Expect(err).NotTo(HaveOccurred())

					Expect(hostVeth.Name).To(Equal(testHostVeth))
					Expect(containerVeth.Name).To(Equal(testPeerVeth))
				})
			})
		})
	})

	Describe("MoveToNetworkNamespace", func() {
		var (
			containerVeth *net.Interface
			parentPid     int
			pid           int
		)

		BeforeEach(func() {
			var err error
			_, containerVeth, err = veth.Create(testHostVeth, testPeerVeth)
			Expect(err).NotTo(HaveOccurred())

			createNetNamespace(netNamespaceName)
			parentPid, pid = runCmdInNetNamespace(netNamespaceName, "sleep 1000")
		})

		AfterEach(func() {
			killCmd(parentPid)
			destroyNetNamespace(netNamespaceName)
		})

		It("moves the container's side of the veth into the namespace identified by the pid", func() {
			err := veth.MoveToNetworkNamespace(containerVeth, pid)
			Expect(err).NotTo(HaveOccurred())

			ensureOutputForCommand(fmt.Sprintf("ip netns exec %s ip addr", netNamespaceName), testPeerVeth)
		})

		Context("when the veth doesn't exist", func() {
			It("returns a descriptive error", func() {
				nonexistentVeth := &net.Interface{Name: "nonexistentVeth"}
				err := veth.MoveToNetworkNamespace(nonexistentVeth, pid)

				Expect(err.Error()).To(ContainSubstring("Link not found"))
			})
		})

		Context("when the process doesn't exist", func() {
			It("returns a descriptive error", func() {
				err := veth.MoveToNetworkNamespace(containerVeth, -1)

				Expect(err.Error()).To(ContainSubstring("no such process"))
			})
		})
	})
})

func createNetNamespace(name string) {
	cmdString := fmt.Sprintf("ip netns add %s", name)
	cmd := exec.Command("sh", "-c", cmdString)
	Expect(cmd.Run()).To(Succeed())
}

func destroyNetNamespace(name string) {
	cmdString := fmt.Sprintf("ip netns delete %s", name)
	cmd := exec.Command("sh", "-c", cmdString)
	Expect(cmd.Run()).To(Succeed())
}

func killCmd(pid int) {
	process, err := os.FindProcess(pid)
	Expect(err).NotTo(HaveOccurred())

	Expect(process.Kill()).To(Succeed())
}

func runCmdInNetNamespace(netNamespaceName string, cmdPathAndArgs string) (int, int) {
	cmdString := fmt.Sprintf("ip netns exec %s %s", netNamespaceName, cmdPathAndArgs)
	cmd := exec.Command("sh", "-c", cmdString)
	Expect(cmd.Start()).To(Succeed())

	parentPid := cmd.Process.Pid

	// super gross
	cmd = exec.Command("sh", "-c", fmt.Sprintf("ps --ppid %d | tail -n 1 | awk '{print $1}'", parentPid))
	pidBytes, err := cmd.Output()
	Expect(err).NotTo(HaveOccurred())

	pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	Expect(err).NotTo(HaveOccurred())

	return parentPid, pid
}

func ensureOutputForCommand(command, expectedOutput string) {
	stdout := gbytes.NewBuffer()
	cmd := exec.Command("sh", "-c", command)
	_, err := gexec.Start(cmd, stdout, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	Eventually(stdout).Should(gbytes.Say(expectedOutput))
}
