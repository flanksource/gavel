package verify

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// livePromptIDs collects every x-prompt-id the config schema stamps. That is the
// same marker the settings UI uses to bind a field to its prompt descriptor, and
// each ID is by construction the dotted .gavel.yaml path of the operation's
// PromptSpec field — so it is the list the docs have to agree with.
func livePromptIDs() []string {
	var ids []string
	var walk func(node any)
	walk = func(node any) {
		switch typed := node.(type) {
		case map[string]any:
			if id, ok := typed["x-prompt-id"].(string); ok {
				ids = append(ids, id)
			}
			for _, child := range typed {
				walk(child)
			}
		case []any:
			for _, child := range typed {
				walk(child)
			}
		}
	}
	walk(gavelConfigSchema())
	return ids
}

// docFiles are the hand-written references for .gavel.yaml. gavel.schema.json is
// generated and already guarded by TestConfigSchema_GoldenMatchesCommitted; this
// prose is not, and it drifted for two months after the operation-scoped spec
// rename shipped — still telling readers to configure commit.messagePrompt and a
// top-level verify: section that no longer parse.
var docFiles = []string{
	"SCHEMA.md",
	"site/content/docs/prompt-overrides.mdx",
}

// repoFile reads a path relative to the repo root. go test runs each package in
// its own source directory, so the root is one level up from verify/.
func repoFile(name string) string {
	GinkgoHelper()
	content, err := os.ReadFile(filepath.Join("..", name))
	Expect(err).NotTo(HaveOccurred())
	return string(content)
}

var headingPattern = regexp.MustCompile(`^#{2,3} `)

// withoutRetiredSection drops the "Retired keys" reference table, the one place
// a dead key is supposed to be named.
func withoutRetiredSection(content string) string {
	var kept []string
	skipping := false
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "### Retired keys") {
			skipping = true
			continue
		}
		if skipping {
			if !headingPattern.MatchString(line) {
				continue
			}
			skipping = false
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}

// documentsRetiredKey reports whether the docs still teach a retired key as
// something to configure.
//
// A dotted path is always written as one complete backticked span, so that is
// the match. A single-segment key cannot be matched the same way: `verify` is
// also the name of a lifecycle step and appears legitimately all over the todos
// section, and `verify.GavelConfig` names a Go type. It only counts as retired
// usage when it appears as a top-level key in a YAML example, which is exactly
// the shape the removed section had.
func documentsRetiredKey(body, path string) bool {
	if strings.Contains(path, ".") {
		return strings.Contains(body, "`"+path+"`")
	}
	return regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(path) + `:`).MatchString(body)
}

var _ = Describe("config documentation", func() {
	It("documents every overridable prompt by its live config path", func() {
		ids := livePromptIDs()
		Expect(ids).NotTo(BeEmpty())
		for _, name := range docFiles {
			content := repoFile(name)
			for _, id := range ids {
				Expect(content).To(ContainSubstring(id),
					"%s does not mention the %s prompt override", name, id)
			}
		}
	})

	It("does not present a retired key as a live setting", func() {
		for _, name := range docFiles {
			body := withoutRetiredSection(repoFile(name))
			for path := range LegacyConfigFields {
				Expect(documentsRetiredKey(body, path)).To(BeFalse(),
					"%s still documents the retired key %s outside the retired-keys table", name, path)
			}
		}
	})
})
