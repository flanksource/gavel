package testrunner

import (
	"github.com/flanksource/gavel/git/changeset"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Shared selection options", func() {
	It("preserves explicit files and refs for the common resolver", func() {
		opts := RunOptions{WorkDir: "/project", Changed: true, Since: "staged", ChangedFiles: []string{"pkg/a.go"}}
		Expect(opts.ChangeOptions()).To(Equal(changeset.Options{Dir: "/project", Changed: true, Since: "staged", Files: []string{"pkg/a.go"}}))
		Expect(opts.hasChangeSelector()).To(BeTrue())
		Expect(RunOptions{}.hasChangeSelector()).To(BeFalse())
	})
})
