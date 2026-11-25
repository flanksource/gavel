package github

import (
	"fmt"
	"strings"

	"github.com/flanksource/gavel/models"
)

const codeQLAnnotationsQuery = `query($id: ID!, $after: String) {
  node(id: $id) { ... on CheckRun {
    annotations(first: 100, after: $after) {
      pageInfo { hasNextPage endCursor }
      nodes { path title message annotationLevel location { start { line column } end { line column } } }
    }
  } }
}`

type checkAnnotation struct {
	Path, Title, Message, AnnotationLevel string
	Location                              struct {
		Start, End struct{ Line, Column int }
	}
}

func FetchCodeQLFindings(opts Options, nodeID string) ([]models.Violation, error) {
	if nodeID == "" {
		return nil, fmt.Errorf("CodeQL check is missing its node ID")
	}
	var findings []models.Violation
	var after any
	seen := map[string]bool{}
	for {
		var response struct {
			Data struct {
				Node *struct {
					Annotations *struct {
						Nodes    []checkAnnotation
						PageInfo graphQLPageInfo
					}
				}
			}
		}
		if err := opts.auditQuery(codeQLAnnotationsQuery, map[string]any{"id": nodeID, "after": after}, &response); err != nil {
			return nil, fmt.Errorf("parse CodeQL annotations: %w", err)
		}
		if response.Data.Node == nil || response.Data.Node.Annotations == nil {
			return nil, fmt.Errorf("CodeQL check %s annotations are missing", nodeID)
		}
		page := response.Data.Node.Annotations
		for _, annotation := range page.Nodes {
			finding, err := annotation.violation()
			if err != nil {
				return nil, err
			}
			findings = append(findings, finding)
		}
		if !page.PageInfo.HasNextPage {
			return findings, nil
		}
		cursor := page.PageInfo.EndCursor
		if cursor == "" || seen[cursor] {
			return nil, fmt.Errorf("CodeQL annotation pagination did not advance")
		}
		seen[cursor] = true
		after = cursor
	}
}

func (a checkAnnotation) violation() (models.Violation, error) {
	severities := map[string]models.ViolationSeverity{"FAILURE": models.SeverityError, "WARNING": models.SeverityWarning, "NOTICE": models.SeverityInfo}
	severity, ok := severities[a.AnnotationLevel]
	if !ok {
		return models.Violation{}, fmt.Errorf("unknown CodeQL annotation level %q", a.AnnotationLevel)
	}
	if a.Path == "" || a.Location.Start.Line < 1 || a.Location.End.Line < a.Location.Start.Line {
		return models.Violation{}, fmt.Errorf("invalid CodeQL annotation location for %q", a.Path)
	}
	message := a.Message
	if a.Title != "" {
		message = strings.TrimSpace(a.Title + ": " + message)
	}
	return models.Violation{File: a.Path, Line: a.Location.Start.Line, Column: a.Location.Start.Column, EndLine: a.Location.End.Line, EndColumn: a.Location.End.Column, Message: models.StringPtr(message), Source: "codeql", Severity: severity}, nil
}
