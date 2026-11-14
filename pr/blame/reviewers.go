package blame

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hmarr/codeowners"
)

type collector struct {
	opts       Options
	identities map[string]Identity
	reviewers  map[string]*Reviewer
}

func (c *collector) owners(rules codeowners.Ruleset, path string) error {
	rule, err := rules.Match(path)
	if err != nil {
		return fmt.Errorf("match CODEOWNERS for %s: %w", path, err)
	}
	if rule == nil {
		return nil
	}
	for _, owner := range rule.Owners {
		identity := Identity{Login: owner.String()}
		if owner.Type == codeowners.EmailOwner {
			identity = Identity{Email: owner.Value}
			if resolved, exists := c.identities[strings.ToLower(owner.Value)]; exists {
				identity = resolved
			}
		}
		c.record(identity, Evidence{Path: path, CodeOwner: true, Rule: rule.RawPattern(), RuleLine: rule.LineNumber})
	}
	return nil
}

func (c *collector) add(ctx context.Context, a author, evidence Evidence) error {
	if !c.opts.Since.IsZero() && a.committedAt < c.opts.Since.Unix() {
		return nil
	}
	key := strings.ToLower(a.Email)
	identity, exists := c.identities[key]
	if !exists {
		identity = a.Identity
		if c.opts.ResolveAuthor != nil {
			login, err := c.opts.ResolveAuthor(ctx, a.commit)
			if err != nil {
				return fmt.Errorf("resolve author of %s: %w", a.commit, err)
			}
			if login != "" {
				identity.Login = "@" + strings.TrimPrefix(login, "@")
			}
		}
		c.identities[key] = identity
	}
	c.record(identity, evidence)
	return nil
}

func (c *collector) record(identity Identity, evidence Evidence) {
	if c.excluded(identity) {
		delete(c.reviewers, strings.ToLower(identity.Email))
		delete(c.reviewers, strings.ToLower(identity.Login))
		delete(c.reviewers, strings.ToLower(identity.ID))
		return
	}
	key := identity.key()
	if email := strings.ToLower(identity.Email); email != "" && email != key {
		if previous := c.reviewers[email]; previous != nil {
			delete(c.reviewers, email)
			for _, file := range previous.Files {
				c.record(identity, file)
			}
		}
	}
	reviewer := c.reviewers[key]
	if reviewer == nil {
		reviewer = &Reviewer{Identity: identity}
		c.reviewers[key] = reviewer
	}
	if identity.Email != "" && reviewer.Email == "" {
		reviewer.Name, reviewer.Email = identity.Name, identity.Email
	}
	if identity.ID != "" {
		reviewer.ID = identity.ID
	}
	if evidence.Path == "" {
		return
	}
	for i := range reviewer.Files {
		if reviewer.Files[i].Path == evidence.Path {
			mergeEvidence(&reviewer.Files[i], evidence)
			return
		}
	}
	reviewer.Files = append(reviewer.Files, evidence)
}

func mergeEvidence(target *Evidence, source Evidence) {
	target.Lines += source.Lines
	target.AdjacentLines += source.AdjacentLines
	target.Commits += source.Commits
	if source.CodeOwner {
		target.CodeOwner, target.Rule, target.RuleLine = true, source.Rule, source.RuleLine
	}
}

func (c *collector) excluded(identity Identity) bool {
	for _, value := range []string{identity.Login, identity.Email, identity.Name} {
		if strings.Contains(strings.ToLower(value), "[bot]") {
			return true
		}
	}
	for _, excluded := range c.opts.Exclude {
		if strings.EqualFold(strings.TrimPrefix(identity.Login, "@"), strings.TrimPrefix(excluded, "@")) || strings.EqualFold(identity.Email, excluded) || strings.EqualFold(identity.ID, excluded) {
			return true
		}
	}
	return false
}

func (c *collector) ranked() []Reviewer {
	result := make([]Reviewer, 0, len(c.reviewers))
	for _, reviewer := range c.reviewers {
		for _, file := range reviewer.Files {
			if file.CodeOwner {
				reviewer.CodeOwnerFiles++
			}
			reviewer.Lines += file.Lines
			reviewer.AdjacentLines += file.AdjacentLines
			reviewer.Commits += file.Commits
		}
		sort.Slice(reviewer.Files, func(i, j int) bool { return reviewer.Files[i].Path < reviewer.Files[j].Path })
		result = append(result, *reviewer)
	}
	sort.Slice(result, func(i, j int) bool {
		a, b := result[i], result[j]
		for _, pair := range [][2]int{{a.requiredPolicies(), b.requiredPolicies()}, {len(a.Policies), len(b.Policies)}, {a.CodeOwnerFiles, b.CodeOwnerFiles}, {a.Lines, b.Lines}, {a.AdjacentLines, b.AdjacentLines}, {a.Commits, b.Commits}, {len(a.Files), len(b.Files)}} {
			if pair[0] != pair[1] {
				return pair[0] > pair[1]
			}
		}
		return a.Identity.key() < b.Identity.key()
	})
	return result
}

func (i Identity) key() string {
	for _, key := range []string{i.Login, i.Email, i.ID} {
		if key != "" {
			return strings.ToLower(key)
		}
	}
	return ""
}

func (r Reviewer) requiredPolicies() int {
	count := 0
	for _, policy := range r.Policies {
		if policy.Required {
			count++
		}
	}
	return count
}
