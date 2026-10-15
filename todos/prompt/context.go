package prompt

import (
	"embed"
	"fmt"
	"io/fs"

	"github.com/flanksource/captain/pkg/captainconfig"
	"github.com/flanksource/captain/pkg/promptcontext"
)

//go:embed hooks/*.prompt
var contextHooks embed.FS

// ContextSources are the prompt.context hook sources a todo run applies: the
// hooks gavel ships (hooks/*.prompt), then Captain's configured prompt dirs and
// <workDir>/.captain/prompts, so a user directory can shadow or disable a
// shipped hook by file name.
func ContextSources(cfg captainconfig.Config, workDir string) ([]promptcontext.Source, error) {
	shipped, err := fs.Sub(contextHooks, "hooks")
	if err != nil {
		return nil, err
	}
	defaults, err := promptcontext.DefaultSources(cfg, workDir)
	if err != nil {
		return nil, fmt.Errorf("resolve prompt context hook dirs: %w", err)
	}
	return append([]promptcontext.Source{{FS: shipped, Label: "gavel"}}, defaults...), nil
}
