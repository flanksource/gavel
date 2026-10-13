package processruntime

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestProcessRuntime(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "Process runtime") }

func ptr[T any](value T) *T { return &value }

var _ = Describe("Runtime configuration", func() {
	It("reports sorted environment names without their values", func() {
		Expect((Options{Env: map[string]string{"Z_KEY": "private-z", "A_KEY": "private-a"}, Nice: ptr(5)}).Summary()).To(Equal(Summary{Env: []string{"A_KEY", "Z_KEY"}, Nice: ptr(5)}))
	})
	It("merges environment keys and retains explicit scalar resets", func() {
		base := Options{Env: map[string]string{"KEEP": "base", "CLEAR": "secret"}, CPU: ptr(200.0), Mem: ptr("2GiB"), Nice: ptr(10)}
		override := Options{Env: map[string]string{"CLEAR": "", "ADD": "new"}, CPU: ptr(0.0), Mem: ptr(""), Nice: ptr(0)}
		merged := base.Merge(override)
		Expect(merged).To(Equal(Options{Env: map[string]string{"KEEP": "base", "CLEAR": "", "ADD": "new"}, CPU: ptr(0.0), Mem: ptr(""), Nice: ptr(0)}))
		merged.Env["KEEP"] = "changed"
		Expect(base.Env["KEEP"]).To(Equal("base"))
	})

	DescribeTable("rejects invalid runtime fields", func(opts Options, field string) {
		Expect(opts.Validate()).To(MatchError(ContainSubstring(field)))
	},
		Entry("environment key", Options{Env: map[string]string{"BAD=KEY": "value"}}, "env"),
		Entry("environment value", Options{Env: map[string]string{"KEY": "value\x00"}}, "env"),
		Entry("niceness", Options{Nice: ptr(20)}, "nice"),
		Entry("negative CPU", Options{CPU: ptr(-1.0)}, "cpu"),
		Entry("memory quantity", Options{Mem: ptr("nonsense")}, "mem"),
		Entry("negative timeout", Options{Timeout: ptr("-1s")}, "timeout"),
		Entry("zero timeout", Options{Timeout: ptr("0s")}, "timeout"),
	)
})
