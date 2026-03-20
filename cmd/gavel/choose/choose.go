package choose

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/paginator"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Option func(*model)

func WithHeader(h string) Option { return func(m *model) { m.header = h } }
func WithLimit(n int) Option     { return func(m *model) { m.limit = n } }
func WithShowHelp(b bool) Option { return func(m *model) { m.showHelp = b } }
func WithHeight(n int) Option    { return func(m *model) { m.height = n } }

// WithDetailFunc sets a callback to render the detail panel for the item at index i.
// When set and the terminal is wide enough, a detail pane appears to the right.
func WithDetailFunc(fn func(i int) string) Option {
	return func(m *model) { m.detailFunc = fn }
}

type item struct {
	text     string
	selected bool
	order    int
}

type keymap struct {
	Down, Up, Right, Left, Home, End key.Binding
	ToggleAll, Toggle                key.Binding
	Abort, Quit, Submit              key.Binding
}

func (k keymap) FullHelp() [][]key.Binding { return nil }
func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Toggle,
		key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑↓", "navigate")),
		k.Submit,
		k.ToggleAll,
	}
}

type model struct {
	header       string
	items        []item
	quitting     bool
	submitted    bool
	index        int
	limit        int // 0 = unlimited
	numSelected  int
	currentOrder int
	height       int // items per page
	paginator    paginator.Model
	showHelp     bool
	help         help.Model
	keymap       keymap
	detailFunc   func(int) string
	termWidth    int
	termHeight   int

	cursorStyle       lipgloss.Style
	headerStyle       lipgloss.Style
	itemStyle         lipgloss.Style
	selectedItemStyle lipgloss.Style
}

const minDetailWidth = 100

func defaultKeymap(multiSelect bool) keymap {
	km := keymap{
		Down:  key.NewBinding(key.WithKeys("down", "j", "ctrl+n")),
		Up:    key.NewBinding(key.WithKeys("up", "k", "ctrl+p")),
		Right: key.NewBinding(key.WithKeys("right", "l")),
		Left:  key.NewBinding(key.WithKeys("left", "h")),
		Home:  key.NewBinding(key.WithKeys("g", "home")),
		End:   key.NewBinding(key.WithKeys("G", "end")),
		ToggleAll: key.NewBinding(
			key.WithKeys("a", "ctrl+a"),
			key.WithHelp("ctrl+a", "select all"),
			key.WithDisabled(),
		),
		Toggle: key.NewBinding(
			key.WithKeys("space", "tab", "x"),
			key.WithHelp("x", "toggle"),
			key.WithDisabled(),
		),
		Abort:  key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "abort")),
		Quit:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "quit")),
		Submit: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit")),
	}
	if multiSelect {
		km.Toggle.SetEnabled(true)
		km.ToggleAll.SetEnabled(true)
	}
	return km
}

func newModel(items []string, opts ...Option) model {
	its := make([]item, len(items))
	for i, text := range items {
		its[i] = item{text: text}
	}

	m := model{
		items:             its,
		height:            10,
		showHelp:          true,
		help:              help.New(),
		cursorStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("212")),
		headerStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("99")).Bold(true),
		itemStyle:         lipgloss.NewStyle(),
		selectedItemStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("212")),
	}

	for _, opt := range opts {
		opt(&m)
	}

	multiSelect := m.limit != 1
	m.keymap = defaultKeymap(multiSelect)

	if m.limit == 0 {
		m.limit = len(items)
	}

	p := paginator.New(
		paginator.WithPerPage(m.height),
	)
	p.Type = paginator.Dots
	p.SetTotalPages(len(items))
	m.paginator = p

	return m
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		return m, nil
	case tea.KeyPressMsg:
		start, end := m.paginator.GetSliceBounds(len(m.items))
		km := m.keymap
		switch {
		case key.Matches(msg, km.Down):
			m.index++
			if m.index >= len(m.items) {
				m.index = 0
				m.paginator.Page = 0
			}
			if m.index >= end {
				m.paginator.NextPage()
			}
		case key.Matches(msg, km.Up):
			m.index--
			if m.index < 0 {
				m.index = len(m.items) - 1
				m.paginator.Page = m.paginator.TotalPages - 1
			}
			if m.index < start {
				m.paginator.PrevPage()
			}
		case key.Matches(msg, km.Right):
			m.index = clamp(m.index+m.height, 0, len(m.items)-1)
			m.paginator.NextPage()
		case key.Matches(msg, km.Left):
			m.index = clamp(m.index-m.height, 0, len(m.items)-1)
			m.paginator.PrevPage()
		case key.Matches(msg, km.End):
			m.index = len(m.items) - 1
			m.paginator.Page = m.paginator.TotalPages - 1
		case key.Matches(msg, km.Home):
			m.index = 0
			m.paginator.Page = 0
		case key.Matches(msg, km.ToggleAll):
			if m.limit <= 1 {
				break
			}
			if m.numSelected < len(m.items) && m.numSelected < m.limit {
				m = m.selectAll()
			} else {
				m = m.deselectAll()
			}
		case key.Matches(msg, km.Quit), key.Matches(msg, km.Abort):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, km.Toggle):
			if m.limit == 1 {
				break
			}
			if m.items[m.index].selected {
				m.items[m.index].selected = false
				m.numSelected--
			} else if m.numSelected < m.limit {
				m.items[m.index].selected = true
				m.items[m.index].order = m.currentOrder
				m.numSelected++
				m.currentOrder++
			}
		case key.Matches(msg, km.Submit):
			m.quitting = true
			if m.limit == 1 && m.numSelected < 1 {
				m.items[m.index].selected = true
			}
			m.submitted = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.paginator, cmd = m.paginator.Update(msg)
	return m, cmd
}

func (m model) selectAll() model {
	for i := range m.items {
		if m.numSelected >= m.limit {
			break
		}
		if m.items[i].selected {
			continue
		}
		m.items[i].selected = true
		m.items[i].order = m.currentOrder
		m.numSelected++
		m.currentOrder++
	}
	return m
}

func (m model) deselectAll() model {
	for i := range m.items {
		m.items[i].selected = false
		m.items[i].order = 0
	}
	m.numSelected = 0
	m.currentOrder = 0
	return m
}

func (m model) showDetail() bool {
	return m.detailFunc != nil && m.termWidth >= minDetailWidth
}

func (m model) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}

	listContent := m.renderList()

	if m.showDetail() {
		return tea.NewView(m.renderSplitView(listContent))
	}

	var parts []string
	if m.header != "" {
		parts = append(parts, m.headerStyle.Render(m.header))
	}
	parts = append(parts, listContent)
	if m.showHelp {
		parts = append(parts, "", m.help.View(m.keymap))
	}

	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, parts...))
}

func (m model) renderList() string {
	var s strings.Builder
	start, end := m.paginator.GetSliceBounds(len(m.items))

	cursor := "> "
	selectedPrefix := "[x] "
	unselectedPrefix := "[ ] "
	cursorPrefix := "[ ] "

	for i, it := range m.items[start:end] {
		if i == m.index%m.height {
			s.WriteString(m.cursorStyle.Render(cursor))
		} else {
			s.WriteString(strings.Repeat(" ", lipgloss.Width(cursor)))
		}

		if it.selected {
			s.WriteString(m.selectedItemStyle.Render(selectedPrefix + it.text))
		} else if i == m.index%m.height {
			s.WriteString(m.cursorStyle.Render(cursorPrefix + it.text))
		} else {
			s.WriteString(m.itemStyle.Render(unselectedPrefix + it.text))
		}
		if i != m.height-1 {
			s.WriteRune('\n')
		}
	}

	if m.paginator.TotalPages > 1 {
		s.WriteString(strings.Repeat("\n", m.height-m.paginator.ItemsOnPage(len(m.items))+1))
		s.WriteString("  " + m.paginator.View())
	}

	return s.String()
}

func (m model) renderSplitView(listContent string) string {
	listWidth := m.termWidth * 2 / 5
	detailWidth := m.termWidth - listWidth - 3 // 3 for border

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	// Available height for content (reserve for header + help + border)
	contentHeight := m.termHeight - 4
	if m.header != "" {
		contentHeight -= 1
	}
	if m.showHelp {
		contentHeight -= 2
	}
	if contentHeight < 5 {
		contentHeight = 5
	}

	listPane := lipgloss.NewStyle().
		Width(listWidth).
		Height(contentHeight).
		Render(listContent)

	detailText := ""
	if m.detailFunc != nil {
		detailText = m.detailFunc(m.index)
	}

	detailPane := borderStyle.
		Width(detailWidth).
		Height(contentHeight).
		PaddingLeft(1).
		PaddingRight(1).
		Render(truncateHeight(detailText, contentHeight-2))

	body := lipgloss.JoinHorizontal(lipgloss.Top, listPane, detailPane)

	var parts []string
	if m.header != "" {
		parts = append(parts, m.headerStyle.Render(m.header))
	}
	parts = append(parts, body)
	if m.showHelp {
		parts = append(parts, m.help.View(m.keymap))
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func truncateHeight(s string, maxLines int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= maxLines {
		return s
	}
	return strings.Join(lines[:maxLines], "\n")
}

// Run presents an interactive multi-select list and returns indices of selected items.
// Returns nil slice if user cancels (Esc/Ctrl+C).
func Run(items []string, opts ...Option) ([]int, error) {
	m := newModel(items, opts...)
	tm, err := tea.NewProgram(m, tea.WithOutput(os.Stderr)).Run()
	if err != nil {
		return nil, err
	}
	result := tm.(model)
	if !result.submitted {
		return nil, nil
	}
	var indices []int
	for i, it := range result.items {
		if it.selected {
			indices = append(indices, i)
		}
	}
	return indices, nil
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// FormatTODOListItem formats a TODO for the two-line list display.
func FormatTODOListItem(title, priority, status, path string) string {
	line1 := title
	if priority != "" || status != "" {
		line1 += "  "
		if priority != "" {
			line1 += priority
		}
		if status != "" {
			if priority != "" {
				line1 += "  "
			}
			line1 += status
		}
	}
	if path != "" {
		line1 += "\n       " + path
	}
	return line1
}

// FormatTODODetail renders the full detail panel content for a TODO.
func FormatTODODetail(opts DetailOptions) string {
	var s strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33"))

	s.WriteString(titleStyle.Render(opts.Title))
	s.WriteRune('\n')
	s.WriteRune('\n')

	if opts.Priority != "" {
		s.WriteString(labelStyle.Render("Priority: ") + opts.Priority + "\n")
	}
	if opts.Status != "" {
		s.WriteString(labelStyle.Render("Status:   ") + opts.Status + "\n")
	}
	if opts.Attempts > 0 {
		attemptStr := fmt.Sprintf("%d", opts.Attempts)
		if opts.LastRun != "" {
			attemptStr += " (last: " + opts.LastRun + ")"
		}
		s.WriteString(labelStyle.Render("Attempts: ") + attemptStr + "\n")
	}
	if opts.Language != "" {
		s.WriteString(labelStyle.Render("Language: ") + opts.Language + "\n")
	}
	if opts.Branch != "" {
		s.WriteString(labelStyle.Render("Branch:   ") + opts.Branch + "\n")
	}

	if opts.PRNumber > 0 {
		s.WriteRune('\n')
		s.WriteString(sectionStyle.Render("PR Info") + "\n")
		s.WriteString(labelStyle.Render("  PR: ") + fmt.Sprintf("#%d", opts.PRNumber) + "\n")
		if opts.PRAuthor != "" {
			s.WriteString(labelStyle.Render("  Author: ") + opts.PRAuthor + "\n")
		}
	}

	if len(opts.Paths) > 0 {
		s.WriteRune('\n')
		s.WriteString(sectionStyle.Render("Paths") + "\n")
		for _, p := range opts.Paths {
			s.WriteString("  " + p + "\n")
		}
	}

	if len(opts.Tests) > 0 {
		s.WriteRune('\n')
		s.WriteString(sectionStyle.Render("Verification Tests") + "\n")
		for _, t := range opts.Tests {
			s.WriteString("  • " + t + "\n")
		}
	}

	if opts.Implementation != "" {
		s.WriteRune('\n')
		s.WriteString(sectionStyle.Render("Implementation") + "\n")
		lines := strings.Split(opts.Implementation, "\n")
		if len(lines) > 10 {
			lines = lines[:10]
			lines = append(lines, "...")
		}
		for _, line := range lines {
			s.WriteString("  " + line + "\n")
		}
	}

	if opts.PRComment != "" {
		s.WriteRune('\n')
		s.WriteString(sectionStyle.Render("Review Comment") + "\n")
		lines := strings.Split(opts.PRComment, "\n")
		if len(lines) > 8 {
			lines = lines[:8]
			lines = append(lines, "...")
		}
		for _, line := range lines {
			s.WriteString("  " + line + "\n")
		}
	}

	return s.String()
}

// DetailOptions holds the data needed to render a TODO detail panel.
type DetailOptions struct {
	Title          string
	Priority       string
	Status         string
	Attempts       int
	LastRun        string
	Language       string
	Branch         string
	PRNumber       int
	PRAuthor       string
	PRComment      string
	Paths          []string
	Tests          []string
	Implementation string
}
