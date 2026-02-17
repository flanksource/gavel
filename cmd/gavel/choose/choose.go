package choose

import (
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Option func(*model)

func WithHeader(h string) Option { return func(m *model) { m.header = h } }
func WithLimit(n int) Option     { return func(m *model) { m.limit = n } }
func WithShowHelp(b bool) Option { return func(m *model) { m.showHelp = b } }
func WithHeight(n int) Option    { return func(m *model) { m.height = n } }

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
	height       int
	paginator    paginator.Model
	showHelp     bool
	help         help.Model
	keymap       keymap

	cursorStyle       lipgloss.Style
	headerStyle       lipgloss.Style
	itemStyle         lipgloss.Style
	selectedItemStyle lipgloss.Style
}

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
			key.WithKeys(" ", "tab", "x"),
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

	p := paginator.New()
	p.Type = paginator.Dots
	p.PerPage = m.height
	p.SetTotalPages(len(items))
	m.paginator = p

	return m
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m, nil
	case tea.KeyMsg:
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

func (m model) View() string {
	if m.quitting {
		return ""
	}

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

	var parts []string
	if m.header != "" {
		parts = append(parts, m.headerStyle.Render(m.header))
	}
	parts = append(parts, s.String())
	if m.showHelp {
		parts = append(parts, "", m.help.View(m.keymap))
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
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
