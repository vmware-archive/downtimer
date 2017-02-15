package main_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Downtimer", func() {
	var (
		binaryPath string
	)
	BeforeSuite(func() {
		var err error
		binaryPath, err = gexec.Build("github.com/pivotal-cf/downtimer")
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("general operation", func() {
		// TODO: Write better tests
		It("probes pivotal.io", func() {
			command := exec.Command(binaryPath, "-u", "http://pivotal.io", "-d", "4s", "-i", "2s")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session, 4).Should(gexec.Exit(0))
		})
	})

	Describe("commandline opts", func() {
		It("returns 1 on invalid params", func() {
			command := exec.Command(binaryPath, "invalid", "input")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
		})

		Context("when bosh parameters are supplied", func() {
			It("requires all four bosh parameters if any are specified", func() {
				command := exec.Command(binaryPath, "-u", "http://pivotal.io", "-d", "3s", "-b", "bosh-director.pivotal.io")
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(gexec.Exit(1))
			})
			It("checks bosh if bosh director/user/password/ca-certs is specified", func() {
				command := exec.Command(binaryPath,
					"-u", "http://pivotal.io",
					"-d", "5s",
					"-b", "bosh-director.pivotal.io",
					"-U", "bosh-user",
					"-P", "bosh-password",
					"-c", "/dev/null",
				)
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).ShouldNot(gexec.Exit())
			})
		})
	})
})
