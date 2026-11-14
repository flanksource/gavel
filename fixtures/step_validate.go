package fixtures

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"sort"
	"strings"
)

// stepRefPatterns find `steps.<id>` references in templates and CEL:
// `.steps.id` / `steps.id` and `steps['id']` / `steps["id"]`.
var stepRefPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\bsteps\.([A-Za-z_][A-Za-z0-9_]*)`),
	regexp.MustCompile(`\bsteps\[["']([^"']+)["']\]`),
}

// validateSteps checks what is knowable before anything runs, in file order:
// every kind is registered and accepts the step's config and envelope (wait
// needs Observe, background needs Start), ids are unique, `steps.<id>`
// references name an earlier step, and no two steps export one name unless the
// later one sets override.
func validateSteps(steps []parsedStep) error {
	seen := map[string]bool{}
	kinds := map[string]bool{}
	exporters := map[string]string{}
	for i, step := range steps {
		kinds[step.spec.Kind] = true
		if err := validateStep(step, seen, exporters); err != nil {
			return step.diagnostic(err)
		}
		if err := validateRootID(*step.spec, steps[i+1:]); err != nil {
			return fmt.Errorf("%s: %w", step.label, err)
		}
		if step.spec.ID != "" {
			seen[step.spec.ID] = true
		}
	}
	specs := make([]StepSpec, len(steps))
	for i, step := range steps {
		specs[i] = *step.spec
	}
	for _, name := range slices.Sorted(maps.Keys(kinds)) {
		kind, err := DefaultKinds.Lookup(name)
		if err != nil {
			return err
		}
		if planner, ok := kind.(StepPlanner); ok {
			if err := planner.PrepareSteps(specs); err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		}
	}
	for i, step := range steps {
		*step.spec = specs[i]
	}
	return nil
}

func validateStep(step parsedStep, seen map[string]bool, exporters map[string]string) error {
	spec := step.spec
	if spec.ID != "" && seen[spec.ID] {
		return fmt.Errorf("duplicate step id %q", spec.ID)
	}
	kind, err := DefaultKinds.Lookup(spec.Kind)
	if err != nil {
		return err
	}
	if err := validateKindConfig(kind, spec.Config); err != nil {
		return err
	}
	if _, ok := kind.(Observer); spec.Wait != nil && !ok {
		return fmt.Errorf("kind %q cannot wait: it does not implement Observe", spec.Kind)
	}
	if spec.Wait != nil {
		if _, err := parseJitter(spec.Wait.Jitter); err != nil {
			return err
		}
	}
	if _, ok := kind.(BackgroundKind); spec.Background && !ok {
		return fmt.Errorf("kind %q cannot run in the background: it does not implement Start", spec.Kind)
	}
	for _, ref := range stepReferences(*spec) {
		if !seen[ref] {
			return fmt.Errorf("references steps.%s, which no earlier step defines", ref)
		}
	}
	owner := firstNonEmpty(spec.ID, step.label)
	for _, name := range slices.Sorted(maps.Keys(spec.Export)) {
		if name == exportAll || spec.Export[name] == "" {
			continue
		}
		if prior, ok := exporters[name]; ok && prior != owner && !spec.Override {
			return fmt.Errorf("%s is already exported by step %q; set override: true to replace it", name, prior)
		}
		exporters[name] = owner
	}
	return nil
}

// validateRootID refuses an id that takes a reserved root, and an id that
// cannot be a root (not a CEL identifier) which a later step uses as one —
// `stage-approval.approved` would parse as a subtraction.
func validateRootID(spec StepSpec, later []parsedStep) error {
	if spec.ID == "" {
		return nil
	}
	if slices.Contains(ReservedRoots, spec.ID) {
		return fmt.Errorf("step id %q collides with the reserved root %q (reserved: %s)", spec.ID, spec.ID, strings.Join(ReservedRoots, ", "))
	}
	if IsRootID(spec.ID) {
		return nil
	}
	id := regexp.QuoteMeta(spec.ID)
	celRoot := regexp.MustCompile(`(?:^|[^\w.\]"'/:-])` + id + `\.`)
	templateRoot := regexp.MustCompile(`\{\{-?\s*\.` + id + `\.`)
	for _, step := range later {
		uses := slices.ContainsFunc(stepTexts(*step.spec), templateRoot.MatchString) ||
			slices.ContainsFunc(stepCELTexts(*step.spec), celRoot.MatchString)
		if uses {
			return fmt.Errorf("step id %q is used as a root by %s but is not a valid CEL identifier; rename it (e.g. %s) or read it as steps[%q]",
				spec.ID, step.label, camelID(spec.ID), spec.ID)
		}
	}
	return nil
}

// camelID suggests a root-safe spelling of an id: stage-approval → stageApproval.
func camelID(id string) string {
	parts := strings.FieldsFunc(id, func(r rune) bool { return r == '-' || r == '.' || r == ' ' })
	for i := 1; i < len(parts); i++ {
		parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
	}
	return strings.Join(parts, "")
}

// stepCELTexts are a step's own CEL expressions: if, cel, wait and retry.
func stepCELTexts(spec StepSpec) []string {
	texts := []string{spec.If, spec.CEL}
	if spec.Wait != nil {
		texts = append(texts, spec.Wait.Until, spec.Wait.FailWhen)
	}
	if spec.Retry != nil {
		texts = append(texts, spec.Retry.When)
	}
	return texts
}

// stepTexts are the strings of a step that can reference other steps: its
// config, its CEL expressions and its export templates.
func stepTexts(spec StepSpec) []string {
	texts := append([]string{strings.ReplaceAll(string(spec.Config), `\"`, `"`)}, stepCELTexts(spec)...)
	for _, value := range spec.Export {
		texts = append(texts, value)
	}
	return texts
}

// stepReferences lists the step ids a step's config, if, CEL, wait, retry and
// export templates refer to, sorted and de-duplicated.
func stepReferences(spec StepSpec) []string {
	found := map[string]bool{}
	for _, text := range stepTexts(spec) {
		for _, pattern := range stepRefPatterns {
			for _, match := range pattern.FindAllStringSubmatch(text, -1) {
				found[match[1]] = true
			}
		}
	}
	refs := make([]string, 0, len(found))
	for ref := range found {
		refs = append(refs, ref)
	}
	sort.Strings(refs)
	return refs
}
