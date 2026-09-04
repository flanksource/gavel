package portable

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPortableMarkdownRoundTripsSupportedFields(t *testing.T) {
	issueID := uuid.MustParse("11111111-2222-4333-8444-555555555555")
	workspaceID := uuid.MustParse("aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee")
	body := `Keep the portable description.

## Acceptance Criteria

- [ ] IDs survive export and import`
	issue := native.Issue{
		ID: issueID, WorkspaceID: workspaceID,
		Title: "Round-trip portable fields", Body: body,
		Verification: "## Focused tests\n\n```bash\ngo test ./todos/portable\n```\n\n## Assertions\n\n- export remains lossless",
		Labels:       []string{"database", "todos"},
		Priority:     native.PriorityHigh,
		Status:       native.StatusVerified,
	}

	content, err := exportMarkdown(issue)
	require.NoError(t, err)
	dir := t.TempDir()
	path := filepath.Join(dir, "portable.md")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	parsed, err := todos.ParseTODO(path)
	require.NoError(t, err)
	assert.Equal(t, issue.Body, parsed.MarkdownBody)
	assert.Equal(t, issue.Verification, parsed.VerificationMarkdown)
	imported, err := importIssueFromTODO(parsed, dir, workspaceID)
	require.NoError(t, err)
	assert.Equal(t, issue.ID, imported.ID)
	assert.Equal(t, issue.Title, imported.Title)
	assert.Equal(t, issue.Body, imported.Body)
	assert.Equal(t, issue.Verification, imported.Verification)
	assert.ElementsMatch(t, issue.Labels, imported.Labels)
	assert.Equal(t, issue.Priority, imported.Priority)
	assert.Equal(t, issue.Status, imported.Status)
	assert.Equal(t, issueID.String(), metadataString(parsed.Metadata, metadataID))
}

func TestPortableImportDerivesStableIDWithoutMutatingFile(t *testing.T) {
	workspaceID := uuid.MustParse("aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee")
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "legacy.md")
	todo := &types.TODO{
		FilePath: path,
		TODOFrontmatter: types.TODOFrontmatter{
			Title: "Legacy file", Priority: types.PriorityMedium, Status: types.StatusPending,
		},
		MarkdownBody: "body",
	}
	first, err := importIssueFromTODO(todo, dir, workspaceID)
	require.NoError(t, err)
	second, err := importIssueFromTODO(todo, dir, workspaceID)
	require.NoError(t, err)
	assert.Equal(t, first.ID, second.ID)

	otherWorkspace := uuid.MustParse("bbbbbbbb-cccc-4ddd-8eee-ffffffffffff")
	other, err := importIssueFromTODO(todo, dir, otherWorkspace)
	require.NoError(t, err)
	assert.NotEqual(t, first.ID, other.ID)
}

func TestPortableMappingRejectsTransientOrLossyValues(t *testing.T) {
	for _, status := range []types.Status{
		types.StatusInProgress, types.StatusReview, types.StatusAsk,
		types.StatusFailed, types.StatusUnverified, types.StatusSkipped,
	} {
		_, err := importStatus(status)
		require.ErrorContains(t, err, "not portable", status)
	}
	_, err := exportStatus(native.StatusCancelled)
	require.ErrorContains(t, err, "not representable")
	_, err = exportPriority(native.PriorityCritical)
	require.ErrorContains(t, err, "not representable")
}

func TestPortableExportRefusesUnrelatedCollision(t *testing.T) {
	id := uuid.MustParse("11111111-2222-4333-8444-555555555555")
	path := filepath.Join(t.TempDir(), "collision.md")
	require.NoError(t, os.WriteFile(path, []byte("unrelated"), 0o644))
	require.ErrorContains(t, refuseUnrelatedCollision(path, id, false), "--force")
	require.NoError(t, refuseUnrelatedCollision(path, id, true))

	frontmatter := types.TODOFrontmatter{
		Title: "Matching", Priority: types.PriorityMedium, Status: types.StatusPending,
	}
	frontmatter.Metadata = map[string]any{metadataID: id.String()}
	content, err := todos.WriteFrontmatter(&frontmatter, "body")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	require.NoError(t, refuseUnrelatedCollision(path, id, false))
}

func TestPortableDirectoryResolvesRelativeToExplicitWorkspace(t *testing.T) {
	workDir := filepath.Join(t.TempDir(), "workspace")
	resolved, err := portableDirectory("import directory", workDir, "backup/todos")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(workDir, "backup", "todos"), resolved)

	resolved, err = portableDirectory("export directory", workDir, "")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(workDir, DefaultDirectory), resolved)
}
