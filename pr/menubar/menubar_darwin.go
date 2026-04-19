//go:build darwin && cgo

package menubar

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sync"

	"github.com/caseymrm/menuet"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/ui"
)

type MenuBar struct {
	srv          *ui.Server
	mu           sync.RWMutex
	prs          github.PRSearchResults
	done         chan struct{}
	DashboardURL string
}

func New(srv *ui.Server) *MenuBar {
	return &MenuBar{
		srv:  srv,
		done: make(chan struct{}),
	}
}

func (m *MenuBar) SetPRs(prs github.PRSearchResults) {
	m.mu.Lock()
	m.prs = prs
	m.mu.Unlock()
}

func (m *MenuBar) Run() error {
	runtime.LockOSThread()

	app := menuet.App()
	app.Name = "Gavel PRs"
	app.Label = "com.flanksource.gavel.prs"
	app.SetMenuState(&menuet.MenuState{
		Title: "...",
		// Start without the unread dot; updateTitle() will flip it on the
		// first poll subscription tick.
		Image: m.iconURL(false),
	})
	app.Children = m.menuItems

	go m.pollUpdates()

	app.RunApplication()
	close(m.done)
	return nil
}

// iconURL returns the menubar icon URL served by the running UI server.
// menuet's NSImage loader only recognises "http"-prefixed names as remote
// resources; bare file paths are treated as bundled resource names and
// fail for un-bundled CLIs, so we serve the PNG over the local HTTP server.
// When hasUnread is true, returns a variant with a red dot overlay in the
// corner. Returns an empty string when no dashboard URL is available, which
// makes menuet render title-only (no image).
func (m *MenuBar) iconURL(hasUnread bool) string {
	if m.DashboardURL == "" {
		return ""
	}
	if hasUnread {
		return m.DashboardURL + "/brand/menubar-unread.png"
	}
	return m.DashboardURL + "/brand/menubar.png"
}

func (m *MenuBar) Done() <-chan struct{} {
	return m.done
}

func (m *MenuBar) pollUpdates() {
	ch := m.srv.Subscribe()
	for prs := range ch {
		m.SetPRs(prs)
		m.updateTitle()
	}
}

func (m *MenuBar) updateTitle() {
	m.mu.RLock()
	prs := m.prs
	m.mu.RUnlock()

	var failed int
	for _, pr := range prs {
		if pr.CheckStatus != nil && pr.CheckStatus.Failed > 0 {
			failed++
		}
	}
	unreadCount := len(m.srv.UnreadMap(prs))

	var title string
	switch {
	case m.srv.IsPaused():
		title = "⏸"
	case failed > 0:
		title = fmt.Sprintf("✗ %d/%d", failed, len(prs))
	case unreadCount > 0:
		title = fmt.Sprintf("• %d", unreadCount)
	default:
		// Everything read and nothing failing — icon alone.
		title = ""
	}

	menuet.App().SetMenuState(&menuet.MenuState{
		Title: title,
		Image: m.iconURL(unreadCount > 0),
	})
}

func (m *MenuBar) menuItems() []menuet.MenuItem {
	m.mu.RLock()
	prs := m.prs
	m.mu.RUnlock()
	unread := m.srv.UnreadMap(prs)

	var items []menuet.MenuItem

	header := fmt.Sprintf("%d Pull Requests", len(prs))
	if n := len(unread); n > 0 {
		header = fmt.Sprintf("%d Pull Requests · %d unread", len(prs), n)
	}
	items = append(items, menuet.MenuItem{
		Text:     header,
		FontSize: 12,
	})
	items = append(items, menuet.MenuItem{Type: menuet.Separator})

	groups := groupByRepo(prs)
	for _, g := range groups {
		if len(groups) > 1 {
			items = append(items, menuet.MenuItem{
				Text:     g.repo,
				FontSize: 11,
			})
		}
		for _, pr := range g.items {
			icon := statePrefix(pr)
			key := fmt.Sprintf("%s#%d", pr.Repo, pr.Number)
			marker := "  "
			if unread[key] {
				marker = "• "
			}
			title := fmt.Sprintf("%s%s #%d %s", marker, icon, pr.Number, pr.Title)
			if len(title) > 60 {
				title = title[:57] + "..."
			}
			url := pr.URL
			repo, number := pr.Repo, pr.Number
			items = append(items, menuet.MenuItem{
				Text: title,
				Clicked: func() {
					if err := m.srv.MarkSeen(repo, number); err != nil {
						logger.Warnf("mark seen %s#%d: %v", repo, number, err)
					}
					openURL(url)
				},
			})
		}
	}

	items = append(items, menuet.MenuItem{Type: menuet.Separator})
	if len(unread) > 0 {
		items = append(items, menuet.MenuItem{
			Text: "Mark all as read",
			Clicked: func() {
				if err := m.srv.MarkAllSeen(); err != nil {
					logger.Warnf("mark all seen: %v", err)
				}
				m.updateTitle()
			},
		})
	}
	if m.DashboardURL != "" {
		dashURL := m.DashboardURL
		items = append(items, menuet.MenuItem{
			Text: "Open Dashboard",
			Clicked: func() {
				openURL(dashURL)
			},
		})
	}
	paused := m.srv.IsPaused()
	pauseText := "Pause Polling"
	if paused {
		pauseText = "Resume Polling"
	}
	items = append(items, menuet.MenuItem{
		Text:  pauseText,
		State: paused,
		Clicked: func() {
			m.srv.TogglePause()
			m.updateTitle()
		},
	})
	items = append(items, menuet.MenuItem{
		Text: "Refresh",
		Clicked: func() {
			select {
			case m.srv.RefreshCh() <- struct{}{}:
			default:
			}
		},
	})
	items = append(items, menuet.MenuItem{
		Text: "Quit",
		Clicked: func() {
			os.Exit(0)
		},
	})

	return items
}

func statePrefix(pr github.PRListItem) string {
	if pr.CheckStatus != nil && pr.CheckStatus.Failed > 0 {
		return "✗"
	}
	if pr.CheckStatus != nil && pr.CheckStatus.Running > 0 {
		return "●"
	}
	if pr.IsDraft {
		return "○"
	}
	return "✓"
}

type repoGroup struct {
	repo  string
	items []github.PRListItem
}

func groupByRepo(prs github.PRSearchResults) []repoGroup {
	order := make([]string, 0)
	groups := make(map[string]*repoGroup)
	for _, item := range prs {
		if _, ok := groups[item.Repo]; !ok {
			order = append(order, item.Repo)
			groups[item.Repo] = &repoGroup{repo: item.Repo}
		}
		groups[item.Repo].items = append(groups[item.Repo].items, item)
	}
	result := make([]repoGroup, len(order))
	for i, repo := range order {
		result[i] = *groups[repo]
	}
	return result
}

func openURL(url string) {
	_ = exec.Command("open", url).Start()
}
