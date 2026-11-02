package linters

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/flanksource/gavel/models"
)

type RuleExclusion struct {
	Source  string
	Pattern string
}

type RuleExclusions []RuleExclusion

type RuleExcluder interface {
	RuleExclusionArgs(patterns []string) ([]string, error)
}

func ParseRuleExclusions(values []string, sources map[string]string) (RuleExclusions, error) {
	var rules RuleExclusions
	seen := map[RuleExclusion]bool{}
	for _, value := range values {
		for _, selector := range strings.Split(value, ",") {
			selector = strings.TrimSpace(selector)
			if selector == "" {
				return nil, fmt.Errorf("--exclude: rule selector must not be empty")
			}
			rule := RuleExclusion{Pattern: selector}
			if source, pattern, qualified := strings.Cut(selector, ":"); qualified {
				canonical, exists := sources[strings.TrimSpace(source)]
				if !exists || strings.TrimSpace(pattern) == "" {
					return nil, fmt.Errorf("--exclude: invalid source or empty rule in %q", selector)
				}
				rule.Source, rule.Pattern = canonical, strings.TrimSpace(pattern)
			}
			if !seen[rule] {
				rules = append(rules, rule)
				seen[rule] = true
			}
		}
	}
	return rules, nil
}

func (rules RuleExclusions) Filter(source string, violations []models.Violation) []models.Violation {
	if len(rules) == 0 {
		return violations
	}
	patterns := make([]*regexp.Regexp, len(rules))
	for i, rule := range rules {
		patterns[i] = regexp.MustCompile("(?s)^" + strings.ReplaceAll(regexp.QuoteMeta(rule.Pattern), `\*`, ".*") + "$")
	}
	kept := violations[:0]
	for _, violation := range violations {
		violationSource := violation.Source
		if violationSource == "" {
			violationSource = source
		}
		excluded := false
		for i, rule := range rules {
			if violation.Rule != nil && (rule.Source == "" || rule.Source == violationSource) && patterns[i].MatchString(violation.Rule.Method) {
				excluded = true
				break
			}
		}
		if !excluded {
			kept = append(kept, violation)
		}
	}
	return kept
}

func (rules RuleExclusions) NativeFor(source string, single bool) []string {
	var patterns []string
	seen := map[string]bool{}
	for _, rule := range rules {
		if (rule.Source == source || (rule.Source == "" && single)) && !seen[rule.Pattern] {
			patterns = append(patterns, rule.Pattern)
			seen[rule.Pattern] = true
		}
	}
	return patterns
}

func ExactRuleExclusions(patterns []string) []string {
	var exact []string
	for _, pattern := range patterns {
		if !strings.Contains(pattern, "*") {
			exact = append(exact, pattern)
		}
	}
	return exact
}

func (rules RuleExclusions) FilterResults(results []*LinterResult) {
	for _, result := range results {
		if result != nil {
			result.Violations = rules.Filter(result.Linter, result.Violations)
		}
	}
}
