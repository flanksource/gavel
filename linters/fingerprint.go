package linters

import (
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/flanksource/gavel/models"
)

// AssignFingerprints sets each violation's Fingerprint: a hash of the linter,
// rule, anchor and whitespace-normalised message, without the line, so it
// survives code shifting above the finding. The anchor is the most specific of
// the violation's uir symbol, declaration witness, markdown section (within its
// file) or file. Repeats of one fingerprint are numbered in line order.
func AssignFingerprints(result *LinterResult) {
	if result == nil {
		return
	}
	order := make([]int, len(result.Violations))
	for i := range order {
		order[i] = i
	}
	slices.SortStableFunc(order, func(a, b int) int {
		left, right := result.Violations[a], result.Violations[b]
		return cmp.Or(cmp.Compare(left.Line, right.Line), cmp.Compare(left.Column, right.Column))
	})
	seen := map[string]int{}
	for _, i := range order {
		v := &result.Violations[i]
		base := fingerprintBase(result, *v)
		if n := seen[base]; n > 0 {
			v.Fingerprint = fmt.Sprintf("%s-%d", base, n)
		} else {
			v.Fingerprint = base
		}
		seen[base]++
	}
}

func fingerprintBase(result *LinterResult, v models.Violation) string {
	rule := ""
	if v.Rule != nil {
		rule = v.Rule.Method
	}
	message := ""
	if v.Message != nil {
		message = strings.Join(strings.Fields(*v.Message), " ")
	}
	sum := sha256.Sum256([]byte(strings.Join([]string{cmp.Or(result.Linter, v.Source), rule, violationAnchor(result.WorkDir, v), message}, "\x00")))
	return hex.EncodeToString(sum[:8])
}

func violationAnchor(workDir string, v models.Violation) string {
	file := filepath.ToSlash(v.File)
	if workDir != "" && filepath.IsAbs(v.File) {
		if rel, err := filepath.Rel(workDir, v.File); err == nil {
			file = filepath.ToSlash(rel)
		}
	}
	switch anchor := v.Anchor; {
	case anchor != nil && anchor.Symbol != "":
		return "symbol:" + anchor.Symbol
	case anchor != nil && anchor.Declaration != "":
		return "decl:" + anchor.Declaration
	case anchor != nil && anchor.Section != "":
		return "doc:" + file + "#" + anchor.Section
	default:
		return "file:" + file
	}
}
