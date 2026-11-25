package jira

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/externalissues"
	"github.com/flanksource/gavel/externalissues/syncdoc"
)

type snapshotField struct {
	externalissues.Field
	Schema json.RawMessage `json:"schema,omitempty"`
}

type issueSnapshot struct {
	Version      int                      `json:"version"`
	Issue        externalissues.Issue     `json:"issue"`
	Fields       map[string]snapshotField `json:"fields"`
	IgnoreFields []string                 `json:"ignoreFields"`
}

func saveSnapshot(issue Issue, ignores []string, stage string, doc *syncdoc.Document) (string, error) {
	snapshot := issueSnapshot{Version: 1, Issue: issue.Issue, Fields: map[string]snapshotField{}, IgnoreFields: ignores}
	snapshot.Issue.Fields = nil
	for id, field := range issue.Fields {
		snapshot.Fields[id] = snapshotField{Field: field, Schema: field.Schema}
	}
	data, err := json.Marshal(snapshot)
	if err != nil {
		return "", err
	}
	path := filepath.Join("issues", "cas", syncdoc.Hash(data)+".json")
	if err := os.MkdirAll(filepath.Join(stage, filepath.Dir(path)), 0700); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(stage, path), data, 0600); err != nil {
		return "", err
	}
	doc.Metadata.Snapshot = &syncdoc.Artifact{Path: filepath.ToSlash(path), SHA256: syncdoc.Hash(data)}
	return path, nil
}

func readSnapshot(path string, doc *syncdoc.Document) (*Issue, error) {
	if doc.Metadata.Snapshot == nil {
		return nil, nil
	}
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	defer root.Close()
	data, err := syncdoc.ReadArtifact(root, *doc.Metadata.Snapshot)
	if err != nil {
		return nil, fmt.Errorf("read issue snapshot %s: %w", path, err)
	}
	var snapshot issueSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, err
	}
	if snapshot.Version != 1 || snapshot.Issue.Site != doc.Metadata.Source.Site || snapshot.Issue.Key != doc.Metadata.Source.Key || snapshot.Issue.ID != doc.Metadata.Source.ID {
		return nil, fmt.Errorf("invalid issue snapshot for %s", path)
	}
	issue := &Issue{Issue: snapshot.Issue, cachedIgnoreFields: snapshot.IgnoreFields, savedPath: path, savedMarkdown: doc.Body()}
	issue.Fields = make(externalissues.Fields, len(snapshot.Fields))
	for id, field := range snapshot.Fields {
		field.Field.Schema = field.Schema
		issue.Fields[id] = field.Field
	}
	return issue, nil
}

func loadSavedIssue(path string, full bool) (Issue, error) {
	doc, err := readSyncDocument(path)
	if err != nil {
		return Issue{}, err
	}
	issue, err := readSnapshot(path, doc)
	if err != nil {
		return Issue{}, err
	}
	if issue == nil {
		return Issue{}, fmt.Errorf("sync did not publish an issue snapshot: %s", path)
	}
	issue.Full = full
	return *issue, nil
}
