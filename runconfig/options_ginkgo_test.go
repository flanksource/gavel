package runconfig

import (
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRunConfig(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "Command configuration") }

var _ = Describe("Configured command options", func() {
	It("applies explicit false, zero and empty lists while preserving caller fields", func() {
		no, zero, timeout := false, 0, "2s"
		configured := Test{Recursive: &no, Nodes: &zero, Tags: []string{}, TestTimeout: &timeout}
		actual := struct {
			Recursive   bool
			Nodes       int
			Tags        []string
			TestTimeout time.Duration
		}{true, 4, []string{"integration"}, time.Second}
		Expect(Apply(&actual, configured, map[string]bool{"test-timeout": true})).To(Succeed())
		Expect(actual).To(Equal(struct {
			Recursive   bool
			Nodes       int
			Tags        []string
			TestTimeout time.Duration
		}{false, 0, []string{}, time.Second}))
	})
	It("rejects run-level options in a per-tool override", func() {
		concurrency := 2
		Expect(Validate(Test{Concurrency: &concurrency}, true)).To(MatchError(ContainSubstring("concurrency")))
	})
	It("resolves ignore globs against the configured working directory", func() {
		root := GinkgoT().TempDir()
		workDir := filepath.Join(root, "project")
		actual := struct {
			Ignore  []string
			WorkDir string
		}{WorkDir: root}
		Expect(Apply(&actual, Test{WorkDir: &workDir, Ignore: []string{filepath.Join(workDir, "generated/**")}}, nil)).To(Succeed())
		Expect(actual.Ignore).To(Equal([]string{"generated/**"}))
	})
})
