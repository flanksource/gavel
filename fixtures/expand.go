package fixtures

import (
	"fmt"
	"os"
)

// ExpandVars replaces $VAR and ${VAR} references in s with values from vars.
// Unknown variables are left as literal "$VAR" so that shell variables in
// commands (e.g. $HOME) pass through to bash execution unchanged.
func ExpandVars(s string, vars map[string]any) string {
	return os.Expand(s, func(key string) string {
		if v, ok := vars[key]; ok {
			return fmt.Sprintf("%v", v)
		}
		return "$" + key
	})
}
