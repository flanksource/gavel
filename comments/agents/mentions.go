package agents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/verify"
)

// Mentionables are a fresh run of each runtime family's default agent model,
// as @<family>, and, when the query asks, the scope's live and resumable
// sessions, named by mentionName. A scope without a local working directory
// (comments.ErrNoWorkdir) has no session mentionables; any other failure to
// list them is returned.
func (d delivery) Mentionables(ctx context.Context, q comments.MentionQuery) ([]comments.Mentionable, error) {
	runtimes, err := runtimeMentionables(q.Rule, api.RuntimeCatalog())
	if err != nil || !q.Sessions {
		return runtimes, err
	}
	dir, err := d.Workdir(ctx, comments.Target{ScopeURI: q.Target.ScopeURI})
	if errors.Is(err, comments.ErrNoWorkdir) {
		return runtimes, nil
	}
	if err != nil {
		return nil, err
	}
	records, err := d.Bridge.Live(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing live agent sessions: %w", err)
	}
	sessions, err := d.sessions(ctx, comments.SessionQuery{TenantID: q.TenantID, ScopeURI: q.Target.ScopeURI}, dir, records)
	if err != nil {
		return nil, err
	}
	slugs := map[string]string{}
	for _, record := range records {
		slugs[record.ID] = record.Slug
	}
	reserved := map[string]bool{comments.MentionAgent: true}
	for _, m := range runtimes {
		reserved[strings.ToLower(m.Name)] = true
	}
	return append(runtimes, sessionMentionables(sessions, slugs, reserved)...), nil
}

// sessionMentionables name each session @<mentionName>. A name reserved for
// @agent or a runtime family is theirs: a session it would collide with is not
// listed. Sessions sharing one name are each listed as <name>-<short id>, never
// by the ambiguous name.
func sessionMentionables(sessions []comments.AgentSession, slugs map[string]string, reserved map[string]bool) []comments.Mentionable {
	named := map[string][]comments.AgentSession{}
	var order []string
	for _, session := range sessions {
		name := mentionName(session, slugs[session.ID])
		key := strings.ToLower(name)
		if reserved[key] {
			continue
		}
		if _, seen := named[key]; !seen {
			order = append(order, key)
		}
		named[key] = append(named[key], session)
	}
	var found []comments.Mentionable
	for _, key := range order {
		for _, session := range named[key] {
			name := mentionName(session, slugs[session.ID])
			if len(named[key]) > 1 {
				name += "-" + shortID(session.ID)
			}
			target := comments.DeliverTarget{Kind: "session", SessionID: session.ID}
			found = append(found, comments.Mentionable{ID: "session:" + session.ID, Name: name, Kind: "agent", Route: comments.Route{Source: "mention:session:" + session.ID, Target: target}})
		}
	}
	return found
}

// mentionName is captain's slug for a session; for a managed session without
// one, its label (the thread it started for) as a slug; else its id.
func mentionName(session comments.AgentSession, slug string) string {
	if slug != "" {
		return slug
	}
	if label := slugify(session.Title); session.Managed && session.Title != session.ID && label != "" {
		return label
	}
	return session.ID
}

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

// slugify is text as a lowercase, dash-separated @name of at most 40 characters.
func slugify(text string) string {
	slug := strings.Trim(nonSlug.ReplaceAllString(strings.ToLower(text), "-"), "-")
	if len(slug) > 40 {
		slug = strings.TrimRight(slug[:40], "-")
	}
	return slug
}

func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

// runtimeMentionables are a fresh run per family with an enabled agent mode:
// rule's prompt, presets, variables, instructions, edits and spec (the default
// comment agent's without a rule) with only the spec's model replaced by the
// family's default agent model.
func runtimeMentionables(rule *comments.DeliverTarget, families []api.RuntimeFamily) ([]comments.Mentionable, error) {
	base := comments.DeliverTarget{Kind: "run", Prompt: verify.CommentAgent{}.PromptOrDefault()}
	if rule != nil {
		base = *rule
	}
	options, err := decodeRunOptions(base.RunOptions)
	if err != nil {
		return nil, fmt.Errorf("the comment agent rule's run options: %w", err)
	}
	spec := api.Spec{}
	if options.Spec != nil {
		spec = *options.Spec
	}
	var found []comments.Mentionable
	for _, family := range families {
		i := slices.IndexFunc(family.Modes, func(m api.RuntimeModeEntry) bool { return m.Mode == string(api.ModeAgent) })
		if i < 0 || family.Modes[i].Disabled || family.Modes[i].DefaultModel == "" {
			continue
		}
		spec.Model = api.Model{Name: family.Modes[i].DefaultModel, Mode: api.ModeAgent}
		options.Spec = &spec
		encoded, err := json.Marshal(options)
		if err != nil {
			return nil, fmt.Errorf("encode the %s run options: %w", family.Family, err)
		}
		target := comments.DeliverTarget{Kind: "run", Prompt: base.Prompt, RunOptions: encoded, Fresh: true, Edits: base.Edits}
		found = append(found, comments.Mentionable{ID: "runtime:" + family.Family, Name: family.Family, Kind: "agent", Route: comments.Route{Source: "mention:runtime:" + family.Family, Target: target}})
	}
	return found, nil
}
