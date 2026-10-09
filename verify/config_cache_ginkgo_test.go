package verify

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"

	"github.com/flanksource/captain/pkg/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LoadGavelConfig cache", func() {
	var home, repo, sub string

	BeforeEach(func() {
		home, repo = GinkgoT().TempDir(), GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		// A .git directory pins repo as the git root, so the consulted chain is
		// exactly home, repo, and (for sub) repo/sub.
		Expect(os.Mkdir(filepath.Join(repo, ".git"), 0o755)).To(Succeed())
		sub = filepath.Join(repo, "sub")
		Expect(os.Mkdir(sub, 0o755)).To(Succeed())
	})

	writeConfig := func(dir, content string) {
		Expect(os.WriteFile(filepath.Join(dir, GavelConfigFileName), []byte(content), 0o600)).To(Succeed())
	}
	loadCounting := func(dir string) (GavelConfig, int64) {
		before := gavelConfigParses.Load()
		cfg, err := LoadGavelConfig(dir)
		Expect(err).NotTo(HaveOccurred())
		return cfg, gavelConfigParses.Load() - before
	}

	It("serves an unchanged chain without re-parsing any file", func() {
		writeConfig(home, "procfile: {mem: 1g}\n")
		writeConfig(repo, "procfile: {profile: dev}\n")

		first, parses := loadCounting(sub)
		Expect(parses).To(BeEquivalentTo(2))

		second, parses := loadCounting(sub)
		Expect(parses).To(BeZero())
		Expect(second).To(Equal(first))
		Expect(second.Procfile).To(Equal(ProcfileConfig{Profile: "dev", Mem: "1g"}))
	})

	DescribeTable("re-parses when a consulted file changes, appears, or disappears",
		func(initial, mutate func(home, repo, sub string), wantProfile string, wantParses int) {
			initial(home, repo, sub)
			_, _ = loadCounting(sub)

			mutate(home, repo, sub)
			cfg, parses := loadCounting(sub)
			Expect(cfg.Procfile.Profile).To(Equal(wantProfile))
			Expect(parses).To(BeEquivalentTo(wantParses))
		},
		Entry("repo config rewritten with identical size",
			func(_, repo, _ string) { writeConfig(repo, "procfile: {profile: aaa}\n") },
			func(_, repo, _ string) { writeConfig(repo, "procfile: {profile: bbb}\n") },
			"bbb", 1),
		Entry("repo config rewritten with a different size",
			func(_, repo, _ string) { writeConfig(repo, "procfile: {profile: dev}\n") },
			func(_, repo, _ string) { writeConfig(repo, "procfile: {profile: staging}\n") },
			"staging", 1),
		Entry("previously missing home config appears",
			func(_, _, _ string) {},
			func(home, _, _ string) { writeConfig(home, "procfile: {profile: home}\n") },
			"home", 1),
		Entry("previously missing target-directory config appears",
			func(_, repo, _ string) { writeConfig(repo, "procfile: {profile: repo}\n") },
			func(_, _, sub string) { writeConfig(sub, "procfile: {profile: sub}\n") },
			"sub", 2),
		Entry("target-directory config is removed",
			func(_, repo, sub string) {
				writeConfig(repo, "procfile: {profile: repo}\n")
				writeConfig(sub, "procfile: {profile: sub}\n")
			},
			func(_, _, sub string) { Expect(os.Remove(filepath.Join(sub, GavelConfigFileName))).To(Succeed()) },
			"repo", 1),
	)

	It("re-parses when a new git root moves the chain", func() {
		writeConfig(repo, "procfile: {profile: repo}\n")
		writeConfig(sub, "procfile: {mem: 2g}\n")
		cfg, _ := loadCounting(sub)
		Expect(cfg.Procfile).To(Equal(ProcfileConfig{Profile: "repo", Mem: "2g"}))

		Expect(os.Mkdir(filepath.Join(sub, ".git"), 0o755)).To(Succeed())
		cfg, parses := loadCounting(sub)
		Expect(cfg.Procfile).To(Equal(ProcfileConfig{Mem: "2g"}))
		Expect(parses).To(BeEquivalentTo(1))
	})

	It("isolates the cached config from mutations of a returned copy", func() {
		writeConfig(repo, `
pr: {base: main}
procfile: {env: {K: v}}
commit: {gitignore: ["*.log"]}
ai:
  temperature: 0.5
  cliArgs: {extra: {deep: [a]}}
todos:
  lifecycle:
    steps: [{name: build, with: {args: [x]}}]
`)
		_, _ = loadCounting(repo)
		mutated, parses := loadCounting(repo)
		Expect(parses).To(BeZero(), "the mutated copy must be a cache hit")
		mutated.PR.Base = "mutated"
		mutated.Procfile.Env["K"] = "mutated"
		mutated.Commit.GitIgnore[0] = "mutated"
		*mutated.AI.Temperature = 1
		mutated.AI.CLIArgs["extra"].(map[string]any)["deep"].([]any)[0] = "mutated"
		mutated.Todos.Lifecycle.Steps[0]["with"].(map[string]any)["args"].([]any)[0] = "mutated"

		fresh, parses := loadCounting(repo)
		Expect(parses).To(BeZero())
		Expect(fresh.PR.Base).To(Equal("main"))
		Expect(fresh.Procfile.Env).To(Equal(map[string]string{"K": "v"}))
		Expect(fresh.Commit.GitIgnore).To(Equal([]string{"*.log"}))
		Expect(*fresh.AI.Temperature).To(Equal(0.5))
		Expect(fresh.AI.CLIArgs).To(Equal(map[string]any{"extra": map[string]any{"deep": []any{"a"}}}))
		Expect(fresh.Todos.Lifecycle.Steps).To(Equal([]map[string]any{
			{"name": "build", "with": map[string]any{"args": []any{"x"}}},
		}))
	})

	It("hands concurrent callers independent copies", func() {
		writeConfig(repo, "procfile: {env: {K: v}}\n")
		_, _ = loadCounting(repo)
		var wg sync.WaitGroup
		for i := range 8 {
			wg.Go(func() {
				defer GinkgoRecover()
				cfg, err := LoadGavelConfig(repo)
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg.Procfile.Env).To(HaveKeyWithValue("K", "v"))
				cfg.Procfile.Env["K"] = fmt.Sprint(i)
			})
		}
		wg.Wait()
		cfg, _ := loadCounting(repo)
		Expect(cfg.Procfile.Env).To(Equal(map[string]string{"K": "v"}))
	})

	It("never caches a failed load as success", func() {
		invalid := "ai: {temperature: 3}\n"
		writeConfig(repo, invalid)
		for range 2 {
			before := gavelConfigParses.Load()
			_, err := LoadGavelConfig(repo)
			Expect(err).To(MatchError(ContainSubstring("temperature")))
			Expect(gavelConfigParses.Load() - before).To(BeEquivalentTo(1))
		}

		writeConfig(repo, "ai: {temperature: 1}\n")
		cfg, parses := loadCounting(repo)
		Expect(*cfg.AI.Temperature).To(Equal(1.0))
		Expect(parses).To(BeEquivalentTo(1))
	})
})

var _ = Describe("GavelConfig type graph", func() {
	// merge.Clone copies unexported fields as they stand, so an unexported
	// reference field would be shared between the cache and every caller. The
	// only one reachable is written by assignment after the config is loaded,
	// never populated by decoding; a new one must be reviewed before it lands.
	It("has no unexported reference fields beyond the reviewed ones", func() {
		var found []string
		walkConfigType(reflect.TypeOf(GavelConfig{}), "GavelConfig", map[reflect.Type]bool{
			reflect.TypeOf(api.ModelProvider{}): true, // configPolicy Shared: a catalog entry, copied by reference
		}, &found)
		Expect(found).To(Equal([]string{
			"GavelConfig.AI.Prompt.Attachments[].preparedBytes ([]uint8)",
		}))
	})
})

func walkConfigType(t reflect.Type, path string, seen map[reflect.Type]bool, found *[]string) {
	switch t.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Array:
		walkConfigType(t.Elem(), path+"[]", seen, found)
	case reflect.Map:
		walkConfigType(t.Elem(), path+"{}", seen, found)
	case reflect.Struct:
		if seen[t] {
			return
		}
		seen[t] = true
		for i := range t.NumField() {
			field := t.Field(i)
			if field.IsExported() {
				walkConfigType(field.Type, path+"."+field.Name, seen, found)
				continue
			}
			switch field.Type.Kind() {
			case reflect.Pointer, reflect.Slice, reflect.Map, reflect.Interface, reflect.Chan, reflect.Func:
				*found = append(*found, fmt.Sprintf("%s.%s (%s)", path, field.Name, field.Type))
			}
		}
	}
}
