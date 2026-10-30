package changeset

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Git commit resolution", func() {
	DescribeTable("accepts only a full object ID from Git", func(output string, valid bool) {
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, "git"), []byte("#!/bin/sh\nprintf '%s\\n' \"$GAVEL_TEST_GIT_OUTPUT\"\n"), 0700)).To(Succeed())
		GinkgoT().Setenv("PATH", dir)
		GinkgoT().Setenv("GAVEL_TEST_GIT_OUTPUT", output)
		sha, err := commit(context.Background(), dir, "HEAD")
		if valid {
			Expect(err).NotTo(HaveOccurred())
			Expect(sha).To(Equal(output))
		} else {
			Expect(err).To(MatchError(ContainSubstring("invalid Git commit ID")))
			Expect(sha).To(BeEmpty())
		}
	},
		Entry("SHA-1", strings.Repeat("a", 40), true),
		Entry("SHA-256", strings.Repeat("b", 64), true),
		Entry("an abbreviated ID", "abcdef0", false),
		Entry("an option", "--output=outside", false),
		Entry("an empty response", "", false),
		Entry("multiple IDs", strings.Repeat("a", 40)+"\n"+strings.Repeat("b", 40), false),
	)
})
