package status

import (
	"fmt"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
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
	for _, f := range r.Files {
		t = t.Add(f.prettyRow(pathWidth)).NewLine()
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
		return t.Append("repomap error: "+f.RepomapError.Error(), styleError)
	}

	if f.State == StateConflict {
		t = t.Append("⚠ conflict", styleConflicted)
		return t
	}
	if f.State == StateUntracked && f.FileMap == nil {
		t = t.Append("(untracked)", styleMuted)
		return t
	}
	if f.WorkKind == KindDeleted || f.StagedKind == KindDeleted {
		t = t.Append("(deleted)", styleMuted)
		if f.FileMap == nil {
			return t
		}
		t = t.Space()
	}

	if f.FileMap == nil {
		return t.Add(prettyTestLintBadges(f))
	}

	scopes := make([]string, 0, len(f.FileMap.Scopes))
	for _, s := range f.FileMap.Scopes {
		scopes = append(scopes, string(s))
	}
	if len(scopes) > 0 {
		t = t.Append("· ", styleMuted).Append(strings.Join(scopes, " · "), styleScope)
	}
	if f.FileMap.Language != "" && !containsScope(scopes, f.FileMap.Language) {
		t = t.Space().Append(f.FileMap.Language, styleLanguage)
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

	return t
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
