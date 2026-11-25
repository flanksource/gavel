//go:build !darwin || !cgo || !gavel_menubar

package main

import (
	"github.com/flanksource/gavel/pr/ui"
)

const menuBarSupported = false

func runMenuBar(_ *ui.Server, _ string) error {
	panic("runMenuBar called in a build without menu bar support")
}
