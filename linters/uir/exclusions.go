package uir

import "github.com/flanksource/gavel/linters"

func (u *UIR) RuleExclusionArgs(patterns []string) ([]string, error) {
	var args []string
	for _, rule := range linters.ExactRuleExclusions(patterns) {
		args = append(args, "--exclude", rule)
	}
	return args, nil
}
