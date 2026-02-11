package kubernetes

import (
	"fmt"
	"strings"

	"github.com/flanksource/gavel/models/kubernetes"
	"github.com/flanksource/gavel/utils"
)

func ExtractScalingChanges(patches []kubernetes.ExtendedPatch, before, after map[string]interface{}) *kubernetes.Scaling {
	var scaling kubernetes.Scaling
	hasChanges := false

	for _, patch := range patches {
		path := patch.Path

		if strings.Contains(path, "/replicas") {
			hasChanges = true
			if before != nil {
				if oldReplicas := getIntFromPath(before, path); oldReplicas != nil {
					scaling.Replicas = oldReplicas
				}
			}
			if after != nil {
				if newReplicas := getIntFromPath(after, path); newReplicas != nil {
					scaling.NewReplicas = newReplicas
				}
			}
		}

		if strings.Contains(path, "/resources") && strings.Contains(path, "/cpu") {
			hasChanges = true
			if before != nil {
				scaling.OldCPU = getStringFromPath(before, path)
			}
			if after != nil {
				scaling.NewCPU = getStringFromPath(after, path)
			}
		}

		if strings.Contains(path, "/resources") && strings.Contains(path, "/memory") {
			hasChanges = true
			if before != nil {
				scaling.OldMemory = getStringFromPath(before, path)
			}
			if after != nil {
				scaling.NewMemory = getStringFromPath(after, path)
			}
		}
	}

	if hasChanges {
		return &scaling
	}
	return nil
}

func ExtractAllVersionChanges(patches []kubernetes.ExtendedPatch, before, after map[string]interface{}, patterns []string) []kubernetes.VersionChange {
	var versionChanges []kubernetes.VersionChange

	if len(patterns) == 0 {
		patterns = []string{"**.image", "**.tag", "**.version", "**.appVersion", "**.imageTag"}
	}

	for _, patch := range patches {
		fieldPath := strings.TrimPrefix(patch.Path, "/")
		fieldPath = strings.ReplaceAll(fieldPath, "/", ".")

		matchesPattern := false
		for _, pattern := range patterns {
			if utils.MatchesPattern(fieldPath, pattern) {
				matchesPattern = true
				break
			}
		}

		if !matchesPattern {
			continue
		}

		oldValue := getStringFromPath(before, patch.Path)
		newValue := getStringFromPath(after, patch.Path)

		if oldValue == "" && newValue == "" {
			continue
		}

		vc := detectVersionChange(oldValue, newValue, fieldPath)
		if vc != nil {
			versionChanges = append(versionChanges, *vc)
		}
	}

	return versionChanges
}

func detectVersionChange(oldValue, newValue, fieldPath string) *kubernetes.VersionChange {
	if strings.Contains(oldValue, "/") || strings.Contains(newValue, "/") ||
		strings.Contains(oldValue, "@") || strings.Contains(newValue, "@") ||
		strings.Contains(oldValue, ":") || strings.Contains(newValue, ":") {
		return detectImageVersionChange(oldValue, newValue, fieldPath)
	}

	if utils.IsSemver(oldValue) || utils.IsSemver(newValue) {
		vc := kubernetes.AnalyzeVersionChange(oldValue, newValue)
		vc.FieldPath = fieldPath
		vc.ValueType = "semver"
		return &vc
	}

	if utils.IsSHA256(oldValue) || utils.IsSHA256(newValue) {
		return &kubernetes.VersionChange{
			OldVersion: oldValue,
			NewVersion: newValue,
			FieldPath:  fieldPath,
			ValueType:  "sha256",
			ChangeType: kubernetes.VersionChangeUnknown,
		}
	}

	if utils.IsGitSHA(oldValue) || utils.IsGitSHA(newValue) {
		return &kubernetes.VersionChange{
			OldVersion: oldValue,
			NewVersion: newValue,
			FieldPath:  fieldPath,
			ValueType:  "git-sha",
			ChangeType: kubernetes.VersionChangeUnknown,
		}
	}

	return nil
}

func detectImageVersionChange(oldImage, newImage, fieldPath string) *kubernetes.VersionChange {
	oldRef := utils.ParseImageReference(oldImage)
	newRef := utils.ParseImageReference(newImage)

	var oldVersion, newVersion string
	var digest string
	valueType := "unknown"

	if oldRef.Digest != "" || newRef.Digest != "" {
		digest = newRef.Digest
		if oldRef.Tag != "" || newRef.Tag != "" {
			oldVersion = oldRef.Tag
			newVersion = newRef.Tag
			valueType = "combined"
		} else {
			oldVersion = oldRef.Digest
			newVersion = newRef.Digest
			valueType = "sha256"
		}
	} else {
		oldVersion = oldRef.Tag
		newVersion = newRef.Tag
		valueType = "semver"
	}

	vc := kubernetes.AnalyzeVersionChange(oldVersion, newVersion)
	vc.FieldPath = fieldPath
	vc.ValueType = valueType
	vc.Digest = digest

	return &vc
}

func ExtractEnvironmentChanges(patches []kubernetes.ExtendedPatch, before, after map[string]interface{}) *kubernetes.EnvironmentChange {
	oldEnv := extractEnvVars(before)
	newEnv := extractEnvVars(after)

	if len(oldEnv) > 0 || len(newEnv) > 0 {
		return &kubernetes.EnvironmentChange{
			Old: oldEnv,
			New: newEnv,
		}
	}
	return nil
}

func getIntFromPath(m map[string]interface{}, path string) *int {
	val := getValueFromPath(m, path)
	if intVal, ok := val.(int); ok {
		return &intVal
	}
	if floatVal, ok := val.(float64); ok {
		intVal := int(floatVal)
		return &intVal
	}
	if uintVal, ok := val.(uint64); ok {
		intVal := int(uintVal)
		return &intVal
	}
	if int64Val, ok := val.(int64); ok {
		intVal := int(int64Val)
		return &intVal
	}
	return nil
}

func getStringFromPath(m map[string]interface{}, path string) string {
	val := getValueFromPath(m, path)
	if strVal, ok := val.(string); ok {
		return strVal
	}
	return ""
}

func getValueFromPath(m map[string]interface{}, path string) interface{} {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	current := interface{}(m)

	for _, part := range parts {
		if currentMap, ok := current.(map[string]interface{}); ok {
			current = currentMap[part]
		} else if currentArray, ok := current.([]interface{}); ok {
			var index int
			_, err := fmt.Sscanf(part, "%d", &index)
			if err != nil || index < 0 || index >= len(currentArray) {
				return nil
			}
			current = currentArray[index]
		} else {
			return nil
		}
	}
	return current
}

func extractEnvVars(content map[string]interface{}) map[string]string {
	envVars := make(map[string]string)
	if content == nil {
		return envVars
	}

	if spec, ok := content["spec"].(map[string]interface{}); ok {
		if containers, ok := spec["containers"].([]interface{}); ok {
			for _, container := range containers {
				if containerMap, ok := container.(map[string]interface{}); ok {
					if env, ok := containerMap["env"].([]interface{}); ok {
						for _, envVar := range env {
							if envMap, ok := envVar.(map[string]interface{}); ok {
								if name, ok := envMap["name"].(string); ok {
									if value, ok := envMap["value"].(string); ok {
										envVars[name] = value
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return envVars
}
