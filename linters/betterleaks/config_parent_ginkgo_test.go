package betterleaks

import (
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// A shared .gavel.yaml / .betterleaks.toml in the directory that holds several
// checkouts (example/<repo>) applies to each of them.
var _ = Describe("Betterleaks git-root-parent config layer", func() {
	const fakeRule = "acme-fake-token"
	var parent, repo string

	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		parent = filepath.Join(GinkgoT().TempDir(), "example")
		repo = filepath.Join(parent, "repo")
		Expect(os.MkdirAll(filepath.Join(repo, ".git"), 0o755)).To(Succeed())
	})

	It("discovers the parent's native config before the repo's", func() {
		parentNative := filepath.Join(parent, ".betterleaks.toml")
		repoNative := filepath.Join(repo, ".betterleaks.toml")
		for _, path := range []string{parentNative, repoNative} {
			Expect(os.WriteFile(path, []byte("title = \"cfg\"\n"), 0o644)).To(Succeed())
		}

		Expect(DiscoverConfigs(repo)).To(Equal([]string{parentNative, repoNative}))
	})

	It("filters a violation the parent .gavel.yaml lint.ignore names", func() {
		ignore := "lint: {ignore: [{rule: " + fakeRule + ", source: betterleaks}]}\n"
		Expect(os.WriteFile(filepath.Join(parent, verify.GavelConfigFileName), []byte(ignore), 0o644)).To(Succeed())
		violations, err := (&Betterleaks{}).parseFindings([]byte(`[{"RuleID":"` + fakeRule + `","File":"secrets.txt","StartLine":1}]`))
		Expect(err).NotTo(HaveOccurred())
		results := []*linters.LinterResult{{Linter: "betterleaks", WorkDir: repo, Violations: violations}}

		cfg, err := verify.LoadGavelConfig(repo)
		Expect(err).NotTo(HaveOccurred())

		Expect(linters.FilterIgnoredViolations(results, cfg.Lint.Ignore)).To(Equal(1))
		Expect(results[0].Violations).To(BeEmpty())
	})
})
