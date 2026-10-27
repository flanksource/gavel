package verify

import (
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Hierarchical configuration", func() {
	var repo, parent, target string
	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		repo = GinkgoT().TempDir()
		parent = filepath.Join(repo, "services")
		target = filepath.Join(parent, "api")
		Expect(os.Mkdir(filepath.Join(repo, ".git"), 0o755)).To(Succeed())
		Expect(os.MkdirAll(target, 0o755)).To(Succeed())
	})
	write := func(dir, text string) {
		Expect(os.WriteFile(filepath.Join(dir, GavelConfigFileName), []byte(text), 0o600)).To(Succeed())
	}
	It("merges intermediate directories and reports the same sources in inspection", func() {
		write(repo, "runtime: {env: {ROOT: root}}\n")
		write(parent, "runtime: {env: {PARENT: parent}}\n")
		write(target, "runtime: {env: {CHILD: child}}\n")
		cfg, err := LoadGavelConfig(target)
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.Runtime.Env).To(Equal(map[string]string{"ROOT": "root", "PARENT": "parent", "CHILD": "child"}))
		trace, err := LoadGavelConfigTrace(target)
		Expect(err).NotTo(HaveOccurred())
		Expect(trace.Merged).To(Equal(cfg))
		Expect(trace.Sources).To(HaveLen(3))
	})
	It("loads command options and independent commit repair configuration", func() {
		write(parent, "test: {options: {recursive: false, nodes: 0, ignore: []}}\nlint: {options: {fix: false, tags: [integration]}}\ncommit: {fix: {model: example-model}, lint: {aiFix: false, aiFixMaxIterations: 2}}\n")
		_, err := LoadGavelConfig(target)
		Expect(err).NotTo(HaveOccurred())
	})
	It("persists checked directories and reuses decoded layers after clearing memory", func() {
		write(parent, "runtime: {env: {PARENT: parent}}\n")
		first, err := LoadGavelConfig(target)
		Expect(err).NotTo(HaveOccurred())
		Expect(filepath.Join(repo, ".gavel", "config-cache.json")).To(BeAnExistingFile())
		gavelConfigCache.Lock()
		gavelConfigCache.entries = map[string]gavelConfigCacheEntry{}
		gavelConfigCache.Unlock()
		before := gavelConfigParses.Load()
		second, err := LoadGavelConfig(target)
		Expect(err).NotTo(HaveOccurred())
		Expect(second).To(Equal(first))
		Expect(gavelConfigParses.Load()).To(Equal(before))
	})
	It("reuses the cross process configuration cache", func() {
		if directory := os.Getenv("GAVEL_CONFIG_CACHE_PROBE"); directory != "" {
			before := gavelConfigParses.Load()
			cfg, err := LoadGavelConfig(directory)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Runtime.Env["CACHED"]).To(Equal("parent"))
			Expect(gavelConfigParses.Load()).To(Equal(before))
			return
		}
		write(parent, "runtime: {env: {CACHED: parent}}\n")
		_, err := LoadGavelConfig(target)
		Expect(err).NotTo(HaveOccurred())
		executable, err := os.Executable()
		Expect(err).NotTo(HaveOccurred())
		command := exec.Command(executable, "-test.run=^TestGavelConfigDecoding$", "-ginkgo.focus=reuses the cross process configuration cache", "-ginkgo.no-color")
		command.Env = append(os.Environ(), "GAVEL_CONFIG_CACHE_PROBE="+target)
		output, err := command.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(output))
	})
	It("preserves explicit resets and nearer category settings over ancestor tool settings", func() {
		write(repo, "lint: {linters: {ruff: {options: {fix: true, tags: [integration]}}}}\n")
		write(parent, "lint: {options: {fix: false, tags: []}}\n")
		cfg, err := LoadGavelConfig(target)
		Expect(err).NotTo(HaveOccurred())
		options := cfg.LintOptionsFor("ruff")
		Expect(options.Fix).To(HaveValue(BeFalse()))
		Expect(options.Tags).To(BeEmpty())
		Expect(options.Tags).NotTo(BeNil())
	})
})
