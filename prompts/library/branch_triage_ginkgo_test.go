package library_test

import (
	"encoding/json"
	"io/fs"
	"strings"

	promptlib "github.com/flanksource/captain/pkg/ai/prompt"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/prompts/library"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

var _ = Describe("branch-triage.prompt", func() {
	const repo = "/src/acme's checkout"

	render := func(data map[string]any) api.Spec {
		tpl, err := promptlib.LoadFS(library.FS, "branch-triage.prompt")
		Expect(err).NotTo(HaveOccurred())
		spec, _, err := tpl.Render(promptlib.RenderOptions{Declared: true, Data: data})
		Expect(err).NotTo(HaveOccurred())
		return spec
	}

	It("runs as an agent in the audited checkout", func() {
		spec := render(map[string]any{"repo": repo})
		Expect(spec.Setup.Cwd).To(Equal(repo))
		Expect(spec.Model.Mode).To(Equal(api.ModeAgent))
	})

	DescribeTable("keeps the audit read-only",
		func(tool string, policy api.ToolPolicy) {
			Expect(render(map[string]any{"repo": repo}).Permissions.Tools).To(HaveKeyWithValue(tool, policy))
		},
		Entry(nil, "Edit", api.ToolPolicy("deny")),
		Entry(nil, "Write", api.ToolPolicy("deny")),
		Entry(nil, "Bash(gh:*)", api.ToolPolicy("deny")),
		Entry(nil, "Bash(git fetch:*)", api.ToolPolicy("deny")),
		Entry(nil, "Bash(git push:*)", api.ToolPolicy("deny")),
		Entry(nil, "Bash(git rebase:*)", api.ToolPolicy("deny")),
		Entry(nil, "Bash(git branch -D:*)", api.ToolPolicy("deny")),
		Entry(nil, "Bash(git worktree remove:*)", api.ToolPolicy("deny")),
		Entry(nil, "Bash(gavel pr close:*)", api.ToolPolicy("deny")),
		Entry(nil, "Bash(gavel pr status --ai-fix:*)", api.ToolPolicy("deny")),
		Entry(nil, "Bash(git worktree list:*)", api.ToolPolicy("allow")),
		Entry(nil, "Bash(git cherry:*)", api.ToolPolicy("allow")),
		Entry(nil, "Bash(gavel pr list:*)", api.ToolPolicy("allow")),
	)

	It("does not grant git -C, which would prefix any subcommand", func() {
		Expect(render(map[string]any{"repo": repo}).Permissions.Tools).NotTo(HaveKey("Bash(git -C:*)"))
	})

	It("returns one verdict per item from the fixed verdict set", func() {
		var schema struct {
			Properties struct {
				Items struct {
					Items struct {
						Properties struct {
							Kind    struct{ Enum []string } `json:"kind"`
							Verdict struct{ Enum []string } `json:"verdict"`
						} `json:"properties"`
					} `json:"items"`
				} `json:"items"`
			} `json:"properties"`
		}
		Expect(json.Unmarshal(render(map[string]any{"repo": repo}).Prompt.SchemaJSON, &schema)).To(Succeed())
		props := schema.Properties.Items.Items.Properties
		Expect([][]string{props.Kind.Enum, props.Verdict.Enum}).To(Equal([][]string{
			{"pr", "branch", "worktree"},
			{"ready-to-merge", "needs-rebase", "close", "delete", "archive", "active", "needs-attention"},
		}))
	})

	DescribeTable("defaults the remote, base and staleness window when unset",
		func(data map[string]any, expected []string) {
			user := render(data).Prompt.User
			for _, fragment := range expected {
				Expect(user).To(ContainSubstring(fragment))
			}
		},
		Entry("unset", map[string]any{"repo": repo},
			[]string{"git rev-parse --abbrev-ref origin/HEAD", "older than 30 days"}),
		Entry("set", map[string]any{"repo": repo, "base": "upstream/develop", "remote": "upstream", "staleDays": 14},
			[]string{"Base: `upstream/develop`", "Remote: `upstream`", "older than 14 days"}),
	)

	It("is discoverable with valid unrendered frontmatter", func() {
		raw, err := fs.ReadFile(library.FS, "branch-triage.prompt")
		Expect(err).NotTo(HaveOccurred())
		var frontmatter map[string]any
		Expect(yaml.Unmarshal([]byte(strings.SplitN(string(raw), "---\n", 3)[1]), &frontmatter)).To(Succeed())
		Expect(frontmatter).To(HaveKeyWithValue("name", "branch-triage"))
	})
})
