package verify

import (
	"github.com/flanksource/gavel/internal/githubselector"
	"maps"
)

func githubLinterSchema(base map[string]any) map[string]any {
	node := maps.Clone(base)
	props := maps.Clone(base["properties"].(map[string]any))
	orgs := stringArray("Organization names to discover, combined with repos. Includes visible forks and archived repositories. Lists replace inherited lists.")
	orgs["items"] = map[string]any{"type": "string", "pattern": githubOrgName.String()}
	repos := stringArray("Owner/repo names or qualified GitHub repository queries, e.g. org:acme topic:mission-control. Matches are added to organization discovery, then exclusions apply. Lists replace inherited lists. With no selectors, use the checkout origin; an empty query result does not use the origin.")
	repos["items"] = map[string]any{"type": "string", "anyOf": []any{
		map[string]any{"pattern": `^[A-Za-z0-9][A-Za-z0-9-]*/[A-Za-z0-9_.-]+$`},
		map[string]any{"pattern": githubselector.QueryPattern},
	}}
	stale := stringWithDefault("Positive Go duration since a PR's last activity. Native --stale-after overrides this value.", "720h")
	props["config"] = object("GitHub repository audit settings. Does not activate auditing; explicitly select the github linter.", map[string]any{
		"orgs":         orgs,
		"repos":        repos,
		"excludeRepos": stringArray("Case-insensitive doublestar globs against full owner/repo names. Exclusions always win, including with --repo. Use owner/* to exclude an organization. Lists replace inherited lists."),
		"staleAfter":   stale,
	})
	props["enabled"] = boolProp("Does not activate this linter automatically. Select github explicitly with gavel lint github or --linters github.")
	node["properties"] = props
	return node
}
