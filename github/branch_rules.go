package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

func FetchBranchRules(opts Options, branch string) ([]json.RawMessage, error) {
	if strings.TrimSpace(branch) == "" {
		return nil, fmt.Errorf("branch name is required")
	}
	repo, err := opts.resolveRepo()
	if err != nil {
		return nil, err
	}
	token, err := opts.token()
	if err != nil {
		return nil, err
	}
	var rules []json.RawMessage
	for page := 1; ; page++ {
		path := fmt.Sprintf("/repos/%s/rules/branches/%s?per_page=100&page=%d", repo, url.PathEscape(branch), page)
		response, err := cachedGet(context.Background(), opts.cacheStore(), token, path, nil)
		if err != nil {
			return nil, err
		}
		var entries []json.RawMessage
		if err := json.Unmarshal(response.Body, &entries); err != nil {
			return nil, fmt.Errorf("parse rules for %s branch %s: %w", repo, branch, err)
		}
		if entries == nil {
			return nil, fmt.Errorf("rules for %s branch %s were null, expected an array", repo, branch)
		}
		rules = append(rules, entries...)
		if len(entries) < 100 {
			return rules, nil
		}
	}
}
