package commit

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	clickyai "github.com/flanksource/clicky/ai"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gomplate/v3"
	"github.com/ghodss/yaml"
)

//go:embed ai-commit-group.md
var commitGroupPrompt string

type commitGroupSpec struct {
	Label string   `json:"label,omitempty" description:"Short human-readable label for the commit group"`
	Files []string `json:"files" description:"Ordered list of repo-relative file paths assigned to this commit"`
}

type commitGroupPlanSchema struct {
	Groups []commitGroupSpec `json:"groups" description:"Ordered commit groups covering every changed file exactly once"`
}

type commitGroup struct {
	Label   string
	Changes []stagedChange
}

func (g commitGroup) Files() []string {
	files := make([]string, 0, len(g.Changes))
	for _, change := range g.Changes {
		files = append(files, change.Path)
	}
	return files
}

func (g commitGroup) GitPaths() []string {
	seen := make(map[string]struct{}, len(g.Changes)*2)
	var paths []string
	for _, change := range g.Changes {
		for _, path := range change.GitPaths() {
			if _, ok := seen[path]; ok {
				continue
			}
			seen[path] = struct{}{}
			paths = append(paths, path)
		}
	}
	return paths
}

func (g commitGroup) diff() string {
	var patches []string
	for _, change := range g.Changes {
		patches = append(patches, strings.TrimRight(change.Patch, "\n"))
	}
	if len(patches) == 0 {
		return ""
	}
	return strings.Join(patches, "\n") + "\n"
}

func (g commitGroup) labelOrDefault() string {
	if g.Label != "" {
		return g.Label
	}
	if len(g.Changes) == 1 {
		return g.Changes[0].Path
	}
	return fmt.Sprintf("%d files", len(g.Changes))
}

func planCommitGroups(ctx context.Context, opts Options, changes []stagedChange) ([]commitGroupSpec, error) {
	if len(changes) == 0 {
		return nil, nil
	}
	if os.Getenv(testEnvVar) == "1" {
		groups := make([]commitGroupSpec, 0, len(changes))
		for _, change := range changes {
			groups = append(groups, commitGroupSpec{
				Label: filepath.Base(change.Path),
				Files: []string{change.Path},
			})
		}
		return mergeGroupsToMax(groups, opts.Max), nil
	}

	if commitGroupPrompt == "" {
		return nil, fmt.Errorf("commit group prompt template is empty")
	}

	templateData := map[string]any{
		"changes": changes,
	}
	if opts.Max > 0 {
		templateData["max"] = opts.Max
	}
	prompt, err := gomplate.RunTemplate(templateData, gomplate.Template{Template: commitGroupPrompt})
	if err != nil {
		return nil, fmt.Errorf("render commit group prompt: %w", err)
	}

	agent, err := buildAgent(opts)
	if err != nil {
		return nil, err
	}
	resp, err := agent.ExecutePrompt(ctx, clickyai.PromptRequest{
		Name:             "Commit grouping plan",
		Prompt:           prompt,
		StructuredOutput: &commitGroupPlanSchema{},
	})
	if err != nil {
		return nil, fmt.Errorf("execute commit group prompt: %w", err)
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("commit group prompt returned error: %s", resp.Error)
	}

	if groups, ok, err := parseCommitGroupResponse(resp.StructuredData, resp.Result); err != nil {
		return nil, fmt.Errorf("parse commit group response: %w", err)
	} else if ok && len(groups) > 0 {
		return groups, nil
	}

	logger.Warnf("commit grouping prompt did not populate structured output, raw result=%q structuredData=%+v", resp.Result, resp.StructuredData)
	return nil, fmt.Errorf("commit grouping prompt returned no groups")
}

func parseCommitGroupResponse(data any, raw string) ([]commitGroupSpec, bool, error) {
	if len(strings.TrimSpace(raw)) == 0 && data == nil {
		return nil, false, nil
	}

	var schema commitGroupPlanSchema
	if ok, err := decodeCommitGroupSchema(data, &schema); err != nil {
		return nil, false, err
	} else if ok && len(schema.Groups) > 0 {
		return schema.Groups, true, nil
	}

	cleaned := cleanStructuredResult(raw)
	if cleaned == "" {
		return nil, false, nil
	}
	if ok, err := decodeCommitGroupSchema([]byte(cleaned), &schema); err != nil {
		return nil, false, err
	} else if ok && len(schema.Groups) > 0 {
		return schema.Groups, true, nil
	}

	return nil, false, nil
}

func decodeCommitGroupSchema(input any, schema *commitGroupPlanSchema) (bool, error) {
	if input == nil {
		return false, nil
	}

	switch v := input.(type) {
	case *commitGroupPlanSchema:
		if v == nil {
			return false, nil
		}
		*schema = *v
		return true, nil
	case commitGroupPlanSchema:
		*schema = v
		return true, nil
	case []byte:
		if err := yaml.Unmarshal(v, schema); err != nil {
			return false, err
		}
		return true, nil
	case string:
		if err := yaml.Unmarshal([]byte(v), schema); err != nil {
			return false, err
		}
		return true, nil
	default:
		data, err := yaml.Marshal(v)
		if err != nil {
			return false, err
		}
		if err := yaml.Unmarshal(data, schema); err != nil {
			return false, err
		}
		return true, nil
	}
}

func cleanStructuredResult(result string) string {
	result = strings.TrimSpace(result)
	switch {
	case strings.HasPrefix(result, "```yaml"):
		result = strings.TrimPrefix(result, "```yaml")
	case strings.HasPrefix(result, "```json"):
		result = strings.TrimPrefix(result, "```json")
	case strings.HasPrefix(result, "```"):
		result = strings.TrimPrefix(result, "```")
	}
	result = strings.TrimSuffix(result, "```")
	return strings.TrimSpace(result)
}

func validateCommitPlan(specs []commitGroupSpec, changes []stagedChange) ([]commitGroup, error) {
	if len(specs) == 0 {
		return nil, fmt.Errorf("%w: planner returned no groups", ErrInvalidCommitAllPlan)
	}

	available := make(map[string]stagedChange, len(changes))
	for _, change := range changes {
		available[change.Path] = change
	}

	seen := make(map[string]struct{}, len(changes))
	groups := make([]commitGroup, 0, len(specs))
	for i, spec := range specs {
		if len(spec.Files) == 0 {
			return nil, fmt.Errorf("%w: group %d is empty", ErrInvalidCommitAllPlan, i+1)
		}

		group := commitGroup{Label: strings.TrimSpace(spec.Label)}
		groupSeen := make(map[string]struct{}, len(spec.Files))
		for _, file := range spec.Files {
			file = strings.TrimSpace(file)
			if file == "" {
				return nil, fmt.Errorf("%w: group %d contains an empty file entry", ErrInvalidCommitAllPlan, i+1)
			}
			if _, ok := groupSeen[file]; ok {
				return nil, fmt.Errorf("%w: file %s appears multiple times in group %d", ErrInvalidCommitAllPlan, file, i+1)
			}
			groupSeen[file] = struct{}{}
			if _, ok := seen[file]; ok {
				return nil, fmt.Errorf("%w: file %s appears in multiple groups", ErrInvalidCommitAllPlan, file)
			}
			change, ok := available[file]
			if !ok {
				return nil, fmt.Errorf("%w: file %s is not part of the staged change set", ErrInvalidCommitAllPlan, file)
			}
			seen[file] = struct{}{}
			group.Changes = append(group.Changes, change)
		}
		groups = append(groups, group)
	}

	if len(seen) != len(available) {
		missing := make([]string, 0, len(available)-len(seen))
		for file := range available {
			if _, ok := seen[file]; !ok {
				missing = append(missing, file)
			}
		}
		sort.Strings(missing)
		return nil, fmt.Errorf("%w: missing files from plan: %s", ErrInvalidCommitAllPlan, strings.Join(missing, ", "))
	}

	return groups, nil
}

// mergeGroupsToMax collapses trailing groups into the last kept group
// so the total count does not exceed max. If max <= 0, returns groups unchanged.
func mergeGroupsToMax(groups []commitGroupSpec, max int) []commitGroupSpec {
	if max <= 0 || len(groups) <= max {
		return groups
	}
	merged := make([]commitGroupSpec, max)
	copy(merged, groups[:max])
	for _, g := range groups[max:] {
		merged[max-1].Files = append(merged[max-1].Files, g.Files...)
	}
	return merged
}
