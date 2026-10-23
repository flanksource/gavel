package lifecycle

import (
	"fmt"
	"slices"
	"strings"
	"time"

	git "github.com/flanksource/commons-db/git"
	"github.com/flanksource/gavel/lifecycle"
)

var EventTypes = []string{"pr.create", "pr.update", "pr.comment", "git.conflict", "test.failure", "lint.failure"}

type Event struct {
	ID       string    `json:"id"`
	Type     string    `json:"type"`
	Author   string    `json:"author"`
	Body     string    `json:"body"`
	Mentions []string  `json:"mentions"`
	URL      string    `json:"url"`
	At       time.Time `json:"at"`
}

func (e Event) Variables() map[string]any {
	return map[string]any{"id": e.ID, "type": e.Type, "author": e.Author, "body": e.Body, "mentions": e.Mentions, "url": e.URL}
}

type Failure struct {
	Kind    string
	RunID   string
	HeadSHA string
	Summary string
	URL     string
	At      time.Time
}

type Snapshot struct {
	PR       git.PullRequest
	Metadata git.PullRequestMetadata
	Activity []git.PullRequestActivity
	Failures []Failure
}

type Action struct {
	Step    lifecycle.Step     `json:"step"`
	Event   Event              `json:"event"`
	Reason  string             `json:"reason"`
	LastRun *lifecycle.StepRun `json:"lastRun,omitempty"`
}

func (s Snapshot) Subject() map[string]any {
	return map[string]any{
		"url": s.PR.URL, "number": s.PR.Number, "state": string(s.PR.State),
		"title": s.PR.Title, "body": s.PR.Body, "author": s.PR.Author,
		"head_sha": s.PR.HeadSHA, "base_sha": s.Metadata.BaseSHA,
		"head": s.PR.Head, "base": s.PR.Base, "draft": s.Metadata.Draft,
		"mergeability": s.Metadata.Mergeability,
	}
}

func Events(s Snapshot) ([]Event, error) {
	if s.PR.Number <= 0 || s.PR.URL == "" || s.PR.HeadSHA == "" {
		return nil, fmt.Errorf("PR lifecycle requires a PR URL, positive number and head SHA")
	}
	if s.PR.State != git.PullRequestOpen {
		return []Event{}, nil
	}
	events := []Event{
		{ID: "create:" + s.PR.URL, Type: "pr.create", Author: s.PR.Author, Body: s.PR.Body, URL: s.PR.URL, At: s.PR.CreatedAt},
		{ID: "update:" + s.PR.HeadSHA + ":" + s.Metadata.BaseSHA, Type: "pr.update", Body: s.PR.Body, URL: s.PR.URL},
	}
	if s.Metadata.Mergeability == "conflicts" {
		events = append(events, Event{ID: "conflict:" + s.PR.HeadSHA + ":" + s.Metadata.BaseSHA, Type: "git.conflict", URL: s.PR.URL})
	}
	for _, activity := range s.Activity {
		if activity.Kind != git.PullRequestComment {
			continue
		}
		if activity.ID == "" {
			return nil, fmt.Errorf("PR comment activity has no source identity: %s", activity.URL)
		}
		at := activity.CreatedAt
		if activity.UpdatedAt.After(at) {
			at = activity.UpdatedAt
		}
		events = append(events, Event{ID: "comment:" + activity.ID + ":" + activity.UpdatedAt.Format(time.RFC3339Nano), Type: "pr.comment", Author: activity.Author, Body: activity.Body, Mentions: activity.Mentions, URL: activity.URL, At: at})
	}
	for _, failure := range s.Failures {
		if failure.HeadSHA != s.PR.HeadSHA {
			continue
		}
		if failure.RunID == "" || !slices.Contains([]string{"test", "lint"}, failure.Kind) {
			return nil, fmt.Errorf("PR CI failure needs a source run identity and test/lint classification")
		}
		events = append(events, Event{ID: failure.Kind + ":" + failure.RunID, Type: failure.Kind + ".failure", Body: failure.Summary, URL: failure.URL, At: failure.At})
	}
	for i := range events {
		if events[i].Mentions == nil {
			events[i].Mentions = []string{}
		}
	}
	return events, nil
}

func Actions(engine *lifecycle.Engine, snapshot Snapshot, runs []lifecycle.StepRun) ([]Action, error) {
	events, err := Events(snapshot)
	if err != nil {
		return nil, err
	}
	actions := []Action{}
	for _, event := range events {
		applicable, err := engine.Applicable(lifecycle.Context{Subject: snapshot.Subject(), Event: event.Variables(), Runs: runs})
		if err != nil {
			return nil, fmt.Errorf("%s: %w", event.Type, err)
		}
		if len(applicable) == 0 {
			continue
		}
		action := Action{Step: applicable[0], Event: event, Reason: strings.TrimSpace(applicable[0].When)}
		for i := len(runs) - 1; i >= 0; i-- {
			if runs[i].Step == action.Step.Name {
				last := runs[i]
				action.LastRun = &last
				break
			}
		}
		actions = append(actions, action)
	}
	order := engine.Definition().StepNames()
	slices.SortStableFunc(actions, func(a, b Action) int {
		if priority := slices.Index(order, a.Step.Name) - slices.Index(order, b.Step.Name); priority != 0 {
			return priority
		}
		return b.Event.At.Compare(a.Event.At)
	})
	return actions, nil
}
