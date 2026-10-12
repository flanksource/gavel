package fixtures

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("fixture SQL profile", func() {
	It("sums statement time and retains the slowest query", func() {
		path := filepath.Join(GinkgoT().TempDir(), "sqlprofile.jsonl")
		Expect(os.WriteFile(path, []byte("{\"duration_ns\":10000000,\"rows\":2,\"slow\":false}\n{\"duration_ns\":30000000,\"rows\":1,\"slow\":true}\n"), 0o600)).To(Succeed())
		profile, err := ReadSQLProfile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(profile).To(Equal(&SQLProfile{Path: path, QueryCount: 2, SlowQueryCount: 1, TotalDurationMS: 40, MaxQueryMS: 30}))
	})

	It("rejects a malformed event without a duration", func() {
		path := filepath.Join(GinkgoT().TempDir(), "sqlprofile.jsonl")
		Expect(os.WriteFile(path, []byte("{}\n"), 0o600)).To(Succeed())
		_, err := ReadSQLProfile(path)
		Expect(err).To(MatchError(ContainSubstring("missing or negative duration")))
	})

	It("retains SQL statements and bound parameters", func() {
		path := filepath.Join(GinkgoT().TempDir(), "sqlprofile.jsonl")
		Expect(os.WriteFile(path, []byte("{\"duration_ns\":10000000,\"rows\":1,\"slow\":false,\"sql\":\"SELECT * FROM users WHERE id = ?\",\"params\":[\"42\"]}\n"), 0o600)).To(Succeed())
		profile, err := ReadSQLProfile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(profile.Statements).To(Equal([]SQLProfileStatement{{DurationMS: 10, Rows: 1, SQL: "SELECT * FROM users WHERE id = ?", Params: []string{"42"}}}))
	})

	It("represents a statement without parameters as an empty list", func() {
		path := filepath.Join(GinkgoT().TempDir(), "sqlprofile.jsonl")
		Expect(os.WriteFile(path, []byte("{\"duration_ns\":10000000,\"sql\":\"SELECT 1\"}\n"), 0o600)).To(Succeed())
		profile, err := ReadSQLProfile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(profile.Statements[0].Params).To(Equal([]string{}))
	})
})
