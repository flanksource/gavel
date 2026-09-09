package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("project action command options", func() {
	provider := projectActionCommandProvider{}

	It("derives non-mutating lint fields from the real Cobra command", func() {
		definition, err := provider.Schema("lint")

		Expect(err).NotTo(HaveOccurred())
		properties := definition.Schema["properties"].(map[string]any)
		Expect(properties).To(HaveKey("linters"))
		Expect(properties).To(HaveKey("changed"))
		Expect(properties).To(HaveKey("summary"))
		Expect(properties).NotTo(HaveKey("work-dir"))
		Expect(properties).NotTo(HaveKey("triage"))
		Expect(properties).NotTo(HaveKey("fix"))
		Expect(properties).NotTo(HaveKey("ai-fix"))
		Expect(properties).NotTo(HaveKey("model"))
		Expect(properties).NotTo(HaveKey("api-key"))
		Expect(properties).NotTo(HaveKey("fallback"))
		Expect(properties).NotTo(HaveKey("mode"))
	})

	It("serializes typed test values in deterministic flag order", func() {
		args, err := provider.Args("test", map[string]any{
			"framework":    []any{"go", "ginkgo"},
			"paths":        []any{"./testrunner"},
			"recursive":    false,
			"test-timeout": "2m",
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(args).To(Equal([]string{
			"--framework=go",
			"--framework=ginkgo",
			"--recursive=false",
			"--test-timeout=2m",
			"./testrunner",
		}))
	})

	It("derives safe non-interactive commit fields and serializes selected files", func() {
		definition, err := provider.Schema("commit")

		Expect(err).NotTo(HaveOccurred())
		properties := definition.Schema["properties"].(map[string]any)
		Expect(properties).To(HaveKey("files"))
		Expect(properties).To(HaveKey("message"))
		Expect(properties).To(HaveKey("commit-all"))
		Expect(properties).To(HaveKey("lint"))
		Expect(properties).NotTo(HaveKey("interactive"))
		Expect(properties).NotTo(HaveKey("batch"))
		Expect(properties).NotTo(HaveKey("push"))
		Expect(properties).NotTo(HaveKey("fixup"))
		Expect(properties).NotTo(HaveKey("since"))
		Expect(properties).NotTo(HaveKey("work-dir"))
		Expect(definition.Defaults).To(HaveKeyWithValue("precommit", "fail"))
		Expect(definition.Schema["description"]).To(ContainSubstring("selected project files"))

		args, err := provider.Args("commit", map[string]any{
			"files":     []any{"one.go", "two.go"},
			"lint":      "true",
			"message":   "fix: selected files",
			"precommit": "fail",
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(args).To(Equal([]string{
			"--lint=true",
			"--message=fix: selected files",
			"--precommit=fail",
			"one.go",
			"two.go",
		}))
	})

	It("rejects excluded, unknown, and invalid option values", func() {
		_, err := provider.Args("lint", map[string]any{"fix": true})
		Expect(err).To(MatchError(ContainSubstring("unsupported lint option")))

		_, err = provider.Args("test", map[string]any{"surprise": true})
		Expect(err).To(MatchError(ContainSubstring("unsupported test option")))

		_, err = provider.Args("test", map[string]any{"test-timeout": "soon"})
		Expect(err).To(MatchError(ContainSubstring("test-timeout")))
	})
})
