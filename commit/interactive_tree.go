package commit

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/status"
	"github.com/flanksource/repomap"
)

// Style strings reused from status/pretty.go so the tree picker colorizes
// chips identically to `gavel status`. Resolved to ANSI escapes via
// clicky.Text(...).ANSI().
const (
	styleStaged       = "text-green-500 font-bold"
	styleModified     = "text-yellow-500 font-bold"
	styleDeleted      = "text-red-500 font-bold"
	styleUntracked    = "text-purple-500 font-bold"
	styleConflicted   = "text-red-600 font-bold underline"
	styleScope        = "text-cyan-600"
	styleLanguage     = "text-muted italic"
	styleMuted        = "text-muted"
	styleCursor       = "text-cyan-600 font-bold"
	styleHeader       = "font-bold"
	styleHelp         = "text-muted"
	styleCheckOn      = "text-green-500 font-bold"
	styleCheckPartial = "text-muted"
)

func styled(s, style string) string { return clicky.Text(s, style).ANSI() }

// runTreePicker is the production tree picker entry point. Tests stub
// runTreePickerFunc instead so they can drive the model directly without
// a real terminal.
func runTreePicker(candidates []status.FileStatus) ([]string, error) {
	model := newTreeModel(candidates)
	if len(model.visible) == 0 {
		return nil, ErrInteractiveEmpty
	}

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return nil, ErrInteractiveNonTTY
	}
	defer tty.Close()

	prog := tea.NewProgram(model, tea.WithInput(tty), tea.WithOutput(tty), tea.WithAltScreen())
	final, err := prog.Run()
	if err != nil {
		if errors.Is(err, tea.ErrInterrupted) {
			return nil, ErrInteractiveCancelled
		}
		return nil, fmt.Errorf("tree picker: %w", err)
	}
	finished := final.(treeModel)
	if finished.cancelled {
		return nil, ErrInteractiveCancelled
	}
	return finished.selectedPaths(), nil
}

// treeNode is one row in the tree. Either a directory or a file.
type treeNode struct {
	Name     string
	Path     string
	Depth    int
	IsDir    bool
	Expanded bool
	Selected bool
	Children []*treeNode
	File     *status.FileStatus
	Parent   *treeNode
}

// treeModel is the bubble-tea model + the pure state container that tests
// drive directly via Update().
type treeModel struct {
	root      *treeNode
	visible   []*treeNode
	cursor    int
	width     int
	height    int
	cancelled bool
	submitted bool
}

func newTreeModel(files []status.FileStatus) treeModel {
	root := &treeNode{Name: "", Path: "", IsDir: true, Expanded: true}
	for i := range files {
		insertFile(root, &files[i])
	}
	sortTree(root)
	m := treeModel{root: root, height: 20, width: 80}
	m.rebuildVisible()
	return m
}

// insertFile walks/creates directory nodes for the file's path and attaches
// the file as a leaf node.
func insertFile(root *treeNode, f *status.FileStatus) {
	parts := strings.Split(f.Path, "/")
	current := root
	for i, segment := range parts {
		isLast := i == len(parts)-1
		if isLast {
			leaf := &treeNode{
				Name:   segment,
				Path:   f.Path,
				Depth:  i + 1,
				IsDir:  false,
				File:   f,
				Parent: current,
			}
			current.Children = append(current.Children, leaf)
			return
		}
		dirPath := strings.Join(parts[:i+1], "/")
		child := findChild(current, segment)
		if child == nil {
			child = &treeNode{
				Name:     segment,
				Path:     dirPath,
				Depth:    i + 1,
				IsDir:    true,
				Expanded: true,
				Parent:   current,
			}
			current.Children = append(current.Children, child)
		}
		current = child
	}
}

func findChild(n *treeNode, name string) *treeNode {
	for _, c := range n.Children {
		if c.Name == name && c.IsDir {
			return c
		}
	}
	return nil
}

// sortTree sorts each node's children: directories first, then files,
// alphabetical within each.
func sortTree(n *treeNode) {
	sort.SliceStable(n.Children, func(i, j int) bool {
		a, b := n.Children[i], n.Children[j]
		if a.IsDir != b.IsDir {
			return a.IsDir && !b.IsDir
		}
		return a.Name < b.Name
	})
	for _, c := range n.Children {
		if c.IsDir {
			sortTree(c)
		}
	}
}

func (m *treeModel) rebuildVisible() {
	m.visible = m.visible[:0]
	for _, c := range m.root.Children {
		appendVisible(&m.visible, c)
	}
	if m.cursor >= len(m.visible) {
		m.cursor = max(0, len(m.visible)-1)
	}
}

func appendVisible(out *[]*treeNode, n *treeNode) {
	*out = append(*out, n)
	if n.IsDir && n.Expanded {
		for _, c := range n.Children {
			appendVisible(out, c)
		}
	}
}

// Init satisfies tea.Model.
func (m treeModel) Init() tea.Cmd { return nil }

func (m treeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

// handleKey is split out so tests can call Update with synthetic
// tea.KeyMsg values without spinning up a Program.
func (m treeModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc", "q":
		m.cancelled = true
		return m, tea.Quit
	case "enter":
		m.submitted = true
		return m, tea.Quit
	case "down", "j":
		if m.cursor < len(m.visible)-1 {
			m.cursor++
		}
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "right", "l":
		if n := m.currentNode(); n != nil && n.IsDir && !n.Expanded {
			n.Expanded = true
			m.rebuildVisible()
		}
	case "left", "h":
		if n := m.currentNode(); n != nil && n.IsDir && n.Expanded {
			n.Expanded = false
			m.rebuildVisible()
		} else if n != nil && n.Parent != nil && n.Parent != m.root {
			// Jump cursor to parent so user can collapse the parent next.
			for i, v := range m.visible {
				if v == n.Parent {
					m.cursor = i
					break
				}
			}
		}
	case " ":
		if n := m.currentNode(); n != nil {
			toggleNode(n)
		}
	case "a":
		if n := m.currentNode(); n != nil {
			toggleNode(n.containerOrSelf())
		}
	case "g":
		m.toggleByLanguage("Go")
	case "t":
		m.toggleByScope(repomap.ScopeTypeTest)
	case "ctrl+a":
		// invert: select all if any unselected, else clear all
		toggleNode(m.root)
	}
	return m, nil
}

func (m treeModel) currentNode() *treeNode {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return nil
	}
	return m.visible[m.cursor]
}

func (n *treeNode) containerOrSelf() *treeNode {
	if n.IsDir {
		return n
	}
	if n.Parent != nil {
		return n.Parent
	}
	return n
}

// toggleNode flips the selected state of a node. For directories, the new
// state is propagated to all leaf descendants. The new state is "true" iff
// at least one leaf is currently unselected.
func toggleNode(n *treeNode) {
	target := !allLeavesSelected(n)
	setLeavesSelected(n, target)
}

func allLeavesSelected(n *treeNode) bool {
	if !n.IsDir {
		return n.Selected
	}
	if len(n.Children) == 0 {
		return false
	}
	for _, c := range n.Children {
		if !allLeavesSelected(c) {
			return false
		}
	}
	return true
}

func anyLeafSelected(n *treeNode) bool {
	if !n.IsDir {
		return n.Selected
	}
	for _, c := range n.Children {
		if anyLeafSelected(c) {
			return true
		}
	}
	return false
}

func setLeavesSelected(n *treeNode, v bool) {
	if !n.IsDir {
		n.Selected = v
		return
	}
	for _, c := range n.Children {
		setLeavesSelected(c, v)
	}
}

// toggleByLanguage flips selection on every file whose FileMap.Language
// matches. Sets the new value to true if any matching file is currently
// unselected, else false.
func (m treeModel) toggleByLanguage(language string) {
	matches := collectLeavesWhere(m.root, func(f *status.FileStatus) bool {
		return f.FileMap != nil && strings.EqualFold(f.FileMap.Language, language)
	})
	flipLeaves(matches)
}

func (m treeModel) toggleByScope(scope repomap.ScopeType) {
	matches := collectLeavesWhere(m.root, func(f *status.FileStatus) bool {
		if f.FileMap == nil {
			return false
		}
		for _, s := range f.FileMap.Scopes {
			if s == scope {
				return true
			}
		}
		return false
	})
	flipLeaves(matches)
}

func collectLeavesWhere(n *treeNode, pred func(*status.FileStatus) bool) []*treeNode {
	var out []*treeNode
	var walk func(n *treeNode)
	walk = func(n *treeNode) {
		if !n.IsDir {
			if n.File != nil && pred(n.File) {
				out = append(out, n)
			}
			return
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(n)
	return out
}

func flipLeaves(leaves []*treeNode) {
	if len(leaves) == 0 {
		return
	}
	target := false
	for _, l := range leaves {
		if !l.Selected {
			target = true
			break
		}
	}
	for _, l := range leaves {
		l.Selected = target
	}
}

func (m treeModel) selectedPaths() []string {
	var out []string
	var walk func(n *treeNode)
	walk = func(n *treeNode) {
		if !n.IsDir {
			if n.Selected {
				out = append(out, n.Path)
			}
			return
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(m.root)
	return out
}

// View renders the tree as a single string with ANSI escape sequences
// embedded inline (via clicky.Text(...).ANSI()). The styling matches
// `gavel status` so users see consistent colors across both commands.
func (m treeModel) View() string {
	var b strings.Builder
	selectedCount := len(m.selectedPaths())
	totalLeaves := countLeaves(m.root)

	b.WriteString(styled("Select files to commit", styleHeader))
	b.WriteString("  ")
	b.WriteString(styled(fmt.Sprintf("(%d / %d selected)", selectedCount, totalLeaves), styleMuted))
	b.WriteByte('\n')
	b.WriteString(styled(
		"  space=toggle  a=toggle folder  g=toggle Go  t=toggle tests  ctrl+a=all  enter=confirm  esc=cancel",
		styleHelp,
	))
	b.WriteString("\n\n")

	pageSize := max(m.height-5, 5)
	start := 0
	if m.cursor >= pageSize {
		start = m.cursor - pageSize + 1
	}
	end := min(start+pageSize, len(m.visible))

	for i := start; i < end; i++ {
		b.WriteString(renderRow(m.visible[i], i == m.cursor))
		b.WriteByte('\n')
	}
	return b.String()
}

func countLeaves(n *treeNode) int {
	if !n.IsDir {
		return 1
	}
	total := 0
	for _, c := range n.Children {
		total += countLeaves(c)
	}
	return total
}

func renderRow(n *treeNode, isCursor bool) string {
	cursor := "  "
	if isCursor {
		cursor = styled("▶ ", styleCursor)
	}
	indent := strings.Repeat("  ", n.Depth-1)
	check := checkboxFor(n)
	name := n.Name
	if n.IsDir {
		expand := "▾ "
		if !n.Expanded {
			expand = "▸ "
		}
		name = expand + name + "/"
	}
	if isCursor {
		name = styled(name, styleHeader)
	}
	row := fmt.Sprintf("%s%s%s %s", cursor, indent, check, name)
	if !n.IsDir && n.File != nil {
		row += "  " + chipsFor(*n.File)
	}
	return row
}

func checkboxFor(n *treeNode) string {
	if !n.IsDir {
		if n.Selected {
			return styled("[x]", styleCheckOn)
		}
		return "[ ]"
	}
	switch {
	case allLeavesSelected(n):
		return styled("[x]", styleCheckOn)
	case anyLeafSelected(n):
		return styled("[~]", styleCheckPartial)
	default:
		return "[ ]"
	}
}

func chipsFor(f status.FileStatus) string {
	parts := make([]string, 0, 5)

	if label := stateGlyph(f); label != "" {
		parts = append(parts, styled(label, stateStyle(f.State)))
	}
	if f.FileMap != nil {
		if lang := strings.TrimSpace(f.FileMap.Language); lang != "" {
			parts = append(parts, styled(lang, styleLanguage))
		}
		for _, s := range f.FileMap.Scopes {
			parts = append(parts, styled(string(s), styleScope))
		}
	}
	if f.Adds > 0 || f.Dels > 0 {
		delta := styled(fmt.Sprintf("+%d", f.Adds), styleStaged) + " " +
			styled(fmt.Sprintf("-%d", f.Dels), styleDeleted)
		parts = append(parts, delta)
	}

	separator := styled(" · ", styleMuted)
	return strings.Join(parts, separator)
}

func stateGlyph(f status.FileStatus) string {
	switch f.State {
	case status.StateStaged:
		return "staged"
	case status.StateUntracked:
		return "untracked"
	case status.StateBoth:
		return "staged+unstaged"
	case status.StateUnstaged:
		return "unstaged"
	case status.StateConflict:
		return "conflict"
	default:
		return ""
	}
}

func stateStyle(s status.FileState) string {
	switch s {
	case status.StateStaged, status.StateBoth:
		return styleStaged
	case status.StateUnstaged:
		return styleModified
	case status.StateUntracked:
		return styleUntracked
	case status.StateConflict:
		return styleConflicted
	default:
		return styleMuted
	}
}
