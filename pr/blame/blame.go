package blame

import (
	"context"
	"fmt"
	"time"

	"github.com/bluekeyes/go-gitdiff/gitdiff"
)

type Options struct {
	Dir, Base, Head string
	HistoryLimit    int
	Exclude         []string
	ResolveAuthor   func(context.Context, string) (string, error)
	PolicyReviewers []PolicyReviewer
	Since           time.Time
}

type Identity struct {
	ID    string `json:"id,omitempty"`
	Login string `json:"login,omitempty"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

type Evidence struct {
	Path          string `json:"path"`
	Lines         int    `json:"lines"`
	AdjacentLines int    `json:"adjacentLines"`
	Commits       int    `json:"commits"`
	CodeOwner     bool   `json:"codeOwner"`
	Rule          string `json:"rule,omitempty"`
	RuleLine      int    `json:"ruleLine,omitempty"`
}

type Reviewer struct {
	Identity
	CodeOwnerFiles int        `json:"codeOwnerFiles"`
	Lines          int        `json:"lines"`
	AdjacentLines  int        `json:"adjacentLines"`
	Commits        int        `json:"commits"`
	Files          []Evidence `json:"files"`
	Policies       []Policy   `json:"policies,omitempty"`
}

type Policy struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Required bool     `json:"required"`
	Status   string   `json:"status,omitempty"`
	Patterns []string `json:"patterns,omitempty"`
}

type PolicyReviewer struct {
	Identity
	Policy Policy
}

type Result struct {
	Base           string     `json:"base"`
	Head           string     `json:"head"`
	MergeBase      string     `json:"mergeBase"`
	CodeownersFile string     `json:"codeownersFile,omitempty"`
	HistoryLimit   int        `json:"historyLimit"`
	Files          []string   `json:"files"`
	Reviewers      []Reviewer `json:"reviewers"`
	Since          *time.Time `json:"since,omitempty"`
}

func Analyze(ctx context.Context, opts Options) (*Result, error) {
	if opts.Base == "" || opts.Head == "" || opts.HistoryLimit <= 0 {
		return nil, fmt.Errorf("base, head and a positive history limit are required")
	}
	repo := repository{ctx: ctx, dir: opts.Dir}
	result, files, err := repo.comparison(opts)
	if err != nil {
		return nil, err
	}
	if !opts.Since.IsZero() {
		cutoff := opts.Since.UTC().Truncate(time.Second)
		result.Since = &cutoff
	}
	rules, path, err := repo.owners(result.Base)
	if err != nil {
		return nil, err
	}
	result.CodeownersFile = path
	candidates := collector{opts: opts, identities: map[string]Identity{}, reviewers: map[string]*Reviewer{}}
	for _, file := range files {
		path := file.NewName
		if file.IsDelete {
			path = file.OldName
		}
		result.Files = append(result.Files, path)
		if err := candidates.owners(rules, path); err != nil {
			return nil, err
		}
		if !file.IsNew {
			if err := candidates.history(repo, result.MergeBase, file); err != nil {
				return nil, err
			}
		}
	}
	for _, candidate := range opts.PolicyReviewers {
		if candidate.Identity.key() == "" || candidate.Policy.ID <= 0 {
			return nil, fmt.Errorf("policy reviewer requires an identity and positive policy ID")
		}
		candidates.record(candidate.Identity, Evidence{})
		if reviewer := candidates.reviewers[candidate.Identity.key()]; reviewer != nil {
			reviewer.Policies = append(reviewer.Policies, candidate.Policy)
		}
	}
	result.Reviewers = candidates.ranked()
	return result, nil
}

func (c *collector) history(repo repository, base string, file *gitdiff.File) error {
	path := file.NewName
	if file.IsDelete {
		path = file.OldName
	}
	history, err := repo.history(base, file.OldName, c.opts)
	if err != nil {
		return err
	}
	for _, author := range history {
		if err := c.add(repo.ctx, author, Evidence{Path: path, Commits: 1}); err != nil {
			return err
		}
	}
	if file.IsBinary || len(file.TextFragments) == 0 || !file.OldMode.IsRegular() {
		return nil
	}
	lines, err := repo.blame(base, file.OldName)
	if err != nil {
		return err
	}
	for line, adjacent := range selectedLines(file, len(lines)) {
		evidence := Evidence{Path: path, Lines: 1}
		if adjacent {
			evidence.Lines, evidence.AdjacentLines = 0, 1
		}
		if err := c.add(repo.ctx, lines[line-1], evidence); err != nil {
			return err
		}
	}
	return nil
}

func selectedLines(file *gitdiff.File, length int) map[int]bool {
	selected := map[int]bool{}
	for _, fragment := range file.TextFragments {
		start, count := int(fragment.OldPosition), int(fragment.OldLines)
		if count == 0 {
			for _, line := range []int{start, start + 1} {
				if line > 0 && line <= length {
					if _, exists := selected[line]; !exists {
						selected[line] = true
					}
				}
			}
		} else {
			for line := start; line < start+count; line++ {
				selected[line] = false
			}
		}
	}
	return selected
}
