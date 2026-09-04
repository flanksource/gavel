//go:build !darwin || !cgo || !gavel_menubar

package main

import (
	"fmt"

	"github.com/flanksource/gavel/pr/ui"
)

func runMenuBar(_ *ui.Server, _ string) error {
	return fmt.Errorf("menu bar requires a macOS build with CGO_ENABLED=1 and -tags gavel_menubar (or task build:prod)")
}
