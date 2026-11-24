package commit

import (
	"context"
	"strings"
	"unicode"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/status"
)

type treePreview struct {
	ctx     context.Context
	cancel  context.CancelFunc
	node    *treeNode
	request uint64
	pane    viewport.Model
	title   string
}

type treePatchMsg struct {
	request uint64
	content string
	label   string
}

type treePatchRequest struct {
	id   uint64
	file status.FileStatus
}

func newTreePreview(ctx context.Context) *treePreview {
	pane := viewport.New()
	pane.KeyMap.Left.SetKeys("shift+left")
	pane.KeyMap.Right.SetKeys("shift+right")
	return &treePreview{ctx: ctx, pane: pane, title: "Patch"}
}

func (m treeModel) refreshPreview() tea.Cmd {
	p := m.preview
	if p == nil {
		return nil
	}
	node := m.currentNode()
	if node != nil && node.File == nil {
		node = nil
	}
	if p.node == node {
		return nil
	}
	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}
	p.node, p.request = node, p.request+1
	p.pane.SetXOffset(0)
	p.pane.GotoTop()
	p.title = "Patch"
	if node == nil {
		p.pane.SetContent("")
		return nil
	}

	p.title = singleLine(node.Path)
	p.pane.SetContent(styled("Loading patch…", styleMuted))
	ctx, cancel := context.WithCancel(p.ctx)
	p.cancel = cancel
	request := treePatchRequest{id: p.request, file: *node.File}
	return func() tea.Msg {
		defer cancel()
		raw, err := readTreePatch(ctx, m.gitRoot, request.file)
		if ctx.Err() != nil {
			return nil
		}
		msg := request.format(ctx, raw, err)
		if ctx.Err() != nil {
			return nil
		}
		return msg
	}
}

func (p *treePreview) apply(msg treePatchMsg) {
	if p.node == nil || msg.request != p.request {
		return
	}
	p.pane.SetContent(msg.content)
	if msg.label != "" {
		p.title = singleLine(p.node.Path) + " · " + msg.label
	}
}

func (r treePatchRequest) format(ctx context.Context, raw string, err error) treePatchMsg {
	result := treePatchMsg{request: r.id}
	switch {
	case err != nil:
		result.content = styled("Patch unavailable. "+singleLine(err.Error()), styleDeleted)
	case raw == "":
		result.content = styled("No net changes to preview", styleMuted)
	default:
		patch, truncated := gavelgit.TruncateDiff(raw)
		if truncated {
			result.label = "Truncated"
		} else if gavelgit.IsBinaryDiff(patch) {
			result.label = "Binary"
		}
		patch = terminalText(strings.ReplaceAll(patch, "\r\n", "\n"))
		result.content, err = renderTreePatch(ctx, r.file, patch)
		if err != nil {
			result.content = styled("Patch unavailable. "+singleLine(err.Error()), styleDeleted)
		}
	}
	return result
}

func (m treeModel) scrollPreview(msg tea.Msg) {
	if p := m.preview; p != nil {
		pane := p.sizedPane(m.layout())
		if pane.Height() > 0 {
			p.pane, _ = pane.Update(msg)
		}
	}
}

func (p *treePreview) sizedPane(layout treeLayout) viewport.Model {
	pane := p.pane
	pane.SetWidth(layout.patchWidth)
	pane.SetHeight(max(0, layout.patchHeight-1))
	pane.SetXOffset(pane.XOffset())
	pane.SetYOffset(pane.YOffset())
	return pane
}

func terminalText(s string) string {
	return strings.Map(func(r rune) rune {
		if r != '\n' && r != '\t' && unicode.IsControl(r) {
			return '�'
		}
		return r
	}, s)
}

func singleLine(s string) string {
	return terminalText(strings.NewReplacer("\r", " ", "\n", " ", "\t", " ").Replace(s))
}
