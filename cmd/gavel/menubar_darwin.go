//go:build darwin && cgo

package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/ui"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

func runMenuBar(srv *ui.Server, dashboardURL string) error {
	if dashboardURL == "" {
		return fmt.Errorf("menu bar requires a dashboard URL")
	}

	var menuWindow application.Window
	hideController := newMenubarHideController(5*time.Second, func() {
		if menuWindow != nil {
			menuWindow.Hide()
		}
	})
	app := application.New(application.Options{
		Name: "Gavel",
		Mac: application.MacOptions{
			ActivationPolicy: application.ActivationPolicyAccessory,
		},
		RawMessageHandler: func(window application.Window, message string, originInfo *application.OriginInfo) {
			if menuWindow == nil || window.ID() != menuWindow.ID() {
				return
			}
			msg, ok := parseMenubarMessage(dashboardURL, message, originInfo)
			if !ok {
				return
			}
			switch msg.Type {
			case menubarOpenExternalMessage:
				if target, ok := parseMenubarExternalTarget(msg.URL); ok {
					openBrowser(target)
				}
			case menubarPointerEnterMessage:
				hideController.cancel()
			case menubarPointerLeaveMessage:
				hideController.schedule()
			}
		},
	})

	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:          "Gavel",
		Width:          760,
		Height:         620,
		URL:            dashboardURL + "/menubar",
		Frameless:      true,
		AlwaysOnTop:    true,
		Hidden:         true,
		DisableResize:  false,
		MinWidth:       520,
		MinHeight:      420,
		BackgroundType: application.BackgroundTypeTransparent,
		Windows: application.WindowsWindow{
			HiddenOnTaskbar: true,
		},
	})
	menuWindow = window

	var blanked atomic.Bool
	blankController := newMenubarBlankController(blankIdle,
		window.IsVisible,
		func() {
			window.SetURL("about:blank")
			blanked.Store(true)
		},
	)

	window.OnWindowEvent(events.Common.WindowShow, func(event *application.WindowEvent) {
		hideController.cancel()
		if blanked.CompareAndSwap(true, false) {
			window.SetURL(dashboardURL + "/menubar")
		}
		blankController.touch()
	})
	window.OnWindowEvent(events.Common.WindowLostFocus, func(event *application.WindowEvent) {
		hideController.cancel()
		window.Hide()
	})

	systray := app.SystemTray.New()
	if runtime.GOOS == "darwin" {
		systray.SetIcon(ui.MenubarIconPNG())
	}
	systray.SetTooltip("Gavel")
	systray.AttachWindow(window)
	systray.WindowOffset(8)
	systray.WindowDebounce(200 * time.Millisecond)

	ch := srv.Subscribe()
	go func() {
		for prs := range ch {
			systray.SetLabel(menuBarLabel(srv, prs))
			systray.SetTooltip(menuBarTooltip(srv, prs))
		}
	}()

	menu := app.NewMenu()
	menu.Add("Open Dashboard").OnClick(func(ctx *application.Context) {
		openBrowser(dashboardURL)
	})
	menu.Add("Refresh").OnClick(func(ctx *application.Context) {
		select {
		case srv.RefreshCh() <- struct{}{}:
		default:
		}
	})
	menu.Add("Quit").OnClick(func(ctx *application.Context) { app.Quit() })
	systray.SetMenu(menu)

	return app.Run()
}

// blankIdle is how long the menubar popover may sit hidden before its webview
// page is navigated to about:blank to reclaim its (otherwise unbounded) memory.
const blankIdle = 5 * time.Minute

const menubarOpenExternalMessage = "gavel:open-external"
const menubarPointerEnterMessage = "gavel:pointer-enter"
const menubarPointerLeaveMessage = "gavel:pointer-leave"

type menubarMessage struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

func parseMenubarMessage(dashboardURL string, message string, originInfo *application.OriginInfo) (menubarMessage, bool) {
	if !isMenubarMessageOrigin(dashboardURL, originInfo) {
		return menubarMessage{}, false
	}

	var msg menubarMessage
	if err := json.Unmarshal([]byte(message), &msg); err != nil {
		return menubarMessage{}, false
	}
	switch msg.Type {
	case menubarOpenExternalMessage, menubarPointerEnterMessage, menubarPointerLeaveMessage:
		return msg, true
	default:
		return menubarMessage{}, false
	}
}

func parseMenubarExternalURL(dashboardURL string, message string, originInfo *application.OriginInfo) (string, bool) {
	msg, ok := parseMenubarMessage(dashboardURL, message, originInfo)
	if !ok || msg.Type != menubarOpenExternalMessage {
		return "", false
	}
	return parseMenubarExternalTarget(msg.URL)
}

func parseMenubarExternalTarget(rawURL string) (string, bool) {
	target, err := url.Parse(rawURL)
	if err != nil || target.Scheme == "" || target.Host == "" {
		return "", false
	}
	switch strings.ToLower(target.Scheme) {
	case "http", "https":
		return target.String(), true
	default:
		return "", false
	}
}

func isMenubarMessageOrigin(dashboardURL string, originInfo *application.OriginInfo) bool {
	if originInfo == nil || !originInfo.IsMainFrame || originInfo.Origin == "" {
		return false
	}
	dashboard, err := url.Parse(dashboardURL)
	if err != nil {
		return false
	}
	origin, err := url.Parse(originInfo.Origin)
	if err != nil {
		return false
	}
	return strings.EqualFold(origin.Scheme, dashboard.Scheme) && strings.EqualFold(origin.Host, dashboard.Host)
}

type menubarHideController struct {
	delay time.Duration
	hide  func()

	mu         sync.Mutex
	generation int
	timer      *time.Timer
}

func newMenubarHideController(delay time.Duration, hide func()) *menubarHideController {
	return &menubarHideController{delay: delay, hide: hide}
}

func (c *menubarHideController) schedule() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.generation++
	generation := c.generation
	if c.timer != nil {
		c.timer.Stop()
	}
	c.timer = time.AfterFunc(c.delay, func() {
		c.mu.Lock()
		if generation != c.generation {
			c.mu.Unlock()
			return
		}
		c.timer = nil
		c.mu.Unlock()

		if c.hide != nil {
			c.hide()
		}
	})
}

func (c *menubarHideController) cancel() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.generation++
	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}
}

// menubarBlankController frees the resident webview page once the popover has
// been idle (hidden) for delay. The menubar window is only hidden, never
// destroyed, so its DOM/JS heap grows unbounded across the daemon's lifetime;
// navigating it to about:blank discards that heap, and the next open reloads
// /menubar. The timer resets on every open ("delay after the last open") and a
// visible popover is never blanked — if the user still has it open when the
// timer fires it reschedules for another idle window instead.
type menubarBlankController struct {
	delay   time.Duration
	visible func() bool
	blank   func()

	mu         sync.Mutex
	generation int
	timer      *time.Timer
}

func newMenubarBlankController(delay time.Duration, visible func() bool, blank func()) *menubarBlankController {
	return &menubarBlankController{delay: delay, visible: visible, blank: blank}
}

// touch (re)starts the idle timer; call it on every WindowShow so the page is
// blanked delay after the most recent open.
func (c *menubarBlankController) touch() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.arm()
}

// arm schedules fire after delay, superseding any pending timer. Caller holds mu.
func (c *menubarBlankController) arm() {
	c.generation++
	generation := c.generation
	if c.timer != nil {
		c.timer.Stop()
	}
	c.timer = time.AfterFunc(c.delay, func() { c.fire(generation) })
}

func (c *menubarBlankController) fire(generation int) {
	c.mu.Lock()
	if generation != c.generation {
		c.mu.Unlock()
		return
	}
	if c.visible != nil && c.visible() {
		c.arm()
		c.mu.Unlock()
		return
	}
	c.timer = nil
	c.mu.Unlock()

	if c.blank != nil {
		c.blank()
	}
}

func menuBarLabel(srv *ui.Server, prs github.PRSearchResults) string {
	var failed int
	for _, pr := range prs {
		if pr.CheckStatus != nil && pr.CheckStatus.Failed > 0 {
			failed++
		}
	}
	unread := len(srv.UnreadMap(prs))
	switch {
	case srv.IsPaused():
		return "paused"
	case failed > 0:
		return fmt.Sprintf("%d/%d", failed, len(prs))
	case unread > 0:
		return fmt.Sprintf("%d", unread)
	default:
		return ""
	}
}

func menuBarTooltip(srv *ui.Server, prs github.PRSearchResults) string {
	var failed int
	for _, pr := range prs {
		if pr.CheckStatus != nil && pr.CheckStatus.Failed > 0 {
			failed++
		}
	}
	unread := len(srv.UnreadMap(prs))
	switch {
	case srv.IsPaused():
		return "Gavel paused"
	case failed > 0:
		return fmt.Sprintf("Gavel: %d of %d pull requests failing", failed, len(prs))
	case unread > 0:
		return fmt.Sprintf("Gavel: %d unread pull requests", unread)
	default:
		return fmt.Sprintf("Gavel: %d pull requests", len(prs))
	}
}
