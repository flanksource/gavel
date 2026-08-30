package labels

import "fmt"

// Removal reports what removing a label actually did.
//
// The two halves are independent because the two scopes behave differently: a
// project removal retires the label outright — the definition goes and the
// label is stripped from every TODO in that workspace — while a global removal
// only drops the shared presentation, leaving each workspace's TODOs carrying a
// label that falls back to its built-in or derived hue. Reporting both counts
// lets one message say which of the two happened without the caller having to
// re-derive it from the scope.
type Removal struct {
	Name string `json:"name"`
	// Definition is true when a stored row was deleted. It is false when the
	// name only ever resolved through a built-in or the hashed fallback, which
	// is not a miss: the label still existed on TODOs and was still removed.
	Definition bool `json:"definition"`
	// Todos is how many TODOs the label was stripped from. Always zero for a
	// global removal, which never touches TODO content.
	Todos int64 `json:"todos"`
}

// Empty reports whether the removal found nothing at all — neither a stored
// definition nor a single TODO carrying the label. Callers treat this as the
// loud miss; anything else did real work.
func (r Removal) Empty() bool { return !r.Definition && r.Todos == 0 }

// String is the one-line summary the CLI and the dashboard both report.
func (r Removal) String() string {
	switch {
	case r.Definition && r.Todos > 0:
		return fmt.Sprintf("removed label %q and stripped it from %s", r.Name, pluralTodos(r.Todos))
	case r.Definition:
		return fmt.Sprintf("removed label %q; no TODO carried it", r.Name)
	case r.Todos > 0:
		return fmt.Sprintf("stripped label %q from %s", r.Name, pluralTodos(r.Todos))
	default:
		return fmt.Sprintf("label %q was not defined and no TODO carried it", r.Name)
	}
}

func pluralTodos(count int64) string {
	if count == 1 {
		return "1 TODO"
	}
	return fmt.Sprintf("%d TODOs", count)
}
