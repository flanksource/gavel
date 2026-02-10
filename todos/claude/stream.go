package claude

import (
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/gavel/claudehistory"
	"github.com/flanksource/gavel/todos"
	claudecode "github.com/severity1/claude-agent-sdk-go"
)

// MessageStreamer handles streaming Claude Code messages and detects user interaction needs.
// It uses ExecutorContext for both internal logging and user-facing notifications.
type MessageStreamer struct {
	ctx     *todos.ExecutorContext
	client  claudecode.Client
	workDir string
}

// NewMessageStreamer creates a new message streamer with the given context and client.
func NewMessageStreamer(ctx *todos.ExecutorContext, client claudecode.Client, workDir string) *MessageStreamer {
	return &MessageStreamer{
		ctx:     ctx,
		client:  client,
		workDir: workDir,
	}
}

// Stream processes messages from Claude, handling interactions via ctx.Ask().
func (s *MessageStreamer) Stream() error {
	msgChan := s.client.ReceiveMessages(s.ctx)

	for {
		select {
		case message := <-msgChan:
			if message == nil {
				s.ctx.Logger.Debugf("Message stream ended")
				return nil // Stream ended
			}

			if err := s.handleMessage(message); err != nil {
				return err
			}

		case <-s.ctx.Done():
			s.ctx.Logger.Debugf("Context cancelled")
			return s.ctx.Err()
		}
	}
}

// handleMessage processes a single message from Claude.
func (s *MessageStreamer) handleMessage(message claudecode.Message) error {
	switch msg := message.(type) {
	case *claudecode.AssistantMessage:
		return s.handleAssistantMessage(msg)

	case *claudecode.SystemMessage:
		return s.handleSystemMessage(msg)

	case *claudecode.ResultMessage:
		return s.handleResultMessage(msg)
	}

	return nil
}

// handleAssistantMessage processes Claude's responses and detects questions.
func (s *MessageStreamer) handleAssistantMessage(msg *claudecode.AssistantMessage) error {
	for _, block := range msg.Content {
		switch b := block.(type) {
		case *claudecode.TextBlock:
			// Log internally
			s.ctx.Logger.Debugf("Assistant: %s", truncateForLog(b.Text))

			// Record in transcript
			s.ctx.GetTranscript().AddExecutorMessage(b.Text, todos.EntryText, nil)

			// Check if Claude is asking a clarifying question
			if question := detectQuestion(b.Text); question != nil {
				// Use executor-agnostic Ask()
				response, err := s.ctx.Ask(*question)
				if err != nil {
					return err
				}

				// Send user's response back to Claude in the same session
				s.ctx.Logger.Debugf("Sending user response to Claude")
				return s.client.Query(s.ctx, response)
			}

		case *claudecode.ThinkingBlock:
			// Log internally
			s.ctx.Logger.Debugf("Thinking: %s", truncateForLog(b.Thinking))

			// Record in transcript
			s.ctx.GetTranscript().AddExecutorMessage(b.Thinking, todos.EntryThinking, nil)

			// Notify user
			s.ctx.Notify(todos.Notification{
				Type:    todos.NotifyThinking,
				Message: b.Thinking,
			})

		case *claudecode.ToolUseBlock:

			// Prepare notification text
			notificationText := claudehistory.ToolUse{
				Tool:  b.Name,
				Input: b.Input,
				// Outcome: b.Output,
			}.Pretty()

			// Log internally
			s.ctx.Logger.Infof("Tool use: %s", b.Name)

			// Record in transcript
			action := fmt.Sprintf("%s(%v)", b.Name, b.Input)
			s.ctx.GetTranscript().AddExecutorMessage(action, todos.EntryAction, map[string]interface{}{
				"tool":   b.Name,
				"action": action,
			})

			// Display to user with chroma-highlighted output
			fmt.Println(notificationText.ANSI())
		}
	}

	return nil
}

// handleSystemMessage processes system-level messages from Claude.
func (s *MessageStreamer) handleSystemMessage(msg *claudecode.SystemMessage) error {
	s.ctx.Logger.Debugf("System message: subtype=%s", msg.Subtype)

	// Check for system-level clarification requests
	if msg.Subtype == "ask_user" || msg.Subtype == "clarification_needed" {
		if questionText, ok := msg.Data["question"].(string); ok {
			question := todos.Question{
				Text:      questionText,
				Timestamp: time.Now(),
				Context:   fmt.Sprintf("System: %s", msg.Subtype),
			}

			response, err := s.ctx.Ask(question)
			if err != nil {
				return err
			}

			return s.client.Query(s.ctx, response)
		}
	}

	return nil
}

// handleResultMessage processes the final result from Claude.
func (s *MessageStreamer) handleResultMessage(msg *claudecode.ResultMessage) error {
	if msg.IsError {
		s.ctx.Logger.Errorf("Claude error: %v", msg.Result)
		s.ctx.Notify(todos.Notification{
			Type:    todos.NotifyError,
			Message: fmt.Sprintf("Error: %v", msg.Result),
		})
		return fmt.Errorf("execution error: %v", msg.Result)
	}

	// Extract metadata
	metadata := map[string]interface{}{
		"turns": msg.NumTurns,
	}

	// Extract token usage
	if msg.Usage != nil {
		if inputTokens, ok := (*msg.Usage)["input_tokens"].(float64); ok {
			if outputTokens, ok := (*msg.Usage)["output_tokens"].(float64); ok {
				tokens := int(inputTokens + outputTokens)
				metadata["tokens"] = tokens
				s.ctx.Logger.Infof("Tokens used: %d (input: %.0f, output: %.0f)", tokens, inputTokens, outputTokens)
			}
		}
	}

	// Extract cost
	if msg.TotalCostUSD != nil {
		metadata["cost"] = *msg.TotalCostUSD
		s.ctx.Logger.Infof("Cost: $%.4f", *msg.TotalCostUSD)
	}

	// Notify user of completion
	s.ctx.Notify(todos.Notification{
		Type:    todos.NotifyCompletion,
		Message: "Execution completed",
		Data:    metadata,
	})

	return nil
}

// detectQuestion checks if text contains a clarifying question from Claude.
func detectQuestion(text string) *todos.Question {
	lowerText := strings.ToLower(text)

	// Common patterns that indicate Claude is asking for clarification
	patterns := []string{
		"i need clarification",
		"could you clarify",
		"which would you prefer",
		"would you like me to",
		"should i",
		"can you specify",
		"please let me know",
		"what would you like",
		"do you want me to",
		"would you prefer",
	}

	for _, pattern := range patterns {
		if strings.Contains(lowerText, pattern) {
			return &todos.Question{
				Text:      text,
				Timestamp: time.Now(),
				Context:   "Detected clarification request",
			}
		}
	}

	return nil
}

// truncateForLog truncates long text for cleaner log output.
func truncateForLog(text string) string {
	const maxLen = 100
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

// getRelativePath converts an absolute path to relative from workDir.
