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

	DescribeTable("keeps metadata errors on their package and retains passing results", func(frames []string, message string) {
		const badPackage = "example.com/suite"
		const goodPackage = "example.com/plain"
		lines := []goTestEvent{
			{Action: "run", Package: goodPackage, Test: "TestPlain"},
			{Action: "pass", Package: goodPackage, Test: "TestPlain"},
			{Action: "pass", Package: goodPackage},
			{Action: "run", Package: badPackage, Test: "TestOrdinary"},
			{Action: "pass", Package: badPackage, Test: "TestOrdinary"},
		}
		for _, frame := range frames {
			lines = append(lines, goTestEvent{Action: "output", Package: badPackage, Output: frame + "\n"})
		}
		lines = append(lines, goTestEvent{Action: "pass", Package: badPackage})
		var input strings.Builder
		for _, event := range lines {
			data, err := json.Marshal(event)
			Expect(err).NotTo(HaveOccurred())
			input.Write(data)
			input.WriteByte('\n')
		}
		results, err := ParseGoTestBatch(strings.NewReader(input.String()), "", map[string]bool{badPackage: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(results[goodPackage].All()).To(ContainElement(And(HaveField("Name", "TestPlain"), HaveField("Passed", true))))
		Expect(results[badPackage].All()).To(ContainElement(And(HaveField("Name", "TestOrdinary"), HaveField("Passed", true))))
		failures := results[badPackage].Failed()
		Expect(failures).To(HaveLen(1))
		Expect(failures[0].Package).To(Equal(badPackage))
		Expect(failures[0].Framework).To(Equal(Ginkgo))
		Expect(failures[0].Message).To(ContainSubstring(message))
	},
		Entry("missing report", nil, "without Ginkgo metadata"),
		Entry("missing frame", []string{"GAVEL_GINKGO_V1:0/2:e30="}, "incomplete Ginkgo metadata"),
		Entry("malformed frame", []string{"GAVEL_GINKGO_V1:invalid"}, "invalid Ginkgo metadata frame"),
		Entry("duplicate frame", []string{"GAVEL_GINKGO_V1:0/2:e30=", "GAVEL_GINKGO_V1:0/2:e30="}, "duplicate Ginkgo metadata chunk"),
		Entry("invalid base64", []string{"GAVEL_GINKGO_V1:0/1:!"}, "decode Ginkgo metadata"),
		Entry("invalid report JSON", []string{"GAVEL_GINKGO_V1:0/1:eA=="}, "invalid character"),
	)

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
