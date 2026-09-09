package todos

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky"
	clickyapi "github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
	"github.com/flanksource/commons/logger"
	"github.com/google/uuid"
)

// ExecutorContext wraps context.Context with logging and user interaction capabilities.
// The Logger field is for executor internal logging (info/debug/error messages).
// UserInteraction is for user-facing notifications and clarifying questions.
type ExecutorContext struct {
	context.Context
	Logger        logger.Logger // For executor internal logging
	interaction   *UserInteraction
	transcript    *ExecutionTranscript
	onSessionID   func(sessionID string)
	onRunPrepared func(RunPreparationResult)
	onRunStart    func(RunStartMetadata)
	onNotices     func(sessionID string, notices []api.Notice)
	// onVerifyProgress receives the definition of done's in-flight reports.
	onVerifyProgress func(api.VerifyReport)
}

// RunStartMetadata is the human-facing run identity recorded on a TODO issue
// when an agent run starts.
type RunStartMetadata struct {
	SessionID     string
	Mode          string
	Driver        string
	Agent         string
	Provider      string
	RuntimeMode   string
	ResolvedModel string
	Effort        string
	// Spec, when set, is the request as it stands after setup has run: the
	// checkout consumed and Cwd pointing at the tree the agent works in. It is
	// reported once, from a hook that trails the setup plugin, and updates the
	// run's persisted rendered spec. Reports that cannot see it leave it nil
	// rather than overwriting with the pre-setup request.
	Spec *api.Spec
}

// UserInteraction handles user-facing communication during TODO execution.
// All output should use clicky.Text() for consistent formatting.
type UserInteraction struct {
	// AskFunc presents questions to the user and returns their response.
	// Questions are formatted with clicky.Text() before display.
	AskFunc func(question Question) (string, error)

	// NotifyFunc displays progress updates to the user.
	// Notifications are formatted with clicky.Text() for pretty output.
	NotifyFunc func(notification Notification)
}

// Question represents a clarification request from the executor to the user.
type Question struct {
	Text      string
	Timestamp time.Time
	Context   string   // Additional context about why the question is being asked
	Options   []string // Optional predefined choices
}

// Pretty returns a formatted text representation of the Question
func (q Question) Pretty() clickyapi.Text {
	result := clicky.Text("").Add(icons.Unknown).Append(" Question", "text-orange-600 font-bold")

	if q.Text != "" {
		result = result.Append(": ", "text-gray-400").Append(q.Text, "text-gray-800 font-medium")
	}

	// Add context if present
	if q.Context != "" {
		result = result.Append("\n  Context: ", "text-gray-500").Append(q.Context, "text-gray-600")
	}

	// Add options if present
	if len(q.Options) > 0 {
		result = result.Append("\n  Options: ", "text-gray-500")
		for i, option := range q.Options {
			if i > 0 {
				result = result.Append(", ", "text-gray-400")
			}
			result = result.Append(option, "text-blue-600")
		}
	}

	// Add timestamp
	result = result.Append("\n  ", "").Add(icons.Info).Append(" ", "").Append(q.Timestamp.Format("15:04:05"), "text-gray-500 text-xs")

	return result
}

// Notification represents a user-facing status update during execution.
type Notification struct {
	Type    NotificationType
	Message string
	Data    map[string]interface{} // Additional structured data
}

// NotificationType categorizes different kinds of notifications.
type NotificationType string

const (
	NotifyThinking   NotificationType = "thinking"   // Executor is reasoning
	NotifyAction     NotificationType = "action"     // Executor performing action (tool use)
	NotifyProgress   NotificationType = "progress"   // Progress update
	NotifyCompletion NotificationType = "completion" // Task completed
	NotifyError      NotificationType = "error"      // Error occurred
	NotifyApproval   NotificationType = "approval"   // Agent is awaiting human approval to use a tool
)

// Pretty returns a formatted text representation of the NotificationType with appropriate styling
func (nt NotificationType) Pretty() clickyapi.Text {
	switch nt {
	case NotifyThinking:
		return clicky.Text("").Add(icons.Lambda).Append(" THINKING", "text-purple-600 font-medium")
	case NotifyAction:
		return clicky.Text("").Add(icons.ArrowRight).Append(" ACTION", "text-blue-600 font-medium")
	case NotifyProgress:
		return clicky.Text("").Add(icons.ArrowRight).Append(" PROGRESS", "text-cyan-600 font-medium")
	case NotifyCompletion:
		return clicky.Text("").Add(icons.Pass).Append(" COMPLETED", "text-green-600 font-bold")
	case NotifyError:
		return clicky.Text("").Add(icons.Error).Append(" ERROR", "text-red-600 font-bold")
	case NotifyApproval:
		return clicky.Text(" APPROVAL", "text-amber-600 font-bold")
	default:
		return clicky.Text(string(nt), "text-gray-500")
	}
}

func file(path string) clickyapi.Text {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	relPath, err := filepath.Rel(cwd, path)
	if err != nil {
		relPath = path
	}

	return clicky.Text("").Add(icons.File).Append(" ", "").Append(relPath, "text-blue-500")
}

// Pretty returns a formatted text representation of the Notification
func (n Notification) Pretty() clickyapi.Text {
	result := n.Type.Pretty()

	if n.Message != "" {
		result = result.Append(": ", "text-gray-400").Append(n.Message, "text-gray-700")
	}

	// Print additional data if present
	if n.Data != nil {

		for k, v := range n.Data {
			switch k {
			case "file":
				result = result.Append(" ").Add(file(v.(string)))
			case "tokens":
				result.Append(fmt.Sprintf(" Tokens: %d", v.(int)), "text-gray-500")
			case "cost":
				result.Append(fmt.Sprintf(" Cost: $%.4f", v.(float64)), "text-gray-500")
			default:
				result = result.Append(" ").Add(clicky.KeyValue(k, v))
			}
		}
	}

	return result
}

// NewExecutorContext creates a new context with logging and user interaction capabilities.
func NewExecutorContext(parent context.Context, log logger.Logger, interaction *UserInteraction) *ExecutorContext {
	return &ExecutorContext{
		Context:     parent,
		Logger:      log,
		interaction: interaction,
		transcript:  NewExecutionTranscript(),
	}
}

// Ask presents a question to the user and waits for their response.
// The question is logged internally and recorded in the transcript.
func (ctx *ExecutorContext) Ask(question Question) (string, error) {
	if ctx.interaction == nil || ctx.interaction.AskFunc == nil {
		return "", fmt.Errorf("no user interaction configured")
	}

	// Log internally
	ctx.Logger.Debugf("Asking user: %s", question.Text)

	// Record question in transcript
	ctx.transcript.AddQuestion(question)

	// Get user response
	response, err := ctx.interaction.AskFunc(question)
	if err != nil {
		ctx.Logger.Errorf("User interaction failed: %v", err)
		return "", fmt.Errorf("user interaction failed: %w", err)
	}

	// Record response in transcript
	ctx.transcript.AddUserResponse(question, response)

	return response, nil
}

// Notify sends a status update to the user.
// The notification is logged internally at the appropriate level and recorded in the transcript.
func (ctx *ExecutorContext) Notify(notification Notification) {
	fmt.Println(notification.Pretty().ANSI())

	// Send to user via NotifyFunc
	if ctx.interaction != nil && ctx.interaction.NotifyFunc != nil {
		ctx.interaction.NotifyFunc(notification)
	}

	// Record in transcript
	ctx.transcript.AddNotification(notification)
}

// GetTranscript returns the complete execution transcript.
func (ctx *ExecutorContext) GetTranscript() *ExecutionTranscript {
	return ctx.transcript
}

// SetSessionIDHook registers a callback an executor invokes (via RecordSessionID)
// as soon as it has finalized the run's session id and before it launches the
// agent. The orchestrator uses it to persist the session id up front so it
// survives a mid-run crash and remains resumable.
func (ctx *ExecutorContext) SetSessionIDHook(fn func(sessionID string)) {
	ctx.onSessionID = fn
}

// RecordSessionID reports the run's session id to the registered hook (if any).
// Executors call it once, before launching the agent. A blank id is ignored.
func (ctx *ExecutorContext) RecordSessionID(sessionID string) {
	if sessionID == "" || ctx.onSessionID == nil {
		return
	}
	ctx.onSessionID(sessionID)
}

// SetRunPreparedHook registers a callback invoked after Captain has admitted
// the run and before the external agent starts.
func (ctx *ExecutorContext) SetRunPreparedHook(fn func(RunPreparationResult)) {
	ctx.onRunPrepared = fn
}

// RecordRunPrepared reports Captain's durable admission identity, and binds the
// admitted run to this execution's context. A TODO can have more than one run
// in flight, so every later callback has to report against the run it was
// prepared for rather than against whichever run the TODO points at now.
func (ctx *ExecutorContext) RecordRunPrepared(result RunPreparationResult) {
	if result.PromptRunID != uuid.Nil {
		ctx.Context = WithPromptRun(ctx.Context, result.PromptRunID)
	}
	// The admission session is bound alongside the run because the two together
	// are the identity a durable tool approval is written against; a broker built
	// mid-execution has no other way to learn them.
	if sessionID, err := uuid.Parse(result.SessionID); err == nil && sessionID != uuid.Nil {
		ctx.Context = WithCaptainSession(ctx.Context, sessionID)
	}
	if result.SessionID == "" || ctx.onRunPrepared == nil {
		return
	}
	ctx.onRunPrepared(result)
}

type promptRunContextKey struct{}

type captainSessionContextKey struct{}

// WithCaptainSession binds Captain's admission session to a context, so a hook
// or broker running inside the execution can address the rows Captain keyed on
// it.
func WithCaptainSession(ctx context.Context, sessionID uuid.UUID) context.Context {
	return context.WithValue(ctx, captainSessionContextKey{}, sessionID)
}

// CaptainSessionFromContext returns the session bound by WithCaptainSession, or
// uuid.Nil outside an admitted run.
func CaptainSessionFromContext(ctx context.Context) uuid.UUID {
	if ctx == nil {
		return uuid.Nil
	}
	sessionID, _ := ctx.Value(captainSessionContextKey{}).(uuid.UUID)
	return sessionID
}

// WithPromptRun binds a prompt run to a context so the provider callbacks made
// while executing it resolve that run and no other.
func WithPromptRun(ctx context.Context, promptRunID uuid.UUID) context.Context {
	return context.WithValue(ctx, promptRunContextKey{}, promptRunID)
}

// PromptRunFromContext returns the run bound by WithPromptRun, or uuid.Nil for
// a caller that is not inside a run (a read surface, or a recovery command).
func PromptRunFromContext(ctx context.Context) uuid.UUID {
	if ctx == nil {
		return uuid.Nil
	}
	promptRunID, _ := ctx.Value(promptRunContextKey{}).(uuid.UUID)
	return promptRunID
}

// SetRunStartHook registers a callback an executor invokes with the resolved
// runtime before dispatch and again if the provider later supplies a session ID.
func (ctx *ExecutorContext) SetRunStartHook(fn func(RunStartMetadata)) {
	ctx.onRunStart = fn
}

// RecordRunStart reports the run metadata to the registered hook (if any).
// Executors call it once before dispatch with the resolved runtime, then again
// when the provider supplies a session ID.
func (ctx *ExecutorContext) RecordRunStart(meta RunStartMetadata) {
	if ctx.onRunStart == nil {
		return
	}
	ctx.onRunStart(meta)
}

// SetVerifyProgressHook registers a callback an executor hands to captain's
// verify Options.Progress, so every in-flight verification report reaches the
// run's persistence without the executor knowing what persistence is.
func (ctx *ExecutorContext) SetVerifyProgressHook(fn func(api.VerifyReport)) {
	ctx.onVerifyProgress = fn
}

// RecordVerifyProgress reports one in-flight verification snapshot to the
// registered hook (if any).
func (ctx *ExecutorContext) RecordVerifyProgress(report api.VerifyReport) {
	if ctx.onVerifyProgress == nil {
		return
	}
	ctx.onVerifyProgress(report)
}

// SetNoticesHook registers a callback an executor invokes (via RecordNotices)
// once the run is over, with everything its lifecycle hooks reported doing.
func (ctx *ExecutorContext) SetNoticesHook(fn func(sessionID string, notices []api.Notice)) {
	ctx.onNotices = fn
}

// RecordNotices reports the run's lifecycle notices to the registered hook (if
// any). Executors call it after the run, when the session id is settled.
func (ctx *ExecutorContext) RecordNotices(sessionID string, notices []api.Notice) {
	if len(notices) == 0 || ctx.onNotices == nil {
		return
	}
	ctx.onNotices(sessionID, notices)
}

// ExecutionTranscript records the complete interaction history during TODO execution.
// This is executor-agnostic and works with any AI system.
type ExecutionTranscript struct {
	Entries []TranscriptEntry
}

// TranscriptEntry represents a single event in the execution transcript.
type TranscriptEntry struct {
	Timestamp time.Time
	Type      EntryType
	Role      string // "executor", "user", "system"
	Content   string
	Metadata  map[string]interface{}
}

// Pretty returns a formatted text representation of the TranscriptEntry
func (te TranscriptEntry) Pretty() clickyapi.Text {
	result := clicky.Text("")

	// Add timestamp
	timestamp := te.Timestamp.Format("15:04:05")
	result = result.Append("[", "text-gray-400").Append(timestamp, "text-gray-500").Append("] ", "text-gray-400")

	// Add type with icon
	result = result.Add(te.Type.Pretty())

	// Add role if present
	if te.Role != "" {
		result = result.Append(" (", "text-gray-400").Append(te.Role, "text-blue-500 font-medium").Append(")", "text-gray-400")
	}

	// Add content (truncated if too long)
	content := te.Content
	result = result.Append(": ", "text-gray-600").Append(content, "text-gray-400 max-w-[120ch]")

	return result
}

// EntryType categorizes different kinds of transcript entries.
type EntryType string

const (
	EntryText         EntryType = "text"          // Regular text from executor
	EntryAction       EntryType = "action"        // Tool use, code execution, etc.
	EntryThinking     EntryType = "thinking"      // Reasoning/planning
	EntryQuestion     EntryType = "question"      // Question to user
	EntryUserResponse EntryType = "user_response" // User's answer
	EntryNotification EntryType = "notification"  // Status notification
)

// Pretty returns a formatted text representation of the EntryType with appropriate styling
func (et EntryType) Pretty() clickyapi.Text {
	switch et {
	case EntryText:
		return clicky.Text("").Add(icons.File).Append(" TEXT", "text-blue-600 font-medium")
	case EntryAction:
		return clicky.Text("").Add(icons.ArrowRight).Append(" ACTION", "text-green-600 font-medium")
	case EntryThinking:
		return clicky.Text("").Add(icons.Lambda).Append(" THINKING", "text-purple-600 font-medium")
	case EntryQuestion:
		return clicky.Text("").Add(icons.Warning).Append(" QUESTION", "text-orange-600 font-medium")
	case EntryUserResponse:
		return clicky.Text("").Add(icons.Info).Append(" RESPONSE", "text-cyan-600 font-medium")
	case EntryNotification:
		return clicky.Text("").Add(icons.Pass).Append(" NOTIFICATION", "text-gray-600 font-medium")
	default:
		return clicky.Text(string(et), "text-gray-500")
	}
}

// NewExecutionTranscript creates a new empty transcript.
func NewExecutionTranscript() *ExecutionTranscript {
	return &ExecutionTranscript{
		Entries: make([]TranscriptEntry, 0),
	}
}

// AddQuestion records a question in the transcript.
func (t *ExecutionTranscript) AddQuestion(q Question) {
	t.Entries = append(t.Entries, TranscriptEntry{
		Timestamp: q.Timestamp,
		Type:      EntryQuestion,
		Role:      "executor",
		Content:   q.Text,
		Metadata: map[string]interface{}{
			"context": q.Context,
			"options": q.Options,
		},
	})
}

// AddUserResponse records a user's response to a question.
func (t *ExecutionTranscript) AddUserResponse(q Question, response string) {
	t.Entries = append(t.Entries, TranscriptEntry{
		Timestamp: time.Now(),
		Type:      EntryUserResponse,
		Role:      "user",
		Content:   response,
		Metadata: map[string]interface{}{
			"question": q.Text,
		},
	})
}

// AddNotification records a notification in the transcript.
func (t *ExecutionTranscript) AddNotification(n Notification) {
	t.Entries = append(t.Entries, TranscriptEntry{
		Timestamp: time.Now(),
		Type:      EntryNotification,
		Role:      "system",
		Content:   n.Message,
		Metadata: map[string]interface{}{
			"notification_type": n.Type,
			"data":              n.Data,
		},
	})
}

// AddExecutorMessage records a message from the executor.
func (t *ExecutionTranscript) AddExecutorMessage(content string, entryType EntryType, metadata map[string]interface{}) {

	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	t.Entries = append(t.Entries, TranscriptEntry{
		Timestamp: time.Now(),
		Type:      entryType,
		Role:      "executor",
		Content:   content,
		Metadata:  metadata,
	})
}
