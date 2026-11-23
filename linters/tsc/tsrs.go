package tsc

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/textproto"
	"os"
	"strconv"
	"strings"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/utils"
)

func tsrsAPIArgs() []string {
	// Build mode still writes incremental build info with --noEmit.
	return []string{"--api", "--async", "--callbacks=writeFile:noop,removeFile:noop"}
}

func (t *TSC) runTSRS(ctx context.Context, binary string) ([]models.Violation, error) {
	checker := *t
	checker.Executable = binary
	_, args, err := linters.PrepareCommand(&checker, checker.RunOptions)
	if err != nil {
		return nil, err
	}
	nativeArgs := checker.nativeArgs()
	if len(nativeArgs) != 0 {
		validationArgs := append(append([]string{"--build"}, nativeArgs...), "--force", "--noEmit", "--dry", "--help")
		if output, err := checker.CommandContext(ctx, t.Name(), binary, validationArgs...).CombinedOutput(); err != nil {
			return nil, fmt.Errorf("validate tsrs build arguments: %w\n%s", err, output)
		}
	}

	diagnostics, err := checker.requestTSRSDiagnostics(ctx, binary, args, nativeArgs)
	if err != nil {
		return nil, err
	}
	gitRoot := utils.GitRoot(t.WorkDir)
	violations := make([]models.Violation, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		converted := diagnostic.toDiagnostic()
		violations = append(violations, converted.toViolation(t.WorkDir, gitRoot))
	}
	return violations, nil
}

func (t *TSC) requestTSRSDiagnostics(ctx context.Context, binary string, apiArgs, nativeArgs []string) ([]tsrsDiagnostic, error) {
	stdin, input, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("open tsrs input: %w", err)
	}
	output, stdout := io.Pipe()

	var stderr bytes.Buffer
	cmd := t.CommandContext(ctx, t.Name(), binary, apiArgs...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = stdin, stdout, t.WrapWriter(&stderr)
	done := make(chan error, 1)
	go func() {
		err := cmd.Run()
		_ = stdin.Close()
		_ = stdout.CloseWithError(err)
		done <- err
	}()

	client := tsrsClient{input: input, output: bufio.NewReader(output)}
	diagnostics, err := client.build(nativeArgs)
	_ = input.Close()
	_ = output.Close()
	if runErr := <-done; runErr != nil {
		return nil, fmt.Errorf("tsrs failed: %w\nStderr:\n%s", runErr, stderr.String())
	}
	return diagnostics, err
}

type tsrsClient struct {
	input  io.Writer
	output *bufio.Reader
	id     int
}

func (c *tsrsClient) build(args []string) ([]tsrsDiagnostic, error) {
	if err := c.call("initialize", nil, nil); err != nil {
		return nil, err
	}
	var orchestrator struct {
		ID int `json:"buildOrchestratorID"`
	}
	params := map[string]any{"rootNames": append(append([]string(nil), args...), "--force", "--noEmit")}
	if err := c.call("createBuildOrchestrator", params, &orchestrator); err != nil {
		return nil, err
	}
	if orchestrator.ID <= 0 {
		return nil, fmt.Errorf("tsrs returned an invalid build orchestrator ID: %d", orchestrator.ID)
	}
	var result struct {
		Status      *int             `json:"status"`
		Diagnostics []tsrsDiagnostic `json:"diagnostics"`
	}
	if err := c.call("build", map[string]int{"buildOrchestratorID": orchestrator.ID}, &result); err != nil {
		return nil, err
	}
	if result.Status == nil {
		return nil, fmt.Errorf("tsrs build returned no status")
	}
	if *result.Status != 0 && len(result.Diagnostics) == 0 {
		return nil, fmt.Errorf("tsrs build failed without diagnostics (exit status %d)", *result.Status)
	}
	return result.Diagnostics, nil
}

func (c *tsrsClient) call(method string, params, result any) error {
	c.id++
	request, err := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": c.id, "method": method, "params": params})
	if err != nil {
		return fmt.Errorf("encode tsrs %s request: %w", method, err)
	}
	if _, err := fmt.Fprintf(c.input, "Content-Length: %d\r\n\r\n%s", len(request), request); err != nil {
		return fmt.Errorf("send tsrs %s request: %w", method, err)
	}
	header, err := textproto.NewReader(c.output).ReadMIMEHeader()
	if err != nil {
		return fmt.Errorf("read tsrs %s response: %w", method, err)
	}
	size, err := strconv.ParseInt(header.Get("Content-Length"), 10, 64)
	if err != nil || size <= 0 {
		return fmt.Errorf("invalid tsrs %s response length: %q", method, header.Get("Content-Length"))
	}
	var response struct {
		ID     int             `json:"id"`
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(io.LimitReader(c.output, size)).Decode(&response); err != nil {
		return fmt.Errorf("decode tsrs %s response: %w", method, err)
	}
	if response.ID != c.id {
		return fmt.Errorf("tsrs %s response ID %d, expected %d", method, response.ID, c.id)
	}
	if response.Error != nil {
		return fmt.Errorf("tsrs %s: %s", method, response.Error.Message)
	}
	if result != nil {
		if err := json.Unmarshal(response.Result, result); err != nil {
			return fmt.Errorf("decode tsrs %s result: %w", method, err)
		}
	}
	return nil
}

type tsrsDiagnostic struct {
	FileName      string `json:"fileName"`
	StartPosition struct {
		Line      int `json:"line"`
		Character int `json:"character"`
	} `json:"startPosition"`
	Text         string           `json:"text"`
	Code         int              `json:"code"`
	Category     int              `json:"category"`
	MessageChain []tsrsDiagnostic `json:"messageChain"`
}

func (d tsrsDiagnostic) toDiagnostic() TSCDiagnostic {
	category := "Message"
	switch d.Category {
	case 0:
		category = "Warning"
	case 1:
		category = "Error"
	}
	return TSCDiagnostic{
		File: d.FileName, Line: d.StartPosition.Line + 1, Column: d.StartPosition.Character + 1,
		Code: d.Code, Category: category, Message: d.message(""),
	}
}

func (d tsrsDiagnostic) message(indent string) string {
	var message strings.Builder
	message.WriteString(indent + d.Text)
	for _, child := range d.MessageChain {
		message.WriteByte('\n')
		message.WriteString(child.message(indent + "  "))
	}
	return message.String()
}
