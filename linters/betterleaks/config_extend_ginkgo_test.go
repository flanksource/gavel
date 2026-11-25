package betterleaks

import (
	"os"
	"path/filepath"

	"github.com/hairyhenderson/toml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// betterleaks rejects a config that sets both extend.useDefault and
// extend.path, so merging a global useDefault layer with a repo layer that
// extends a shared ruleset must inline that ruleset instead of passing the
// (declaring-file-relative) path through.
var _ = Describe("Betterleaks extend.path merging", func() {
	var root, globalCfg, repoCfg string

	write := func(path, body string) {
		Expect(os.MkdirAll(filepath.Dir(path), 0o755)).To(Succeed())
		Expect(os.WriteFile(path, []byte(body), 0o644)).To(Succeed())
	}

	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		GinkgoT().Setenv("XDG_CACHE_HOME", GinkgoT().TempDir())
		root = GinkgoT().TempDir()
		globalCfg = filepath.Join(root, "home", ".betterleaks.toml")
		repoCfg = filepath.Join(root, "work", "repo", ".betterleaks.toml")
		write(globalCfg, "[extend]\nuseDefault = true\n\n[[rules]]\nid = \"global-rule\"\nregex = \"GLOBAL_.*\"\n")
		write(filepath.Join(root, "work", "shared-ruleset.toml"), "[[rules]]\nid = \"acme-brand\"\nregex = \"ACME_.*\"\n\n[[rules]]\nid = \"acme-domain\"\nregex = \"acme\\\\.example\"\n")
	})

	resolve := func() (tomlConfig, error) {
		out, err := ResolveConfig(ResolveConfigOptions{WorkDir: root, Configs: []string{globalCfg, repoCfg}})
		if err != nil {
			return tomlConfig{}, err
		}
		data, err := os.ReadFile(out)
		Expect(err).NotTo(HaveOccurred())
		var merged tomlConfig
		Expect(toml.Unmarshal(data, &merged)).To(Succeed())
		return merged, nil
	}

	ruleIDs := func(cfg tomlConfig) []string {
		ids := make([]string, 0, len(cfg.Rules))
		for _, rule := range cfg.Rules {
			ids = append(ids, rule.ID)
		}
		return ids
	}

	It("inlines a repo layer's relative extend.path so useDefault and path are never both set", func() {
		write(repoCfg, "[extend]\npath = \"../shared-ruleset.toml\"\ndisabledRules = [\"acme-domain\"]\n")

		merged, err := resolve()

		Expect(err).NotTo(HaveOccurred())
		Expect(merged.Extend).To(Equal(tomlExtend{UseDefault: true, DisabledRules: []string{"acme-domain"}}))
		Expect(ruleIDs(merged)).To(Equal([]string{"global-rule", "acme-brand"}))
	})

	// betterleaks only applies extend.disabledRules to rules it loads from an
	// extended config; once rules are inlined into one file they must be dropped here.
	It("drops inlined rules that a later layer disables", func() {
		write(repoCfg, "[extend]\ndisabledRules = [\"global-rule\"]\n")

		merged, err := resolve()

		Expect(err).NotTo(HaveOccurred())
		Expect(ruleIDs(merged)).To(BeEmpty())
		Expect(merged.Extend.DisabledRules).To(Equal([]string{"global-rule"}), "still disables a default rule of that id")
	})

	It("lets the extending config override a rule from the ruleset it extends", func() {
		write(repoCfg, "[extend]\npath = \"../shared-ruleset.toml\"\n\n[[rules]]\nid = \"acme-brand\"\nregex = \"ACME_OVERRIDE_.*\"\n")

		merged, err := resolve()

		Expect(err).NotTo(HaveOccurred())
		Expect(ruleIDs(merged)).To(Equal([]string{"global-rule", "acme-brand", "acme-domain"}))
		Expect(merged.Rules[1].Regex).To(Equal("ACME_OVERRIDE_.*"))
	})

	It("fails loudly when an extend.path target is missing", func() {
		write(repoCfg, "[extend]\npath = \"../missing.toml\"\n")

		_, err := resolve()

		Expect(err).To(MatchError(ContainSubstring(filepath.Join(root, "work", "missing.toml"))))
	})
})
