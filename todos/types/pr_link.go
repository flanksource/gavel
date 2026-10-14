package types

import (
	"fmt"
	"net/url"
	"strings"
)

// Hosted PR providers a TODO can link to.
const (
	PRProviderGitHub = "github"
	PRProviderAzure  = "azure"
)

// PRRef identifies one hosted pull request. Repo is a lower-cased key, not a
// display name: GitHub and Azure DevOps both resolve owner and repository names
// case-insensitively, so two spellings of one PR must key the same row. It is
// built from the URL's escaped path so the backfill SQL
// (internal/database/schema/160_backfill_todo_pr_links.sql) derives exactly the
// same key from a stored URL.
type PRRef struct {
	Provider string `json:"provider"`
	Repo     string `json:"repo"`
	Number   int    `json:"number"`
	URL      string `json:"url"`
}

// PRRefFromURL validates a hosted PR URL with PRFromURL and derives its key.
func PRRefFromURL(raw string) (PRRef, error) {
	pr, err := PRFromURL(raw)
	if err != nil {
		return PRRef{}, err
	}
	u, err := url.Parse(pr.URL)
	if err != nil {
		return PRRef{}, fmt.Errorf("invalid PR URL: %w", err)
	}
	parts := strings.Split(strings.Trim(u.EscapedPath(), "/"), "/")
	ref := PRRef{Number: pr.Number, URL: pr.URL}
	if u.Host == "github.com" {
		ref.Provider, ref.Repo = PRProviderGitHub, strings.ToLower(parts[0]+"/"+parts[1])
		return ref, nil
	}
	ref.Provider, ref.Repo = PRProviderAzure, strings.ToLower(strings.Join(azureRepoKey(u.Host, parts), "/"))
	return ref, nil
}

// azureRepoKey is the organization, optional project and repository of an Azure
// PR path already validated by azuredevops.ParsePRURL: [org/][project/]_git/repo/pullrequest/N.
func azureRepoKey(host string, parts []string) []string {
	git := len(parts) - 4
	scope := parts[:git]
	var org string
	if host == "dev.azure.com" {
		org, scope = scope[0], scope[1:]
	} else {
		org = strings.TrimSuffix(host, ".visualstudio.com")
		if len(scope) > 0 && strings.EqualFold(scope[0], "DefaultCollection") {
			scope = scope[1:]
		}
	}
	return append(append([]string{org}, scope...), parts[git+1])
}

// SamePR reports whether two refs name one pull request, whatever URL spelling
// each was built from.
func (r PRRef) SamePR(other PRRef) bool {
	return r.Provider == other.Provider && r.Repo == other.Repo && r.Number == other.Number
}

// PRRelation is how a linked PR relates to a TODO.
type PRRelation string

const (
	// PRRelationImplements marks a PR that carries the TODO's work. The TODO's
	// primary PR (todo.pr) is always linked with this relation.
	PRRelationImplements PRRelation = "implements"
	// PRRelationRelated marks a PR that touches the same ground without
	// carrying the work.
	PRRelationRelated PRRelation = "related"
)

// ParsePRRelation normalizes a relation name. Empty means related, the claim
// that asserts least.
func ParsePRRelation(value string) (PRRelation, error) {
	switch relation := PRRelation(strings.ToLower(strings.TrimSpace(value))); relation {
	case "":
		return PRRelationRelated, nil
	case PRRelationImplements, PRRelationRelated:
		return relation, nil
	default:
		return "", fmt.Errorf("unknown PR relation %q; valid relations: %s, %s", value, PRRelationImplements, PRRelationRelated)
	}
}

// PRLinkSource records who created a PR link.
type PRLinkSource string

const (
	PRLinkSourceTriage PRLinkSource = "triage"
	PRLinkSourceManual PRLinkSource = "manual"
)

func (s PRLinkSource) Validate() error {
	if s == PRLinkSourceTriage || s == PRLinkSourceManual {
		return nil
	}
	return fmt.Errorf("unknown PR link source %q; valid sources: %s, %s", s, PRLinkSourceTriage, PRLinkSourceManual)
}
