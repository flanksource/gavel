package claude

import (
	"fmt"
	"time"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	claudecode "github.com/severity1/claude-agent-sdk-go"
)

// ClaudeExecutor implements todos.Executor using the Claude Code SDK.
type ClaudeExecutor struct {
	workDir   string
	sessionID string
	options   []claudecode.Option
}

// NewClaudeExecutor creates a new Claude Code executor with the specified configuration.
func NewClaudeExecutor(workDir string, sessionID string, opts ...claudecode.Option) *ClaudeExecutor {
	return &ClaudeExecutor{
		workDir:   workDir,
		sessionID: sessionID,
		options:   opts,
	}
}

// Name returns the executor name for identification.
func (e *ClaudeExecutor) Name() string {
	return "claude-code"
}

// Execute implements the todos.Executor interface using Claude Code SDK.
// It uses the WithClient pattern for automatic resource management.
func (e *ClaudeExecutor) Execute(ctx *todos.ExecutorContext, todo *types.TODO) (*todos.ExecutionResult, error) {
	result := &todos.ExecutionResult{
		ExecutorName: e.Name(),
		Transcript:   ctx.GetTranscript(),
	}

	startTime := time.Now()

	ctx.Logger.Infof("Starting Claude Code session: %s", e.sessionID)
	ctx.Notify(todos.Notification{
		Type:    todos.NotifyProgress,
		Message: fmt.Sprintf("Starting %s session", e.Name()),
	})

	// Build Claude-specific prompt from TODO
	prompt := BuildPrompt(todo)
	ctx.Logger.Debugf("Prompt length: %d characters", len(prompt))

	// Execute with Claude Code SDK using WithClient pattern
	err := claudecode.WithClient(ctx, func(client claudecode.Client) error {
		// Send initial query with session ID for context isolation
		ctx.Logger.Debugf("Sending query to Claude with session ID: %s", e.sessionID)
		if err := client.QueryWithSession(ctx, prompt, e.sessionID); err != nil {
			return fmt.Errorf("query failed: %w", err)
		}

		// Stream messages with interactive context
		ctx.Logger.Debugf("Starting message streaming")
		streamer := NewMessageStreamer(ctx, client, e.workDir)
		return streamer.Stream()
	}, e.options...)

	result.Duration = time.Since(startTime)

	if err != nil {
		ctx.Logger.Errorf("Claude execution failed: %v", err)
		result.Success = false
		result.ErrorMessage = err.Error()
		return result, err
	}

	// Parse result metadata from transcript
	parseResultMetadata(result)

	ctx.Logger.Infof("Claude execution completed successfully in %v", result.Duration)
	result.Success = true
	return result, nil
}

// parseResultMetadata extracts tokens, cost, and other metadata from the transcript.
func parseResultMetadata(result *todos.ExecutionResult) {
	for _, entry := range result.Transcript.Entries {
		if entry.Type == todos.EntryNotification {
			// Extract token usage
			if tokens, ok := entry.Metadata["tokens"].(int); ok {
				result.TokensUsed = tokens
			}
			// Extract cost
			if cost, ok := entry.Metadata["cost"].(float64); ok {
				result.CostUSD = cost
			}
			// Extract turn count
			if turns, ok := entry.Metadata["turns"].(int); ok {
				result.NumTurns = turns
			}
		}
		// Track actions performed
		if entry.Type == todos.EntryAction {
			if action, ok := entry.Metadata["action"].(string); ok {
				result.ActionsPerformed = append(result.ActionsPerformed, action)
			}
		}
	}
}
