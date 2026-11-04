package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("comments CLI registration", func() {
	It("exposes structured creation, native publication, lint and explicit provisioning commands", func() {
		for _, path := range [][]string{{"comments", "create"}, {"comments", "publish"}, {"comments", "provision"}, {"comments", "dispatch"}, {"comment-lint", "preview"}, {"comment-lint", "run"}} {
			command, remaining, err := rootCmd.Find(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(remaining).To(BeEmpty())
			Expect(command.Name()).To(Equal(path[len(path)-1]))
		}
		create, _, err := rootCmd.Find([]string{"comments", "create"})
		Expect(err).NotTo(HaveOccurred())
		Expect(create.Flags().Lookup("target")).NotTo(BeNil())
		publish, _, err := rootCmd.Find([]string{"comments", "publish"})
		Expect(err).NotTo(HaveOccurred())
		Expect(publish.Flags().Lookup("expectedVersion").Value.Type()).To(Equal("int64"))
	})
})
