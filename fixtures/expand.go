package fixtures

import (
	"fmt"
	"regexp"

	"github.com/flanksource/commons/logger"
)

var varReference = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}|\$([A-Za-z_][A-Za-z0-9_]*)`)

// ExpandVars replaces $VAR and ${VAR} references in s with values from vars.
// Everything else is left exactly as written — unknown variables, shell
// parameter expansions such as ${VAR:-default} or ${0#prefix}, and specials
// like $1 or $? — so shell syntax in commands passes through to bash unchanged.
func ExpandVars(s string, vars map[string]any) string {
	return varReference.ReplaceAllStringFunc(s, func(ref string) string {
		match := varReference.FindStringSubmatch(ref)
		key := match[1] + match[2]
		v, ok := vars[key]
		if !ok {
			return ref
		}
		val := fmt.Sprintf("%v", v)
		logger.V(6).Infof("ExpandVars: $%s → %s", key, val)
		return val
	})
}
