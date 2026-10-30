package workflows

import (
	"fmt"

	"github.com/flanksource/gavel/lifecycle"
	prlifecycle "github.com/flanksource/gavel/pr/lifecycle"
	workflow "github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/verify"
)

func Default(kind string) (lifecycle.Lifecycle, error) {
	switch kind {
	case "todos":
		return workflow.Default()
	case "pr":
		return prlifecycle.Default()
	default:
		return lifecycle.Lifecycle{}, fmt.Errorf("unknown lifecycle kind %q; expected todos or pr", kind)
	}
}

func Config(cfg verify.GavelConfig, kind string) (verify.LifecycleConfig, error) {
	switch kind {
	case "todos":
		return cfg.Todos.Lifecycle, nil
	case "pr":
		return cfg.PR.Lifecycle, nil
	default:
		return verify.LifecycleConfig{}, fmt.Errorf("unknown lifecycle kind %q", kind)
	}
}

func Resolve(kind string, override verify.LifecycleConfig, dir string) (lifecycle.Lifecycle, error) {
	base, err := Default(kind)
	if err != nil {
		return base, err
	}
	overlay, err := lifecycle.Read(override, dir)
	if err != nil {
		return base, err
	}
	def := lifecycle.Merge(base, overlay)
	if err := Validate(kind, def); err != nil {
		return lifecycle.Lifecycle{}, err
	}
	def, err = lifecycle.Select(def)
	if err != nil {
		return def, err
	}
	_, err = lifecycle.New(def)
	return def, err
}

func Validate(kind string, def lifecycle.Lifecycle) error {
	switch kind {
	case "todos":
		return workflow.Validate(def)
	case "pr":
		return prlifecycle.Validate(def)
	default:
		return fmt.Errorf("unknown lifecycle kind %q", kind)
	}
}
