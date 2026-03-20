package kubernetes

import (
	"bufio"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/repomap"
	repomapcel "github.com/flanksource/repomap/cel"
	repomapk8s "github.com/flanksource/repomap/kubernetes"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/models/kubernetes"
	"github.com/mattbaird/jsonpatch"
)

type AnalyzerContext interface {
	ReadFile(path, commit string) (string, error)
	GetSeverityConfig() *repomap.SeverityConfig
	GetSeverityEngine() *repomapcel.Engine
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

func AnalyzeKubernetesChanges(ctx AnalyzerContext, commit models.Commit, change *models.CommitChange) error {
	logger.Tracef("[kubernetes] analyzing %s @ %s", change.File, commit.Hash)

	if !repomapk8s.IsYaml(change.File) {
		return nil
	}

	severityEngine := ctx.GetSeverityEngine()

	beforeCommit := commit.Hash + "^"
	afterCommit := commit.Hash

	var beforeContent, afterContent string
	var err error

	if change.Type != models.SourceChangeTypeAdded {
		beforeContent, err = ctx.ReadFile(change.File, beforeCommit)
		if err != nil {
			logger.Errorf("Error reading before %s:%s %w", change.File, beforeCommit, err)
			return nil
		}
	}

	if change.Type != models.SourceChangeTypeDeleted {
		afterContent, err = ctx.ReadFile(change.File, afterCommit)
		if err != nil {
			logger.Errorf("Error reading after %s:%s %w", change.File, afterCommit, err)
			return nil
		}
	}

	var beforeDocs, afterDocs []kubernetes.YAMLDocument

	if beforeContent != "" {
		beforeDocs, err = parseYAMLDocuments(beforeContent)
		if err != nil {
			return fmt.Errorf("error parsing before %s:%s %w", change.File, beforeCommit, err)
		}
	}

	if afterContent != "" {
		afterDocs, err = parseYAMLDocuments(afterContent)
		if err != nil {
			return fmt.Errorf("error parsing after %s:%s %w", change.File, afterCommit, err)
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

		if !repomapk8s.IsKubernetesResource(afterDoc.Content) {
			continue
		}

		afterRef := extractRef(afterDoc)
		var beforeDoc *kubernetes.YAMLDocument
		for i := range beforeDocs {
			if !repomapk8s.IsKubernetesResource(beforeDocs[i].Content) {
				continue
			}
			beforeRef := extractRef(beforeDocs[i])

			if beforeRef.Kind == afterRef.Kind &&
				beforeRef.Name == afterRef.Name &&
				beforeRef.Namespace == afterRef.Namespace {
				beforeDoc = &beforeDocs[i]
				break
			}
		}

		k8sChange, err := createKubernetesChange(commit, change, beforeDoc, afterDoc, severityEngine)
		if err != nil {
			return fmt.Errorf("error creating kubernetes change for %s:%s %w", change.File, afterCommit, err)
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
			if !repomapk8s.IsKubernetesResource(beforeDoc.Content) {
				continue
			}

			if idx < len(afterDocs) && repomapk8s.IsKubernetesResource(afterDocs[idx].Content) {
				continue
			}

			k8sChange, err := createKubernetesChange(commit, change, &beforeDoc, kubernetes.YAMLDocument{}, severityEngine)
			if err != nil {
				return fmt.Errorf("error creating kubernetes change for deleted %s:%s %w", change.File, beforeCommit, err)
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
		change.LinesChanged = models.NewLineRanges(lines)
	}

	if len(change.KubernetesChanges) > 0 {
		var severities []models.Severity
		for _, k8sChange := range change.KubernetesChanges {
			severities = append(severities, models.Severity(k8sChange.Severity))
		}
		change.Severity = models.MaxSeverities(severities)
	}

	return nil
}

func createKubernetesChange(commit models.Commit, change *models.CommitChange, beforeDoc *kubernetes.YAMLDocument, afterDoc kubernetes.YAMLDocument, engine *repomapcel.Engine) (kubernetes.KubernetesChange, error) {
	refDoc := afterDoc
	if afterDoc.Content == nil && beforeDoc != nil {
		refDoc = *beforeDoc
	}

	ref := extractRef(refDoc)

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
		ctx := repomapcel.BuildContext(nil, toRepomapChange(change), toRepomapK8sChange(&k8sChange))
		k8sChange.Severity = kubernetes.ChangeSeverity(engine.Evaluate(ctx))
	} else {
		k8sChange.Severity = DetermineChangeSeverity(changeType, patches, versionChanges)
	}

	return k8sChange, nil
}

func parseYAMLDocuments(content string) ([]kubernetes.YAMLDocument, error) {
	docs, err := repomapk8s.ParseYAMLDocuments(content)
	if err != nil {
		return nil, err
	}
	result := make([]kubernetes.YAMLDocument, len(docs))
	for i, d := range docs {
		result[i] = kubernetes.YAMLDocument{StartLine: d.StartLine, EndLine: d.EndLine, Content: d.Content}
	}
	return result, nil
}

func extractRef(doc kubernetes.YAMLDocument) kubernetes.KubernetesRef {
	r := repomapk8s.ExtractKubernetesRef(repomapk8s.YAMLDocument{
		StartLine: doc.StartLine, EndLine: doc.EndLine, Content: doc.Content,
	})
	return kubernetes.KubernetesRef{
		APIVersion: r.APIVersion, Kind: r.Kind, Namespace: r.Namespace, Name: r.Name,
		JSONPath: r.JSONPath, StartLine: r.StartLine, EndLine: r.EndLine,
		Labels: r.Labels, Annotations: r.Annotations,
	}
}

func toRepomapChange(c *models.CommitChange) *repomap.CommitChange {
	if c == nil {
		return nil
	}
	rc := &repomap.CommitChange{
		File: c.File, Type: repomap.SourceChangeType(c.Type),
		Adds: c.Adds, Dels: c.Dels,
	}
	for _, kc := range c.KubernetesChanges {
		rc.KubernetesChanges = append(rc.KubernetesChanges, *toRepomapK8sChange(&kc))
	}
	return rc
}

func toRepomapK8sChange(kc *kubernetes.KubernetesChange) *repomapk8s.KubernetesChange {
	if kc == nil {
		return nil
	}
	rk := &repomapk8s.KubernetesChange{
		KubernetesRef: repomapk8s.KubernetesRef{
			APIVersion: kc.APIVersion, Kind: kc.Kind, Namespace: kc.Namespace, Name: kc.Name,
			StartLine: kc.StartLine, EndLine: kc.EndLine, Labels: kc.Labels, Annotations: kc.Annotations,
		},
		ChangeType:       repomapk8s.SourceChangeType(kc.ChangeType),
		SourceType:       repomapk8s.KubernetesSourceType(kc.SourceType),
		Severity:         repomapk8s.ChangeSeverity(kc.Severity),
		Before:           kc.Before,
		After:            kc.After,
		FieldsChanged:    kc.FieldsChanged,
		FieldChangeCount: kc.FieldChangeCount,
	}
	for _, p := range kc.Patches {
		rk.Patches = append(rk.Patches, repomapk8s.ExtendedPatch{
			Operation: p.Operation, Path: p.Path, Value: p.Value, OldValue: p.OldValue,
		})
	}
	if kc.Scaling != nil {
		rk.Scaling = &repomapk8s.Scaling{
			Replicas: kc.Scaling.Replicas, NewReplicas: kc.Scaling.NewReplicas,
			OldCPU: kc.Scaling.OldCPU, NewCPU: kc.Scaling.NewCPU,
			OldMemory: kc.Scaling.OldMemory, NewMemory: kc.Scaling.NewMemory,
		}
	}
	for _, vc := range kc.VersionChanges {
		rk.VersionChanges = append(rk.VersionChanges, repomapk8s.VersionChange{
			FieldPath: vc.FieldPath, OldVersion: vc.OldVersion, NewVersion: vc.NewVersion,
			ChangeType: repomapk8s.VersionChangeType(vc.ChangeType),
			ValueType:  vc.ValueType,
		})
	}
	if kc.EnvironmentChange != nil {
		rk.EnvironmentChange = &repomapk8s.EnvironmentChange{
			Old: kc.EnvironmentChange.Old, New: kc.EnvironmentChange.New,
		}
	}
	return rk
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
