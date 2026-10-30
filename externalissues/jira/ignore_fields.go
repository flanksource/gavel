package jira

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/flanksource/commons/collections"
)

const remoteLinksName = "Remote links"

func (options ParseOptions) ignores(id string) bool {
	if len(options.IgnoreFields) == 0 {
		return false
	}
	switch strings.ToLower(id) {
	case "id", "key", "summary", "url", "project":
		return false
	case "comments":
		id = "comment"
	}
	name := options.Catalog[id].Name
	if id == "remoteLinks" {
		name = remoteLinksName
	}
	return collections.MatchItems(id, options.IgnoreFields...) ||
		(name != "" && collections.MatchItems(name, options.IgnoreFields...))
}

func (options ParseOptions) filter(raw json.RawMessage) (json.RawMessage, error) {
	if len(options.IgnoreFields) == 0 {
		return raw, nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, fmt.Errorf("decode Jira issue: %w", err)
	}
	for id := range fields {
		if options.ignores(id) {
			delete(fields, id)
		}
	}
	return json.Marshal(fields)
}
