package fixtures

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFixtureForTest(t *testing.T) {
	// Create a temporary test fixture
	tmpDir := t.TempDir()
	fixtureFile := filepath.Join(tmpDir, "test.md")

	fixtureContent := `---
codeBlocks: [bash, python]
build: make test
---

# Test Section

Test fixture with multiple code blocks.

` + "```bash exitCode=0\necho \"test\"\n```" + `

` + "```python\nprint(\"skip\")\n```" + `
`

	err := os.WriteFile(fixtureFile, []byte(fixtureContent), 0644)
	require.NoError(t, err)

	// Parse using helper
	node, err := ParseFixtureForTest(fixtureFile)
	require.NoError(t, err)
	require.NotNil(t, node)

	// Verify it parsed the structure
	assert.Equal(t, "test.md", node.Name)
	assert.Equal(t, FileNode, node.Type)

	// Note: Full validation of parsed content would require
	// walking the tree and inspecting children. This basic test
	// verifies the helper function works without errors.
}
