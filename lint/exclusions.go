package lint

import (
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/verify"
)

func ParseRuleExclusions(values []string) (linters.RuleExclusions, error) {
	sources := make(map[string]string, len(verify.BuiltinLinterNames)+len(linterAliases)+1)
	for _, name := range verify.BuiltinLinterNames {
		sources[name] = name
	}
	for alias, name := range linterAliases {
		sources[alias] = name
	}
	sources["typescript"] = "tsc"
	sources["prompt"] = "prompt"
	return linters.ParseRuleExclusions(values, sources)
}
