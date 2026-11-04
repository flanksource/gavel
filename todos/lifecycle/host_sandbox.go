// A step's sandbox backend, from the saved Captain config: the kind a backend name resolves to,
// which captain checks the effective spec against, and the selection the step's run executes under.
package lifecycle

import (
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
)

// sandboxBackends are the saved Captain config's sandbox backends; none for a host without one.
func (h *Host) sandboxBackends() captainconfig.SandboxDefaults {
	if h.Saved == nil {
		return captainconfig.SandboxDefaults{}
	}
	return h.Saved.Sandbox
}

// normalizeSandbox gives a step's sandbox that names a backend the backend's kind, which captain
// requires of the effective spec. Nothing else of the spec changes: a step that names no sandbox
// gets none, whatever the saved config's default.
func (h *Host) normalizeSandbox(spec api.Spec) (api.SpecNormalization, error) {
	ref := spec.Sandbox
	if ref == nil || ref.Mode != "" || ref.Backend == "" {
		return api.SpecNormalization{Spec: spec}, nil
	}
	selection, err := h.sandboxBackends().Resolve(ref.Backend)
	if err != nil {
		return api.SpecNormalization{}, err
	}
	resolved := *ref
	resolved.Mode = selection.Kind
	spec.Sandbox = &resolved
	return api.SpecNormalization{
		Spec: spec, Fields: api.FieldPresence{"/sandbox/mode": true},
		Source: api.FieldSource{Kind: api.FieldSourceSaved, Name: "~/.captain.yaml sandbox backend", Key: "sandbox.backends." + ref.Backend},
	}, nil
}

// sandboxSelection is the sandbox a step's run executes under: the backend or kind its spec
// names, with the backend's options from the saved Captain config; none for a run that is
// unsandboxed or sandboxed natively by its agent.
func (h *Host) sandboxSelection(spec api.Spec) (*api.SandboxConfig, error) {
	ref := spec.Sandbox
	if ref == nil || ref.Mode == api.SandboxOff || ref.Mode == api.SandboxNative {
		return nil, nil
	}
	selector := ref.Backend
	if selector == "" {
		selector = string(ref.Mode)
	}
	selection, err := h.sandboxBackends().Resolve(selector)
	if err != nil {
		return nil, err
	}
	return &api.SandboxConfig{Kind: selection.Kind, Name: selection.Name, Options: selection.Options, Agent: ref.Agent, Dispatch: ref.Dispatch}, nil
}
