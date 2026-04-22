package choose

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewModel_DefaultsApplied(t *testing.T) {
	m := newModel([]string{"a", "b", "c"})
	assert.Equal(t, 3, len(m.items))
	assert.Equal(t, 10, m.height)
	assert.True(t, m.showHelp)
	assert.Equal(t, 0, m.index)
}

func TestNewModel_WithOptions(t *testing.T) {
	m := newModel([]string{"a", "b"}, WithHeader("Pick:"), WithHeight(5), WithLimit(1))
	assert.Equal(t, "Pick:", m.header)
	assert.Equal(t, 5, m.height)
	assert.Equal(t, 1, m.limit)
	assert.True(t, m.singleSelect)
}

func TestModel_NavigationWraps(t *testing.T) {
	m := newModel([]string{"a", "b", "c"}, WithLimit(0))

	// Move down past end wraps to 0
	m.index = 2
	updated, _ := m.Update(tea.KeyPressMsg{Code: 'j'})
	m = updated.(model)
	assert.Equal(t, 0, m.index)

	// Move up past 0 wraps to end
	updated, _ = m.Update(tea.KeyPressMsg{Code: 'k'})
	m = updated.(model)
	assert.Equal(t, 2, m.index)
}

func TestModel_ToggleSelection(t *testing.T) {
	m := newModel([]string{"a", "b", "c"}, WithLimit(0))

	// Toggle item 0
	updated, _ := m.Update(tea.KeyPressMsg{Code: ' '})
	m = updated.(model)
	assert.True(t, m.items[0].selected)
	assert.Equal(t, 1, m.numSelected)

	// Toggle again to deselect
	updated, _ = m.Update(tea.KeyPressMsg{Code: ' '})
	m = updated.(model)
	assert.False(t, m.items[0].selected)
	assert.Equal(t, 0, m.numSelected)
}

func TestModel_SubmitReturnsSelected(t *testing.T) {
	m := newModel([]string{"a", "b", "c"}, WithLimit(0))

	// Select item 0
	updated, _ := m.Update(tea.KeyPressMsg{Code: ' '})
	m = updated.(model)

	// Move down and select item 1
	updated, _ = m.Update(tea.KeyPressMsg{Code: 'j'})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyPressMsg{Code: ' '})
	m = updated.(model)

	// Submit
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated.(model)
	assert.True(t, m.submitted)
	assert.True(t, m.items[0].selected)
	assert.True(t, m.items[1].selected)
	assert.False(t, m.items[2].selected)
}

func TestModel_SingleSelectToggleReplacesSelection(t *testing.T) {
	m := newModel([]string{"a", "b", "c"}, WithLimit(1))

	updated, _ := m.Update(tea.KeyPressMsg{Code: ' '})
	m = updated.(model)
	assert.True(t, m.items[0].selected)
	assert.False(t, m.items[1].selected)
	assert.Equal(t, 1, m.numSelected)

	updated, _ = m.Update(tea.KeyPressMsg{Code: 'j'})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyPressMsg{Code: ' '})
	m = updated.(model)
	assert.False(t, m.items[0].selected)
	assert.True(t, m.items[1].selected)
	assert.False(t, m.items[2].selected)
	assert.Equal(t, 1, m.numSelected)
}

func TestModel_SingleSelectSubmitChoosesCurrentItem(t *testing.T) {
	m := newModel([]string{"a", "b", "c"}, WithLimit(1))

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'j'})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated.(model)

	assert.True(t, m.submitted)
	assert.False(t, m.items[0].selected)
	assert.True(t, m.items[1].selected)
	assert.False(t, m.items[2].selected)
}

func TestModel_QuitDoesNotSubmit(t *testing.T) {
	m := newModel([]string{"a", "b"}, WithLimit(0))

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	m = updated.(model)
	assert.True(t, m.quitting)
	assert.False(t, m.submitted)
}

func TestModel_SelectAll(t *testing.T) {
	m := newModel([]string{"a", "b", "c"}, WithLimit(0))

	// ctrl+a to select all
	updated, _ := m.Update(tea.KeyPressMsg{Code: 'a', Mod: tea.ModCtrl})
	m = updated.(model)
	assert.Equal(t, 3, m.numSelected)

	// ctrl+a again to deselect all
	updated, _ = m.Update(tea.KeyPressMsg{Code: 'a', Mod: tea.ModCtrl})
	m = updated.(model)
	assert.Equal(t, 0, m.numSelected)
}

func TestModel_ShowDetail(t *testing.T) {
	called := false
	detailFn := func(i int) string {
		called = true
		return "detail for " + strings.Repeat("x", i)
	}

	m := newModel([]string{"a", "b"}, WithDetailFunc(detailFn))

	// Without terminal width, no detail panel
	assert.False(t, m.showDetail())

	// With wide terminal, detail panel shown
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	m = updated.(model)
	assert.True(t, m.showDetail())

	// Narrow terminal hides detail
	updated, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	m = updated.(model)
	assert.False(t, m.showDetail())

	// Render with detail
	updated, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	m = updated.(model)
	_ = m.View()
	assert.True(t, called)
}

func TestModel_ViewWithoutDetail(t *testing.T) {
	m := newModel([]string{"item1", "item2"}, WithHeader("Test:"), WithLimit(0))

	view := m.View()
	content := view.Content
	assert.Contains(t, content, "Test:")
	assert.Contains(t, content, "item1")
	assert.Contains(t, content, "item2")
}

func TestModel_ViewSingleSelectUsesRadioMarkers(t *testing.T) {
	m := newModel([]string{"item1", "item2"}, WithHeader("Pick one:"), WithLimit(1))

	view := m.View()
	content := view.Content
	assert.Contains(t, content, "( ) item1")
	assert.Contains(t, content, "( ) item2")
	assert.NotContains(t, content, "[ ] item1")
}

func TestModel_WindowSizeMsg(t *testing.T) {
	m := newModel([]string{"a"})

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 150, Height: 40})
	m = updated.(model)
	assert.Equal(t, 150, m.termWidth)
	assert.Equal(t, 40, m.termHeight)
}

func TestFormatTODOListItem(t *testing.T) {
	result := FormatTODOListItem("Fix bug", "⚠ HIGH", "✗ FAIL", "src/auth.go")
	assert.Contains(t, result, "Fix bug")
	assert.Contains(t, result, "⚠ HIGH")
	assert.Contains(t, result, "✗ FAIL")
	assert.Contains(t, result, "src/auth.go")
}

func TestFormatTODOListItem_NoPriority(t *testing.T) {
	result := FormatTODOListItem("Simple task", "", "● PEND", "")
	assert.Contains(t, result, "Simple task")
	assert.Contains(t, result, "● PEND")
	assert.NotContains(t, result, "\n")
}

func TestFormatTODODetail(t *testing.T) {
	detail := FormatTODODetail(DetailOptions{
		Title:          "Fix auth",
		Priority:       "high",
		Status:         "failed",
		Attempts:       3,
		LastRun:        "2h ago",
		PRNumber:       42,
		PRAuthor:       "reviewer",
		Paths:          []string{"src/auth.go", "src/middleware.go"},
		Tests:          []string{"test_auth_flow", "test_token_refresh"},
		Implementation: "Update the auth middleware\nto handle expired tokens\nproperly.",
	})

	assert.Contains(t, detail, "Fix auth")
	assert.Contains(t, detail, "high")
	assert.Contains(t, detail, "failed")
	assert.Contains(t, detail, "3")
	assert.Contains(t, detail, "2h ago")
	assert.Contains(t, detail, "#42")
	assert.Contains(t, detail, "reviewer")
	assert.Contains(t, detail, "src/auth.go")
	assert.Contains(t, detail, "test_auth_flow")
	assert.Contains(t, detail, "Update the auth middleware")
}

func TestFormatTODODetail_Minimal(t *testing.T) {
	detail := FormatTODODetail(DetailOptions{
		Title:  "Minimal",
		Status: "pending",
	})
	require.Contains(t, detail, "Minimal")
	require.Contains(t, detail, "pending")
	require.NotContains(t, detail, "PR Info")
	require.NotContains(t, detail, "Paths")
}

func TestTruncateHeight(t *testing.T) {
	input := "line1\nline2\nline3\nline4\nline5"
	result := truncateHeight(input, 3)
	assert.Equal(t, "line1\nline2\nline3", result)

	// No truncation needed
	short := "a\nb"
	assert.Equal(t, short, truncateHeight(short, 5))
}
