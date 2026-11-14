package golangci

import (
	"slices"
	"strings"

	"github.com/flanksource/gavel/linters"
)

func (g *GolangciLint) RuleExclusionArgs(patterns []string) ([]string, error) {
	var args []string
	for _, rule := range linters.ExactRuleExclusions(patterns) {
		args = append(args, "--disable", rule)
	}
	return args, nil
}

func (g *GolangciLint) excludeEnabledOnly(args []string) []string {
	excluded := linters.ExactRuleExclusions(g.NativeRuleExclusions)
	if len(excluded) == 0 {
		return args
	}
	var filtered []string
	for i := 0; i < len(args); i++ {
		value, joined := strings.CutPrefix(args[i], "--enable-only=")
		if !joined && (args[i] != "--enable-only" || i+1 == len(args)) {
			filtered = append(filtered, args[i])
			continue
		}
		if !joined {
			i++
			value = args[i]
		}
		enabled := strings.Split(value, ",")
		enabled = slices.DeleteFunc(enabled, func(rule string) bool { return slices.Contains(excluded, strings.TrimSpace(rule)) })
		if len(enabled) == 0 {
			filtered = append(filtered, "--default=none")
		} else {
			filtered = append(filtered, "--enable-only="+strings.Join(enabled, ","))
		}
	}
	return filtered
}
