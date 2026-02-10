package claude

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/task"
	flanksourceContext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/todos/types"
)

// ClaudeClient wraps Claude Code CLI execution
type ClaudeClient struct {
	workDir     string
	timeout     time.Duration
	claudePath  string // Path to claude executable
	sessionLogs []string
}

// ClaudeConfig configures the Claude Code client
type ClaudeConfig struct {
	WorkingDir string
	Timeout    time.Duration
	ClaudePath string // Optional: path to claude executable
}

// CompletionResult represents the result of a Claude Code execution
type CompletionResult struct {
	Success      bool
	Output       string
	Error        string
	FilesChanged []string
}

// NewClaudeClient creates a new Claude Code client
func NewClaudeClient(config ClaudeConfig) (*ClaudeClient, error) {
	if config.WorkingDir == "" {
		var err error
		config.WorkingDir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	if config.Timeout == 0 {
		config.Timeout = 10 * time.Minute // Default 10 minute timeout
	}

	claudePath := config.ClaudePath
	if claudePath == "" {
		// Try to find claude in PATH
		path, err := exec.LookPath("claude")
		if err != nil {
			return nil, fmt.Errorf("claude executable not found in PATH. Please install Claude Code CLI or specify path with ClaudePath")
		}
		claudePath = path
	}

	return &ClaudeClient{
		workDir:     config.WorkingDir,
		timeout:     config.Timeout,
		claudePath:  claudePath,
		sessionLogs: make([]string, 0),
	}, nil
}

// SendPrompt sends a prompt to Claude Code and waits for completion
func (c *ClaudeClient) SendPrompt(ctx context.Context, prompt string) (*CompletionResult, error) {
	// Create a task for tracking
	typedTask := clicky.StartTask[*CompletionResult](
		"Execute Claude Code",
		func(taskCtx flanksourceContext.Context, t *task.Task) (*CompletionResult, error) {
			return c.executeClaudeCode(taskCtx, prompt)
		},
		clicky.WithTaskTimeout(c.timeout),
	)

	return typedTask.GetResult()
}

// executeClaudeCode runs the claude CLI with the given prompt
func (c *ClaudeClient) executeClaudeCode(ctx flanksourceContext.Context, prompt string) (*CompletionResult, error) {
	// Write prompt to temporary file
	promptFile, err := c.writePromptFile(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to write prompt file: %w", err)
	}
	defer os.Remove(promptFile)

	ctx.Infof("Executing Claude Code with prompt from: %s", promptFile)

	// Execute claude with the prompt
	cmd := exec.CommandContext(ctx, c.claudePath, "chat", "--message-file", promptFile)
	cmd.Dir = c.workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set environment variables
	cmd.Env = append(os.Environ(),
		"CLAUDE_CODE_WORKING_DIR="+c.workDir,
	)

	// Run command
	err = cmd.Run()

	// Capture output
	output := stdout.String()
	errOutput := stderr.String()

	// Store session logs
	c.sessionLogs = append(c.sessionLogs, output)
	if errOutput != "" {
		c.sessionLogs = append(c.sessionLogs, errOutput)
	}

	result := &CompletionResult{
		Output: output,
		Error:  errOutput,
	}

	// Check for completion signal
	if ExtractCompletionSignal(output) {
		result.Success = true
		ctx.Infof("✅ Implementation completed successfully")
	} else if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Command failed: %v\nStderr: %s", err, errOutput)
		ctx.Errorf("❌ Implementation failed: %v", err)
	} else {
		// Command succeeded but no completion signal
		result.Success = false
		result.Error = "No IMPLEMENTATION_COMPLETE signal found in output"
		ctx.Warnf("⚠️ Command succeeded but no completion signal found")
	}

	// Try to detect changed files from output
	result.FilesChanged = c.extractChangedFiles(output)

	return result, nil
}

// SendRetryPrompt sends a retry prompt with failure details
func (c *ClaudeClient) SendRetryPrompt(ctx context.Context, todo *types.TODO, failureDetails string) (*CompletionResult, error) {
	prompt := BuildRetryPrompt(todo, failureDetails)
	return c.SendPrompt(ctx, prompt)
}

// GetSessionLogs returns the accumulated session logs
func (c *ClaudeClient) GetSessionLogs() string {
	return strings.Join(c.sessionLogs, "\n---\n")
}

// writePromptFile writes the prompt to a temporary file
func (c *ClaudeClient) writePromptFile(prompt string) (string, error) {
	tmpDir := os.TempDir()
	promptFile := filepath.Join(tmpDir, fmt.Sprintf("claude-prompt-%d.txt", time.Now().Unix()))

	if err := os.WriteFile(promptFile, []byte(prompt), 0644); err != nil {
		return "", fmt.Errorf("failed to write prompt file: %w", err)
	}

	return promptFile, nil
}

// extractChangedFiles attempts to extract file paths from Claude's output
func (c *ClaudeClient) extractChangedFiles(output string) []string {
	var files []string

	// Simple heuristic: look for common file modification patterns
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for patterns like "Created file: ...", "Modified: ...", etc.
		prefixes := []string{
			"Created file:",
			"Modified:",
			"Updated:",
			"Edited:",
			"File created:",
			"File modified:",
		}

		for _, prefix := range prefixes {
			if strings.HasPrefix(line, prefix) {
				// Extract the file path
				file := strings.TrimSpace(strings.TrimPrefix(line, prefix))
				if file != "" {
					files = append(files, file)
				}
			}
		}
	}

	return files
}

// Close cleans up any resources (for interface compatibility)
func (c *ClaudeClient) Close() error {
	// Nothing to clean up for CLI-based client
	return nil
}

// ExtractCompletionSignal checks if the output contains the completion signal
func ExtractCompletionSignal(output string) bool {
	return strings.Contains(output, "IMPLEMENTATION_COMPLETE")
}

// ReadCLAUDEMD reads the CLAUDE.md file from the working directory
func ReadCLAUDEMD(workDir string) (string, error) {
	// For now, return a simplified version of key best practices
	// In a full implementation, we would read from ~/.claude/CLAUDE.md or workDir/CLAUDE.md
	return `
## Key Best Practices

**TDD (C-1 MUST)**: Write tests first, then implementation
**Small Functions (C-8 SHOULD NOT)**: Functions < 50 lines, files < 400 lines
**No Unnecessary Abstraction (C-7 SHOULD NOT)**: Only extract functions if reused or improving testability
**Domain Vocabulary (C-2 MUST)**: Use existing naming conventions
**Test Quality (T-1 MUST)**: Unit tests colocated in *_test.go files
**Separation (T-3 MUST)**: Unit tests separate from integration tests
`, nil
}
