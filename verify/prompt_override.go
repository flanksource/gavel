package verify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	dotprompt "github.com/flanksource/captain/pkg/ai/prompt"
	"github.com/flanksource/captain/pkg/api"
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
	// Inline is either a JSON string containing a complete .prompt template, or
	// a structured captain api.Spec object. json.RawMessage preserves which form
	// the user supplied while ghodss/yaml routes .gavel.yaml through JSON.
	Inline json.RawMessage `yaml:"inline,omitempty" json:"inline,omitempty"`
	// File is a path to a .prompt file. Relative paths resolve against the
	// directory of the .gavel.yaml that declared this override.
	File string `yaml:"file,omitempty" json:"file,omitempty"`

	// baseDir is populated by the config loader and survives layered merges. It
	// is intentionally not serialized; programmatic values fall back to the dir
	// passed to Resolve.
	baseDir string
}

// InlinePrompt returns an override containing complete .prompt source text.
func InlinePrompt(text string) PromptOverride {
	raw, _ := json.Marshal(text)
	return PromptOverride{Inline: raw}
}

// StructuredInlinePrompt returns an override containing a captain api.Spec.
func StructuredInlinePrompt(spec api.Spec) (PromptOverride, error) {
	raw, err := json.Marshal(spec)
	if err != nil {
		return PromptOverride{}, err
	}
	return PromptOverride{Inline: raw}, nil
}

// IsZero reports whether no override is configured.
func (o PromptOverride) IsZero() bool {
	return !o.HasInline() && strings.TrimSpace(o.File) == ""
}

// HasInline reports whether an inline string or structured spec is configured.
func (o PromptOverride) HasInline() bool {
	raw := bytes.TrimSpace(o.Inline)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return false
	}
	if raw[0] == '"' {
		var text string
		if err := json.Unmarshal(raw, &text); err == nil {
			return strings.TrimSpace(text) != ""
		}
	}
	return true
}

// InlineText returns the inline value when it is the string form.
func (o PromptOverride) InlineText() (string, bool) {
	raw := bytes.TrimSpace(o.Inline)
	if len(raw) == 0 || raw[0] != '"' {
		return "", false
	}
	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return "", false
	}
	return text, true
}

// UnmarshalJSON accepts either a bare string (shorthand for Inline) or an object
// with inline/file keys. ghodss/yaml routes .gavel.yaml through JSON, so this
// also covers `messagePrompt: "..."` and `messagePrompt: {file: ...}`.
func (o *PromptOverride) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		*o = PromptOverride{}
		return nil
	}
	if data[0] == '"' {
		var text string
		if err := json.Unmarshal(data, &text); err != nil {
			return err
		}
		raw, _ := json.Marshal(text)
		o.Inline, o.File = raw, ""
		return nil
	}
	type alias PromptOverride
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	if bytes.Equal(bytes.TrimSpace(a.Inline), []byte("null")) {
		return fmt.Errorf("inline prompt must be a string or captain api.Spec object, not null")
	}
	*o = PromptOverride(a)
	return nil
}

// Resolve returns the override template text — Inline verbatim, or the contents
// of File (relative paths resolved against dir). When no override is set it
// returns fallback (the embedded default). A configured-but-unreadable File is a
// hard error; it never silently falls back to the default.
func (o PromptOverride) Resolve(dir, fallback string) (string, error) {
	if o.HasInline() {
		return o.resolveInline()
	}
	if o.File != "" {
		path := o.ResolvedFilePath(dir)
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("read prompt override file %q: %w", path, err)
		}
		return string(data), nil
	}
	return fallback, nil
}

// ResolvedFilePath returns File as an absolute path relative to the config
// layer that declared it. Empty File returns an empty path.
func (o PromptOverride) ResolvedFilePath(fallbackDir string) string {
	if o.File == "" || filepath.IsAbs(o.File) {
		return o.File
	}
	baseDir := o.baseDir
	if baseDir == "" {
		baseDir = fallbackDir
	}
	return filepath.Join(baseDir, o.File)
}

func (o PromptOverride) resolveInline() (string, error) {
	raw := bytes.TrimSpace(o.Inline)
	switch raw[0] {
	case '"':
		var text string
		if err := json.Unmarshal(raw, &text); err != nil {
			return "", fmt.Errorf("decode inline prompt string: %w", err)
		}
		return text, nil
	case '{':
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.DisallowUnknownFields()
		var spec api.Spec
		if err := dec.Decode(&spec); err != nil {
			return "", fmt.Errorf("decode inline prompt spec: %w", err)
		}
		if err := spec.Validate(); err != nil {
			return "", fmt.Errorf("validate inline prompt spec: %w", err)
		}

		var frontmatter map[string]any
		if err := json.Unmarshal(raw, &frontmatter); err != nil {
			return "", fmt.Errorf("decode inline prompt frontmatter: %w", err)
		}
		if promptValue, ok := frontmatter["prompt"].(map[string]any); ok {
			delete(promptValue, "user")
			if len(promptValue) == 0 {
				delete(frontmatter, "prompt")
			}
		}
		return (&dotprompt.Document{Frontmatter: frontmatter, Body: spec.Prompt.User}).String()
	default:
		return "", fmt.Errorf("inline prompt must be a string or captain api.Spec object")
	}
}

func setPromptOverrideBaseDirs(cfg *GavelConfig, dir string) {
	promptOverrideType := reflect.TypeOf(PromptOverride{})
	var walk func(reflect.Value)
	walk = func(v reflect.Value) {
		for v.Kind() == reflect.Pointer {
			if v.IsNil() {
				return
			}
			v = v.Elem()
		}
		if v.Type() == promptOverrideType {
			v.Addr().Interface().(*PromptOverride).baseDir = dir
			return
		}
		if v.Kind() != reflect.Struct {
			return
		}
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).PkgPath == "" {
				walk(v.Field(i))
			}
		}
	}
	walk(reflect.ValueOf(cfg))
}
