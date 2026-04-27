package status

import (
	"fmt"
	"sort"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/repomap"
)

// Starship git_status symbols (https://starship.rs/config/#git-status).
const (
	symBranch     = ""
	symStaged     = "+"
	symModified   = "!"
	symDeleted    = "✘"
	symRenamed    = "»"
	symCopied     = "»"
	symUntracked  = "?"
	symConflicted = "="
	symTypeChange = "T"
)

const (
	styleBranch     = "text-purple-600 font-bold"
	styleStaged     = "text-green-500 font-bold"
	styleModified   = "text-yellow-500 font-bold"
	styleDeleted    = "text-red-500 font-bold"
	styleRenamed    = "text-blue-500 font-bold"
	styleConflicted = "text-red-600 font-bold underline"
	styleUntracked  = "text-purple-500 font-bold"
	styleMuted      = "text-muted"
	styleScope      = "text-cyan-600"
	styleLanguage   = "text-muted italic"
	styleError      = "text-red-500"
	styleRunning    = "text-blue-500"
)

func (r *Result) Pretty() api.Text {
	t := clicky.Text("")
	if r == nil {
		return t
	}

	t = t.Add(r.prettyHeader()).NewLine()

	if len(r.Files) == 0 {
		return t.Append("clean", styleMuted).NewLine()
	}

	pathWidth := longestPath(r.Files)
	for i, group := range groupFilesByScopeKey(r.Files) {
		if i > 0 {
			t = t.NewLine()
		}
		t = t.Add(prettyScopeHeader(group.label)).NewLine()
		for _, f := range group.files {
			t = t.Add(f.prettyRow(pathWidth)).NewLine()
		}
	}
	return t
}

func (r *Result) prettyHeader() api.Text {
	counts := r.Counts()
	t := clicky.Text("")
	if r.Branch != "" {
		t = t.Append(symBranch, styleBranch).Space().Append(r.Branch, styleBranch)
	}
	t = t.Space().Append("[", styleMuted)
	first := true
	writeCount := func(sym, style string, n int) {
		if n == 0 {
			return
		}
		if !first {
			t = t.Append(" ", "")
		}
		t = t.Append(sym, style).Append(fmt.Sprintf("%d", n), styleMuted)
		first = false
	}
	writeCount(symStaged, styleStaged, counts.Staged+counts.Both)
	writeCount(symModified, styleModified, counts.Unstaged+counts.Both)
	writeCount(symUntracked, styleUntracked, counts.Untracked)
	writeCount(symConflicted, styleConflicted, counts.Conflict)
	if first {
		t = t.Append("clean", styleMuted)
	}
	t = t.Append("]", styleMuted)
	if counts.Adds > 0 || counts.Dels > 0 {
		t = t.Space().
			Append(fmt.Sprintf("+%d", counts.Adds), styleStaged).
			Space().
			Append(fmt.Sprintf("-%d", counts.Dels), styleDeleted)
	}
	if r.ResultsStale && r.ResultsSHA != "" {
		t = t.Space().
			Append("· stale results (sha ", styleMuted).
			Append(shortSHA(r.ResultsSHA), styleModified).
			Append(")", styleMuted)
	}
	return t
}

func shortSHA(s string) string {
	if len(s) > 8 {
		return s[:8]
	}
	return s
}

func (f FileStatus) prettyRow(pathWidth int) api.Text {
	t := clicky.Text("  ")

	sym, style := symbolFor(f)
	t = t.Append(sym, style).Space()

	path := f.Path
	if f.PreviousPath != "" {
		path = fmt.Sprintf("%s → %s", f.PreviousPath, f.Path)
	}
	pathPad := max(pathWidth-runeLen(path), 0)
	t = t.Append(path, "").Append(strings.Repeat(" ", pathPad+2), "")

	t = t.Add(prettyLineDelta(f)).Space()
	t = t.Add(f.prettyEnrichment())
	return t
}

func prettyLineDelta(f FileStatus) api.Text {
	t := clicky.Text("")
	if f.Adds == 0 && f.Dels == 0 {
		return t.Append("        ", "")
	}
	left := fmt.Sprintf("+%d", f.Adds)
	right := fmt.Sprintf("-%d", f.Dels)
	// Pad the combined delta to a fixed width so columns line up.
	combinedLen := runeLen(left) + 1 + runeLen(right)
	pad := max(8-combinedLen, 0)
	t = t.Append(left, styleStaged).Space().Append(right, styleDeleted)
	if pad > 0 {
		t = t.Append(strings.Repeat(" ", pad), "")
	}
	return t
}

func (f FileStatus) prettyEnrichment() api.Text {
	t := clicky.Text("")

	if f.RepomapError != nil {
		return appendAISummaryState(t.Append("repomap error: "+f.RepomapError.Error(), styleError), f)
	}

	if f.State == StateConflict {
		t = t.Append("⚠ conflict", styleConflicted)
		return appendAISummaryState(t, f)
	}
	if f.State == StateUntracked && f.FileMap == nil {
		t = t.Append("(untracked)", styleMuted)
		return appendAISummaryState(t, f)
	}
	if f.WorkKind == KindDeleted || f.StagedKind == KindDeleted {
		t = t.Append("(deleted)", styleMuted)
		if f.FileMap == nil {
			return appendAISummaryState(t, f)
		}
		t = t.Space()
	}

	if f.FileMap == nil {
		return appendAISummaryState(t.Add(prettyTestLintBadges(f)), f)
	}

	k8s := len(f.FileMap.KubernetesRefs)
	viols := len(f.FileMap.Violations)
	if k8s > 0 || viols > 0 {
		t = t.Append("  ", "")
		if k8s > 0 {
			t = t.Append(fmt.Sprintf("k8s:%d", k8s), styleMuted)
		}
		if viols > 0 {
			if k8s > 0 {
				t = t.Space()
			}
			t = t.Append(fmt.Sprintf("viol:%d", viols), styleDeleted)
		}
	}

	t = t.Add(prettyTestLintBadges(f))

	return appendAISummaryState(t, f)
}

func appendAISummaryState(t api.Text, f FileStatus) api.Text {
	summary := prettyAISummary(f)
	if summary.IsEmpty() {
		return t
	}
	if !t.IsEmpty() {
		t = t.Append("  ", "")
	}
	return t.Add(summary)
}

func prettyAISummary(f FileStatus) api.Text {
	summary := strings.TrimSpace(f.AISummary)
	if summary != "" {
		return clicky.Text(summary, styleMuted)
	}

	switch f.AIStatus {
	case AISummaryStatusPending:
		return clicky.Text("⏳ ai", styleMuted)
	case AISummaryStatusRunning:
		return clicky.Text("⟳ ai", styleRunning)
	case AISummaryStatusFailed:
		return clicky.Text("⚠ ai summary failed", styleError)
	default:
		return clicky.Text("")
	}
}

type scopeGroup struct {
	key   scopeGroupKey
	label string
	files []FileStatus
}

type scopeGroupKey struct {
	label      string
	hasNonTest bool
	hasTest    bool
}

func groupFilesByScopeKey(files []FileStatus) []scopeGroup {
	if len(files) == 0 {
		return nil
	}

	groupsByScope := make(map[scopeGroupKey][]FileStatus)
	for _, file := range files {
		key := scopeKey(file)
		groupsByScope[key] = append(groupsByScope[key], file)
	}

	keys := make([]scopeGroupKey, 0, len(groupsByScope))
	for key := range groupsByScope {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return compareScopeKeys(keys[i], keys[j]) < 0
	})

	groups := make([]scopeGroup, 0, len(keys))
	for _, key := range keys {
		groups = append(groups, scopeGroup{
			key:   key,
			label: key.label,
			files: groupsByScope[key],
		})
	}
	return groups
}

func scopeKey(f FileStatus) scopeGroupKey {
	if f.FileMap == nil {
		return scopeGroupKey{label: string(repomap.ScopeTypeGeneral)}
	}

	parts := make([]string, 0, len(f.FileMap.Scopes)+1)
	seen := map[string]struct{}{}
	if language := strings.TrimSpace(f.FileMap.Language); language != "" {
		parts = append(parts, language)
		seen[strings.ToLower(language)] = struct{}{}
	}

	var (
		nonTest []string
		tests   []string
	)
	for _, scope := range f.FileMap.Scopes {
		label := strings.TrimSpace(string(scope))
		if label == "" {
			continue
		}
		key := strings.ToLower(label)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		if scope == repomap.ScopeTypeTest {
			tests = append(tests, label)
			continue
		}
		nonTest = append(nonTest, label)
	}

	parts = append(parts, nonTest...)
	parts = append(parts, tests...)
	if len(parts) == 0 {
		parts = append(parts, string(repomap.ScopeTypeGeneral))
	}

	return scopeGroupKey{
		label:      strings.Join(parts, " · "),
		hasNonTest: len(nonTest) > 0 || (strings.TrimSpace(f.FileMap.Language) != "" && len(tests) == 0),
		hasTest:    len(tests) > 0,
	}
}

func compareScopeKeys(a, b scopeGroupKey) int {
	aRank := scopeSortRank(a)
	bRank := scopeSortRank(b)
	if aRank != bRank {
		return aRank - bRank
	}
	switch {
	case a.label < b.label:
		return -1
	case a.label > b.label:
		return 1
	default:
		return 0
	}
}

func scopeSortRank(key scopeGroupKey) int {
	switch {
	case key.hasTest && !key.hasNonTest:
		return 2
	case !key.hasNonTest && !key.hasTest:
		return 1
	default:
		return 0
	}
}

func prettyScopeHeader(label string) api.Text {
	return clicky.Text(" ").Append(label, "font-bold "+styleScope)
}

func symbolFor(f FileStatus) (string, string) {
	switch f.State {
	case StateConflict:
		return symConflicted, styleConflicted
	case StateUntracked:
		return symUntracked, styleUntracked
	}

	kind := f.StagedKind
	if kind == KindUnknown {
		kind = f.WorkKind
	}
	switch kind {
	case KindAdded:
		return symStaged, styleStaged
	case KindDeleted:
		return symDeleted, styleDeleted
	case KindRenamed:
		return symRenamed, styleRenamed
	case KindCopied:
		return symCopied, styleRenamed
	case KindTypeChange:
		return symTypeChange, styleModified
	}

	if f.State == StateStaged {
		return symStaged, styleStaged
	}
	return symModified, styleModified
}

func prettyTestLintBadges(f FileStatus) api.Text {
	t := clicky.Text("")
	ts := f.TestStatus
	ls := f.LintStatus

	if ts.Failed == 0 && ts.Passed == 0 && ts.Skipped == 0 &&
		ls.Errors == 0 && ls.Warnings == 0 && ls.Infos == 0 {
		return t
	}

	t = t.Append("  ", "")

	switch {
	case ts.Failed > 0:
		t = t.Append(fmt.Sprintf("⚠ fail:%d", ts.Failed), styleDeleted)
	case ts.Passed > 0:
		t = t.Append(fmt.Sprintf("✓ %d", ts.Passed), styleStaged)
	case ts.Skipped > 0:
		t = t.Append(fmt.Sprintf("~ skip:%d", ts.Skipped), styleMuted)
	}

	if ts.Failed+ts.Passed+ts.Skipped > 0 && ls.Errors+ls.Warnings+ls.Infos > 0 {
		t = t.Space()
	}

	switch {
	case ls.Errors > 0:
		t = t.Append(fmt.Sprintf("⚠ err:%d", ls.Errors), styleDeleted)
	case ls.Warnings > 0:
		t = t.Append(fmt.Sprintf("⚠ warn:%d", ls.Warnings), styleModified)
	case ls.Infos > 0:
		t = t.Append(fmt.Sprintf("· info:%d", ls.Infos), styleMuted)
	}

	return t
}

func containsScope(scopes []string, lang string) bool {
	for _, s := range scopes {
		if strings.EqualFold(s, lang) {
			return true
		}
	}
	return false
}

func longestPath(files []FileStatus) int {
	maxLen := 0
	for _, f := range files {
		p := f.Path
		if f.PreviousPath != "" {
			p = fmt.Sprintf("%s → %s", f.PreviousPath, f.Path)
		}
		if n := runeLen(p); n > maxLen {
			maxLen = n
		}
	}
	return maxLen
}

func runeLen(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}
