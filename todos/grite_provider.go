package todos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/task"
	depsinstaller "github.com/flanksource/deps/pkg/installer"
	depstypes "github.com/flanksource/deps/pkg/types"
	"github.com/flanksource/gavel/todos/types"
)

const statusLabelPrefix = "status:"
const priorityLabelPrefix = "priority:"
const sessionLabelPrefix = "session:"

type GriteCommandRunner func(ctx context.Context, workDir, binary string, args ...string) ([]byte, error)

type GriteProvider struct {
	WorkDir string
	Binary  string
	BinDir  string
	AppDir  string
	Runner  GriteCommandRunner
}

func NewGriteProvider(workDir string) *GriteProvider {
	return &GriteProvider{WorkDir: workDir}
}

func (p *GriteProvider) List(ctx context.Context, filters DiscoveryFilters) (types.TODOS, error) {
	var out types.TODOS
	seen := map[string]bool{}
	for _, state := range griteListStates(filters) {
		args := []string{"issue", "list", "--json"}
		if state != "" {
			args = append(args, "--state", state)
		}
		raw, err := p.run(ctx, args...)
		if err != nil {
			return nil, err
		}
		data, err := decodeGrite[griteIssueListData](raw)
		if err != nil {
			return nil, err
		}
		for _, issue := range data.Issues {
			if seen[issue.IssueID] {
				continue
			}
			seen[issue.IssueID] = true
			todo := todoFromGriteIssue(issue, p.WorkDir)
			if filters.Matches(todo) {
				out = append(out, todo)
			}
		}
	}
	out.Sort()
	return out, nil
}

func (p *GriteProvider) Get(ctx context.Context, ref string) (*types.TODO, error) {
	raw, err := p.run(ctx, "issue", "show", ref, "--json")
	if err != nil {
		return nil, err
	}
	data, err := decodeGrite[griteIssueShowData](raw)
	if err != nil {
		return nil, err
	}
	body := bodyFromGriteEvents(data.Events)
	defaults := frontmatterFromGriteIssue(data.Issue, p.WorkDir)
	todo, err := ParseTODOContent(data.Issue.Title, body, p.WorkDir, defaults)
	if err != nil {
		return nil, err
	}
	applyGriteIdentity(todo, data.Issue)
	todo.ProviderEvents = providerEventsFromGriteEvents(data.Events)
	return todo, nil
}

func (p *GriteProvider) Create(ctx context.Context, req CreateRequest) (*types.TODO, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}
	priority := req.Priority
	if priority == "" {
		priority = types.PriorityMedium
	}
	status := req.Status
	if status == "" {
		status = types.StatusPending
	}
	args := []string{"issue", "create", "--title", title, "--body", strings.TrimSpace(req.Body)}
	for _, label := range griteCreateLabels(priority, status) {
		args = append(args, "--label", label)
	}
	args = append(args, "--json")
	raw, err := p.run(ctx, args...)
	if err != nil {
		return nil, err
	}
	data, err := decodeGrite[griteIssueCreateData](raw)
	if err != nil {
		return nil, err
	}
	id := data.ID()
	if id == "" {
		return nil, fmt.Errorf("grite issue create did not return an issue id")
	}
	todo, err := p.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if status == types.StatusCompleted {
		if err := p.UpdateState(ctx, todo, StateUpdate{Status: &status}); err != nil {
			return nil, err
		}
	}
	return todo, nil
}

func (p *GriteProvider) Delete(ctx context.Context, todo *types.TODO) error {
	id := TODOReference(todo)
	if id == "" {
		return fmt.Errorf("missing grite issue id")
	}
	if todo != nil && todo.ProviderState == "closed" {
		return nil
	}
	_, err := p.run(ctx, "issue", "close", id, "--json")
	if err == nil && todo != nil {
		todo.ProviderState = "closed"
		todo.Status = types.StatusCompleted
	}
	return err
}

func (p *GriteProvider) UpdateState(ctx context.Context, todo *types.TODO, updates StateUpdate) error {
	id := TODOReference(todo)
	if id == "" {
		return fmt.Errorf("missing grite issue id")
	}
	if updates.Priority != nil {
		if err := p.applyPriority(ctx, id, todo, *updates.Priority); err != nil {
			return err
		}
	}
	if updates.SessionID != nil && *updates.SessionID != "" {
		if err := p.applySessionLabel(ctx, id, todo, *updates.SessionID); err != nil {
			return err
		}
	}
	if updates.Status == nil {
		return nil
	}

	status := *updates.Status
	if status == types.StatusCompleted {
		if todo.ProviderState != "closed" {
			if _, err := p.run(ctx, "issue", "close", id, "--json"); err != nil {
				return err
			}
		}
		todo.ProviderState = "closed"
	} else {
		if todo.ProviderState == "closed" {
			if _, err := p.run(ctx, "issue", "reopen", id, "--json"); err != nil {
				return err
			}
		}
		todo.ProviderState = "open"
	}

	for _, label := range existingStatusLabels(todo.Labels) {
		if status != types.StatusCompleted && label == statusLabel(status) {
			continue
		}
		if _, err := p.run(ctx, "issue", "label", "remove", id, "--label", label, "--json"); err != nil {
			return err
		}
		todo.Labels = removeLabel(todo.Labels, label)
	}

	if status != types.StatusCompleted {
		label := statusLabel(status)
		if !hasLabel(todo.Labels, label) {
			if _, err := p.run(ctx, "issue", "label", "add", id, "--label", label, "--json"); err != nil {
				return err
			}
			todo.Labels = append(todo.Labels, label)
			sort.Strings(todo.Labels)
		}
	}

	return nil
}

// applyPriority swaps the issue's priority:* label to the requested priority,
// keeping the in-memory todo's Labels/Priority in sync.
func (p *GriteProvider) applyPriority(ctx context.Context, id string, todo *types.TODO, priority types.Priority) error {
	want := priorityLabel(priority)
	for _, label := range existingPriorityLabels(todo.Labels) {
		if label == want {
			continue
		}
		if _, err := p.run(ctx, "issue", "label", "remove", id, "--label", label, "--json"); err != nil {
			return err
		}
		todo.Labels = removeLabel(todo.Labels, label)
	}
	if !hasLabel(todo.Labels, want) {
		if _, err := p.run(ctx, "issue", "label", "add", id, "--label", want, "--json"); err != nil {
			return err
		}
		todo.Labels = append(todo.Labels, want)
		sort.Strings(todo.Labels)
	}
	todo.Priority = priority
	return nil
}

// applySessionLabel swaps the issue's session:* label to the current run's
// session id, keeping the in-memory todo's Labels in sync. This records which
// agent session worked on the issue so the run can be located and resumed.
func (p *GriteProvider) applySessionLabel(ctx context.Context, id string, todo *types.TODO, sessionID string) error {
	want := sessionLabel(sessionID)
	for _, label := range existingSessionLabels(todo.Labels) {
		if label == want {
			continue
		}
		if _, err := p.run(ctx, "issue", "label", "remove", id, "--label", label, "--json"); err != nil {
			return err
		}
		todo.Labels = removeLabel(todo.Labels, label)
	}
	if !hasLabel(todo.Labels, want) {
		if _, err := p.run(ctx, "issue", "label", "add", id, "--label", want, "--json"); err != nil {
			return err
		}
		todo.Labels = append(todo.Labels, want)
		sort.Strings(todo.Labels)
	}
	return nil
}

func (p *GriteProvider) UpdateLatestFailure(ctx context.Context, todo *types.TODO, result *types.TestResultInfo) error {
	if result == nil {
		return nil
	}
	body, _ := clicky.Format(result, clicky.FormatOptions{Markdown: true})
	return p.comment(ctx, todo, "## Latest Failure\n\n"+body)
}

func (p *GriteProvider) SaveAttempt(ctx context.Context, todo *types.TODO, result *ExecutionResult) error {
	if result == nil {
		return nil
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "## Attempt %d\n\n", todo.Attempts)
	fmt.Fprintf(&sb, "- **Status:** %s\n", result.statusString())
	fmt.Fprintf(&sb, "- **Date:** %s\n", time.Now().Format("2006-01-02 15:04"))
	fmt.Fprintf(&sb, "- **Model:** %s\n", result.ExecutorName)
	if result.Duration > 0 {
		fmt.Fprintf(&sb, "- **Duration:** %s\n", result.Duration.Round(time.Second))
	}
	if result.CostUSD > 0 {
		fmt.Fprintf(&sb, "- **Cost:** $%.4f\n", result.CostUSD)
	}
	if result.TokensUsed > 0 {
		fmt.Fprintf(&sb, "- **Tokens:** %d\n", result.TokensUsed)
	}
	if result.CommitSHA != "" {
		fmt.Fprintf(&sb, "- **Commit:** `%s`\n", result.CommitSHA)
	}
	if result.ErrorMessage != "" {
		fmt.Fprintf(&sb, "- **Error:** %s\n", result.ErrorMessage)
	}
	// The transcript is intentionally NOT written into the issue: it can be
	// re-parsed on demand from the agent's session log via the recorded
	// session:<id> label (see applySessionLabel), which the dashboard's session
	// tab follows live. Embedding the full transcript here only duplicated that
	// data and bloated every issue.
	if sid := sessionIDFromLabels(todo.Labels); sid != "" {
		fmt.Fprintf(&sb, "- **Session:** `%s`\n", sid)
	}
	return p.comment(ctx, todo, sb.String())
}

func (p *GriteProvider) comment(ctx context.Context, todo *types.TODO, body string) error {
	id := TODOReference(todo)
	if id == "" {
		return fmt.Errorf("missing grite issue id")
	}
	_, err := p.run(ctx, "issue", "comment", id, "--body", body, "--json")
	return err
}

func (p *GriteProvider) run(ctx context.Context, args ...string) ([]byte, error) {
	binary, err := p.ensureBinary(ctx)
	if err != nil {
		return nil, err
	}
	runner := p.Runner
	if runner == nil {
		runner = runGriteCommand
	}
	return runner(ctx, p.WorkDir, binary, args...)
}

func (p *GriteProvider) ensureBinary(ctx context.Context) (string, error) {
	if p.Binary != "" {
		return p.Binary, nil
	}
	if path, err := exec.LookPath("grite"); err == nil {
		p.Binary = path
		return path, nil
	}
	if p.BinDir == "" {
		p.BinDir = filepath.Join(p.WorkDir, ".gavel", "bin")
	}
	if p.AppDir == "" {
		p.AppDir = filepath.Join(p.WorkDir, ".gavel", "deps")
	}
	path, err := installGrite(ctx, p.BinDir, p.AppDir)
	if err != nil {
		return "", err
	}
	p.Binary = path
	return path, nil
}

func runGriteCommand(ctx context.Context, workDir, binary string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Dir = workDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return stdout.Bytes(), fmt.Errorf("grite %s: %s", strings.Join(args, " "), msg)
	}
	return stdout.Bytes(), nil
}

func installGrite(_ context.Context, binDir, appDir string) (string, error) {
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return "", err
	}
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", err
	}
	// "any" (not "latest"): accept whatever grite is already installed and avoid
	// pinning to the newest release, which forces a GitHub releases/latest lookup
	// that 403s under API rate limits. Any working grite is fine here.
	const griteVersion = "any"
	cfg := &depstypes.DepsConfig{
		Dependencies: map[string]string{"grite": griteVersion},
		Registry: map[string]depstypes.Package{
			"grite": griteDepsPackage(),
		},
	}
	inst := depsinstaller.NewWithConfig(
		cfg,
		depsinstaller.WithBinDir(binDir),
		depsinstaller.WithAppDir(appDir),
		depsinstaller.WithProgress(false),
	)
	if _, err := inst.InstallWithResult("grite", griteVersion, &task.Task{}); err != nil {
		return "", fmt.Errorf("install grite via deps: %w", err)
	}
	candidate := filepath.Join(binDir, executableName("grite"))
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}
	if path, err := exec.LookPath("grite"); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("grite installed but binary was not found in %s", binDir)
}

func griteDepsPackage() depstypes.Package {
	return depstypes.Package{
		Name:         "grite",
		Manager:      "github_release",
		Repo:         "neul-labs/grite",
		BinaryName:   "grite",
		ChecksumFile: "",
		AssetPatterns: map[string]string{
			"darwin-amd64": "grite-{{.version}}-x86_64-apple-darwin.tar.gz",
			"darwin-arm64": "grite-{{.version}}-aarch64-apple-darwin.tar.gz",
			"linux-amd64":  "grite-{{.version}}-x86_64-unknown-linux-gnu.tar.gz",
			"linux-arm64":  "grite-{{.version}}-aarch64-unknown-linux-gnu.tar.gz",
		},
	}
}

func executableName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func decodeGrite[T any](raw []byte) (T, error) {
	var zero T
	var env struct {
		OK    bool `json:"ok"`
		Data  T    `json:"data"`
		Error *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return zero, err
	}
	if !env.OK {
		if env.Error != nil {
			return zero, fmt.Errorf("grite %s: %s", env.Error.Code, env.Error.Message)
		}
		return zero, fmt.Errorf("grite command failed")
	}
	return env.Data, nil
}

type griteIssue struct {
	IssueID      string   `json:"issue_id"`
	Title        string   `json:"title"`
	State        string   `json:"state"`
	Labels       []string `json:"labels"`
	CreatedTS    int64    `json:"created_ts"`
	UpdatedTS    int64    `json:"updated_ts"`
	CommentCount int      `json:"comment_count"`
}

type griteIssueListData struct {
	Issues []griteIssue `json:"issues"`
	Total  int          `json:"total"`
}

type griteIssueShowData struct {
	Issue  griteIssue   `json:"issue"`
	Events []griteEvent `json:"events"`
}

type griteIssueCreateData struct {
	IssueID string      `json:"issue_id"`
	Issue   *griteIssue `json:"issue"`
}

func (d griteIssueCreateData) ID() string {
	if d.IssueID != "" {
		return d.IssueID
	}
	if d.Issue != nil {
		return d.Issue.IssueID
	}
	return ""
}

type griteEvent struct {
	EventID     string                     `json:"event_id"`
	IssueID     string                     `json:"issue_id"`
	Actor       string                     `json:"actor"`
	TimestampMS int64                      `json:"ts_unix_ms"`
	Parent      *string                    `json:"parent"`
	Kind        map[string]json.RawMessage `json:"kind"`
}

type griteEventPayload struct {
	Body  string `json:"body"`
	Title string `json:"title"`
	Label string `json:"label"`
}

func todoFromGriteIssue(issue griteIssue, workDir string) *types.TODO {
	return &types.TODO{
		FilePath:        "grite:" + issue.IssueID,
		ID:              issue.IssueID,
		ShortID:         shortGriteID(issue.IssueID),
		Provider:        ProviderGrite,
		ProviderState:   issue.State,
		Labels:          append([]string(nil), issue.Labels...),
		TODOFrontmatter: frontmatterFromGriteIssue(issue, workDir),
	}
}

func frontmatterFromGriteIssue(issue griteIssue, workDir string) types.TODOFrontmatter {
	fm := types.TODOFrontmatter{
		Title:    issue.Title,
		Priority: priorityFromLabels(issue.Labels),
		Status:   statusFromGriteIssue(issue.State, issue.Labels),
	}
	fm.CWD = workDir
	if issue.UpdatedTS > 0 {
		t := time.UnixMilli(issue.UpdatedTS)
		fm.LastRun = &t
	}
	if sid := sessionIDFromLabels(issue.Labels); sid != "" {
		fm.LLM = &types.LLM{SessionId: sid}
	}
	return fm
}

func applyGriteIdentity(todo *types.TODO, issue griteIssue) {
	todo.FilePath = "grite:" + issue.IssueID
	todo.ID = issue.IssueID
	todo.ShortID = shortGriteID(issue.IssueID)
	todo.Provider = ProviderGrite
	todo.ProviderState = issue.State
	todo.Labels = append([]string(nil), issue.Labels...)
	if todo.Title == "" {
		todo.Title = issue.Title
	}
	// Carry the recorded session id onto the todo so the dashboard can follow or
	// resume the run. Get() parses the issue body (which may overwrite the
	// label-derived default), so set it authoritatively here from the labels.
	if sid := sessionIDFromLabels(issue.Labels); sid != "" {
		if todo.LLM == nil {
			todo.LLM = &types.LLM{}
		}
		todo.LLM.SessionId = sid
	}
}

func bodyFromGriteEvents(events []griteEvent) string {
	var body string
	for _, event := range events {
		for _, name := range []string{"IssueCreated", "IssueUpdated"} {
			raw, ok := event.Kind[name]
			if !ok {
				continue
			}
			var payload griteEventPayload
			if err := json.Unmarshal(raw, &payload); err == nil && payload.Body != "" {
				body = payload.Body
			}
		}
	}
	return body
}

func providerEventsFromGriteEvents(events []griteEvent) []types.ProviderEvent {
	var out []types.ProviderEvent
	for _, event := range events {
		kindNames := make([]string, 0, len(event.Kind))
		for name := range event.Kind {
			kindNames = append(kindNames, name)
		}
		sort.Strings(kindNames)

		for _, name := range kindNames {
			var payload griteEventPayload
			_ = json.Unmarshal(event.Kind[name], &payload)
			providerEvent := types.ProviderEvent{
				ID:      event.EventID,
				ShortID: shortGriteID(event.EventID),
				Kind:    name,
				Actor:   event.Actor,
				Title:   payload.Title,
				Body:    payload.Body,
				Label:   payload.Label,
			}
			if event.TimestampMS > 0 {
				providerEvent.Timestamp = time.UnixMilli(event.TimestampMS)
			}
			out = append(out, providerEvent)
		}
	}
	return collapseLabelChanges(out)
}

// collapseLabelChanges merges an adjacent LabelRemoved/LabelAdded pair that share
// a label namespace (e.g. priority:medium removed + priority:high added) into a
// single LabelChanged event, so the timeline reads "priority:medium → priority:high"
// instead of two separate lines.
func collapseLabelChanges(events []types.ProviderEvent) []types.ProviderEvent {
	merged := make([]types.ProviderEvent, 0, len(events))
	for i := 0; i < len(events); i++ {
		if i+1 < len(events) {
			if changed, ok := labelChange(events[i], events[i+1]); ok {
				merged = append(merged, changed)
				i++
				continue
			}
		}
		merged = append(merged, events[i])
	}
	return merged
}

// labelChange returns the merged LabelChanged event for a removed/added pair in
// the same namespace, or ok=false when the two events are not such a pair.
func labelChange(first, second types.ProviderEvent) (types.ProviderEvent, bool) {
	removed, added := first, second
	if removed.Kind == "LabelAdded" && added.Kind == "LabelRemoved" {
		removed, added = second, first
	}
	if removed.Kind != "LabelRemoved" || added.Kind != "LabelAdded" {
		return types.ProviderEvent{}, false
	}
	if removed.Label == "" || added.Label == "" {
		return types.ProviderEvent{}, false
	}
	if labelNamespace(removed.Label) != labelNamespace(added.Label) {
		return types.ProviderEvent{}, false
	}
	changed := added
	changed.Kind = "LabelChanged"
	changed.OldLabel = removed.Label
	changed.NewLabel = added.Label
	changed.Label = ""
	return changed, true
}

// labelNamespace returns the portion of a label before the first ':' so that
// priority:medium and priority:high are recognized as the same namespace.
func labelNamespace(label string) string {
	if idx := strings.Index(label, ":"); idx >= 0 {
		return label[:idx]
	}
	return label
}

func shortGriteID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

func griteListStates(filters DiscoveryFilters) []string {
	if len(filters.IncludeStatuses) > 0 {
		onlyCompleted := true
		needsOpen := false
		for _, status := range filters.IncludeStatuses {
			if status != types.StatusCompleted {
				onlyCompleted = false
				needsOpen = true
			}
		}
		if onlyCompleted {
			return []string{"closed"}
		}
		if needsOpen && !includesStatus(filters.IncludeStatuses, types.StatusCompleted) {
			return []string{"open"}
		}
	}
	if includesStatus(filters.ExcludeStatuses, types.StatusCompleted) {
		return []string{"open"}
	}
	return []string{"open", "closed"}
}

func statusFromGriteIssue(state string, labels []string) types.Status {
	if state == "closed" {
		return types.StatusCompleted
	}
	for _, label := range labels {
		if !strings.HasPrefix(label, statusLabelPrefix) {
			continue
		}
		status := types.Status(strings.TrimPrefix(label, statusLabelPrefix))
		if isKnownStatus(status) {
			return status
		}
	}
	return types.StatusPending
}

func priorityFromLabels(labels []string) types.Priority {
	for _, label := range labels {
		if !strings.HasPrefix(label, priorityLabelPrefix) {
			continue
		}
		priority := types.Priority(strings.TrimPrefix(label, priorityLabelPrefix))
		switch priority {
		case types.PriorityHigh, types.PriorityMedium, types.PriorityLow:
			return priority
		}
	}
	return types.PriorityMedium
}

func priorityLabel(priority types.Priority) string {
	return priorityLabelPrefix + string(priority)
}

func existingPriorityLabels(labels []string) []string {
	var out []string
	for _, label := range labels {
		if strings.HasPrefix(label, priorityLabelPrefix) {
			out = append(out, label)
		}
	}
	return out
}

func griteCreateLabels(priority types.Priority, status types.Status) []string {
	labels := []string{}
	if priority != "" {
		labels = append(labels, priorityLabel(priority))
	}
	if status != "" && status != types.StatusCompleted {
		labels = append(labels, statusLabel(status))
	}
	return labels
}

func statusLabel(status types.Status) string {
	return statusLabelPrefix + string(status)
}

func existingStatusLabels(labels []string) []string {
	var out []string
	for _, label := range labels {
		if strings.HasPrefix(label, statusLabelPrefix) {
			out = append(out, label)
		}
	}
	return out
}

func sessionLabel(sessionID string) string {
	return sessionLabelPrefix + sessionID
}

func existingSessionLabels(labels []string) []string {
	var out []string
	for _, label := range labels {
		if strings.HasPrefix(label, sessionLabelPrefix) {
			out = append(out, label)
		}
	}
	return out
}

// sessionIDFromLabels returns the agent session id recorded on the issue via the
// session:<id> label, or "" when none is present. It lets the dashboard locate
// (and resume) the session that worked on the issue without storing the
// transcript in the issue itself.
func sessionIDFromLabels(labels []string) string {
	for _, label := range labels {
		if strings.HasPrefix(label, sessionLabelPrefix) {
			return strings.TrimPrefix(label, sessionLabelPrefix)
		}
	}
	return ""
}

func includesStatus(statuses []types.Status, want types.Status) bool {
	for _, status := range statuses {
		if status == want {
			return true
		}
	}
	return false
}

func isKnownStatus(status types.Status) bool {
	return types.IsKnownStatus(status)
}

func hasLabel(labels []string, want string) bool {
	for _, label := range labels {
		if label == want {
			return true
		}
	}
	return false
}

func removeLabel(labels []string, remove string) []string {
	out := labels[:0]
	for _, label := range labels {
		if label != remove {
			out = append(out, label)
		}
	}
	return out
}
