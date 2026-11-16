package verify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/flanksource/captain/pkg/api"
)

// DefaultCommentAgentPrompt is the captain prompt a comment agent runs when its
// entry names none: the editor's managed comment-working prompt.
const DefaultCommentAgentPrompt = "editor-comments"

// CommentsConfig configures comment threads on the files under this
// .gavel.yaml's directory.
type CommentsConfig struct {
	Agents CommentAgentsConfig `yaml:"agents,omitempty" json:"agents,omitempty"`
}

// CommentAgentsConfig is which agent works the comment threads on pages under
// this layer's directory: Default for every page beneath it, Pages for single
// files, keyed by their path relative to the directory. Layers merge field by
// field (see ResolveCommentAgent): a nearer layer, and a page entry over its
// directory's Default, set only the fields they name.
type CommentAgentsConfig struct {
	Default *CommentAgent           `yaml:"default,omitempty" json:"default,omitempty"`
	Pages   map[string]CommentAgent `yaml:"pages,omitempty" json:"pages,omitempty"`
}

// CommentAgent is the captain run a new human comment thread is handed to.
// Every field is optional so a layer can override one and inherit the rest;
// Enabled false opts its pages out of an agent inherited from further up.
type CommentAgent struct {
	Enabled *bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	// Assign makes the agent the thread's assignee as soon as it is created.
	Assign *bool `yaml:"assign,omitempty" json:"assign,omitempty"`
	// Prompt is the captain prompt the run starts; DefaultCommentAgentPrompt when empty.
	Prompt string `yaml:"prompt,omitempty" json:"prompt,omitempty"`
	// Presets replace the prompt frontmatter's runtime presets when set.
	Presets []string `yaml:"presets,omitempty" json:"presets,omitempty"`
	// Spec is layered over the prompt frontmatter's runtime spec.
	Spec *api.Spec `yaml:"spec,omitempty" json:"spec,omitempty"`
	// Variables feed the prompt template below the host's own.
	Variables map[string]any `yaml:"variables,omitempty" json:"variables,omitempty"`
	// Instructions are appended to the thread the session is given.
	Instructions string `yaml:"instructions,omitempty" json:"instructions,omitempty"`
	// Group is the delivery group key: the threads delivered in one group
	// reuse the live session recorded for it, instead of each starting a run.
	Group string `yaml:"group,omitempty" json:"group,omitempty"`
	// Edits is CommentEditsDirect (the default when empty), letting the agent
	// edit files, or CommentEditsPropose: it answers with a proposed patch.
	Edits string `yaml:"edits,omitempty" json:"edits,omitempty"`
}

// Comment agent edits policies; the values of comments.EditsDirect/EditsPropose.
const (
	CommentEditsDirect  = "direct"
	CommentEditsPropose = "propose"
)

// Validate rejects an edits policy other than direct or propose.
func (a CommentAgent) Validate() error {
	if a.Edits != "" && a.Edits != CommentEditsDirect && a.Edits != CommentEditsPropose {
		return fmt.Errorf("edits must be %q or %q, not %q", CommentEditsDirect, CommentEditsPropose, a.Edits)
	}
	return nil
}

// PromptOrDefault is the prompt the agent's run starts.
func (a CommentAgent) PromptOrDefault() string {
	if a.Prompt == "" {
		return DefaultCommentAgentPrompt
	}
	return a.Prompt
}

// IsEnabled is whether new threads are handed to the agent; unset is off.
func (a CommentAgent) IsEnabled() bool { return a.Enabled != nil && *a.Enabled }

// Assigns is whether the agent is made the thread's assignee; unset is no.
func (a CommentAgent) Assigns() bool { return a.Assign != nil && *a.Assign }

// CommentAgentLayer is one config layer's comment agents, named Label, with
// Dir the slash-separated directory its keys are relative to ("." for the root
// the file paths are relative to).
type CommentAgentLayer struct {
	Label  string
	Dir    string
	Agents CommentAgentsConfig
}

// CommentAgentSource is the entry that set a resolved agent field: a layer's
// default, or its entry for Page.
type CommentAgentSource struct {
	Layer string `json:"layer"`
	Page  string `json:"page,omitempty"`
}

// ResolvedCommentAgent is a file's agent and, per field (its JSON name), the
// entry that last set it. Found is false when no layer configures the file.
type ResolvedCommentAgent struct {
	Agent      CommentAgent                  `json:"agent"`
	Provenance map[string]CommentAgentSource `json:"provenance"`
	Found      bool                          `json:"found"`
}

// ResolveCommentAgent is file's agent: layers apply in order (ancestors
// first), each merging its Default when it covers file, then its page entry
// for file, field by field.
func ResolveCommentAgent(layers []CommentAgentLayer, file string) (ResolvedCommentAgent, error) {
	out := ResolvedCommentAgent{Provenance: map[string]CommentAgentSource{}}
	apply := func(entry CommentAgent, source CommentAgentSource) error {
		data, err := json.Marshal(entry)
		if err != nil {
			return fmt.Errorf("encode the comment agent of %s: %w", source.Layer, err)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(data, &fields); err != nil {
			return fmt.Errorf("decode the comment agent of %s: %w", source.Layer, err)
		}
		for field := range fields {
			out.Provenance[field] = source
		}
		out.Agent, out.Found = Merge(out.Agent, entry), true
		return nil
	}
	for _, layer := range layers {
		rel, covered := relativeTo(layer.Dir, file)
		if !covered {
			continue
		}
		if layer.Agents.Default != nil {
			if err := apply(*layer.Agents.Default, CommentAgentSource{Layer: layer.Label}); err != nil {
				return out, err
			}
		}
		if page, ok := layer.Agents.Pages[rel]; ok {
			if err := apply(page, CommentAgentSource{Layer: layer.Label, Page: rel}); err != nil {
				return out, err
			}
		}
	}
	return out, nil
}

// relativeTo is file relative to dir, both slash-separated, and whether file lies beneath dir.
func relativeTo(dir, file string) (string, bool) {
	if dir == "." || dir == "" {
		return file, true
	}
	rel, ok := strings.CutPrefix(file, dir+"/")
	return rel, ok && rel != ""
}

// Validate rejects a page key that does not name a file beneath the
// .gavel.yaml's directory, so an entry can never route threads elsewhere.
func (c CommentAgentsConfig) Validate() error {
	if c.Default != nil {
		if err := c.Default.Validate(); err != nil {
			return fmt.Errorf("comments.agents.default: %w", err)
		}
	}
	for key, agent := range c.Pages {
		if !IsCommentPageKey(key) {
			return fmt.Errorf("comments.agents.pages: %q must be a relative file path beneath the .gavel.yaml directory", key)
		}
		if err := agent.Validate(); err != nil {
			return fmt.Errorf("comments.agents.pages.%s: %w", key, err)
		}
	}
	return nil
}

// IsCommentPageKey is whether key is a clean, relative, slash-separated path
// beneath its directory.
func IsCommentPageKey(key string) bool {
	return key != "" && key != "." && path.Clean(key) == key && !path.IsAbs(key) && key != ".." && !strings.HasPrefix(key, "../") && !strings.Contains(key, "\\")
}

// Set makes agent the layer's default (page empty) or its entry for page, or
// removes that entry when agent is nil.
func (c *CommentAgentsConfig) Set(page string, agent *CommentAgent) error {
	if agent != nil {
		if err := agent.Validate(); err != nil {
			return err
		}
	}
	switch {
	case page == "":
		c.Default = agent
	case !IsCommentPageKey(page):
		return fmt.Errorf("comment agent page %q must be a relative file path beneath the layer's directory", page)
	case agent == nil:
		delete(c.Pages, page)
		if len(c.Pages) == 0 {
			c.Pages = nil
		}
	default:
		if c.Pages == nil {
			c.Pages = map[string]CommentAgent{}
		}
		c.Pages[page] = *agent
	}
	return nil
}

// CommentAgentEdit turns a layer's entry (nil when it has none) into its new
// entry; a nil result removes the entry.
type CommentAgentEdit func(current *CommentAgent) (*CommentAgent, error)

// ReplaceCommentAgent makes agent the whole entry, or removes it when nil.
func ReplaceCommentAgent(agent *CommentAgent) CommentAgentEdit {
	return func(*CommentAgent) (*CommentAgent, error) { return agent, nil }
}

// PatchCommentAgent applies patch, a JSON merge patch (RFC 7386), to an
// entry: a field that is present replaces that field (objects merge
// recursively), a field set to null is removed so it inherits again, and a
// null patch removes the whole entry. The patch is checked (known fields,
// their types, the edits policy) before it is returned.
func PatchCommentAgent(patch json.RawMessage) (CommentAgentEdit, error) {
	var doc any
	if err := json.Unmarshal(patch, &doc); err != nil {
		return nil, fmt.Errorf("comment agent patch: %w", err)
	}
	if doc == nil {
		return ReplaceCommentAgent(nil), nil
	}
	fields, ok := doc.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("comment agent patch must be a JSON object or null, not %s", patch)
	}
	if _, err := decodeCommentAgent(withoutNulls(fields)); err != nil {
		return nil, fmt.Errorf("comment agent patch: %w", err)
	}
	return func(current *CommentAgent) (*CommentAgent, error) {
		var base any = map[string]any{}
		if current != nil {
			data, err := json.Marshal(current)
			if err != nil {
				return nil, fmt.Errorf("encode the comment agent entry: %w", err)
			}
			if err := json.Unmarshal(data, &base); err != nil {
				return nil, fmt.Errorf("decode the comment agent entry: %w", err)
			}
		}
		merged, err := decodeCommentAgent(mergePatch(base, fields))
		if err != nil {
			return nil, fmt.Errorf("apply the comment agent patch: %w", err)
		}
		return &merged, nil
	}, nil
}

// mergePatch is target with patch applied as RFC 7386 describes.
func mergePatch(target, patch any) any {
	fields, ok := patch.(map[string]any)
	if !ok {
		return patch
	}
	merged, ok := target.(map[string]any)
	if !ok {
		merged = map[string]any{}
	}
	for key, value := range fields {
		if value == nil {
			delete(merged, key)
		} else {
			merged[key] = mergePatch(merged[key], value)
		}
	}
	return merged
}

// withoutNulls is fields with every null member removed, at any depth.
func withoutNulls(fields map[string]any) map[string]any {
	out := map[string]any{}
	for key, value := range fields {
		switch v := value.(type) {
		case nil:
		case map[string]any:
			out[key] = withoutNulls(v)
		default:
			out[key] = v
		}
	}
	return out
}

// decodeCommentAgent decodes doc strictly into a valid CommentAgent.
func decodeCommentAgent(doc any) (CommentAgent, error) {
	var agent CommentAgent
	data, err := json.Marshal(doc)
	if err != nil {
		return agent, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&agent); err != nil {
		return agent, err
	}
	return agent, agent.Validate()
}

// Edit applies edit to the layer's default (page empty) or its entry for page.
func (c *CommentAgentsConfig) Edit(page string, edit CommentAgentEdit) error {
	current := c.Default
	if page != "" {
		current = nil
		if entry, ok := c.Pages[page]; ok {
			current = &entry
		}
	}
	next, err := edit(current)
	if err != nil {
		return err
	}
	return c.Set(page, next)
}

func commentsSchema(specSchema map[string]any) map[string]any {
	edits := stringWithDefault("direct lets the agent edit the project's files; propose has it answer with a proposed patch a person accepts, rejects or asks to revise.", CommentEditsDirect)
	edits["enum"] = []any{CommentEditsDirect, CommentEditsPropose}
	agent := object("A captain run that new human comment threads are handed to.", map[string]any{
		"enabled":      boolProp("Hand new human threads on these pages to the agent. false opts the pages out of an agent inherited from a .gavel.yaml further up."),
		"assign":       boolProp("Make the agent the thread's assignee as soon as the thread is created."),
		"prompt":       stringWithDefault("Captain prompt the run starts.", DefaultCommentAgentPrompt),
		"presets":      stringArray("Runtime presets replacing the prompt frontmatter's presets."),
		"spec":         specNodeSchema(specSchema, "Runtime spec layered over the prompt frontmatter's.", "", "Model the run uses; the prompt's when unset."),
		"variables":    mapObject("Prompt template variables, below the ones the host sets.", map[string]any{}),
		"instructions": stringProp("Instructions appended to the thread the session is given."),
		"group":        stringProp("Delivery group: threads delivered in the same group reuse the group's live session instead of each starting a new run."),
		"edits":        edits,
	})
	return object("Comment thread settings for the files beneath this .gavel.yaml's directory.", map[string]any{
		"agents": object("Which agent works new comment threads. The nearest .gavel.yaml wins, and a page entry wins over its directory's default.", map[string]any{
			"default": agent,
			"pages":   mapObject("Per-file agents, keyed by the file's path relative to this .gavel.yaml's directory.", agent),
		}),
	})
}
