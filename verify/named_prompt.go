package verify

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// NamedPromptSpec is one entry of `todos.prompts`: a PromptSpec plus the
// metadata that makes it selectable by name. A named prompt supplies the
// template and model; Class says which behaviour it runs as, so "which prompt
// runs" and "how the run behaves" stay separate axes.
//
// PromptSpec is a named field rather than embedded for the reason its own Spec
// field documents: PromptSpec declares MarshalJSON/UnmarshalJSON, and embedding
// would promote both onto NamedPromptSpec where they would shadow the default
// struct codec and silently drop class/title/description. The codec below keeps
// the on-disk shape flat — class beside model beside file:
//
//	todos:
//	  prompts:
//	    security:
//	      class: plan
//	      title: Security pass
//	      model: claude-opus-4-5
//	      file: .gavel/prompts/security.prompt
type NamedPromptSpec struct {
	PromptSpec `json:",inline" yaml:",inline"`
	// Class is the behaviour class the prompt executes as: run or plan. Empty
	// defaults to plan, the read-only posture — a prompt that was never told it
	// may commit must not inherit the right to.
	Class string `json:"class,omitempty" yaml:"class,omitempty"`
	// Title and Description are what the dashboard and `gavel todos prompts` show.
	Title       string `json:"title,omitempty" yaml:"title,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// IsEmpty reports whether this entry configures nothing at all.
func (s NamedPromptSpec) IsEmpty() bool {
	return s.PromptSpec.IsEmpty() && s.Class == "" && s.Title == "" && s.Description == ""
}

// Merge layers override onto this entry field-wise. It is what makes a repo
// .gavel.yaml able to re-point one named prompt's model while inheriting the
// home config's title and file.
func (s NamedPromptSpec) Merge(override NamedPromptSpec) NamedPromptSpec {
	merged := NamedPromptSpec{
		PromptSpec:  s.PromptSpec.Merge(override.PromptSpec),
		Class:       s.Class,
		Title:       s.Title,
		Description: s.Description,
	}
	if override.Class != "" {
		merged.Class = override.Class
	}
	if override.Title != "" {
		merged.Title = override.Title
	}
	if override.Description != "" {
		merged.Description = override.Description
	}
	return merged
}

// UnmarshalJSON accepts every shape PromptSpec does — an object, an inline spec
// or .prompt string, or a bare prompt body — and additionally reads the naming
// metadata from the object form. The string forms carry no metadata, so a
// `security: "Review this for..."` entry is a plan-class prompt with no title.
func (s *NamedPromptSpec) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		*s = NamedPromptSpec{}
		return nil
	}

	var spec PromptSpec
	if err := spec.UnmarshalJSON(trimmed); err != nil {
		return err
	}
	*s = NamedPromptSpec{PromptSpec: spec}
	if trimmed[0] == '"' {
		return nil
	}

	// The object form is flat: class/title/description sit beside the spec's own
	// keys, so both halves decode from the same object.
	var meta struct {
		Class       string `json:"class"`
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(trimmed, &meta); err != nil {
		return err
	}
	s.Class, s.Title, s.Description = meta.Class, meta.Title, meta.Description
	return nil
}

// MarshalJSON emits the flat object form so a .gavel.yaml round-trips through
// SaveGavelConfig unchanged. It splices the metadata into whatever PromptSpec
// produced rather than re-deriving the spec's own omissions.
func (s NamedPromptSpec) MarshalJSON() ([]byte, error) {
	raw, err := s.PromptSpec.MarshalJSON()
	if err != nil {
		return nil, err
	}
	flat := map[string]json.RawMessage{}
	if err := json.Unmarshal(raw, &flat); err != nil {
		return nil, fmt.Errorf("flatten named prompt spec: %w", err)
	}
	for key, value := range map[string]string{
		"class":       s.Class,
		"title":       s.Title,
		"description": s.Description,
	} {
		if value == "" {
			continue
		}
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("marshal named prompt %s: %w", key, err)
		}
		flat[key] = encoded
	}
	return json.Marshal(flat)
}
