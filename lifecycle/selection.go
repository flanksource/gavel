package lifecycle

import "fmt"

func Select(def Lifecycle) (Lifecycle, error) {
	byName := make(map[string]Step, len(def.Steps))
	for _, step := range def.Steps {
		byName[step.Name] = step
	}
	disabled := map[string]bool{}
	if def.Disabled != nil {
		for _, name := range *def.Disabled {
			if _, ok := byName[name]; !ok || disabled[name] {
				return Lifecycle{}, fmt.Errorf("lifecycle %s: disabled step %q is unknown or repeated", def.Name, name)
			}
			disabled[name] = true
		}
	}
	steps := make([]Step, 0, len(def.Steps))
	seen := map[string]bool{}
	if def.Order != nil {
		for _, name := range *def.Order {
			step, ok := byName[name]
			if !ok || seen[name] {
				return Lifecycle{}, fmt.Errorf("lifecycle %s: ordered step %q is unknown or repeated", def.Name, name)
			}
			if !disabled[name] {
				steps = append(steps, step)
			}
			seen[name] = true
		}
	}
	for _, step := range def.Steps {
		if !seen[step.Name] && !disabled[step.Name] {
			steps = append(steps, step)
		}
	}
	def.Steps, def.Order, def.Disabled = steps, nil, nil
	return def, nil
}
