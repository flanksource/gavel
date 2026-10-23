package parsers

import (
	"encoding/json"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Go test batch metadata", func() {
	It("retains cached Ginkgo context when filtering results", func() {
		original := Test{Cached: true, Context: GinkgoContext{Labels: []string{"database"}, NumAttempts: 2}, Children: Tests{{Name: "spec", Passed: true}}}
		filtered := original.Filter(TestFilter{})
		Expect(filtered.Cached).To(BeTrue())
		Expect(filtered.Context).To(Equal(original.Context))
	})

	It("fails when a Ginkgo report is missing a frame", func() {
		lines := []goTestEvent{
			{Action: "output", Package: "example.com/suite", Output: "GAVEL_GINKGO_V1:0/2:e30=\n"},
			{Action: "pass", Package: "example.com/suite"},
		}
		var input strings.Builder
		for _, event := range lines {
			data, err := json.Marshal(event)
			Expect(err).NotTo(HaveOccurred())
			input.Write(data)
			input.WriteByte('\n')
		}
		_, err := ParseGoTestBatch(strings.NewReader(input.String()), "", map[string]bool{"example.com/suite": true})
		Expect(err).To(MatchError(ContainSubstring("incomplete Ginkgo metadata")))
	})

	DescribeTable("attributes a test-variant build failure to the package under test",
		func(variant string) {
			const pkg = "example.com/suite"
			const compileError = "labels.go:28:6: Labels already declared through dot-import\n"
			lines := []goTestEvent{
				{Action: "build-output", ImportPath: variant, Output: "# " + variant + "\n"},
				{Action: "build-output", ImportPath: variant, Output: compileError},
				{Action: "build-fail", ImportPath: variant},
				{Action: "start", Package: pkg},
				{Action: "output", Package: pkg, Output: "FAIL\t" + pkg + " [build failed]\n"},
				{Action: "fail", Package: pkg},
			}
			var input strings.Builder
			for _, event := range lines {
				data, err := json.Marshal(event)
				Expect(err).NotTo(HaveOccurred())
				input.Write(data)
				input.WriteByte('\n')
			}
			results, err := ParseGoTestBatch(strings.NewReader(input.String()), "", map[string]bool{pkg: true})
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(HaveLen(1))
			Expect(results[pkg].All()).To(ContainElement(SatisfyAll(
				HaveField("Name", "Build Failed"),
				HaveField("Failed", true),
				HaveField("Message", ContainSubstring(compileError)),
			)))
		},
		Entry("internal test package", "example.com/suite [example.com/suite.test]"),
		Entry("external test package", "example.com/suite_test [example.com/suite.test]"),
	)
})
