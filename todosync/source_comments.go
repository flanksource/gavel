// Package todosync synchronizes source TODO and FIXME comments into native
// PostgreSQL-backed TODOs.
package todosync

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/utils"
)

const (
	SourceTodoLabel         = "source:todo"
	sourceTodoIDLabelPrefix = "source-id:"
	sourceTodoMarkerPrefix  = "source-marker:"
	sourceTodoKind          = "code-comment"
)

var defaultSourceCommentMarkers = []string{"TODO", "FIXME"}

var defaultSourceCommentIgnores = []string{
	".git",
	".gavel",
	".todos",
	".next",
	"build",
	"coverage",
	"dist",
	"node_modules",
	"target",
	"tmp",
	"vendor",
}

type SourceCommentSyncOptions struct {
	WorkDir string
	Paths   []string
	Markers []string
	Ignore  []string
	DryRun  bool
}

type SourceComment struct {
	ID         string `json:"id"`
	Marker     string `json:"marker"`
	Message    string `json:"message,omitempty"`
	Path       string `json:"path"`
	Line       int    `json:"line"`
	Column     int    `json:"column"`
	Raw        string `json:"raw"`
	Occurrence int    `json:"occurrence"`
}

func (c SourceComment) PathRef() string {
	return types.PathRef{File: c.Path, Line: c.Line}.String()
}

type SourceTodoChange struct {
	ID    string `json:"id"`
	Ref   string `json:"ref,omitempty"`
	Title string `json:"title"`
	Path  string `json:"path"`
}

type SourceCommentSyncResult struct {
	DryRun       bool               `json:"dry_run,omitempty"`
	ScannedFiles int                `json:"scanned_files"`
	Matches      int                `json:"matches"`
	Created      []SourceTodoChange `json:"created,omitempty"`
	Updated      []SourceTodoChange `json:"updated,omitempty"`
	Completed    []SourceTodoChange `json:"completed,omitempty"`
	Unchanged    []SourceTodoChange `json:"unchanged,omitempty"`
}

func ScanSourceComments(opts SourceCommentSyncOptions) ([]SourceComment, int, error) {
	workDir, err := sourceWorkDir(opts.WorkDir)
	if err != nil {
		return nil, 0, err
	}
	markers := normalizeMarkers(opts.Markers)
	if len(markers) == 0 {
		return nil, 0, nil
	}
	pattern := markerPattern(markers)
	paths := opts.Paths
	if len(paths) == 0 {
		paths = []string{"."}
	}

	ignore := append([]string(nil), defaultSourceCommentIgnores...)
	ignore = append(ignore, opts.Ignore...)
	occurrences := map[string]int{}
	var comments []SourceComment
	scanned := 0

	for _, arg := range paths {
		root := arg
		if !filepath.IsAbs(root) {
			root = filepath.Join(workDir, root)
		}
		// Bounded + gitignore-aware: a scratch worktree or vendored checkout
		// under workDir would sync a duplicate TODO for every source marker.
		if err := utils.WalkGitIgnoredBounded(root, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return nil
			}
			rel, err := filepath.Rel(workDir, path)
			if err != nil {
				return nil
			}
			rel = filepath.ToSlash(filepath.Clean(rel))
			if rel == "." {
				rel = entry.Name()
			}
			if entry.IsDir() {
				if path != root && shouldIgnoreSourcePath(rel, entry.Name(), ignore) {
					return filepath.SkipDir
				}
				return nil
			}
			if shouldIgnoreSourcePath(rel, entry.Name(), ignore) {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil || bytes.IndexByte(data, 0) >= 0 {
				return nil
			}
			scanned++
			fileComments := extractSourceComments(string(data), rel, pattern, occurrences)
			comments = append(comments, fileComments...)
			return nil
		}); err != nil {
			return nil, scanned, err
		}
	}

	sort.Slice(comments, func(i, j int) bool {
		if comments[i].Path != comments[j].Path {
			return comments[i].Path < comments[j].Path
		}
		if comments[i].Line != comments[j].Line {
			return comments[i].Line < comments[j].Line
		}
		return comments[i].Column < comments[j].Column
	})
	return comments, scanned, nil
}

func SyncSourceComments(ctx context.Context, provider todos.Provider, opts SourceCommentSyncOptions) (*SourceCommentSyncResult, error) {
	comments, scanned, err := ScanSourceComments(opts)
	if err != nil {
		return nil, err
	}
	result := &SourceCommentSyncResult{
		DryRun:       opts.DryRun,
		ScannedFiles: scanned,
		Matches:      len(comments),
	}

	existing, err := provider.List(ctx, todos.DiscoveryFilters{IncludeLabels: []string{SourceTodoLabel}})
	if err != nil {
		return nil, fmt.Errorf("list generated source todos: %w", err)
	}
	existingByID := map[string]*types.TODO{}
	for _, todo := range existing {
		if id := sourceTodoID(todo); id != "" {
			existingByID[id] = todo
		}
	}

	seen := map[string]bool{}
	for _, comment := range comments {
		seen[comment.ID] = true
		change := sourceTodoChange(comment, nil)
		existingTodo := existingByID[comment.ID]
		if existingTodo == nil {
			if !opts.DryRun {
				created, err := provider.Create(ctx, sourceCreateRequest(comment))
				if err != nil {
					return nil, fmt.Errorf("create source todo %s: %w", comment.PathRef(), err)
				}
				change = sourceTodoChange(comment, created)
			}
			result.Created = append(result.Created, change)
			continue
		}

		detail, err := provider.Get(ctx, todos.TODOReference(existingTodo))
		if err != nil {
			return nil, fmt.Errorf("load generated source todo %s: %w", todos.TODOReference(existingTodo), err)
		}
		if detail.ProviderState == "" {
			detail.ProviderState = existingTodo.ProviderState
		}
		if len(detail.Labels) == 0 {
			detail.Labels = existingTodo.Labels
		}
		changed, err := syncExistingSourceTodo(ctx, provider, detail, comment, opts.DryRun)
		if err != nil {
			return nil, err
		}
		change = sourceTodoChange(comment, detail)
		if changed {
			result.Updated = append(result.Updated, change)
		} else {
			result.Unchanged = append(result.Unchanged, change)
		}
	}

	for id, existingTodo := range existingByID {
		if seen[id] {
			continue
		}
		detail, err := provider.Get(ctx, todos.TODOReference(existingTodo))
		if err != nil {
			return nil, fmt.Errorf("load stale source todo %s: %w", todos.TODOReference(existingTodo), err)
		}
		if detail.Status == types.StatusCompleted {
			continue
		}
		change := sourceTodoChange(sourceCommentFromTODO(id, detail), detail)
		if !opts.DryRun {
			completed := types.StatusCompleted
			if err := provider.UpdateState(ctx, detail, todos.StateUpdate{Status: &completed}); err != nil {
				return nil, fmt.Errorf("complete stale source todo %s: %w", todos.TODOReference(detail), err)
			}
		}
		result.Completed = append(result.Completed, change)
	}

	return result, nil
}

func syncExistingSourceTodo(ctx context.Context, provider todos.Provider, todo *types.TODO, comment SourceComment, dryRun bool) (bool, error) {
	wantTitle := sourceTodoTitle(comment)
	wantBody := sourceTodoBody(comment)
	path := types.StringOrSlice{comment.PathRef()}
	sourceLabels := sourceTodoLabels(comment)
	edit := todos.EditRequest{
		Labels:   &sourceLabels,
		Metadata: sourceTodoMetadata(comment),
	}
	needsEdit := false
	if todo.Title != wantTitle {
		edit.Title = &wantTitle
		needsEdit = true
	}
	if strings.TrimSpace(todo.MarkdownBody) != strings.TrimSpace(wantBody) {
		edit.Body = &wantBody
		needsEdit = true
	}
	if !sameStringOrSlice(todo.Path, path) {
		edit.Path = &path
		needsEdit = true
	}
	if needsEdit && !dryRun {
		if err := provider.Edit(ctx, todo, edit); err != nil {
			return false, fmt.Errorf("update source todo %s: %w", todos.TODOReference(todo), err)
		}
	}
	needsState := todo.Status == types.StatusCompleted
	if needsState && !dryRun {
		pending := types.StatusPending
		if err := provider.UpdateState(ctx, todo, todos.StateUpdate{Status: &pending}); err != nil {
			return false, fmt.Errorf("reopen source todo %s: %w", todos.TODOReference(todo), err)
		}
	}
	return needsEdit || needsState, nil
}

func sourceCreateRequest(comment SourceComment) todos.CreateRequest {
	return todos.CreateRequest{
		Title:    sourceTodoTitle(comment),
		Body:     sourceTodoBody(comment),
		Priority: types.PriorityMedium,
		Status:   types.StatusPending,
		Path:     types.StringOrSlice{comment.PathRef()},
		Labels:   sourceTodoLabels(comment),
		Metadata: sourceTodoMetadata(comment),
	}
}

func sourceTodoTitle(comment SourceComment) string {
	message := strings.TrimSpace(comment.Message)
	if message == "" {
		message = comment.PathRef()
	}
	if len(message) > 120 {
		message = strings.TrimSpace(message[:117]) + "..."
	}
	return fmt.Sprintf("%s: %s", strings.ToUpper(comment.Marker), message)
}

func sourceTodoBody(comment SourceComment) string {
	marker := strings.ToUpper(comment.Marker)
	payload := map[string]any{
		"id":      comment.ID,
		"path":    comment.Path,
		"line":    comment.Line,
		"column":  comment.Column,
		"marker":  marker,
		"message": comment.Message,
	}
	encoded, _ := json.Marshal(payload)

	var sb strings.Builder
	sb.WriteString("Generated from a source comment.\n\n")
	fmt.Fprintf(&sb, "- Source: `%s`\n", comment.PathRef())
	fmt.Fprintf(&sb, "- Marker: `%s`\n", marker)
	if strings.TrimSpace(comment.Message) != "" {
		fmt.Fprintf(&sb, "- Message: %s\n", comment.Message)
	}
	if strings.TrimSpace(comment.Raw) != "" {
		sb.WriteString("\n```text\n")
		sb.WriteString(comment.Raw)
		sb.WriteString("\n```\n")
	}
	sb.WriteString("\n<!-- gavel:source-todo ")
	sb.Write(encoded)
	sb.WriteString(" -->\n")
	return sb.String()
}

func sourceTodoLabels(comment SourceComment) []string {
	return []string{
		SourceTodoLabel,
		sourceTodoIDLabelPrefix + comment.ID,
		sourceTodoMarkerPrefix + strings.ToLower(comment.Marker),
	}
}

func sourceTodoMetadata(comment SourceComment) map[string]any {
	return map[string]any{
		"source":        sourceTodoKind,
		"source_id":     comment.ID,
		"source_marker": strings.ToLower(comment.Marker),
	}
}

func sourceTodoChange(comment SourceComment, todo *types.TODO) SourceTodoChange {
	change := SourceTodoChange{
		ID:    comment.ID,
		Title: sourceTodoTitle(comment),
		Path:  comment.PathRef(),
	}
	if todo != nil {
		change.Ref = todos.TODOReference(todo)
		if todo.Title != "" {
			change.Title = todo.Title
		}
		if len(todo.Path) > 0 {
			change.Path = todo.Path[0]
		}
	}
	return change
}

func sourceCommentFromTODO(id string, todo *types.TODO) SourceComment {
	comment := SourceComment{ID: id, Marker: "TODO"}
	if todo != nil {
		comment.Message = strings.TrimPrefix(todo.Title, "TODO: ")
		if marker := sourceTodoMarker(todo); marker != "" {
			comment.Marker = strings.ToUpper(marker)
		}
		if len(todo.Path) > 0 {
			ref := types.ParsePathRef(todo.Path[0])
			comment.Path = ref.File
			comment.Line = ref.Line
		}
	}
	return comment
}

func sourceTodoID(todo *types.TODO) string {
	for _, label := range labelsForSourceTodo(todo) {
		if id, ok := strings.CutPrefix(label, sourceTodoIDLabelPrefix); ok {
			return id
		}
	}
	if todo != nil && todo.Metadata != nil {
		if id, ok := todo.Metadata["source_id"].(string); ok {
			return id
		}
	}
	return ""
}

func sourceTodoMarker(todo *types.TODO) string {
	for _, label := range labelsForSourceTodo(todo) {
		if marker, ok := strings.CutPrefix(label, sourceTodoMarkerPrefix); ok {
			return marker
		}
	}
	if todo != nil && todo.Metadata != nil {
		if marker, ok := todo.Metadata["source_marker"].(string); ok {
			return marker
		}
	}
	return ""
}

func labelsForSourceTodo(todo *types.TODO) []string {
	if todo == nil {
		return nil
	}
	labels := append([]string(nil), todo.Labels...)
	if todo.Metadata == nil {
		return labels
	}
	switch value := todo.Metadata["labels"].(type) {
	case string:
		labels = append(labels, value)
	case []string:
		labels = append(labels, value...)
	case []any:
		for _, item := range value {
			if label, ok := item.(string); ok {
				labels = append(labels, label)
			}
		}
	}
	return labels
}

func normalizeMarkers(markers []string) []string {
	if len(markers) == 0 {
		markers = defaultSourceCommentMarkers
	}
	seen := map[string]bool{}
	var out []string
	for _, marker := range markers {
		marker = strings.ToUpper(strings.TrimSpace(marker))
		if marker == "" || seen[marker] {
			continue
		}
		seen[marker] = true
		out = append(out, marker)
	}
	return out
}

func markerPattern(markers []string) *regexp.Regexp {
	escaped := make([]string, len(markers))
	for i, marker := range markers {
		escaped[i] = regexp.QuoteMeta(marker)
	}
	return regexp.MustCompile(`(?i)\b(` + strings.Join(escaped, "|") + `)\b(?:\(([^)]*)\))?\s*:?(.*)$`)
}

type commentSegment struct {
	text        string
	columnStart int
}

func extractSourceComments(source, path string, pattern *regexp.Regexp, occurrences map[string]int) []SourceComment {
	var comments []SourceComment
	lines := strings.Split(source, "\n")
	for index, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		for _, segment := range sourceCommentSegments(line) {
			found := pattern.FindStringSubmatchIndex(segment.text)
			if found == nil {
				continue
			}
			marker := strings.ToUpper(segment.text[found[2]:found[3]])
			assignee := ""
			if found[4] >= 0 {
				assignee = strings.TrimSpace(segment.text[found[4]:found[5]])
			}
			message := cleanSourceCommentMessage(segment.text[found[6]:found[7]])
			if assignee != "" {
				message = strings.TrimSpace("(" + assignee + ") " + message)
			}
			key := strings.Join([]string{path, marker, normalizeSourceMessage(message)}, "\x00")
			occurrences[key]++
			occurrence := occurrences[key]
			lineNumber := index + 1
			column := segment.columnStart + found[2] + 1
			comments = append(comments, SourceComment{
				ID:         sourceCommentID(path, marker, message, occurrence),
				Marker:     marker,
				Message:    message,
				Path:       path,
				Line:       lineNumber,
				Column:     column,
				Raw:        strings.TrimSpace(line),
				Occurrence: occurrence,
			})
		}
	}
	return comments
}

func sourceCommentSegments(line string) []commentSegment {
	start := findSourceCommentStart(line)
	if start >= 0 {
		return []commentSegment{{text: line[start:], columnStart: start}}
	}
	trimmed := strings.TrimLeft(line, " \t")
	offset := len(line) - len(trimmed)
	if strings.HasPrefix(trimmed, "*") {
		return []commentSegment{{text: trimmed[1:], columnStart: offset + 1}}
	}
	return nil
}

func findSourceCommentStart(line string) int {
	var quote rune
	for i, r := range line {
		if quote != 0 {
			if r == quote && !sourceCharEscaped(line, i) {
				quote = 0
			}
			continue
		}
		switch r {
		case '\'', '"', '`':
			quote = r
			continue
		case '<':
			if strings.HasPrefix(line[i:], "<!--") {
				return i + len("<!--")
			}
		case '/':
			if strings.HasPrefix(line[i:], "//") || strings.HasPrefix(line[i:], "/*") {
				return i + 2
			}
		case '#':
			if i == 0 || isSourceCommentSpace(line[i-1]) {
				return i + 1
			}
		case '-':
			if strings.HasPrefix(line[i:], "--") && (i == 0 || isSourceCommentSpace(line[i-1])) {
				return i + 2
			}
		case ';':
			if i == 0 || isSourceCommentSpace(line[i-1]) {
				return i + 1
			}
		}
	}
	return -1
}

func sourceCharEscaped(line string, index int) bool {
	slashes := 0
	for i := index - 1; i >= 0 && line[i] == '\\'; i-- {
		slashes++
	}
	return slashes%2 == 1
}

func isSourceCommentSpace(b byte) bool {
	return b == ' ' || b == '\t'
}

func cleanSourceCommentMessage(message string) string {
	message = strings.TrimSpace(message)
	for _, suffix := range []string{"*/", "-->"} {
		message = strings.TrimSpace(strings.TrimSuffix(message, suffix))
	}
	message = strings.TrimSpace(strings.TrimPrefix(message, "-"))
	message = strings.TrimSpace(strings.TrimPrefix(message, ":"))
	return message
}

func normalizeSourceMessage(message string) string {
	return strings.ToLower(strings.Join(strings.Fields(message), " "))
}

func sourceCommentID(path, marker, message string, occurrence int) string {
	key := fmt.Sprintf("%s\x00%s\x00%s\x00%d", path, strings.ToUpper(marker), normalizeSourceMessage(message), occurrence)
	hash := sha1.Sum([]byte(key))
	return hex.EncodeToString(hash[:])[:12]
}

func sourceWorkDir(workDir string) (string, error) {
	if workDir == "" {
		return os.Getwd()
	}
	return filepath.Abs(workDir)
}

func shouldIgnoreSourcePath(rel, base string, patterns []string) bool {
	rel = filepath.ToSlash(strings.TrimPrefix(rel, "./"))
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(filepath.ToSlash(pattern))
		if pattern == "" {
			continue
		}
		if pattern == base || pattern == rel {
			return true
		}
		if ok, _ := filepath.Match(pattern, base); ok {
			return true
		}
		if strings.Contains(pattern, "/") {
			if ok, _ := filepath.Match(pattern, rel); ok {
				return true
			}
		}
	}
	return false
}

func sameStringOrSlice(a, b types.StringOrSlice) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
