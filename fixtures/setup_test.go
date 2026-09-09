package fixtures

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/commons-db/shell"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeFixtureFile writes a markdown fixture into a temp dir and parses it,
// returning the frontmatter the parser bound.
func parseFixtureFrontMatter(t *testing.T, body string) *FrontMatter {
	t.Helper()
	path := filepath.Join(t.TempDir(), "setup.md")
	require.NoError(t, os.WriteFile(path, []byte(body), 0644))

	file, err := os.Open(path)
	require.NoError(t, err)
	defer func() { _ = file.Close() }()

	fm, _, err := parseFrontMatter(file)
	require.NoError(t, err)
	require.NotNil(t, fm)
	fm.CleanMetadata()
	return fm
}

// TestFrontMatterSetupIsNotSwallowedIntoMetadata is the regression test for the
// goccy inline-map bug: before Setup was a real field, `setup:` fell into the
// `yaml:",inline"` Metadata map and silently became a template variable. It must
// bind to the typed field and leave no trace in Metadata.
func TestFrontMatterSetupIsNotSwallowedIntoMetadata(t *testing.T) {
	fm := parseFixtureFrontMatter(t, `---
setup:
  cwd: .
  dotenv: [.env.test]
exec: echo
---

# Body
`)

	require.NotNil(t, fm.Setup, "setup: did not bind to the typed field")
	assert.Equal(t, ".", fm.Setup.Cwd)
	assert.Equal(t, []string{".env.test"}, fm.Setup.DotEnv)
	assert.Nil(t, fm.Metadata["setup"], "setup: leaked into Metadata as a template var")
}

// TestFrontMatterSetupDecodesJSONTaggedConnections pins why SetupSpec routes
// through ghodss/yaml: connection.KubernetesConnection spells its field
// `json:"connection"` with no yaml tag, so a goccy-native decode would bind it by
// lowercased field name and drop the value.
func TestFrontMatterSetupDecodesJSONTaggedConnections(t *testing.T) {
	fm := parseFixtureFrontMatter(t, `---
setup:
  connections:
    kubernetes:
      connection: connection://kubernetes/sandbox
---

# Body
`)

	require.NotNil(t, fm.Setup)
	require.NotNil(t, fm.Setup.Connections.Kubernetes)
	assert.Equal(t, "connection://kubernetes/sandbox", fm.Setup.Connections.Kubernetes.ConnectionName)
}

func TestFrontMatterSetupCheckoutAndWorktree(t *testing.T) {
	fm := parseFixtureFrontMatter(t, `---
setup:
  checkout:
    mode: local
    path: .
    worktree:
      mode: new
      uncommitted: skip
---

# Body
`)

	require.NotNil(t, fm.Setup)
	require.NotNil(t, fm.Setup.Checkout)
	assert.Equal(t, shell.CheckoutLocal, fm.Setup.Checkout.Mode)
	require.NotNil(t, fm.Setup.Checkout.Worktree)
	assert.Equal(t, shell.WorktreeNew, fm.Setup.Checkout.Worktree.Mode)
	assert.Equal(t, shell.CloneSkip, fm.Setup.Checkout.Worktree.Uncommitted)
}

// A setup is prepared once per file and shared by every test in it, so a per-test
// block cannot be honoured. Rejecting it loudly matters because the per-test
// frontmatter parser otherwise drops keys it does not recognise without a word.
func TestPerTestSetupIsRejected(t *testing.T) {
	body := "### command: uses setup\n\n" +
		"```bash\necho hi\n```\n\n" +
		"```frontmatter\nsetup:\n  cwd: .\n```\n"

	_, err := parseMarkdownWithGoldmarkTree(body, nil, t.TempDir())
	require.Error(t, err, "per-test setup: was accepted")
	assert.Contains(t, err.Error(), "file-level frontmatter only")
	assert.Contains(t, err.Error(), "uses setup", "error does not name the offending test")
}

// TestApplyWorktreeDefaults pins the one place the JSON-schema defaults, the docs
// and the runtime must agree. `base` is defaulted before `uncommitted`, because
// whether uncommitted work can be replayed is derived from it.
func TestApplyWorktreeDefaults(t *testing.T) {
	tests := []struct {
		name        string
		setup       *shell.Setup
		wantBase    string
		wantUncomm  shell.CloneMode
		wantIgnored shell.CloneMode
		wantWarning bool
	}{
		{
			name: "bare new worktree gets HEAD, clone, clone",
			setup: &shell.Setup{Checkout: &shell.Checkout{
				Worktree: &shell.Worktree{Mode: shell.WorktreeNew},
			}},
			wantBase:    "HEAD",
			wantUncomm:  shell.CloneClone,
			wantIgnored: shell.CloneClone,
		},
		{
			name: "non-HEAD base degrades uncommitted and warns",
			setup: &shell.Setup{Checkout: &shell.Checkout{
				Worktree: &shell.Worktree{Mode: shell.WorktreeNew, Base: "main"},
			}},
			wantBase:    "main",
			wantUncomm:  shell.CloneSkip,
			wantIgnored: shell.CloneClone,
			wantWarning: true,
		},
		{
			name: "explicit uncommitted survives a non-HEAD base",
			setup: &shell.Setup{Checkout: &shell.Checkout{
				Worktree: &shell.Worktree{Mode: shell.WorktreeNew, Base: "main", Uncommitted: shell.CloneClone},
			}},
			wantBase:    "main",
			wantUncomm:  shell.CloneClone,
			wantIgnored: shell.CloneClone,
		},
		{
			name: "explicit values are never overwritten",
			setup: &shell.Setup{Checkout: &shell.Checkout{
				Worktree: &shell.Worktree{
					Mode: shell.WorktreeNew, Base: "v1.0.0",
					Uncommitted: shell.CloneSkip, Ignored: shell.CloneSkip,
				},
			}},
			wantBase:    "v1.0.0",
			wantUncomm:  shell.CloneSkip,
			wantIgnored: shell.CloneSkip,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings := applyWorktreeDefaults(tt.setup)
			wt := tt.setup.Checkout.Worktree
			assert.Equal(t, tt.wantBase, wt.Base)
			assert.Equal(t, tt.wantUncomm, wt.Uncommitted)
			assert.Equal(t, tt.wantIgnored, wt.Ignored)
			assert.Empty(t, wt.Prefix, "prefix stays empty so commons-db autogenerates the branch")
			assert.Equal(t, tt.wantWarning, len(warnings) > 0, "warnings = %v", warnings)
		})
	}
}

// A checkout that declares no worktree, or no checkout at all, must come back
// untouched — there is nothing to clone into, so defaulting would invent one.
func TestApplyWorktreeDefaultsLeavesNonWorktreeSetupsAlone(t *testing.T) {
	noWorktree := &shell.Setup{Checkout: &shell.Checkout{Mode: shell.CheckoutLocal, Path: "."}}
	assert.Empty(t, applyWorktreeDefaults(noWorktree))
	assert.Nil(t, noWorktree.Checkout.Worktree)

	modeNone := &shell.Setup{Checkout: &shell.Checkout{
		Worktree: &shell.Worktree{Mode: shell.WorktreeNone},
	}}
	assert.Empty(t, applyWorktreeDefaults(modeNone))
	assert.Empty(t, modeNone.Checkout.Worktree.Base)
	assert.Empty(t, modeNone.Checkout.Worktree.Uncommitted)

	assert.Empty(t, applyWorktreeDefaults(&shell.Setup{}))
	assert.Empty(t, applyWorktreeDefaults(nil))
}
