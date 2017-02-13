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

	Describe("commandline opts", func() {
		It("returns 1 on invalid params", func() {
			command := exec.Command(binaryPath, "invalid", "input")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
		})
	})
})
