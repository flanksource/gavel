package linters

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
)

// RunReport runs a linter speaking gavel's report protocol and parses what it
// reports (ParseReport). A command that cannot start is an error; one that
// exits non-zero is judged by its report.
func (o RunOptions) RunReport(ctx context.Context, name, binary string, args ...string) (*LinterResult, error) {
	command := o.CommandContext(ctx, name, binary, args...)
	var stdout, stderr bytes.Buffer
	command.Stdout = o.WrapWriter(&stdout)
	command.Stderr = o.WrapWriter(&stderr)
	exitCode := 0
	if err := command.Run(); err != nil {
		var exited *exec.ExitError
		if !errors.As(err, &exited) {
			return nil, fmt.Errorf("%s execution: %w: %s", name, err, stderr.String())
		}
		exitCode = exited.ExitCode()
	}
	result, err := ParseReport(stdout.Bytes(), exitCode)
	if err != nil {
		return result, fmt.Errorf("%s: %w; stderr: %s", name, err, stderr.String())
	}
	return result, nil
}

// ParseReport reads a gavel-protocol linter's report: a JSON array holding
// exactly one LinterResult, written by a process that exits 0 when the report
// has no failures (HasFailures) and 1 when it has some. Once the report
// decodes it is returned with any error, so its findings survive a failed
// execution.
func ParseReport(output []byte, exitCode int) (*LinterResult, error) {
	var results []*LinterResult
	if err := json.Unmarshal(output, &results); err != nil {
		return nil, fmt.Errorf("decode lint report: %w", err)
	}
	if len(results) != 1 || results[0] == nil {
		return nil, fmt.Errorf("expected one linter result, got %d", len(results))
	}
	result := results[0]
	if exitCode != 0 && exitCode != 1 {
		return result, fmt.Errorf("unexpected exit status %d", exitCode)
	}
	if result.Skipped || result.TimedOut || !result.Success || result.Error != "" {
		return result, fmt.Errorf("lint execution failed: %s", result.Error)
	}
	if (exitCode == 1) != HasFailures(results) {
		return result, errors.New("lint report disagrees with exit status")
	}
	return result, nil
}
