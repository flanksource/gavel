package ui

import (
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"

	dotprompt "github.com/flanksource/captain/pkg/ai/prompt"
	"github.com/flanksource/gavel/prompts"
	promptregistry "github.com/flanksource/gavel/prompts/registry"
	"github.com/flanksource/gavel/verify"
)

// promptEffective is what gavel would actually run for the scope's directory
// once every .gavel.yaml layer is merged. It is distinct from the layer the
// editor writes to: that layer can be untouched ("default") while a lower layer
// (~/.gavel.yaml, the git root) supplies the override that really runs.
type promptEffective struct {
	Source  string `json:"source"`           // default | inline | file
	Origin  string `json:"origin,omitempty"` // layer supplying the override: user-home | git-root | target-directory
	Path    string `json:"path,omitempty"`
	Raw     string `json:"raw"`
	Version string `json:"version"`
}

func effectivePromptView(desc prompts.Prompt, dir string) (*promptEffective, error) {
	trace, err := verify.LoadGavelConfigTrace(dir)
	if err != nil {
		return nil, err
	}
	ov, err := promptOverridePtr(&trace.Merged, desc.ConfigPath)
	if err != nil {
		return nil, err
	}
	raw, err := promptSpecRaw(ov, trace.TargetDir, desc.Default)
	if err != nil {
		return nil, err
	}
	view := &promptEffective{Source: overrideSource(ov), Raw: raw, Version: promptSourceVersion(raw)}
	if ov.File != "" {
		view.Path = ov.ResolvedFilePath(trace.TargetDir)
	}
	for _, source := range trace.Sources {
		cfg := source.Config
		if layer, err := promptOverridePtr(&cfg, desc.ConfigPath); err == nil && !layer.IsEmpty() {
			view.Origin = source.Origin
		}
	}
	return view, nil
}

// checkPromptBaseRaw refuses a save whose merge base is no longer the layer's
// current text: another editor, a CLI run, or a lower layer changed it after the
// dialog loaded, and merging onto the stale base would silently discard that.
func checkPromptBaseRaw(ov *verify.PromptSpec, dir string, desc prompts.Prompt, baseRaw *string) (int, error) {
	if baseRaw == nil {
		return 0, nil
	}
	current, err := promptSpecRaw(ov, dir, desc.Default)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if *baseRaw != current {
		return http.StatusConflict, fmt.Errorf(
			"prompt %s changed since it was loaded (version %s, now %s); reload before saving",
			desc.ID, promptSourceVersion(*baseRaw), promptSourceVersion(current))
	}
	return 0, nil
}

// rejectLossyInline refuses to store a document inline when doing so would
// change frontmatter api.Spec cannot hold (output, input, config, …). An inline
// override is persisted as the typed spec: with the default body it keeps
// running the default template, so those keys survive only while they still
// equal the default's; with its own body the runtime renders that body alone
// and every such key is gone. A file override keeps the document verbatim, so
// the error steers the caller there.
func rejectLossyInline(text, def string) error {
	doc, err := dotprompt.Parse(text)
	if err != nil {
		return err
	}
	_, _, defaultFrontmatter, _ := promptregistry.ParsePromptSource(def)
	bodyUnchanged := bodyMatchesDefault(doc.Body, def)
	modeled := modeledSpecKeys()
	var changed []string
	for key := range unionKeys(doc.Frontmatter, defaultFrontmatter) {
		if modeled[key] || (bodyUnchanged && reflect.DeepEqual(doc.Frontmatter[key], defaultFrontmatter[key])) {
			continue
		}
		changed = append(changed, key)
	}
	if len(changed) == 0 {
		return nil
	}
	sort.Strings(changed)
	return fmt.Errorf(
		"an inline override cannot change the frontmatter key(s) %s; save this prompt as a file to keep them",
		strings.Join(changed, ", "))
}

func unionKeys(maps ...map[string]any) map[string]bool {
	keys := map[string]bool{}
	for _, m := range maps {
		for key := range m {
			keys[key] = true
		}
	}
	return keys
}
