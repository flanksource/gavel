package lifecycle_test

import (
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/lifecycle"
	lifecycle1 "github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("lifecycle editing", func() {
	load := func(document string) (lifecycle.Lifecycle, error) {
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, "lifecycle.yaml"), []byte(document), 0o600)).To(Succeed())
		return lifecycle1.LoadWith(verify.LifecycleConfig{File: "lifecycle.yaml"}, dir)
	}

	It("reorders inherited steps and disables triage without replacing its prompt", func() {
		def, err := load("order: [run, plan]\ndisabled: [triage]\n")
		Expect(err).NotTo(HaveOccurred())
		Expect(def.StepNames()).To(HaveExactElements("run", "plan", "verify", "triage.new"))
	})

	It("refuses to disable the required verification step", func() {
		_, err := load("disabled: [verify]\n")
		Expect(err).To(MatchError(ContainSubstring("verify")))
	})

	It("refuses an order containing an unknown step", func() {
		_, err := load("order: [missing-step]\n")
		Expect(err).To(MatchError(ContainSubstring("missing-step")))
	})

	It("refuses duplicate ordering entries", func() {
		_, err := load("order: [run, run]\n")
		Expect(err).To(MatchError(ContainSubstring("run")))
	})
})
