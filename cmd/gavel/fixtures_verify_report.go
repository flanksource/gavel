package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky"
	"github.com/spf13/cobra"

	"github.com/flanksource/gavel/fixtures/verifier"
	"github.com/flanksource/gavel/verify"
)

// VerifyReportFormat is the only --format `gavel fixtures run` implements: the
// external fixture-runner contract captain dispatches `workflow.verify.fixture`
// through (~/.captain.yaml `verify.fixtureRunner`).
//
// The contract is one process, three streams: the fixture markdown arrives on
// stdin, `--cwd` and repeated `--changed` arrive on argv, and stdout carries
// NDJSON — zero or more {"progress": <VerifyReport>} lines while the checks run,
// then exactly one {"report": <VerifyReport>}.
//
// A failing definition of done is a verdict, not a runner failure: it exits 0
// with a report that says so. A non-zero exit means no verdict was produced.
const VerifyReportFormat = "captain-verify-report"

var fixturesRunChanged []string

var fixturesRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute a fixture document from stdin and answer with a captain verification report",
	Long: "Reads a fixture markdown document on stdin, executes it in --cwd, and writes\n" +
		"NDJSON progress snapshots followed by exactly one verification report on stdout.\n" +
		"This is captain's external fixture-runner contract; --format " + VerifyReportFormat + " is required.",
	Args:         cobra.NoArgs,
	RunE:         runFixturesVerifyReport,
	SilenceUsage: true,
}

func runFixturesVerifyReport(cmd *cobra.Command, _ []string) error {
	if format := clicky.Flags.ResolveFormat(); format != VerifyReportFormat {
		return fmt.Errorf("gavel fixtures run implements only --format %s (got %q)", VerifyReportFormat, format)
	}
	document, err := io.ReadAll(cmd.InOrStdin())
	if err != nil {
		return fmt.Errorf("read fixture document from stdin: %w", err)
	}
	if strings.TrimSpace(string(document)) == "" {
		return fmt.Errorf("no fixture document on stdin")
	}
	cwd, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	out := &ndjsonWriter{out: cmd.OutOrStdout()}
	report, err := runVerifyReport(cmd, out, string(document), cwd, fixturesRunChanged)
	if err != nil {
		return err
	}
	return out.write("report", report)
}

// runVerifyReport executes the document and streams its progress to out,
// returning the verdict's report. A verifier that could not produce a verdict at
// all returns an error, which the caller turns into a non-zero exit.
func runVerifyReport(cmd *cobra.Command, out *ndjsonWriter, document, cwd string, changed []string) (api.VerifyReport, error) {
	// An `ai` step in the document grades on the same chain a todo's definition
	// of done resolves: .gavel.yaml todos.verify over the ai: base. The
	// document's own `ai:` front matter still overrides it.
	cfg, err := verify.LoadGavelConfig(cwd)
	if err != nil {
		return api.VerifyReport{}, fmt.Errorf("load .gavel.yaml: %w", err)
	}
	graderSpec := cfg.AI.Merge(cfg.Todos.Verify)

	fixtureVerifier, err := verifier.NewVerifier(document)
	if err != nil {
		return api.VerifyReport{}, err
	}
	fixtureVerifier.Spec = &graderSpec

	fixtureVerifier.SetProgress(func(progress api.VerifyReport) {
		// A progress line that cannot be written is not worth failing a run over:
		// the verdict line is the contract, and it is written by the caller on the
		// same stream, which will report the same failure there.
		_ = out.write("progress", progress)
	})

	verdict, err := fixtureVerifier.Verify(cmd.Context(), cwd, changed)
	if err != nil {
		return api.VerifyReport{}, err
	}
	if verdict.Report == nil {
		return api.VerifyReport{}, fmt.Errorf("fixture verifier produced a verdict with no report")
	}
	return *verdict.Report, nil
}

// ndjsonWriter is the one writer of the protocol stream. Progress snapshots
// arrive from the engine's goroutine while the verdict is written by the
// command, so every line goes through the same lock, or the two could
// interleave inside a line.
type ndjsonWriter struct {
	mu  sync.Mutex
	out io.Writer
}

// write emits one protocol line: {"<key>": <report>}.
func (w *ndjsonWriter) write(key string, report api.VerifyReport) error {
	line, err := json.Marshal(map[string]api.VerifyReport{key: report})
	if err != nil {
		return fmt.Errorf("encode %s line: %w", key, err)
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, err := w.out.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("write %s line: %w", key, err)
	}
	if flusher, ok := w.out.(interface{ Sync() error }); ok && w.out == os.Stdout {
		_ = flusher.Sync()
	}
	return nil
}

func init() {
	fixturesRunCmd.Flags().StringArrayVar(&fixturesRunChanged, "changed", nil,
		"Path the agent changed this turn; repeat once per file (empty means verify the whole tree)")
	fixturesCmd.AddCommand(fixturesRunCmd)
}
