//go:build !darwin || (darwin && !cgo)

package main

import (
	"fmt"

	"github.com/flanksource/gavel/pr/ui"
)

func runMenuBar(_ *ui.Server, _ string) error {
	return fmt.Errorf("menu bar requires a macOS build with CGO_ENABLED=1")
}
