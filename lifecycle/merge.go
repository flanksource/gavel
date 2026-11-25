package lifecycle

// Merge overlays a partial lifecycle onto a base. Steps merge by name and keep
// the base's order; new steps are appended in the overlay's order.
func Merge(base, overlay Lifecycle) Lifecycle {
	merged := Lifecycle{Name: base.Name, Subject: map[string]string{}, Order: base.Order, Disabled: base.Disabled}
	if overlay.Order != nil {
		merged.Order = overlay.Order
	}
	if overlay.Disabled != nil {
		merged.Disabled = overlay.Disabled
	}
	if overlay.Name != "" {
		merged.Name = overlay.Name
	}
	for name, typ := range base.Subject {
		merged.Subject[name] = typ
	}
	for name, typ := range overlay.Subject {
		merged.Subject[name] = typ
	}
	replaced := map[string]Step{}
	for _, step := range overlay.Steps {
		replaced[step.Name] = step
	}
	seen := map[string]bool{}
	for _, step := range base.Steps {
		if override, ok := replaced[step.Name]; ok {
			step = override
		}
		merged.Steps = append(merged.Steps, step)
		seen[step.Name] = true
	}
	for _, step := range overlay.Steps {
		if !seen[step.Name] {
			merged.Steps = append(merged.Steps, step)
		}
	}
	return merged
}
