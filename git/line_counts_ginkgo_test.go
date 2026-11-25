package git

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Diff line counts", func() {
	DescribeTable("excludes non-SLOC files from both patch summaries",
		func(path string, wantAdds, wantDels int) {
			patch := fmt.Sprintf("diff --git a/%[1]s b/%[1]s\n--- a/%[1]s\n+++ b/%[1]s\n@@ -1 +1 @@\n-old\n+new\n", path)
			files := parseCommitFiles(patch)
			Expect(files).To(Equal([]CommitFile{{Path: path, Status: "modified", Adds: wantAdds, Dels: wantDels}}))
			changes, err := ParsePatch(patch)
			Expect(err).NotTo(HaveOccurred())
			Expect(changes).To(HaveLen(1))
			Expect(changes[0].Adds).To(Equal(wantAdds))
			Expect(changes[0].Dels).To(Equal(wantDels))
			if wantAdds == 0 {
				Expect(changes[0].LinesChanged).To(BeEmpty())
			}
		},
		Entry("source", "main.go", 1, 1),
		Entry("documentation", "README.md", 1, 1),
		Entry("unknown text", "notes.custom", 1, 1),
		Entry("vector image", "image.svg", 0, 0),
		Entry("PDF", "report.pdf", 0, 0),
		Entry("uppercase image extension", "image.PNG", 0, 0),
		Entry("archive", "archive.zip", 0, 0),
		Entry("asset path with spaces", "assets/my image.svg", 0, 0),
	)

	DescribeTable("classifies Git numstat path encodings",
		func(raw, destination string) {
			adds, dels, file, binary, ok := parseNumstatLine("10\t5\t" + raw)
			Expect(ok).To(BeTrue())
			Expect(file).To(Equal(destination))
			agg := make(map[string]*groupAccumulator)
			aggregate(commitMeta{subject: "feat: assets"}, []numstatRow{{file: file, adds: adds, dels: dels, isBinary: binary}}, agg, []GroupBy{GroupByType}, false)
			Expect(finalizeGroupSummaries(agg)[0].Count).To(Equal(Count{Commits: 1, Files: 1}))
		},
		Entry("quoted unicode", `"assets/caf\303\251.svg"`, "assets/café.svg"),
		Entry("renamed asset", "assets/{old.svg => new.png}", "assets/new.png"),
		Entry("renamed folder", "{before => after}/image.svg", "after/image.svg"),
	)

	It("counts non-SLOC files in grouped summaries without counting their lines", func() {
		agg := make(map[string]*groupAccumulator)
		aggregate(commitMeta{subject: "feat: update assets"}, []numstatRow{
			{file: "main.go", adds: 2, dels: 1},
			{file: "image.svg", adds: 100, dels: 50},
			{file: "report.pdf", adds: 20, dels: 10},
			{file: "binary.go", adds: 10, dels: 5, isBinary: true},
		}, agg, []GroupBy{GroupByType}, false)
		Expect(finalizeGroupSummaries(agg)).To(Equal([]GroupSummary{{
			Group: []string{"feat"}, Count: Count{Commits: 1, Files: 4, Adds: 2, Dels: 1},
		}}))
	})

	It("never treats binary patch payload as added or deleted source lines", func() {
		patch := "diff --git a/blob b/blob\nGIT binary patch\nliteral 2\n+payload\n-payload\n"
		Expect(parseCommitFiles(patch)).To(Equal([]CommitFile{{Path: "blob", Status: "modified", Binary: true}}))
		changes, err := ParsePatch(patch)
		Expect(err).NotTo(HaveOccurred())
		Expect(changes).To(HaveLen(1))
		Expect(changes[0].Adds).To(BeZero())
		Expect(changes[0].Dels).To(BeZero())
		Expect(changes[0].LinesChanged).To(BeEmpty())
	})
})
