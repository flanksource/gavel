package launch

import "github.com/flanksource/gavel/resultstore"

func linterProvenance(config *LintConfig) []resultstore.LinterConfiguration {
	if config == nil {
		return nil
	}
	if len(config.Linters) == 0 {
		return []resultstore.LinterConfiguration{{ID: "configured", Engine: "discovery"}}
	}
	result := make([]resultstore.LinterConfiguration, 0, len(config.Linters))
	for _, entry := range config.Linters {
		item := resultstore.LinterConfiguration{ID: entry.ID, Engine: entry.Linter}
		if entry.Linter == "prompt" {
			item.Prompt, _ = entry.Options["prompt"].(string)
			item.PromptVersion, _ = entry.Options["promptVersion"].(string)
			item.Presets = append(item.Presets, stringValues(entry.Options["presets"])...)
		}
		result = append(result, item)
	}
	return result
}

func stringValues(value any) []string {
	if values, ok := value.([]string); ok {
		return values
	}
	var result []string
	if values, ok := value.([]any); ok {
		for _, value := range values {
			if text, ok := value.(string); ok {
				result = append(result, text)
			}
		}
	}
	return result
}
