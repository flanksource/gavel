package fixtures

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flanksource/gavel/fixtures/record"
)

// TestFrontMatterRecordIsNotSwallowedIntoMetadata mirrors the `setup:` regression
// test: FrontMatter.Metadata is `yaml:",inline"`, so a key without a typed field
// silently becomes a template variable instead of failing.
func TestFrontMatterRecordIsNotSwallowedIntoMetadata(t *testing.T) {
	fm := parseFixtureFrontMatter(t, `---
record: [ansi, http]
exec: echo
---

# Body
`)

	require.NotNil(t, fm.Record, "record: did not bind to the typed field")
	assert.Equal(t, []record.Kind{record.KindANSI, record.KindHTTP}, fm.Record.Kinds())
	assert.Nil(t, fm.Metadata["record"], "record: leaked into Metadata as a template var")
}

func TestRecordBindsFromCommandBlockFrontMatter(t *testing.T) {
	fixtures, err := parseMarkdownWithGoldmark(`
### command: echo hi

`+"```yaml\nrecord: ansi\n```"+`

`+"```bash\necho hi\n```"+`
`, nil, t.TempDir())
	require.NoError(t, err)
	require.Len(t, fixtures, 1)

	assert.Equal(t, []record.Kind{record.KindANSI}, fixtures[0].Test.Record.Kinds())
}

// TestRecordPerTestReplacesFileLevel pins the MergeInto semantic: a per-test
// block replaces the file's wholesale rather than unioning with it, so
// `record: none` on one test is a real opt-out.
func TestRecordPerTestReplacesFileLevel(t *testing.T) {
	fileLevel, err := record.Parse("ansi,http")
	require.NoError(t, err)
	perTest, err := record.Parse("sql")
	require.NoError(t, err)

	merged := ExecFixtureBase{Record: fileLevel}.MergeInto(ExecFixtureBase{Record: perTest})
	assert.Equal(t, []record.Kind{record.KindSQL}, merged.Record.Kinds(), "the file's ansi and http must not leak through")

	optOut, err := record.Parse("none")
	require.NoError(t, err)
	merged = ExecFixtureBase{Record: fileLevel}.MergeInto(ExecFixtureBase{Record: optOut})
	assert.True(t, merged.Record.IsEmpty(), "`record: none` must start nothing")

	merged = ExecFixtureBase{Record: fileLevel}.MergeInto(ExecFixtureBase{})
	assert.Equal(t, []record.Kind{record.KindANSI, record.KindHTTP}, merged.Record.Kinds(),
		"a test that says nothing inherits the file's recorders")
}

func TestRecordBindsFromTableColumn(t *testing.T) {
	fixtures, err := parseMarkdownWithGoldmark(`
| Name | Args | Record |
|------|------|--------|
| records ansi | --help | ansi |
| records nothing | --help | none |
`, nil, t.TempDir())
	require.NoError(t, err)
	require.Len(t, fixtures, 2)

	assert.Equal(t, []record.Kind{record.KindANSI}, fixtures[0].Test.Record.Kinds())
	require.NotNil(t, fixtures[1].Test.Record, "an explicit `none` is a declaration, not an absence")
	assert.True(t, fixtures[1].Test.Record.IsEmpty())
}

func TestRecordTypoIsARedParse(t *testing.T) {
	_, err := parseMarkdownWithGoldmark(`
| Name | Args | Record |
|------|------|--------|
| typo | --help | tcpdump |
`, nil, t.TempDir())
	require.Error(t, err, "a mistyped recorder must fail the parse, not silently record nothing")
	assert.Contains(t, err.Error(), "tcpdump")
}
