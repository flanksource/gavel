package ruff

import (
	"regexp"
	"strings"
)

var excludedRuleCode = regexp.MustCompile(`^[A-Z]+[0-9]+$`)
var excludedRulePrefix = regexp.MustCompile(`^[A-Z]+[0-9]*\*$`)

func (r *Ruff) RuleExclusionArgs(patterns []string) ([]string, error) {
	var codes []string
	for _, pattern := range patterns {
		if excludedRuleCode.MatchString(pattern) || excludedRulePrefix.MatchString(pattern) {
			codes = append(codes, strings.TrimSuffix(pattern, "*"))
		}
	}
	if len(codes) == 0 {
		return nil, nil
	}
	return []string{"--ignore", strings.Join(codes, ",")}, nil
}
