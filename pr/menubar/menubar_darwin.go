//go:build darwin

package menubar

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sync"

	"github.com/caseymrm/menuet"
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

func (m *MenuBar) Run() {
	runtime.LockOSThread()

	app := menuet.App()
	app.Name = "Gavel PRs"
	app.Label = "com.flanksource.gavel.prs"
	app.SetMenuState(&menuet.MenuState{
		Title: "PR: ...",
	})
	app.Children = m.menuItems

	go m.pollUpdates()

	app.RunApplication()
	close(m.done)
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

	var title string
	if failed > 0 {
		title = fmt.Sprintf("PR: %d/%d", failed, len(prs))
	} else if len(prs) > 0 {
		title = "PR: ✓"
	} else {
		title = "PR: 0"
	}

	menuet.App().SetMenuState(&menuet.MenuState{
		Title: title,
	})
}

func (m *MenuBar) menuItems() []menuet.MenuItem {
	m.mu.RLock()
	prs := m.prs
	m.mu.RUnlock()

	var items []menuet.MenuItem

	items = append(items, menuet.MenuItem{
		Text:    fmt.Sprintf("%d Pull Requests", len(prs)),
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
			title := fmt.Sprintf("%s #%d %s", icon, pr.Number, pr.Title)
			if len(title) > 60 {
				title = title[:57] + "..."
			}
			url := pr.URL
			items = append(items, menuet.MenuItem{
				Text: title,
				Clicked: func() {
					openURL(url)
				},
			})
		}
	}

	items = append(items, menuet.MenuItem{Type: menuet.Separator})
	if m.DashboardURL != "" {
		dashURL := m.DashboardURL
		items = append(items, menuet.MenuItem{
			Text: "Open Dashboard",
			Clicked: func() {
				openURL(dashURL)
			},
		})
	}
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
