package github

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github/activity"
)

// orgDefaultBranchStatusQuery fetches a page of an org's repositories ordered by
// most-recently-pushed, each with its default branch's latest commit status
// rollup. One GraphQL call covers up to `first` repos and their CI status —
// there is no org-wide Actions REST endpoint, so this batched query is how we
// avoid an N+1 of per-repo `actions/runs` calls.
const orgDefaultBranchStatusQuery = `query($org: String!, $first: Int!, $after: String) {
  organization(login: $org) {
    repositories(first: $first, after: $after, orderBy: {field: PUSHED_AT, direction: DESC}) {
      pageInfo { hasNextPage endCursor }
      nodes {
        nameWithOwner
        isPrivate
        isArchived
        url
        pushedAt
        defaultBranchRef {
          name
          target {
            ... on Commit {
              statusCheckRollup {
                state
                contexts(first: 100) {
                  nodes {
                    __typename
                    ... on CheckRun { name status conclusion detailsUrl checkSuite { workflowRun { workflow { name } } } }
                    ... on StatusContext { context state targetUrl }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}`

// orgRepoPageSize is the repos-per-GraphQL-page. 100 is GitHub's max for a
// connection; combined with PUSHED_AT ordering + early stop it usually means a
// single request for a recency-filtered sweep.
const orgRepoPageSize = 100

// defaultMaxRepos caps the sweep so a giant org with everything pushed inside
// the window can't fan out unbounded. Hit is logged, never silently truncated.
const defaultMaxRepos = 500

// OrgStatusOptions parameterizes a default-branch CI sweep over an org.
type OrgStatusOptions struct {
	Org        string
	Since      time.Time // only repos pushed at/after this are included; zero = no filter
	MaxRepos   int       // 0 => defaultMaxRepos
	FailedOnly bool      // keep only repos whose default branch is currently failing
}

// RepoBranchStatus is one repo's default-branch CI status.
type RepoBranchStatus struct {
	Repo          string        `json:"repo"`
	Private       bool          `json:"private"`
	DefaultBranch string        `json:"defaultBranch"`
	URL           string        `json:"url"`
	PushedAt      time.Time     `json:"pushedAt"`
	CheckStatus   *CheckSummary `json:"checkStatus,omitempty"`
}

// OrgBranchStatusResults is a renderable list of per-repo default-branch status.
type OrgBranchStatusResults struct {
	Org   string             `json:"org"`
	Repos []RepoBranchStatus `json:"repos"`
}

type orgStatusResponse struct {
	Data struct {
		Organization struct {
			Repositories struct {
				PageInfo struct {
					HasNextPage bool   `json:"hasNextPage"`
					EndCursor   string `json:"endCursor"`
				} `json:"pageInfo"`
				Nodes []orgRepoNode `json:"nodes"`
			} `json:"repositories"`
		} `json:"organization"`
	} `json:"data"`
}

type orgRepoNode struct {
	NameWithOwner    string    `json:"nameWithOwner"`
	IsPrivate        bool      `json:"isPrivate"`
	IsArchived       bool      `json:"isArchived"`
	URL              string    `json:"url"`
	PushedAt         time.Time `json:"pushedAt"`
	DefaultBranchRef *struct {
		Name   string `json:"name"`
		Target struct {
			StatusCheckRollup *graphQLStatusCheckRollup `json:"statusCheckRollup"`
		} `json:"target"`
	} `json:"defaultBranchRef"`
}

// FetchOrgDefaultBranchStatus returns each repo's default-branch CI status for
// the given org, newest-pushed first. Repos pushed before opt.Since are dropped;
// because the GraphQL query orders by PUSHED_AT DESC, paging stops as soon as a
// page's last repo predates the cutoff. Archived repos and repos with no default
// branch (empty repos) are skipped. Private repos are included when the token
// has access (requires `repo` scope).
func FetchOrgDefaultBranchStatus(opts Options, opt OrgStatusOptions) (OrgBranchStatusResults, *RateLimit, error) {
	token, err := opts.token()
	if err != nil {
		return OrgBranchStatusResults{}, nil, err
	}

	maxRepos := opt.MaxRepos
	if maxRepos <= 0 {
		maxRepos = defaultMaxRepos
	}

	var (
		out    = OrgBranchStatusResults{Org: opt.Org}
		lastRL *RateLimit
		cursor *string
		capped bool
		tooOld bool
	)

	for {
		vars := map[string]any{"org": opt.Org, "first": orgRepoPageSize, "after": cursor}
		body, rl, err := postGraphQL(token, graphqlEndpoint(), activity.KindGraphQL, orgDefaultBranchStatusQuery, vars)
		if rl != nil {
			lastRL = rl
		}
		if err != nil {
			return out, lastRL, err
		}

		var page orgStatusResponse
		if err := json.Unmarshal(body, &page); err != nil {
			return out, lastRL, fmt.Errorf("parse org status response: %w", err)
		}

		conn := page.Data.Organization.Repositories
		for _, node := range conn.Nodes {
			if !opt.Since.IsZero() && node.PushedAt.Before(opt.Since) {
				tooOld = true
				continue
			}
			status, ok := node.toRepoBranchStatus()
			if !ok {
				continue
			}
			if opt.FailedOnly && !status.isFailing() {
				continue
			}
			out.Repos = append(out.Repos, status)
			if len(out.Repos) >= maxRepos {
				capped = true
				break
			}
		}

		if capped || tooOld || !conn.PageInfo.HasNextPage {
			break
		}
		next := conn.PageInfo.EndCursor
		cursor = &next
	}

	if capped {
		logger.Warnf("org %s default-branch status capped at %d repos; pass a tighter --since or raise the cap", opt.Org, maxRepos)
	}
	return out, lastRL, nil
}

// toRepoBranchStatus converts a GraphQL repo node into a RepoBranchStatus.
// Returns ok=false for archived repos and empty repos (no default branch) so
// the caller skips them.
func (n orgRepoNode) toRepoBranchStatus() (RepoBranchStatus, bool) {
	if n.IsArchived || n.DefaultBranchRef == nil {
		return RepoBranchStatus{}, false
	}
	return RepoBranchStatus{
		Repo:          n.NameWithOwner,
		Private:       n.IsPrivate,
		DefaultBranch: n.DefaultBranchRef.Name,
		URL:           n.URL,
		PushedAt:      n.PushedAt,
		CheckStatus:   summarizeRollup(n.DefaultBranchRef.Target.StatusCheckRollup),
	}, true
}

func (s RepoBranchStatus) isFailing() bool {
	return s.CheckStatus != nil && s.CheckStatus.Failed > 0
}

// statusIcon picks a single rollup glyph for the repo's default branch: green
// when everything passed, red on any failure, yellow while running, gray when
// pending or unknown (no checks configured).
func (s RepoBranchStatus) statusIcon() api.Text {
	cs := s.CheckStatus
	switch {
	case cs == nil || (cs.Passed == 0 && cs.Failed == 0 && cs.Running == 0 && cs.Pending == 0):
		return clicky.Text("○", "text-gray-500")
	case cs.Failed > 0:
		return clicky.Text("●", "text-red-600")
	case cs.Running > 0:
		return clicky.Text("●", "text-yellow-600")
	case cs.Pending > 0:
		return clicky.Text("○", "text-gray-500")
	default:
		return clicky.Text("●", "text-green-600")
	}
}

func (s RepoBranchStatus) Pretty() api.Text {
	text := clicky.Text("  ", "").
		Add(s.statusIcon()).
		Append(" "+s.Repo, "font-bold")
	if s.Private {
		text = text.Append(" private", "text-gray-500")
	}
	if s.DefaultBranch != "" {
		text = text.Append(" @"+s.DefaultBranch, "text-gray-500")
	}
	if s.CheckStatus != nil {
		text = text.Append(" ", "").Add(s.CheckStatus.PrettySummary())
		for _, f := range s.CheckStatus.Failures {
			text = text.NewLine().Add(f.Pretty("    "))
		}
	}
	return text
}

func (r OrgBranchStatusResults) Pretty() api.Text {
	if len(r.Repos) == 0 {
		return clicky.Text("No repositories with recent pushes found", "text-gray-500")
	}
	header := strings.TrimSpace(fmt.Sprintf("CI on default branch — %s (%d repos)", r.Org, len(r.Repos)))
	text := clicky.Text(header, "font-bold")
	for i, repo := range r.Repos {
		if i > 0 {
			text = text.NewLine().Append(prListDivider, "text-gray-500")
		}
		text = text.NewLine().Add(repo.Pretty())
	}
	return text
}
