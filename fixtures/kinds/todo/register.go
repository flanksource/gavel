// Package todo registers the `todo` and `todo_prompt` fixture step kinds: HTTP
// clients that drive a TODO through a running gavel dashboard (create, triage,
// plan review, lifecycle runs) and observe it until a fixture's wait verdict.
package todo

import "github.com/flanksource/gavel/fixtures"

func init() {
	fixtures.RegisterKind(todoKind{})
	fixtures.RegisterKind(promptKind{})
}

// Both kinds render their own config: `server`, `ref` and `dir` are expanded,
// while a todo body is sent verbatim — it is markdown that may hold an exec
// fixture whose `$VAR` and `{{…}}` belong to that fixture, not this one.
func (todoKind) TemplatesOwnConfig()   {}
func (promptKind) TemplatesOwnConfig() {}
