/* Copyright (C) 2017-Present Pivotal Software, Inc. All rights reserved.

This program and the accompanying materials are made available under
the terms of the under the Apache License, Version 2.0 (the "License”);
you may not use this file except in compliance with the License.

You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License. */

package main_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
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
			Eventually(session, 5).Should(gexec.Exit(0))
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
				Eventually(session.Err).Should(gbytes.Say("all bosh options must be specified"))
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
