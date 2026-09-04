package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/fixtures"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type testANSIOptions struct {
	Width    int      `json:"width,omitempty" flag:"width" help:"PTY width in columns (0 = detect from terminal, fallback 120)"`
	Height   int      `json:"height,omitempty" flag:"height" help:"PTY height in rows (0 = detect from terminal, fallback 40)"`
	Interval string   `json:"interval,omitempty" flag:"snapshot-interval" default:"100ms" help:"How often to settle a screen snapshot (e.g. 50ms, 250ms)"`
	Args     []string `json:"-" args:"true"`
}

func (o testANSIOptions) Help() api.Textable {
	const (
		heading   = "font-bold text-purple-600"
		flagStyle = "text-cyan-600 font-bold"
		muted     = "text-muted"
	)
	return clicky.Text("gavel test ansi", heading).Space().
		Append("— run a command under a fixed-size PTY and capture its timed output", muted).NewLine().NewLine().
		Append("Records the merged stdout/stderr stream as asciinema-style timed events plus a", "").NewLine().
		Append("timeline of width-aware settled-screen snapshots. Useful for debugging TUI redraw", "").NewLine().
		Append("bugs (cursor-up under-counts, wrap-induced smears, duplicated frames).", "").NewLine().NewLine().
		Append("USAGE", heading).NewLine().
		Append("  gavel test ansi [out.json] -- <command> [args...]", flagStyle).NewLine().
		Append("  A command after ", muted).Append("--", flagStyle).Append(" is required. Without out.json the JSON is written to stdout.", muted).NewLine().NewLine().
		Append("FLAGS", heading).NewLine().
		Append("  --width", flagStyle).Append("             PTY columns (0 = detect, fallback 120)", muted).NewLine().
		Append("  --height", flagStyle).Append("            PTY rows (0 = detect, fallback 40)", muted).NewLine().
		Append("  --snapshot-interval", flagStyle).Append(" snapshot cadence (default 100ms)", muted).NewLine().NewLine().
		Append("EXAMPLES", heading).NewLine().
		Append("  gavel test ansi --width 120 out.json -- gavel status --ai .", flagStyle).NewLine().
		Append("  gavel test ansi -- gavel test ./pkg", flagStyle).Append("   (JSON to stdout)", muted).NewLine()
}

var testANSICmd *cobra.Command

func runTestANSI(opts testANSIOptions) (any, error) {
	dash := testANSICmd.Flags().ArgsLenAtDash()
	if dash < 0 {
		return nil, fmt.Errorf("a command is required after `--` (e.g. gavel test ansi out.json -- gavel status)")
	}
	outPaths := opts.Args[:dash]
	command := opts.Args[dash:]
	if len(command) == 0 {
		return nil, fmt.Errorf("no command given after `--`")
	}
	if len(outPaths) > 1 {
		return nil, fmt.Errorf("at most one output path may precede `--`, got %d: %v", len(outPaths), outPaths)
	}
	outPath := ""
	if len(outPaths) == 1 {
		outPath = outPaths[0]
	}

	interval, err := time.ParseDuration(opts.Interval)
	if err != nil {
		return nil, fmt.Errorf("invalid --snapshot-interval %q: %w", opts.Interval, err)
	}

	width, height := resolveANSISize(opts.Width, opts.Height)

	capture, err := fixtures.CaptureANSI(fixtures.CaptureOptions{
		Width:            width,
		Height:           height,
		SnapshotInterval: interval,
		Command:          command,
		// The settled screen and its duplicate lines are the whole point of this
		// command, as opposed to a fixture that only wants a PTY.
		Snapshots: true,
	})
	if err != nil {
		return nil, err
	}

	if outPath == "" {
		if err := encodeANSICapture(os.Stdout, capture); err != nil {
			return nil, fmt.Errorf("write capture to stdout: %w", err)
		}
		return nil, nil
	}
	if err := writeANSICapture(outPath, capture); err != nil {
		return nil, fmt.Errorf("write %s: %w", outPath, err)
	}
	return ansiCaptureSummary(capture, outPath), nil
}

func encodeANSICapture(w io.Writer, capture *fixtures.Capture) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(capture)
}

func writeANSICapture(path string, capture *fixtures.Capture) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	cleanup := func(cause error) error {
		closeErr := tmp.Close()
		if errors.Is(closeErr, os.ErrClosed) {
			closeErr = nil
		}
		removeErr := os.Remove(tmpPath)
		if os.IsNotExist(removeErr) {
			removeErr = nil
		}
		return errors.Join(cause, closeErr, removeErr)
	}
	if err = tmp.Chmod(0o644); err != nil {
		return cleanup(err)
	}
	if err = encodeANSICapture(tmp, capture); err != nil {
		return cleanup(err)
	}
	if err = tmp.Close(); err != nil {
		return cleanup(err)
	}
	if err = os.Rename(tmpPath, path); err != nil {
		return cleanup(err)
	}
	return nil
}

// resolveANSISize fills in 0-valued width/height from the controlling terminal,
// falling back to a sensible fixed size when stdout is not a TTY.
func resolveANSISize(width, height int) (int, int) {
	if width > 0 && height > 0 {
		return width, height
	}
	tw, th, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || tw <= 0 || th <= 0 {
		tw, th = 120, 40
	}
	if width <= 0 {
		width = tw
	}
	if height <= 0 {
		height = th
	}
	return width, height
}

type ansiSummary struct {
	Out           string   `json:"out"`
	Command       []string `json:"command"`
	Width         int      `json:"width"`
	Height        int      `json:"height"`
	ExitCode      int      `json:"exit_code"`
	DurationMs    int64    `json:"duration_ms"`
	Events        int      `json:"events"`
	Snapshots     int      `json:"snapshots"`
	Duplicates    int      `json:"duplicate_lines"`
	MaxLineWidth  int      `json:"max_line_width"`
	WidthOverflow bool     `json:"width_overflow"`
}

// ansiCaptureSummary builds a small digest for pretty output. MaxLineWidth is
// measured from the UNWRAPPED settled screen (width 0) so a value greater than
// the PTY width flags content that soft-wraps — the precondition for the
// cursor-up-under-count smear.
func ansiCaptureSummary(c *fixtures.Capture, outPath string) ansiSummary {
	maxW := c.MaxLineWidth()
	return ansiSummary{
		Out:           outPath,
		Command:       c.Command,
		Width:         c.Width,
		Height:        c.Height,
		ExitCode:      c.ExitCode,
		DurationMs:    c.DurationMs,
		Events:        len(c.Events),
		Snapshots:     len(c.Snapshots),
		Duplicates:    len(c.Final.Duplicates),
		MaxLineWidth:  maxW,
		WidthOverflow: maxW > c.Width,
	}
}

func init() {
	testANSICmd = clicky.AddNamedCommand("ansi", testCmd, testANSIOptions{}, runTestANSI)
	testANSICmd.Use = "ansi [out.json] -- <command>"
	testANSICmd.Short = "Capture timed PTY/ANSI output of a command for redraw debugging"
	testANSICmd.Flags().SetInterspersed(true)
}
