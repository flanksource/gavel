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
	exporters := map[string]string{}
	for _, step := range steps {
		if err := validateStep(step, seen, exporters); err != nil {
			return fmt.Errorf("%s: %w", step.label, err)
		}
		if step.spec.ID != "" {
			seen[step.spec.ID] = true
		}
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
	for _, ref := range stepReferences(spec) {
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

// stepReferences lists the step ids a step's config, CEL, wait, retry and
// export templates refer to, sorted and de-duplicated.
func stepReferences(spec StepSpec) []string {
	texts := []string{strings.ReplaceAll(string(spec.Config), `\"`, `"`), spec.CEL}
	if spec.Wait != nil {
		texts = append(texts, spec.Wait.Until, spec.Wait.FailWhen)
	}
	if spec.Retry != nil {
		texts = append(texts, spec.Retry.When)
	}
	for _, value := range spec.Export {
		texts = append(texts, value)
	}
	found := map[string]bool{}
	for _, text := range texts {
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
