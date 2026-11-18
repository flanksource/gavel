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
	return clicky.Text(`Capture a command's merged stdout and stderr in a fixed-size terminal.

The JSON includes asciinema-style timed events and settled screen snapshots
for debugging terminal redraws. Put the command and its arguments after --.
Gavel writes the JSON to stdout when you omit the output path.`)
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
	testANSICmd.Use = "ansi [flags] [out.json] -- <command> [args...]"
	testANSICmd.Short = "Capture timed PTY/ANSI output of a command for redraw debugging"
	testANSICmd.Example = "  gavel test ansi --width 120 out.json -- gavel test ./testrunner\n  gavel test ansi -- gavel version"
	testANSICmd.Flags().SetInterspersed(true)
}
