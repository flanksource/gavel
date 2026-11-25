package comments

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// MentionAgent is the Mentionable ID and name of a target's comment agent rule.
const MentionAgent = "agent"

// Mentionable is an agent a comment body addresses with an @Name token: the
// body's thread is delivered through Route instead of its automatic route.
type Mentionable struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// Kind is "agent", as clicky-ui's CommentMentionable kind.
	Kind string `json:"kind"`
	// Route is where the mention delivers the thread; an Assign route makes
	// its agent the thread's assignee first, as the rule's automatic route does.
	Route Route `json:"-"`
}

// MentionQuery asks a host's Delivery for the agents a comment on Target can
// address, besides its rule's own @agent.
type MentionQuery struct {
	TenantID string
	Target   Target
	// Rule is the run Target's comment agent rule configures; nil when none does.
	Rule *DeliverTarget
	// Sessions adds the scope's live and resumable sessions, which the host
	// lists from its agent runtime. Without it only the agents that need no
	// lookup, the runtime families, are offered. A scope without a local
	// working directory has no session mentionables.
	Sessions bool
}

// mentionToken is an @name token: not inside a word or address, and ending on a word character.
var mentionToken = regexp.MustCompile(`(?:^|[^\w@./-])@(\w(?:[\w.-]*\w)?)`)

// codeFence opens or closes a fenced code block; blockquoteLine starts a quoted line.
var (
	codeFence      = regexp.MustCompile("^ {0,3}(`{3,}|~{3,})")
	blockquoteLine = regexp.MustCompile(`^ {0,3}>`)
)

// MentionedIn are the mentionables whose @Name is a token of body, matched
// case-insensitively, in mentionables order, each once. A token inside inline
// code, a fenced code block or a blockquote is quoted text, not a mention.
func MentionedIn(body string, mentionables []Mentionable) []Mentionable {
	return mentioned(mentionTokens(body), mentionables)
}

func mentioned(tokens map[string]bool, mentionables []Mentionable) []Mentionable {
	var found []Mentionable
	for _, m := range mentionables {
		if tokens[strings.ToLower(m.Name)] && !slices.ContainsFunc(found, func(f Mentionable) bool { return f.ID == m.ID }) {
			found = append(found, m)
		}
	}
	return found
}

// mentionTokens are body's @name tokens, lowercased, outside code and quotes.
func mentionTokens(body string) map[string]bool {
	tokens := map[string]bool{}
	for _, match := range mentionToken.FindAllStringSubmatch(prose(body), -1) {
		tokens[strings.ToLower(match[1])] = true
	}
	return tokens
}

// prose is body with its fenced code blocks and blockquote lines blanked and
// its inline code spans removed; line boundaries are kept.
func prose(body string) string {
	lines := strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n")
	open := ""
	for i, line := range lines {
		if marker := codeFence.FindStringSubmatch(line); marker != nil {
			if open == "" {
				open = marker[1]
			} else if marker[1][0] == open[0] && len(marker[1]) >= len(open) {
				open = ""
			}
			lines[i] = ""
			continue
		}
		if open != "" || blockquoteLine.MatchString(line) {
			lines[i] = ""
		}
	}
	return withoutCodeSpans(strings.Join(lines, "\n"))
}

// withoutCodeSpans replaces each inline code span (a run of backticks closed
// by a run of the same length) with a space.
func withoutCodeSpans(text string) string {
	var out strings.Builder
	for i := 0; i < len(text); {
		if text[i] != '`' {
			out.WriteByte(text[i])
			i++
			continue
		}
		n := backtickRun(text, i)
		end := closingBackticks(text, i+n, n)
		if end < 0 {
			out.WriteString(text[i : i+n])
			i += n
			continue
		}
		out.WriteByte(' ')
		i = end + n
	}
	return out.String()
}

func backtickRun(text string, at int) int {
	n := 0
	for at+n < len(text) && text[at+n] == '`' {
		n++
	}
	return n
}

// closingBackticks is where a run of exactly n backticks starts at or after from; -1 for none.
func closingBackticks(text string, from, n int) int {
	for j := from; j < len(text); {
		if text[j] != '`' {
			j++
			continue
		}
		run := backtickRun(text, j)
		if run == n {
			return j
		}
		j += run
	}
	return -1
}

// Mentionables lists the agents a comment on t can address: its comment agent
// rule as @agent, and what the host's Delivery offers (runtimes, sessions).
func (s *Service) Mentionables(ctx context.Context, t Target) ([]Mentionable, error) {
	if err := s.requireDelivery(); err != nil {
		return nil, err
	}
	if t.ScopeURI == "" || t.URI == "" {
		return nil, fmt.Errorf("%w: mentionables need a scope and target", ErrInvalid)
	}
	p, err := s.principal(ctx)
	if err != nil {
		return nil, err
	}
	if err = s.access.Authorize(ctx, p, "mentionables", Target{ScopeURI: t.ScopeURI, URI: t.URI}); err != nil {
		return nil, err
	}
	return s.mentionables(ctx, p, t, true)
}

// mentionables are t's rule as @agent and what the host's Delivery offers;
// sessions adds the scope's sessions, which costs the host a runtime lookup.
func (s *Service) mentionables(ctx context.Context, p Principal, t Target, sessions bool) ([]Mentionable, error) {
	rule, err := s.agentRule(ctx, t)
	if err != nil {
		return nil, err
	}
	q, found := MentionQuery{TenantID: p.TenantID, Target: t, Sessions: sessions}, []Mentionable{}
	if rule != nil {
		q.Rule = &rule.Target
		found = append(found, Mentionable{ID: MentionAgent, Name: MentionAgent, Kind: "agent", Route: *rule})
	}
	offered, err := s.delivery.Mentionables(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("listing the agents a comment on %s can mention: %w", t.URI, err)
	}
	return append(found, offered...), nil
}

// mentionRoute is the route body, written on t by p, is delivered through by
// its one agent mention; nil when it mentions none. Only a person's mentions
// deliver: an agent's are prose, so agents never hand threads to each other by
// quoting. The static names (@agent, the runtime families) are matched first;
// sessions are listed only for a token they leave unmatched. Two different
// agents in one body are ErrInvalid.
func (s *Service) mentionRoute(ctx context.Context, p Principal, t Target, body string) (*Route, error) {
	tokens := mentionTokens(body)
	if s.delivery == nil || p.Actor.Kind != "user" || len(tokens) == 0 {
		return nil, nil
	}
	all, err := s.mentionables(ctx, p, t, false)
	if err != nil {
		return nil, err
	}
	if unmatched(tokens, all) {
		if all, err = s.mentionables(ctx, p, t, true); err != nil {
			return nil, err
		}
	}
	found := mentioned(tokens, all)
	if len(found) > 1 {
		names := make([]string, 0, len(found))
		for _, m := range found {
			names = append(names, "@"+m.Name)
		}
		return nil, fmt.Errorf("%w: the comment mentions %s; address one agent", ErrInvalid, strings.Join(names, ", "))
	}
	if len(found) == 0 {
		return nil, nil
	}
	return s.normalizeRoute(found[0].Route)
}

// unmatched is whether a token names none of mentionables.
func unmatched(tokens map[string]bool, mentionables []Mentionable) bool {
	names := map[string]bool{}
	for _, m := range mentionables {
		names[strings.ToLower(m.Name)] = true
	}
	for token := range tokens {
		if !names[token] {
			return true
		}
	}
	return false
}

// agentRule is t's comment agent rule, nil when the host's routes configure none for t.
func (s *Service) agentRule(ctx context.Context, t Target) (*Route, error) {
	rules, ok := s.routes.(AgentRules)
	if !ok {
		return nil, nil
	}
	r, found, err := rules.AgentRule(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("resolving the comment agent rule of %s: %w", t.URI, err)
	}
	if !found {
		return nil, nil
	}
	return &r, nil
}

// withRuleEdits is target with an empty edits policy set to t's rule's, or
// EditsDirect when no rule sets one; a target's own policy wins.
func (s *Service) withRuleEdits(ctx context.Context, t Target, target DeliverTarget) (DeliverTarget, error) {
	if target.Edits != "" {
		return target, nil
	}
	rule, err := s.agentRule(ctx, t)
	if err != nil {
		return target, err
	}
	target.Edits = EditsDirect
	if rule != nil && rule.Target.Edits != "" {
		target.Edits = rule.Target.Edits
	}
	return target, nil
}
