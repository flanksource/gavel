package main

import (
	"bytes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("connection catalog", func() {
	It("registers all shared sources and preserves database-free file operations", func() {
		command := newConnectionsCommand()
		output := new(bytes.Buffer)
		command.SetOut(output)
		GinkgoT().Setenv("XDG_CONFIG_HOME", GinkgoT().TempDir())
		previous := workingDir
		workingDir = GinkgoT().TempDir()
		DeferCleanup(func() { workingDir = previous })
		command.SetArgs([]string{"list"})
		Expect(command.Execute()).To(Succeed())
		Expect(command.Flag("source").Usage).To(ContainSubstring("server"))
		add, _, err := command.Find([]string{"add", "http"})
		Expect(err).NotTo(HaveOccurred())
		Expect(add.Name()).To(Equal("http"))
	})
})
