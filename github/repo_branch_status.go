package github

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/github/activity"
)

const repoDefaultBranchStatusQuery = `query($owner: String!, $name: String!) {
  repository(owner: $owner, name: $name) { ...BranchStatus }
}` + branchStatusFragment

func FetchDefaultBranchStatus(opts Options) (RepoBranchStatus, error) {
	repo, err := opts.resolveRepo()
	if err != nil {
		return RepoBranchStatus{}, err
	}
	parts := strings.Split(repo, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return RepoBranchStatus{}, fmt.Errorf("invalid repository %q: expected owner/repo", repo)
	}
	token, err := opts.token()
	if err != nil {
		return RepoBranchStatus{}, err
	}
	body, _, err := postGraphQL(token, graphqlEndpoint(), activity.KindGraphQL, repoDefaultBranchStatusQuery, map[string]any{"owner": parts[0], "name": parts[1]})
	if err != nil {
		return RepoBranchStatus{}, err
	}
	var response struct {
		Data struct {
			Repository *orgRepoNode `json:"repository"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return RepoBranchStatus{}, fmt.Errorf("parse %s default branch status: %w", repo, err)
	}
	if response.Data.Repository == nil || response.Data.Repository.DefaultBranchRef == nil {
		return RepoBranchStatus{}, fmt.Errorf("repository %s has no default branch or is not accessible", repo)
	}
	if response.Data.Repository.IsArchived {
		return RepoBranchStatus{}, fmt.Errorf("repository %s is archived", repo)
	}
	status, _ := response.Data.Repository.toRepoBranchStatus()
	return status, nil
}
