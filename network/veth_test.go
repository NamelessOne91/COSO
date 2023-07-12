package network

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/vishvananda/netlink"
)

const (
	testHostVeth = "test-veth"
	testPeerVeth = "test-peer-veth"
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
})
