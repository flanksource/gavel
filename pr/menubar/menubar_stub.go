//go:build !darwin

package menubar

import (
	"fmt"

	"github.com/flanksource/gavel/pr/ui"
)

type MenuBar struct {
	DashboardURL string
}

func New(_ *ui.Server) *MenuBar {
	return &MenuBar{}
}

func (m *MenuBar) Run() error {
	return fmt.Errorf("menu bar is only supported on macOS")
}

func (m *MenuBar) Done() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}
