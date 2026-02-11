package kubernetes

import (
	"bufio"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/git/rules"
	. "github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/models/kubernetes"
	"github.com/flanksource/gavel/repomap"
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/commons/logger"
	"github.com/mattbaird/jsonpatch"
)

type AnalyzerContext interface {
	ReadFile(path, commit string) (string, error)
	GetSeverityConfig() *rules.SeverityConfig
	GetSeverityEngine() *rules.Engine
}

func FindAffectedDocuments(documents []kubernetes.YAMLDocument, changedLines []int) []int {
	var affected []int

	for i, doc := range documents {
		for _, line := range changedLines {
			if line >= doc.StartLine && line <= doc.EndLine {
				affected = append(affected, i)
				break
			}
		}
	}

	return affected
}

type ChangedLines struct {
	AddedLines   []int
	DeletedLines []int
}

func ExtractChangedLines(patch string) ChangedLines {
	result := ChangedLines{
		AddedLines:   []int{},
		DeletedLines: []int{},
	}
	scanner := bufio.NewScanner(strings.NewReader(patch))
	oldLineNum := 0
	newLineNum := 0

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "@@") {
			parts := strings.Split(line, " ")
			if len(parts) >= 3 {
				oldRange := strings.TrimPrefix(parts[1], "-")
				newRange := strings.TrimPrefix(parts[2], "+")
				_, _ = fmt.Sscanf(oldRange, "%d", &oldLineNum)
				_, _ = fmt.Sscanf(newRange, "%d", &newLineNum)
			}
			continue
		}

		if strings.HasPrefix(line, "diff") || strings.HasPrefix(line, "index") ||
			strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") {
			continue
		}

		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			result.AddedLines = append(result.AddedLines, newLineNum)
			newLineNum++
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			result.DeletedLines = append(result.DeletedLines, oldLineNum)
			oldLineNum++
		} else {
			oldLineNum++
			newLineNum++
		}
	}

	return result
}

type JSONPatches []kubernetes.ExtendedPatch

func jsonPointerToDotNotation(path string) string {
	path = strings.TrimPrefix(path, "/")
	return strings.ReplaceAll(path, "/", ".")
}

func formatValue(val interface{}) string {
	if val == nil {
		return "null"
	}
	switch v := val.(type) {
	case string:
		return v
	case int, int32, int64:
		return fmt.Sprintf("%d", v)
	case float64:
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%v", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		if jsonBytes, err := json.Marshal(v); err == nil {
			return string(jsonBytes)
		}
		return fmt.Sprintf("%v", v)
	}
}

func (j JSONPatches) Pretty() api.Text {
	if len(j) == 0 {
		return clicky.Text("No changes", "muted")
	}

	changes := []string{}
	for _, patch := range j {
		dotPath := jsonPointerToDotNotation(patch.Path)

		switch patch.Operation {
		case "add":
			newVal := formatValue(patch.Value)
			changes = append(changes, fmt.Sprintf("%s = %s", dotPath, newVal))
		case "remove":
			oldVal := formatValue(patch.OldValue)
			changes = append(changes, fmt.Sprintf("%s %s removed", dotPath, oldVal))
		case "replace":
			oldVal := formatValue(patch.OldValue)
			newVal := formatValue(patch.Value)
			changes = append(changes, fmt.Sprintf("%s %s → %s", dotPath, oldVal, newVal))
		}
	}

	return clicky.Text(strings.Join(changes, ", "))
}

func GenerateJSONPatches(before, after map[string]interface{}) (JSONPatches, error) {
	beforeJSON, err := json.Marshal(before)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal before JSON: %w", err)
	}

	afterJSON, err := json.Marshal(after)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal after JSON: %w", err)
	}

	patches, err := jsonpatch.CreatePatch(beforeJSON, afterJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to create JSON patch: %w", err)
	}

	extendedPatches := make([]kubernetes.ExtendedPatch, len(patches))
	for i, patch := range patches {
		extendedPatches[i] = kubernetes.ExtendedPatch{
			Operation: patch.Operation,
			Path:      patch.Path,
			Value:     patch.Value,
		}

		if patch.Operation == "replace" || patch.Operation == "remove" {
			extendedPatches[i].OldValue = getValueFromPath(before, patch.Path)
		}
	}

	return extendedPatches, nil
}

func DetermineChangeType(before, after map[string]interface{}) kubernetes.SourceChangeType {
	if len(before) == 0 {
		return kubernetes.SourceChangeTypeAdded
	}
	if len(after) == 0 {
		return kubernetes.SourceChangeTypeDeleted
	}
	return kubernetes.SourceChangeTypeModified
}

func DetermineChangeSeverity(changeType kubernetes.SourceChangeType, patches JSONPatches, versionChanges []kubernetes.VersionChange) kubernetes.ChangeSeverity {
	if changeType == kubernetes.SourceChangeTypeDeleted {
		return kubernetes.ChangeSeverityHigh
	}

	for _, vc := range versionChanges {
		if vc.ValueType == "sha256" || vc.ValueType == "git-sha" || vc.ValueType == "combined" {
			return kubernetes.ChangeSeverityMedium
		}
	}

	for _, patch := range patches {
		path := patch.Path

		if strings.Contains(path, "/replicas") ||
			strings.Contains(path, "/image") ||
			strings.Contains(path, "/resources/limits") ||
			strings.Contains(path, "/resources/requests") {
			return kubernetes.ChangeSeverityMedium
		}

		if strings.Contains(path, "/env") ||
			strings.Contains(path, "/volumes") ||
			strings.Contains(path, "/volumeMounts") {
			return kubernetes.ChangeSeverityMedium
		}
	}

	return kubernetes.ChangeSeverityLow
}

func GenerateSummary(ref kubernetes.KubernetesRef, changeType kubernetes.SourceChangeType, patches JSONPatches) api.Text {
	t := clicky.Text("")

	switch changeType {
	case kubernetes.SourceChangeTypeAdded, kubernetes.SourceChangeTypeDeleted:
		return t.Append(changeType.Pretty())
	}

	return t.Append(patches.Pretty())
}

func ExtractFieldPaths(patches []kubernetes.ExtendedPatch) []string {
	var paths []string
	for _, patch := range patches {
		paths = append(paths, jsonPointerToDotNotation(patch.Path))
	}
	return paths
}

func AnalyzeKubernetesChanges(ctx AnalyzerContext, commit Commit, change *CommitChange) error {
	logger.Tracef("[kubernetes] analyzing %s @ %s", change.File, commit.Hash)

	if !repomap.IsYaml(change.File) {
		return nil
	}

	severityEngine := ctx.GetSeverityEngine()

	beforeCommit := commit.Hash + "^"
	afterCommit := commit.Hash

	var beforeContent, afterContent string
	var err error

	if change.Type != SourceChangeTypeAdded {
		beforeContent, err = ctx.ReadFile(change.File, beforeCommit)
		if err != nil {
			logger.Errorf("Error reading before %s:%s %w", change.File, beforeCommit, err)
			return nil
		}
	}

	if change.Type != SourceChangeTypeDeleted {
		afterContent, err = ctx.ReadFile(change.File, afterCommit)
		if err != nil {
			logger.Errorf("Error reading after %s:%s %w", change.File, afterCommit, err)
			return nil
		}
	}

	var beforeDocs, afterDocs []kubernetes.YAMLDocument

	if beforeContent != "" {
		beforeDocs, err = repomap.ParseYAMLDocuments(beforeContent)
		if err != nil {
			return fmt.Errorf("Error parsing before %s:%s %w", change.File, beforeCommit, err)
		}
	}

	if afterContent != "" {
		afterDocs, err = repomap.ParseYAMLDocuments(afterContent)
		if err != nil {
			return fmt.Errorf("Error parsing after %s:%s %w", change.File, afterCommit, err)
		}
	}

	filePatch := commit.GetFilePatch(change.File)
	changedLineInfo := ExtractChangedLines(filePatch)

	affectedIndices := FindAffectedDocuments(afterDocs, changedLineInfo.AddedLines)

	changedLinesSet := make(map[int]struct{}, len(changedLineInfo.AddedLines))

	for _, idx := range affectedIndices {
		if idx >= len(afterDocs) {
			continue
		}

		afterDoc := afterDocs[idx]

		if !repomap.IsKubernetesResource(afterDoc.Content) {
			continue
		}

		afterRef := repomap.ExtractKubernetesRef(afterDoc)
		var beforeDoc *kubernetes.YAMLDocument
		for i := range beforeDocs {
			if !repomap.IsKubernetesResource(beforeDocs[i].Content) {
				continue
			}
			beforeRef := repomap.ExtractKubernetesRef(beforeDocs[i])

			if beforeRef.Kind == afterRef.Kind &&
				beforeRef.Name == afterRef.Name &&
				beforeRef.Namespace == afterRef.Namespace {
				beforeDoc = &beforeDocs[i]
				break
			}
		}

		k8sChange, err := createKubernetesChange(commit, change, beforeDoc, afterDoc, severityEngine)
		if err != nil {
			return fmt.Errorf("Error creating Kubernetes change for %s:%s %w", change.File, afterCommit, err)
		}

		for line := afterDoc.StartLine; line <= afterDoc.EndLine; line++ {
			changedLinesSet[line] = struct{}{}
		}

		change.KubernetesChanges = append(change.KubernetesChanges, k8sChange)
	}

	if len(changedLineInfo.DeletedLines) > 0 {
		deletedIndices := FindAffectedDocuments(beforeDocs, changedLineInfo.DeletedLines)

		for _, idx := range deletedIndices {
			if idx >= len(beforeDocs) {
				continue
			}

			beforeDoc := beforeDocs[idx]
			if !repomap.IsKubernetesResource(beforeDoc.Content) {
				continue
			}

			if idx < len(afterDocs) && repomap.IsKubernetesResource(afterDocs[idx].Content) {
				continue
			}

			k8sChange, err := createKubernetesChange(commit, change, &beforeDoc, kubernetes.YAMLDocument{}, severityEngine)
			if err != nil {
				return fmt.Errorf("Error creating Kubernetes change for deleted %s:%s %w", change.File, beforeCommit, err)
			}

			for line := beforeDoc.StartLine; line <= beforeDoc.EndLine; line++ {
				changedLinesSet[line] = struct{}{}
			}

			change.KubernetesChanges = append(change.KubernetesChanges, k8sChange)
		}
	}

	if len(changedLinesSet) > 0 {
		lines := make([]int, 0, len(changedLinesSet))
		for line := range changedLinesSet {
			lines = append(lines, line)
		}
		change.LinesChanged = NewLineRanges(lines)
	}

	if len(change.KubernetesChanges) > 0 {
		var severities []Severity
		for _, k8sChange := range change.KubernetesChanges {
			severities = append(severities, Severity(k8sChange.Severity))
		}
		change.Severity = MaxSeverities(severities)
	}

	return nil
}

func createKubernetesChange(commit Commit, change *CommitChange, beforeDoc *kubernetes.YAMLDocument, afterDoc kubernetes.YAMLDocument, engine *rules.Engine) (kubernetes.KubernetesChange, error) {
	refDoc := afterDoc
	if afterDoc.Content == nil && beforeDoc != nil {
		refDoc = *beforeDoc
	}

	ref := repomap.ExtractKubernetesRef(refDoc)

	ref.StartLine = refDoc.StartLine
	ref.EndLine = refDoc.EndLine

	var changeType kubernetes.SourceChangeType
	var beforeContent, afterContent map[string]interface{}

	if beforeDoc == nil || beforeDoc.Content == nil {
		changeType = kubernetes.SourceChangeTypeAdded
		afterContent = afterDoc.Content
	} else if afterDoc.Content == nil {
		changeType = kubernetes.SourceChangeTypeDeleted
		beforeContent = beforeDoc.Content
	} else {
		changeType = kubernetes.SourceChangeTypeModified
		beforeContent = beforeDoc.Content
		afterContent = afterDoc.Content
	}

	var patches []kubernetes.ExtendedPatch
	if changeType == kubernetes.SourceChangeTypeModified {
		var err error
		patches, err = GenerateJSONPatches(beforeContent, afterContent)
		if err != nil {
			patches = []kubernetes.ExtendedPatch{}
		}
	}

	fieldPaths := ExtractFieldPaths(patches)

	sourceType := DetermineSourceType(change.File, afterContent)

	var scaling *kubernetes.Scaling
	var versionChanges []kubernetes.VersionChange
	var envChange *kubernetes.EnvironmentChange

	if changeType == kubernetes.SourceChangeTypeModified {
		scaling = ExtractScalingChanges(patches, beforeContent, afterContent)
		versionChanges = ExtractAllVersionChanges(patches, beforeContent, afterContent, nil)
		envChange = ExtractEnvironmentChanges(patches, beforeContent, afterContent)
	}

	k8sChange := kubernetes.KubernetesChange{
		KubernetesRef:     ref,
		ChangeType:        changeType,
		SourceType:        sourceType,
		Patches:           patches,
		Scaling:           scaling,
		VersionChanges:    versionChanges,
		EnvironmentChange: envChange,
		Before:            beforeContent,
		After:             afterContent,
		FieldsChanged:     fieldPaths,
		FieldChangeCount:  len(patches),
	}

	if engine != nil {
		ctx := rules.BuildContext(nil, change, &k8sChange)
		k8sChange.Severity = kubernetes.ChangeSeverity(engine.Evaluate(ctx))
	} else {
		k8sChange.Severity = DetermineChangeSeverity(changeType, patches, versionChanges)
	}

	return k8sChange, nil
}

func DetermineSourceType(filePath string, content map[string]interface{}) kubernetes.KubernetesSourceType {
	if strings.Contains(filePath, "kustomize") || filepath.Base(filePath) == "kustomization.yaml" {
		return kubernetes.Kustomize
	}

	if strings.Contains(filePath, "helm") || strings.Contains(filePath, "charts") || strings.Contains(filePath, "templates") {
		return kubernetes.Helm
	}

	if content != nil {
		if apiVersion, ok := content["apiVersion"].(string); ok {
			if strings.Contains(apiVersion, "flux") {
				return kubernetes.Flux
			}
			if strings.Contains(apiVersion, "argoproj.io") {
				return kubernetes.ArgoCD
			}
		}
	}

	return kubernetes.YAML
}
