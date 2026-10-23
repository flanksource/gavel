package eslint

import (
	"encoding/json"

	"github.com/flanksource/gavel/linters"
)

func (e *ESLint) RuleExclusionArgs(patterns []string) ([]string, error) {
	rules := map[string]string{}
	for _, rule := range linters.ExactRuleExclusions(patterns) {
		rules[rule] = "off"
	}
	if len(rules) == 0 {
		return nil, nil
	}
	data, err := json.Marshal(rules)
	if err != nil {
		return nil, err
	}
	return []string{"--rule", string(data)}, nil
}
