// Package all registers every fixture step kind gavel ships. Import it for its
// side effect wherever fixtures run — the gavel binary, the definition-of-done
// verifier, a TODO's checks — so a fixture document resolves the same kinds
// from every entry point:
//
//	import _ "github.com/flanksource/gavel/fixtures/kinds/all"
package all

import (
	// browser, playwright, agent-browser
	_ "github.com/flanksource/gavel/fixtures/kinds/browser"
	// services
	_ "github.com/flanksource/gavel/fixtures/kinds/services"
	// todo, todo_prompt
	_ "github.com/flanksource/gavel/fixtures/kinds/todo"
	// exec, test, lint, ai
	_ "github.com/flanksource/gavel/fixtures/types"
)
