package todos

import (
	"fmt"
	"strings"
	"testing"

	"github.com/flanksource/gavel/todos/types"
)

func indexTODO(shortID, title, body string) *types.TODO {
	return &types.TODO{
		ID:              "id-" + shortID,
		ShortID:         shortID,
		MarkdownBody:    body,
		TODOFrontmatter: types.TODOFrontmatter{Title: title, Status: types.StatusPending, Priority: types.PriorityHigh},
	}
}

func TestBuildBacklogIndexRendersEntries(t *testing.T) {
	index := BuildBacklogIndex([]*types.TODO{
		indexTODO("ab12cd", "Fix the parser panic", "The parser panics on an unterminated string."),
	}, nil, nil)

	want := "- ab12cd  Fix the parser panic  (pending, high)\n  > The parser panics on an unterminated string."
	if index != want {
		t.Fatalf("index =\n%q\nwant\n%q", index, want)
	}
}

// Titles alone cannot separate two TODOs that both say "parser crash", so the
// excerpt is what makes the index worth reading before opening anything.
func TestBuildBacklogIndexExcerptsTheBody(t *testing.T) {
	long := strings.TrimSpace(strings.Repeat("the parser panics on an unterminated string ", 20))
	index := BuildBacklogIndex([]*types.TODO{indexTODO("ab12cd", "Parser", long)}, nil, nil)

	excerpt := strings.TrimPrefix(strings.Split(index, "\n")[1], "  > ")
	if !strings.HasSuffix(excerpt, "…") {
		t.Errorf("a truncated excerpt should say so: %q", excerpt)
	}
	if runes := len([]rune(excerpt)); runes > BacklogExcerptRunes+1 {
		t.Errorf("excerpt is %d runes, want at most %d plus the ellipsis", runes, BacklogExcerptRunes)
	}
	if strings.HasSuffix(strings.TrimSuffix(excerpt, "…"), " ") {
		t.Errorf("excerpt should not end mid-space: %q", excerpt)
	}
}

func TestBuildBacklogIndexFlattensAndOmitsBodies(t *testing.T) {
	index := BuildBacklogIndex([]*types.TODO{
		indexTODO("ab12cd", "Multi", "First line.\n\nSecond   line."),
		indexTODO("ff0011", "Empty", "   "),
	}, nil, nil)

	if !strings.Contains(index, "  > First line. Second line.") {
		t.Errorf("a multi-line body must flatten to one row:\n%s", index)
	}
	for _, line := range strings.Split(index, "\n") {
		if strings.Contains(line, "ff0011") && strings.Contains(index, "ff0011\n  >") {
			t.Errorf("a TODO with no body should carry no excerpt line:\n%s", index)
		}
	}
	if strings.Count(index, "  > ") != 1 {
		t.Errorf("want exactly one excerpt line:\n%s", index)
	}
}

// The others in a bulk request are being triaged concurrently, and a fold naming
// one that its own verdict already closed is refused — so the agent has to be able
// to tell them apart.
func TestBuildBacklogIndexMarksTheBatch(t *testing.T) {
	index := BuildBacklogIndex([]*types.TODO{
		indexTODO("ab12cd", "In the batch", "body"),
		indexTODO("ff0011", "Not in the batch", "body"),
	}, nil, []string{"AB12CD"})

	if !strings.Contains(index, "(pending, high)  [in this batch]") {
		t.Errorf("the batch entry was not marked (match is case-insensitive):\n%s", index)
	}
	if strings.Count(index, "[in this batch]") != 1 {
		t.Errorf("only the selected entry should be marked:\n%s", index)
	}
}

// A TODO cannot duplicate itself, and seeing its own row invites the agent to say
// it does.
func TestBuildBacklogIndexExcludesTheTriagedTODO(t *testing.T) {
	self := indexTODO("ab12cd", "Being triaged", "body")
	index := BuildBacklogIndex([]*types.TODO{self, indexTODO("ff0011", "Other", "body")}, []*types.TODO{self}, nil)

	if strings.Contains(index, "ab12cd") {
		t.Errorf("the triaged TODO must not appear in its own backlog:\n%s", index)
	}
	if !strings.Contains(index, "ff0011") {
		t.Errorf("the rest of the backlog should remain:\n%s", index)
	}
	if BuildBacklogIndex([]*types.TODO{self}, []*types.TODO{self}, nil) != "" {
		t.Error("an index with nothing left in it should render empty, not a heading with no rows")
	}
}

// An agent told it is seeing the whole backlog will assert "no duplicate" with a
// confidence a truncated list does not support.
func TestBuildBacklogIndexReportsTruncation(t *testing.T) {
	candidates := make([]*types.TODO, 0, MaxBacklogEntries+3)
	for i := range MaxBacklogEntries + 3 {
		candidates = append(candidates, indexTODO(fmt.Sprintf("id%04d", i), "Todo", "body"))
	}
	index := BuildBacklogIndex(candidates, nil, nil)

	if strings.Count(index, "\n- ") != MaxBacklogEntries-1 {
		t.Errorf("index listed %d entries, want %d", strings.Count(index, "\n- ")+1, MaxBacklogEntries)
	}
	if !strings.Contains(index, "(3 further TODOs are not listed") {
		t.Errorf("truncation must be stated in the text:\n%s", index[len(index)-200:])
	}
}
