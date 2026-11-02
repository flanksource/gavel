package oxlint

import "github.com/flanksource/gavel/linters"

func (o *Oxlint) RuleExclusionArgs(patterns []string) ([]string, error) {
	var args []string
	for _, rule := range linters.ExactRuleExclusions(patterns) {
		switch rule {
		case "all", "correctness", "suspicious", "pedantic", "perf", "style", "restriction", "nursery":
			continue
		}
		args = append(args, "--allow", rule)
	}
	return args, nil
}
