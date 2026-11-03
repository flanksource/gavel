package lint

import (
	"context"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/processruntime"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func runtimeValue[T any](value T) *T { return &value }

var _ = Describe("Linter runtime capabilities", func() {
	var dir string
	BeforeEach(func() {
		dir = GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
	})

	DescribeTable("rejects process controls before running an embedded linter", func(runtime processruntime.Options) {
		linter := &entityLinter{}
		result := linters.RunLinterWithTask(context.Background(), nil, linter, linters.RunOptions{WorkDir: dir, Runtime: runtime})
		Expect(result.Error).To(ContainSubstring("fake-entities runtime: in-process linters do not support"))
		Expect(linter.runs).To(BeZero())
		Expect(linter.opts).To(BeNil())
	},
		Entry("environment", processruntime.Options{Env: map[string]string{"EXAMPLE": "value"}}),
		Entry("niceness", processruntime.Options{Nice: runtimeValue(5)}),
		Entry("CPU", processruntime.Options{CPU: runtimeValue(100.0)}),
		Entry("memory", processruntime.Options{Mem: runtimeValue("1GiB")}),
	)

	DescribeTable("accepts context timeouts for JSON and embedded linters", func(name, timeout string) {
		var linter linters.Linter = &entityLinter{}
		if name == "json" {
			var exists bool
			linter, exists = builtinLinterRegistry(dir).Get(name)
			Expect(exists).To(BeTrue())
		}
		_, _, err := linters.PrepareCommand(linter, linters.RunOptions{WorkDir: dir, Runtime: processruntime.Options{Timeout: &timeout}})
		Expect(err).NotTo(HaveOccurred())
	},
		Entry("JSON timeout", "json", "5s"),
		Entry("JSON timeout disabled", "json", ""),
		Entry("embedded timeout", "embedded", "5s"),
		Entry("embedded timeout disabled", "embedded", ""),
	)
})
