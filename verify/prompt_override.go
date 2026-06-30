package verify

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PromptOverride replaces one of Gavel's built-in AI prompt templates. The
// defaults ship embedded in the binary; a project (or the user's ~/.gavel.yaml)
// overrides them either inline or by pointing at a file. Inline wins when both
// are set. In .gavel.yaml a bare string is shorthand for an inline override:
//
//	verify:
//	  promptTemplate: "You are a strict reviewer. {{scopeInstruction}}"
//	commit:
//	  messagePrompt:
//	    file: .gavel/prompts/commit-message.prompt
type PromptOverride struct {
	// Inline is the prompt template text used verbatim.
	Inline string `yaml:"inline,omitempty" json:"inline,omitempty"`
	// File is a path to a .prompt file. Relative paths resolve against the
	// directory of the .gavel.yaml the resolver is invoked for.
	File string `yaml:"file,omitempty" json:"file,omitempty"`
}

// IsZero reports whether no override is configured.
func (o PromptOverride) IsZero() bool {
	return strings.TrimSpace(o.Inline) == "" && strings.TrimSpace(o.File) == ""
}

// UnmarshalJSON accepts either a bare string (shorthand for Inline) or an object
// with inline/file keys. ghodss/yaml routes .gavel.yaml through JSON, so this
// also covers `messagePrompt: "..."` and `messagePrompt: {file: ...}`.
func (o *PromptOverride) UnmarshalJSON(data []byte) error {
	if strings.HasPrefix(strings.TrimSpace(string(data)), "\"") {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		o.Inline, o.File = s, ""
		return nil
	}
	type alias PromptOverride
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*o = PromptOverride(a)
	return nil
}

// Resolve returns the override template text — Inline verbatim, or the contents
// of File (relative paths resolved against dir). When no override is set it
// returns fallback (the embedded default). A configured-but-unreadable File is a
// hard error; it never silently falls back to the default.
func (o PromptOverride) Resolve(dir, fallback string) (string, error) {
	if strings.TrimSpace(o.Inline) != "" {
		return o.Inline, nil
	}
	if o.File != "" {
		path := o.File
		if !filepath.IsAbs(path) {
			path = filepath.Join(dir, path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("read prompt override file %q: %w", path, err)
		}
		return string(data), nil
	}
	return fallback, nil
}
