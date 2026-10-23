package runners

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/testrunner/parsers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ginkgo overlay", func() {
	It("replays structured suite metadata through the Go test cache", func() {
		workDir, err := filepath.Abs("../..")
		Expect(err).NotTo(HaveOccurred())
		pkg := "./testrunner/runners/testdata/overlayfixture"
		importPath := "github.com/flanksource/gavel/testrunner/runners/testdata/overlayfixture"
		overlay, err := BuildGinkgoOverlay(workDir, []string{pkg}, nil)
		Expect(err).NotTo(HaveOccurred())
		defer os.Remove(overlay)

		for run := 0; run < 2; run++ {
			cmd := exec.Command("go", "test", "-json", "-v", "-overlay="+overlay, pkg)
			cmd.Dir = workDir
			output, err := cmd.Output()
			Expect(err).NotTo(HaveOccurred(), string(output))
			var marker, cached bool
			for _, line := range strings.Split(string(output), "\n") {
				var event struct {
					Action string
					Output string
				}
				if json.Unmarshal([]byte(line), &event) != nil || event.Action != "output" {
					continue
				}
				marker = marker || strings.Contains(event.Output, "GAVEL_GINKGO_V1:")
				cached = cached || strings.Contains(event.Output, "(cached)")
			}
			Expect(marker).To(BeTrue(), "missing metadata on run %d", run+1)
			if run == 1 {
				Expect(cached).To(BeTrue(), "second run was not cached")
			}
			results, err := parsers.ParseGoTestBatch(strings.NewReader(string(output)), workDir, map[string]bool{importPath: true})
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(HaveKey(importPath))
			var ordinary, spec *parsers.Test
			for _, suite := range results[importPath] {
				for i := range suite.Tests {
					test := &suite.Tests[i]
					if test.Name == "TestOrdinary" {
						ordinary = test
					}
					if test.Name == "reports metadata" {
						spec = test
					}
				}
			}
			Expect(ordinary).NotTo(BeNil())
			Expect(spec).NotTo(BeNil())
			Expect(spec.Suite).To(Equal([]string{"container"}))
			Expect(spec.Framework).To(Equal(parsers.Ginkgo))
			Expect(spec.Cached).To(Equal(cached))
			ctx, ok := spec.Context.(parsers.GinkgoContext)
			Expect(ok).To(BeTrue())
			Expect(ctx.Labels).To(ConsistOf("inherited", "leaf"))
			Expect(ctx.ReportEntries).To(HaveLen(1))
			Expect(string(ctx.ReportEntries[0].Value)).To(MatchJSON(`{"kind":"Widget"}`))
		}
	})

	It("compiles the reporter into a package declaring names its imports would redeclare", func() {
		workDir, err := filepath.Abs("../..")
		Expect(err).NotTo(HaveOccurred())
		pkg := "./testrunner/runners/testdata/overlaycollision"
		overlay, err := BuildGinkgoOverlay(workDir, []string{pkg}, nil)
		Expect(err).NotTo(HaveOccurred())
		defer os.Remove(overlay)

		cmd := exec.Command("go", "test", "-v", "-overlay="+overlay, pkg)
		cmd.Dir = workDir
		output, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(output))
		Expect(string(output)).To(ContainSubstring("GAVEL_GINKGO_V1:"))
	})

	It("rejects Ginkgo v1 before creating an overlay", func() {
		workDir, err := filepath.Abs("../..")
		Expect(err).NotTo(HaveOccurred())
		_, err = BuildGinkgoOverlay(workDir, []string{"./testrunner/runners/testdata/overlayv1"}, nil)
		Expect(err).To(MatchError(ContainSubstring("Ginkgo v1 is unsupported")))
	})

	It("places the reporter in the external package when a suite spans internal and external test packages", func() {
		dir := GinkgoT().TempDir()
		for name, source := range map[string]string{
			"a_internal_test.go": "package sample\nimport \"github.com/onsi/ginkgo/v2\"\n",
			"b_external_test.go": "package sample_test\nimport \"github.com/onsi/ginkgo/v2\"\n",
			"c_internal_test.go": "package sample\nimport \"github.com/onsi/ginkgo/v2\"\n",
		} {
			Expect(os.WriteFile(filepath.Join(dir, name), []byte(source), 0o600)).To(Succeed())
		}
		Expect(ginkgoPackageName(dir, nil)).To(Equal("sample_test"))
	})

	It("rejects Ginkgo files from unrelated packages in one directory", func() {
		dir := GinkgoT().TempDir()
		for name, source := range map[string]string{
			"a_test.go": "package sample\nimport \"github.com/onsi/ginkgo/v2\"\n",
			"b_test.go": "package other\nimport \"github.com/onsi/ginkgo/v2\"\n",
		} {
			Expect(os.WriteFile(filepath.Join(dir, name), []byte(source), 0o600)).To(Succeed())
		}
		_, err := ginkgoPackageName(dir, nil)
		Expect(err).To(MatchError(ContainSubstring("multiple Ginkgo test packages")))
	})
})
