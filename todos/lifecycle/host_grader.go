package lifecycle

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

func (h *Host) finalizeVerification(ctx context.Context, todo *types.TODO, lc Context, class types.RunMode, spec *api.Spec, trace []api.SpecLayer) error {
	verification, ok := lc.Subject["verification"].(map[string]any)
	if !ok {
		return fmt.Errorf("lifecycle subject: verification must be an object")
	}
	original, ok := verification["document"].(string)
	if !ok {
		return fmt.Errorf("lifecycle subject: verification.document must be a string")
	}
	if original == "" {
		return nil
	}
	// Verification uses its resolved request; generation uses an independent
	// grader so the implementer's session and runtime never grade its own work.
	grader := *spec
	if class != types.ModeVerify {
		var err error
		if grader, err = h.graderSpec(ctx, todo); err != nil {
			return err
		}
	}
	dod, err := todos.BuildDefinitionOfDone(todos.DefinitionOfDoneOptions{
		WorkDir: h.WorkDir, Todos: []*types.TODO{todo}, Grader: grader,
	})
	if err != nil {
		return err
	}
	verification["document"] = dod.Fixture
	replaceVerificationDocument(spec, original, dod.Fixture)
	for i := range trace {
		replaceVerificationDocument(&trace[i].Spec, original, dod.Fixture)
	}
	return nil
}

func replaceVerificationDocument(spec *api.Spec, original, resolved string) {
	if spec.Workflow == nil || spec.Workflow.Verify == nil {
		return
	}
	if fixture := spec.Workflow.Verify.Fixture; strings.Contains(fixture, original) {
		workflow := *spec.Workflow
		verify := *workflow.Verify
		verify.Fixture = strings.ReplaceAll(fixture, original, resolved)
		workflow.Verify = &verify
		spec.Workflow = &workflow
	}
}
