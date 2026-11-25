package commit

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

const (
	commitPreviewSplitWidth = 100
	patchZoneID             = "patch"
)

type treeLayout struct {
	header                  []string
	treeWidth, treeHeight   int
	patchWidth, patchHeight int
	sideBySide              bool
}

func (m treeModel) layout() treeLayout {
	width := max(1, m.width)
	header := strings.Split(ansi.Wrap(m.renderHeader(), width, ""), "\n")
	header = header[:min(len(header), max(0, m.height-1))]
	body := max(0, m.height-len(header))
	layout := treeLayout{header: header, treeWidth: width, treeHeight: body}
	if m.preview == nil || body < 3 {
		return layout
	}
	if width >= commitPreviewSplitWidth {
		layout.sideBySide = true
		layout.treeWidth = min(64, width*2/5)
		layout.patchWidth, layout.patchHeight = width-layout.treeWidth-3, body
	} else {
		layout.treeHeight = body / 2
		layout.patchWidth, layout.patchHeight = width, body-layout.treeHeight
	}
	return layout
}

func (m treeModel) View() tea.View {
	layout := m.layout()
	tree := m.renderTree(layout)
	var body string
	if layout.patchHeight == 0 {
		body = tree
	} else if layout.sideBySide {
		divider := strings.TrimSuffix(strings.Repeat(" │ \n", layout.treeHeight), "\n")
		body = lipgloss.JoinHorizontal(lipgloss.Top, tree, divider, m.renderPatch(layout))
	} else {
		body = tree + "\n" + m.renderPatch(layout)
	}
	content := strings.Join(append(layout.header, body), "\n")
	view := tea.NewView(content)
	view.AltScreen = true
	if m.zones != nil {
		view.MouseMode = tea.MouseModeAllMotion
		view.SetContent(m.zones.Scan(content))
	}
	return view
}

func (m treeModel) rowRange(height int) (start, end int) {
	if height <= 0 || len(m.visible) == 0 {
		return 0, 0
	}
	start = max(m.cursor-height+1, min(m.cursor, m.treeOffset))
	start = max(0, min(start, len(m.visible)-height))
	return start, min(start+height, len(m.visible))
}

func (m treeModel) renderTree(layout treeLayout) string {
	rows := make([]string, layout.treeHeight)
	start, end := m.rowRange(layout.treeHeight)
	filterActive := normalizedFilter(m.filterQuery) != ""
	for i := start; i < end; i++ {
		node := m.visible[i]
		row := renderRow(node, i == m.cursor, filterActive)
		row = fitTreeLine(row, layout.treeWidth)
		if m.zones != nil {
			row = m.zones.Mark(rowZoneID(node), row)
		}
		rows[i-start] = row
	}
	if len(rows) > 0 && len(m.visible) == 0 {
		rows[0] = fitTreeLine(styled("  No files match the current filter", styleHelp), layout.treeWidth)
	}
	if layout.sideBySide {
		for i, row := range rows {
			if row == "" {
				rows[i] = strings.Repeat(" ", layout.treeWidth)
			}
		}
	}
	return strings.Join(rows, "\n")
}

func (m treeModel) renderPatch(layout treeLayout) string {
	title := fitTreeLine(styled(m.preview.title, styleHeader), layout.patchWidth)
	content := m.preview.sizedPane(layout).View()
	if content != "" {
		title += "\n" + content
	}
	if m.zones == nil {
		return title
	}
	return m.zones.Mark(patchZoneID, title)
}

func fitTreeLine(line string, width int) string {
	line = ansi.Truncate(line, width, "…")
	return line + strings.Repeat(" ", max(0, width-ansi.StringWidth(line)))
}

func rowZoneID(node *treeNode) string {
	return fmt.Sprintf("row:%p", node)
}

func (m *treeModel) handleMouse(msg tea.MouseMsg) {
	if m.zones == nil || m.ignorePrompt != nil || m.filtering {
		return
	}
	switch event := msg.(type) {
	case tea.MouseWheelMsg:
		if m.zones.Get(patchZoneID).InBounds(msg) {
			m.scrollPreview(event)
		} else if m.mouseRow(msg) >= 0 {
			switch event.Button {
			case tea.MouseWheelUp:
				m.cursor = max(0, m.cursor-3)
			case tea.MouseWheelDown:
				m.cursor = min(len(m.visible)-1, m.cursor+3)
			}
		}
	case tea.MouseMotionMsg:
		if i := m.mouseRow(msg); i >= 0 {
			m.cursor = i
		}
	case tea.MouseClickMsg:
		if i := m.mouseRow(msg); i >= 0 && event.Button == tea.MouseLeft {
			m.cursor = i
			node := m.visible[i]
			x, _ := m.zones.Get(rowZoneID(node)).Pos(msg)
			checkboxStart := 2 * node.Depth
			if x >= checkboxStart && x < checkboxStart+3 {
				toggleNode(node)
			}
		}
	}
}

func (m treeModel) mouseRow(msg tea.MouseMsg) int {
	layout := m.layout()
	start, end := m.rowRange(layout.treeHeight)
	i := start + msg.Mouse().Y - len(layout.header)
	if i >= start && i < end && msg.Mouse().X < layout.treeWidth &&
		m.zones.Get(rowZoneID(m.visible[i])).InBounds(msg) {
		return i
	}
	return -1
}
