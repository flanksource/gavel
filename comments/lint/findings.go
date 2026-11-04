package lint

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/flanksource/gavel/comments"
)

type Finding struct {
	ID      string          `json:"id"`
	Body    string          `json:"bodyMarkdown"`
	Target  comments.Target `json:"target"`
	Content json.RawMessage `json:"content,omitempty"`
}
type MaterializeRequest struct {
	RunID      string
	WorkflowID string
	Findings   []Finding
}

// Materialize explicitly saves findings locally. It does not publish provider comments.
func Materialize(ctx context.Context, service *comments.Service, req MaterializeRequest) ([]comments.Comment, error) {
	if service == nil || req.RunID == "" {
		return nil, fmt.Errorf("service and run identity are required")
	}
	seen := map[string]bool{}
	for _, f := range req.Findings {
		if f.ID == "" || seen[f.ID] {
			return nil, fmt.Errorf("finding identities must be nonempty and unique")
		}
		seen[f.ID] = true
	}
	out := make([]comments.Comment, 0, len(req.Findings))
	for _, f := range req.Findings {
		c, err := service.Create(ctx, comments.CreateRequest{Target: f.Target, Body: f.Body, WorkflowID: req.WorkflowID, RequestID: "lint:" + req.RunID + ":" + f.ID, Content: f.Content, Links: []comments.Link{{Relation: "execution", URI: "gavel://execution/" + url.PathEscape(req.RunID)}}})
		if err != nil {
			return out, err
		}
		out = append(out, c)
	}
	return out, nil
}
