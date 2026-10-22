package changeset

import (
	"encoding/json"
	"fmt"
)

func githubComparison(env Environment) (Comparison, error) {
	path := env.Lookup("GITHUB_EVENT_PATH")
	if path == "" {
		return Comparison{}, fmt.Errorf("GITHUB_EVENT_PATH is required")
	}
	data, err := env.ReadFile(path)
	if err != nil {
		return Comparison{}, err
	}
	var event struct {
		Before, After string
		PullRequest   *struct {
			Base, Head struct{ SHA string }
		} `json:"pull_request"`
		MergeGroup *struct {
			Base string `json:"base_sha"`
			Head string `json:"head_sha"`
		} `json:"merge_group"`
	}
	if err := json.Unmarshal(data, &event); err != nil {
		return Comparison{}, fmt.Errorf("decode event: %w", err)
	}
	switch env.Lookup("GITHUB_EVENT_NAME") {
	case "pull_request", "pull_request_target":
		if event.PullRequest == nil || event.PullRequest.Base.SHA == "" || event.PullRequest.Head.SHA == "" {
			return Comparison{}, fmt.Errorf("event requires pull_request base.sha and head.sha")
		}
		return Comparison{Kind: "pull-request", Base: event.PullRequest.Base.SHA, Head: event.PullRequest.Head.SHA}, nil
	case "merge_group":
		if event.MergeGroup == nil || event.MergeGroup.Base == "" || event.MergeGroup.Head == "" {
			return Comparison{}, fmt.Errorf("event requires merge_group base_sha and head_sha")
		}
		return Comparison{Kind: "range", Base: event.MergeGroup.Base, Head: event.MergeGroup.Head}, nil
	case "push":
		if event.After == "" || zeroSHA(event.After) {
			return Comparison{}, fmt.Errorf("push event has no built commit")
		}
		return previousComparison(event.Before, event.After), nil
	default:
		return Comparison{Kind: "commit", Head: env.Lookup("GITHUB_SHA")}, nil
	}
}
