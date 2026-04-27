package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/flanksource/gavel/cmd/gavel/choose"
	"github.com/flanksource/gavel/internal/prompting"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/verify"
)

type violationType struct {
	Source  string
	Rule    string
	Count   int
	Files   []string
	Example string
}

func (v violationType) Label() string {
	if v.Rule != "" {
		return fmt.Sprintf("%s/%s  (%d violations, %d files)", v.Source, v.Rule, v.Count, len(v.Files))
	}
	return fmt.Sprintf("%s  (%d violations, %d files)", v.Source, v.Count, len(v.Files))
}

func collectViolationTypes(results []*linters.LinterResult) []violationType {
	type key struct{ source, rule string }
	index := map[key]*violationType{}
	var order []key

	for _, r := range results {
		if r == nil {
			continue
		}
		for _, v := range r.Violations {
			rule := ""
			if v.Rule != nil {
				rule = v.Rule.Method
			}
			k := key{source: v.Source, rule: rule}
			vt, ok := index[k]
			if !ok {
				vt = &violationType{Source: v.Source, Rule: rule}
				index[k] = vt
				order = append(order, k)
			}
			vt.Count++
			if vt.Example == "" && v.Message != nil {
				vt.Example = *v.Message
			}
			found := false
			for _, f := range vt.Files {
				if f == v.File {
					found = true
					break
				}
			}
			if !found {
				vt.Files = append(vt.Files, v.File)
			}
		}
	}

	result := make([]violationType, 0, len(order))
	for _, k := range order {
		result = append(result, *index[k])
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})
	return result
}

func runTriage(results []*linters.LinterResult, workDir string) ([]verify.LintIgnoreRule, error) {
	prompting.Prepare()

	types := collectViolationTypes(results)
	if len(types) == 0 {
		return nil, nil
	}

	items := make([]string, len(types))
	for i, vt := range types {
		items[i] = vt.Label()
	}

	detailFunc := func(i int) string {
		if i < 0 || i >= len(types) {
			return ""
		}
		return formatViolationDetail(types[i], workDir)
	}

	selected, err := choose.Run(items,
		choose.WithHeader("Select violation types to ignore:"),
		choose.WithLimit(0),
		choose.WithDetailFunc(detailFunc),
	)
	if err != nil {
		return nil, err
	}
	if len(selected) == 0 {
		return nil, nil
	}

	var rules []verify.LintIgnoreRule
	for _, idx := range selected {
		vt := types[idx]
		newRules, err := triageAction(vt, workDir)
		if err != nil {
			return rules, err
		}
		rules = append(rules, newRules...)
	}

	return rules, nil
}

func triageAction(vt violationType, workDir string) ([]verify.LintIgnoreRule, error) {
	var options []string
	ruleLabel := vt.Source
	if vt.Rule != "" {
		ruleLabel = vt.Source + "/" + vt.Rule
	}

	options = append(options, fmt.Sprintf("Ignore all '%s' violations everywhere", ruleLabel))

	relFiles := relPaths(vt.Files, workDir)
	if len(relFiles) <= 5 {
		for _, f := range relFiles {
			if vt.Rule != "" {
				options = append(options, fmt.Sprintf("Ignore '%s' in %s only", ruleLabel, f))
			}
		}
	}

	header := fmt.Sprintf("Action for %s (%d violations):", ruleLabel, vt.Count)
	indices, err := choose.Run(options,
		choose.WithHeader(header),
		choose.WithLimit(1),
	)
	if err != nil {
		return nil, err
	}
	if len(indices) == 0 {
		return nil, nil
	}

	choice := indices[0]
	if choice == 0 {
		r := verify.LintIgnoreRule{Source: vt.Source}
		if vt.Rule != "" {
			r.Rule = vt.Rule
		}
		return []verify.LintIgnoreRule{r}, nil
	}

	fileIdx := choice - 1
	if fileIdx < len(relFiles) {
		return []verify.LintIgnoreRule{{
			Rule:   vt.Rule,
			Source: vt.Source,
			File:   relFiles[fileIdx],
		}}, nil
	}

	return nil, nil
}

func formatViolationDetail(vt violationType, workDir string) string {
	var s strings.Builder
	fmt.Fprintf(&s, "Source:     %s\n", vt.Source)
	if vt.Rule != "" {
		fmt.Fprintf(&s, "Rule:       %s\n", vt.Rule)
	}
	fmt.Fprintf(&s, "Violations: %d\n", vt.Count)
	fmt.Fprintf(&s, "Files:      %d\n", len(vt.Files))

	if vt.Example != "" {
		fmt.Fprintf(&s, "\nExample: %s\n", vt.Example)
	}

	relFiles := relPaths(vt.Files, workDir)
	s.WriteString("\nAffected files:\n")
	limit := 15
	for i, f := range relFiles {
		if i >= limit {
			fmt.Fprintf(&s, "  ... and %d more\n", len(relFiles)-limit)
			break
		}
		fmt.Fprintf(&s, "  %s\n", f)
	}

	return s.String()
}

func relPaths(files []string, workDir string) []string {
	out := make([]string, len(files))
	for i, f := range files {
		if rel, err := filepath.Rel(workDir, f); err == nil {
			out[i] = rel
		} else {
			out[i] = f
		}
	}
	return out
}
