package lint

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/flanksource/gavel/linters"
	jsonlint "github.com/flanksource/gavel/linters/json"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("JSON lint integration", func() {
	var root string
	BeforeEach(func() {
		root = GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
	})

	It("runs and previews the built-in adapter without an external binary", func() {
		git, err := exec.LookPath("git")
		Expect(err).NotTo(HaveOccurred())
		bin := GinkgoT().TempDir()
		Expect(os.Symlink(git, filepath.Join(bin, "git"))).To(Succeed())
		GinkgoT().Setenv("PATH", bin)
		Expect(os.WriteFile(filepath.Join(root, "bad.json"), []byte(`{bad}`), 0o600)).To(Succeed())
		results, err := Execute(Options{Context: context.Background(), WorkDir: root, Linters: []string{"json"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveExactElements(And(HaveField("Success", true), HaveField("Violations", HaveLen(1)))))
		Expect(linters.HasFailures(results)).To(BeTrue())
		Expect(DryRun(Options{WorkDir: root, Linters: []string{"json"}})).To(Succeed())
	})

	It("activates automatically when matching files exist", func() {
		adapter := jsonlint.New(root)
		selected, _ := shouldSelectLinter(root, verify.GavelConfig{}, adapter, false)
		Expect(selected).To(BeFalse())
		Expect(os.WriteFile(filepath.Join(root, "data.yaml"), []byte("enabled: true\n"), 0o600)).To(Succeed())
		selected, _ = shouldSelectLinter(root, verify.GavelConfig{}, adapter, false)
		Expect(selected).To(BeTrue())
		Expect(builtinLinterRegistry(root).Has("json")).To(BeTrue())
	})

	It("activates for matching files in a nested directory", func() {
		nested := filepath.Join(root, "configs")
		Expect(os.Mkdir(nested, 0o700)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(nested, "data.yaml"), []byte("enabled: true\n"), 0o600)).To(Succeed())
		selected, _ := shouldSelectLinter(root, verify.GavelConfig{}, jsonlint.New(root), false)
		Expect(selected).To(BeTrue())
	})

	It("resolves configured schema paths from the config and honors CLI overrides", func() {
		Expect(os.WriteFile(filepath.Join(root, ".gavel.yaml"), []byte("lint:\n  linters:\n    json:\n      options:\n        schema: schema.json\n"), 0o600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "schema.json"), []byte(`false`), 0o600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "accept.json"), []byte(`true`), 0o600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "data.json"), []byte(`{}`), 0o600)).To(Succeed())
		opts := Options{WorkDir: root, Linters: []string{"json"}, Files: []string{"data.json"}}
		results, err := Execute(opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveExactElements(HaveField("Violations", HaveLen(1))))
		opts.Schema = filepath.Join(root, "accept.json")
		results, err = Execute(opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveExactElements(HaveField("Violations", BeEmpty())))
	})

	It("preserves configured schema URLs", func() {
		Expect(os.WriteFile(filepath.Join(root, ".gavel.yaml"), []byte("lint:\n  linters:\n    json:\n      options:\n        schema: https://example.org/schema.json\n"), 0o600)).To(Succeed())
		cfg, err := verify.LoadGavelConfig(root)
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.LintOptionsFor("json").Schema).To(HaveValue(Equal("https://example.org/schema.json")))
	})

	It("resolves CLI schemas before git-root grouping", func() {
		nested := filepath.Join(root, "nested")
		Expect(os.Mkdir(nested, 0o700)).To(Succeed())
		Expect(os.Mkdir(filepath.Join(root, ".git"), 0o700)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(nested, "schema.json"), []byte(`false`), 0o600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(nested, "data.json"), []byte(`{}`), 0o600)).To(Succeed())
		results, err := Run(context.Background(), Options{WorkDir: nested, Linters: []string{"json"}, Files: []string{"data.json"}, Schema: "schema.json"})
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveExactElements(And(HaveField("Success", true), HaveField("Violations", HaveLen(1)))))
	})

	It("keeps explicit file selection and ignores during discovery", func() {
		for _, name := range []string{"wanted.json", "ignored.json", "other.json"} {
			Expect(os.WriteFile(filepath.Join(root, name), []byte(`{bad}`), 0o600)).To(Succeed())
		}
		results, err := Execute(Options{WorkDir: root, Linters: []string{"json"}, Files: []string{"wanted.json", "ignored.json"}, Ignore: []string{"ignored.json"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveExactElements(HaveField("Violations", HaveExactElements(HaveField("File", filepath.Join(root, "wanted.json"))))))
	})

	DescribeTable("expands quoted YAML, JSON and NDJSON glob selectors through the pipeline", func(patterns []string) {
		nested := filepath.Join(root, "configs")
		Expect(os.Mkdir(nested, 0o700)).To(Succeed())
		for name, contents := range map[string]string{"data.yaml": "broken: [\n", "data.json": `{bad}`, "data.ndjson": "{bad}\n"} {
			Expect(os.WriteFile(filepath.Join(nested, name), []byte(contents), 0o600)).To(Succeed())
		}
		results, err := Run(context.Background(), Options{WorkDir: root, Linters: []string{"json"}, Files: patterns})
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveExactElements(And(HaveField("Success", true), HaveField("FileCount", 3), HaveField("Violations", HaveLen(3)))))
	}, Entry("directory", []string{"configs/*.yaml", "configs/*.json", "configs/*.ndjson"}), Entry("recursive", []string{"**/*.yaml", "**/*.json", "**/*.ndjson"}))
})
