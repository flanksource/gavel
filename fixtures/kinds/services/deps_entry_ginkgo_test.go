package services

import (
	"github.com/flanksource/deps/start"
	"github.com/flanksource/deps/start/state"

	"github.com/flanksource/gavel/fixtures"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("deps entry options", func() {
	ginkgo.DescribeTable("exports the host from the actual connection endpoint", func(endpoint, host string) {
		instance := &start.Instance{State: &state.State{}}
		instance.Connection.URL = endpoint
		outputs, err := depsOutputs(Entry{Name: "db"}, instance, 15432)
		Expect(err).NotTo(HaveOccurred())
		Expect(outputs).To(HaveKeyWithValue("host", host))
		Expect(outputs).To(HaveKeyWithValue("url", endpoint))
	}, ginkgo.Entry("remote SQL Server", "sqlserver://acme:example@moshe-home:15432?database=master", "moshe-home"), ginkgo.Entry("local PostgreSQL", "postgres://localhost:15432/example", "localhost"), ginkgo.Entry("IPv6", "postgres://[::1]:15432/example", "::1"))
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
		Expect(options.BindAddress).To(BeEmpty(), "deps must choose binding for the Docker daemon host")
	})
})
