package services

import (
	"github.com/flanksource/deps/start"

	"github.com/flanksource/gavel/fixtures"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("deps entry options", func() {
	ginkgo.It("hands deps the step's ChildEnv as the host env, under the entry's own env", func() {
		const fixtureVar, entryVar = "GAVEL_SPEC_FIXTURE_VAR", "GAVEL_SPEC_ENTRY_VAR"
		step := &fixtures.StepContext{
			Scope:  fixtures.NewStepScope([]string{fixtureVar + "=from-fixture", entryVar + "=from-fixture"}),
			TmpDir: ginkgo.GinkgoT().TempDir(),
		}
		entry := Entry{Name: "db", Deps: "postgres", Env: map[string]string{entryVar: "from-entry"}}

		options, err := start.ResolveOptions(depsOptions(step, entry, ginkgo.GinkgoT().TempDir(), 15432))

		Expect(err).NotTo(HaveOccurred())
		Expect(options.HostEnv).To(Equal(step.ChildEnv()))
		Expect(options.HostEnv).To(HaveKeyWithValue(fixtureVar, "from-fixture"))
		Expect(options.Env).To(Equal(map[string]string{entryVar: "from-entry"}))
	})
})
