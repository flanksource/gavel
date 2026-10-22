package shims

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

const marker = "# gavel-shim"

// Environment variables the shim script passes to `gavel lint shim-exec`.
const (
	EnvTimeout = "GAVEL_SHIM_TIMEOUT"
	EnvCPU     = "GAVEL_SHIM_CPU"
	EnvMem     = "GAVEL_SHIM_MEM"
	EnvNice    = "GAVEL_SHIM_NICE"
	// EnvBypass=off runs the real tool with its arguments unchanged and no limits.
	EnvBypass = "GAVEL_SHIMS"
)

// Settings are the limits baked into a shim script.
type Settings struct {
	Timeout string
	CPU     float64
	Mem     string
	Nice    int
}

// Render returns the POSIX shell shim for tool, invoking gavel by absolute path
// so the shim never depends on gavel itself being on PATH.
func Render(tool, gavelPath string, settings Settings) string {
	return fmt.Sprintf("#!/bin/sh\n%s tool=%s\n%s=%s %s=%s %s=%s %s=%d exec %s lint shim-exec %s \"$@\"\n",
		marker, tool,
		EnvTimeout, shellQuote(settings.Timeout),
		EnvCPU, shellQuote(fmt.Sprint(settings.CPU)),
		EnvMem, shellQuote(settings.Mem),
		EnvNice, settings.Nice,
		shellQuote(gavelPath), tool)
}

// IsShim reports whether path is a script written by Render.
func IsShim(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close() //nolint:errcheck
	head := make([]byte, 256)
	n, _ := io.ReadFull(f, head)
	return bytes.Contains(head[:n], []byte("\n"+marker+" "))
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}
