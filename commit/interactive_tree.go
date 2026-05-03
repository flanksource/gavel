package commit

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/status"
	"github.com/flanksource/repomap"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
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

// treePickerResult captures the outcome of an interactive tree picker run.
// Selected is the set of paths the user marked for staging. RmCached is the
// subset of files the user just added to .gitignore that were already tracked
// (state != untracked) — the caller must run `git rm --cached` on them so the
// new ignore actually takes effect on subsequent git status calls.
type treePickerResult struct {
	Selected []string
	RmCached []string
}

// runTreePicker is the production tree picker entry point. Tests stub
// runTreePickerFunc instead so they can drive the model directly without
// a real terminal.
func runTreePicker(candidates []status.FileStatus, gitRoot string) (treePickerResult, error) {
	model := newTreeModel(candidates)
	model.gitRoot = gitRoot
	model.appendGitIgnore = appendGitIgnore
	if len(model.visible) == 0 {
		return treePickerResult{}, ErrInteractiveEmpty
	}

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return treePickerResult{}, ErrInteractiveNonTTY
	}
	defer tty.Close()

	prog := tea.NewProgram(model, tea.WithInput(tty), tea.WithOutput(tty), tea.WithAltScreen())
	final, err := prog.Run()
	if err != nil {
		if errors.Is(err, tea.ErrInterrupted) {
			return treePickerResult{}, ErrInteractiveCancelled
		}
		return treePickerResult{}, fmt.Errorf("tree picker: %w", err)
	}
	finished := final.(treeModel)
	if finished.cancelled {
		return treePickerResult{}, ErrInteractiveCancelled
	}
	return treePickerResult{
		Selected: finished.selectedPaths(),
		RmCached: finished.pendingRmCached,
	}, nil
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

// gitIgnoreWriter is the indirection point for tests: the tree model calls
// it to append entries to {gitRoot}/.gitignore. Production wires this to
// commit.appendGitIgnore.
type gitIgnoreWriter func(gitRoot string, entries []string) ([]string, error)

// ignorePromptState tracks the inline submenu opened by the `i` keybinding.
// While non-nil, the tree model intercepts keys to apply the user's choice
// instead of using the normal navigation handler.
type ignorePromptState struct {
	node *treeNode // the row that was highlighted when the user pressed `i`
}

// treeModel is the bubble-tea model + the pure state container that tests
// drive directly via Update().
type treeModel struct {
	root            *treeNode
	visible         []*treeNode
	cursor          int
	width           int
	height          int
	filtering       bool
	filterQuery     string
	cancelled       bool
	submitted       bool
	gitRoot         string
	appendGitIgnore gitIgnoreWriter
	ignorePrompt    *ignorePromptState
	statusLine      string
	pendingRmCached []string
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
	query := normalizedFilter(m.filterQuery)
	if query != "" {
		for _, c := range m.root.Children {
			appendVisibleFiltered(&m.visible, c, query)
		}
		if len(m.visible) == 0 {
			m.cursor = 0
			return
		}
		if m.cursor >= len(m.visible) {
			m.cursor = len(m.visible) - 1
		}
		return
	}
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

func appendVisibleFiltered(out *[]*treeNode, n *treeNode, query string) bool {
	if nodeMatchesFilter(n, query) {
		*out = append(*out, n)
		if n.IsDir {
			for _, c := range n.Children {
				appendVisibleAll(out, c)
			}
		}
		return true
	}

	if !n.IsDir {
		return false
	}

	var childMatches []*treeNode
	for _, c := range n.Children {
		var visibleChild []*treeNode
		if appendVisibleFiltered(&visibleChild, c, query) {
			childMatches = append(childMatches, visibleChild...)
		}
	}
	if len(childMatches) == 0 {
		return false
	}
	*out = append(*out, n)
	*out = append(*out, childMatches...)
	return true
}

func appendVisibleAll(out *[]*treeNode, n *treeNode) {
	*out = append(*out, n)
	if n.IsDir {
		for _, c := range n.Children {
			appendVisibleAll(out, c)
		}
	}
}

func normalizedFilter(query string) string {
	return strings.ToLower(strings.TrimSpace(query))
}

func nodeMatchesFilter(n *treeNode, query string) bool {
	if query == "" {
		return true
	}
	haystack := strings.ToLower(n.Path + " " + n.Name)
	if n.File != nil {
		haystack += " " + strings.ToLower(stateGlyph(*n.File))
		if n.File.FileMap != nil {
			haystack += " " + strings.ToLower(n.File.FileMap.Language)
			for _, s := range n.File.FileMap.Scopes {
				haystack += " " + strings.ToLower(string(s))
			}
		}
	}
	return strings.Contains(haystack, query)
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
	if m.ignorePrompt != nil {
		return m.handleIgnoreKey(msg)
	}
	if m.filtering {
		return m.handleFilterKey(msg)
	}
	switch msg.String() {
	case "ctrl+c", "esc", "q":
		m.cancelled = true
		return m, tea.Quit
	case "/":
		m.filtering = true
		m.statusLine = ""
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
	case "i":
		if n := m.currentNode(); n != nil {
			m.ignorePrompt = &ignorePromptState{node: n}
			m.statusLine = ""
		}
	}
	return m, nil
}

func (m treeModel) handleFilterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		m.cancelled = true
		return m, tea.Quit
	case tea.KeyEsc:
		m.filtering = false
		m.filterQuery = ""
		m.rebuildVisible()
		return m, nil
	case tea.KeyEnter:
		m.filtering = false
		return m, nil
	case tea.KeyBackspace:
		m.filterQuery = trimLastRune(m.filterQuery)
		m.rebuildVisible()
		return m, nil
	case tea.KeyCtrlU:
		m.filterQuery = ""
		m.rebuildVisible()
		return m, nil
	}

	if len(msg.Runes) > 0 {
		m.filterQuery += string(msg.Runes)
		m.rebuildVisible()
	}
	return m, nil
}

func trimLastRune(s string) string {
	if s == "" {
		return ""
	}
	r := []rune(s)
	return string(r[:len(r)-1])
}

func (m treeModel) currentNode() *treeNode {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return nil
	}
	return m.visible[m.cursor]
}

// handleIgnoreKey runs while the inline ignore submenu is open. It applies
// the user's chosen action to {gitRoot}/.gitignore via m.appendGitIgnore,
// removes the matched leaves from the tree, and closes the submenu. Esc or
// any unrelated key cancels without writing.
func (m treeModel) handleIgnoreKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	node := m.ignorePrompt.node
	m.ignorePrompt = nil
	switch msg.String() {
	case "esc", "ctrl+c":
		m.statusLine = ""
	case "f":
		if node.IsDir {
			m.statusLine = styled("cannot ignore as file: cursor is on a folder", styleHelp)
			break
		}
		m.applyGitIgnoreEntry(node.Path)
	case "d":
		entry := folderIgnoreEntry(node)
		if entry == "" {
			m.statusLine = styled("no parent folder to ignore", styleHelp)
			break
		}
		m.applyGitIgnoreEntry(entry)
	case "e":
		ext := extensionGlob(node)
		if ext == "" {
			m.statusLine = styled("no extension to ignore", styleHelp)
			break
		}
		m.applyGitIgnoreEntry(ext)
	}
	return m, nil
}

// folderIgnoreEntry returns the .gitignore folder pattern for the given node
// (e.g. "junk/" for a node under junk/). Returns "" when the node is at the
// repo root and therefore has no enclosing folder.
func folderIgnoreEntry(n *treeNode) string {
	if n.IsDir {
		if n.Path == "" {
			return ""
		}
		return strings.TrimSuffix(filepath.ToSlash(n.Path), "/") + "/"
	}
	dir := filepath.Dir(filepath.ToSlash(n.Path))
	if dir == "." || dir == "" {
		return ""
	}
	return dir + "/"
}

// extensionGlob returns "*<ext>" (e.g. "*.log") when the node is a file with
// a non-empty extension. Returns "" for directories or extension-less files
// (Makefile, LICENSE, etc.) so the caller can disable the action.
func extensionGlob(n *treeNode) string {
	if n.IsDir {
		return ""
	}
	ext := filepath.Ext(n.Name)
	if ext == "" || ext == "." {
		return ""
	}
	return "*" + ext
}

// applyGitIgnoreEntry writes a single entry to .gitignore via the configured
// writer, prunes leaves matched by the new pattern, schedules already-tracked
// matches for `git rm --cached`, and updates the status line.
func (m *treeModel) applyGitIgnoreEntry(entry string) {
	if m.appendGitIgnore == nil {
		m.statusLine = styled("internal: gitignore writer not wired", styleDeleted)
		return
	}
	written, err := m.appendGitIgnore(m.gitRoot, []string{entry})
	if err != nil {
		m.statusLine = styled(fmt.Sprintf("gitignore write failed: %v", err), styleDeleted)
		return
	}

	matched := m.collectIgnoredLeaves(entry)
	for _, leaf := range matched {
		if leaf.File == nil {
			continue
		}
		if leaf.File.State != status.StateUntracked {
			m.pendingRmCached = appendIfMissing(m.pendingRmCached, leaf.Path)
		}
	}
	m.pruneLeaves(matched)
	m.rebuildVisible()
	if m.cursor >= len(m.visible) {
		m.cursor = max(0, len(m.visible)-1)
	}
	if len(m.visible) == 0 {
		m.filterQuery = ""
		m.rebuildVisible()
	}

	switch {
	case len(written) == 0:
		m.statusLine = styled(fmt.Sprintf("already ignored: %s (%d files removed from picker)", entry, len(matched)), styleHelp)
	default:
		m.statusLine = styled(fmt.Sprintf("ignored: %s (%d files removed from picker)", entry, len(matched)), styleStaged)
	}
}

// collectIgnoredLeaves returns every leaf in the tree whose path is matched
// by the given gitignore pattern. Uses go-git's gitignore matcher so we get
// the same semantics as the precommit gitignore check.
func (m treeModel) collectIgnoredLeaves(pattern string) []*treeNode {
	matcher := gitignore.NewMatcher([]gitignore.Pattern{gitignore.ParsePattern(pattern, nil)})
	var out []*treeNode
	var walk func(n *treeNode)
	walk = func(n *treeNode) {
		if !n.IsDir {
			if matcher.Match(splitGitPath(n.Path), false) {
				out = append(out, n)
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

// pruneLeaves removes the given leaves from the tree, then prunes any
// directory left with no descendants.
func (m treeModel) pruneLeaves(leaves []*treeNode) {
	if len(leaves) == 0 {
		return
	}
	mark := make(map[*treeNode]struct{}, len(leaves))
	for _, l := range leaves {
		mark[l] = struct{}{}
	}
	var clean func(n *treeNode)
	clean = func(n *treeNode) {
		kept := n.Children[:0]
		for _, c := range n.Children {
			if !c.IsDir {
				if _, drop := mark[c]; drop {
					continue
				}
				kept = append(kept, c)
				continue
			}
			clean(c)
			if len(c.Children) > 0 {
				kept = append(kept, c)
			}
		}
		n.Children = kept
	}
	clean(m.root)
}

func appendIfMissing(slice []string, v string) []string {
	for _, s := range slice {
		if s == v {
			return slice
		}
	}
	return append(slice, v)
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
	visibleLeaves := countVisibleLeaves(m.visible)
	filterActive := normalizedFilter(m.filterQuery) != ""

	b.WriteString(styled("Select files to commit", styleHeader))
	b.WriteString("  ")
	b.WriteString(styled(fmt.Sprintf("(%d / %d selected)", selectedCount, totalLeaves), styleMuted))
	if filterActive {
		b.WriteString("  ")
		b.WriteString(styled(fmt.Sprintf("filter=%q (%d files)", m.filterQuery, visibleLeaves), styleMuted))
	}
	b.WriteByte('\n')
	if m.ignorePrompt != nil {
		b.WriteString(renderIgnorePrompt(m.ignorePrompt.node))
	} else if m.filtering {
		b.WriteString(styled(
			fmt.Sprintf("  filter: %s  type=search  backspace=delete  ctrl+u=clear  enter=keep  esc=clear",
				m.filterQuery),
			styleHelp,
		))
	} else {
		b.WriteString(styled(
			"  /=filter  space=toggle  a=toggle folder  g=toggle Go  t=toggle tests  ctrl+a=all  i=ignore  enter=confirm  esc=cancel",
			styleHelp,
		))
	}
	if m.statusLine != "" {
		b.WriteByte('\n')
		b.WriteString("  ")
		b.WriteString(m.statusLine)
	}
	b.WriteString("\n\n")

	pageSize := max(m.height-5, 5)
	start := 0
	if m.cursor >= pageSize {
		start = m.cursor - pageSize + 1
	}
	end := min(start+pageSize, len(m.visible))

	for i := start; i < end; i++ {
		b.WriteString(renderRow(m.visible[i], i == m.cursor, filterActive))
		b.WriteByte('\n')
	}
	if len(m.visible) == 0 {
		b.WriteString(styled("  No files match the current filter", styleHelp))
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

func countVisibleLeaves(nodes []*treeNode) int {
	total := 0
	for _, n := range nodes {
		if !n.IsDir {
			total++
		}
	}
	return total
}

// renderIgnorePrompt builds the inline submenu shown when the user pressed
// `i`. It dynamically dims options that don't apply to the highlighted node
// (e.g. a top-level file has no folder; an extension-less file has no glob).
func renderIgnorePrompt(n *treeNode) string {
	var parts []string
	parts = append(parts, styled("ignore", styleHeader)+":")
	if !n.IsDir {
		parts = append(parts, styled(fmt.Sprintf("f=file (%s)", n.Path), styleHelp))
	} else {
		parts = append(parts, styled("f=file (n/a — folder selected)", styleMuted))
	}
	if entry := folderIgnoreEntry(n); entry != "" {
		parts = append(parts, styled(fmt.Sprintf("d=folder (%s)", entry), styleHelp))
	} else {
		parts = append(parts, styled("d=folder (n/a — top level)", styleMuted))
	}
	if ext := extensionGlob(n); ext != "" {
		parts = append(parts, styled(fmt.Sprintf("e=ext (%s)", ext), styleHelp))
	} else {
		parts = append(parts, styled("e=ext (n/a — no extension)", styleMuted))
	}
	parts = append(parts, styled("esc=cancel", styleHelp))
	return "  " + strings.Join(parts, "  ")
}

func renderRow(n *treeNode, isCursor bool, forceExpanded bool) string {
	cursor := "  "
	if isCursor {
		cursor = styled("▶ ", styleCursor)
	}
	indent := strings.Repeat("  ", n.Depth-1)
	check := checkboxFor(n)
	name := n.Name
	if n.IsDir {
		expand := "▾ "
		if !forceExpanded && !n.Expanded {
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
	if !f.ModifiedAt.IsZero() {
		if age := status.HumanAge(time.Since(f.ModifiedAt)); age != "" {
			parts = append(parts, styled(age+" ago", styleMuted))
		}
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
