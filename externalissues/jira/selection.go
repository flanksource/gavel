package jira

import (
	"encoding/json"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/commons/collections"
	"github.com/flanksource/commons/duration"
)

var projectTarget = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
var issueTarget = regexp.MustCompile(`^([A-Z][A-Z0-9_]*)-([1-9][0-9]*)$`)
var numericTarget = regexp.MustCompile(`^[1-9][0-9]*$`)

type issuePredicate func(json.RawMessage, map[string]FieldDefinition) (bool, error)

func (s *syncSelection) predicate(seen map[string]bool) issuePredicate {
	return func(raw json.RawMessage, catalog map[string]FieldDefinition) (bool, error) {
		if s.fields == nil {
			if err := s.resolve(catalog); err != nil {
				return false, err
			}
		}
		var issue struct {
			ID     json.RawMessage            `json:"id"`
			Key    string                     `json:"key"`
			Fields map[string]json.RawMessage `json:"fields"`
		}
		if err := json.Unmarshal(raw, &issue); err != nil {
			return false, err
		}
		if issue.Key == "" || issue.Fields == nil {
			return false, fmt.Errorf("Jira issue response requires a key and fields object")
		}
		if seen[issue.Key] {
			return false, nil
		}
		seen[issue.Key] = true
		issue.Fields["key"] = json.RawMessage(strconv.Quote(issue.Key))
		if issue.ID != nil {
			issue.Fields["id"] = issue.ID
		}
		fields, err := json.Marshal(issue.Fields)
		if err != nil {
			return false, err
		}
		return s.matches(fields)
	}
}

func syncTargetQuery(target string) (string, error) {
	target = strings.ToUpper(strings.TrimSpace(target))
	if projectTarget.MatchString(target) {
		return fmt.Sprintf("project = %q", target), nil
	}
	if issueTarget.MatchString(target) || numericTarget.MatchString(target) {
		return "", nil
	}
	start, end, ok := strings.Cut(target, "..")
	a, b := issueTarget.FindStringSubmatch(start), issueTarget.FindStringSubmatch(end)
	if !ok || a == nil || b == nil || a[1] != b[1] {
		return "", fmt.Errorf("invalid Jira target %q: expected project, issue key/ID, or APP-10..APP-50", target)
	}
	first, err := strconv.ParseUint(a[2], 10, 64)
	if err != nil {
		return "", fmt.Errorf("invalid range start: %w", err)
	}
	last, err := strconv.ParseUint(b[2], 10, 64)
	if err != nil {
		return "", fmt.Errorf("invalid range end: %w", err)
	}
	if first > last {
		return "", fmt.Errorf("reversed Jira range %q", target)
	}
	return fmt.Sprintf("project = %q AND key >= %q AND key <= %q", a[1], start, end), nil
}

func parseSyncSince(value string, now time.Time) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02"} {
		if at, err := time.Parse(layout, value); err == nil {
			return at, nil
		}
	}
	span, err := duration.ParseDuration(strings.TrimPrefix(value, "now-"))
	if err != nil || span <= 0 {
		return time.Time{}, fmt.Errorf("invalid --since %q: expected a positive duration, ISO date, or RFC3339 timestamp", value)
	}
	return now.Add(-time.Duration(span)), nil
}

type syncSelection struct {
	since    time.Time
	patterns map[string][]string
	fields   map[string][]string
}

func (options SyncOptions) selection(now time.Time) (*syncSelection, error) {
	since, err := parseSyncSince(options.Since, now)
	if err != nil {
		return nil, err
	}
	selection := &syncSelection{since: since, patterns: map[string][]string{}}
	for _, filter := range options.Filters {
		name, pattern, ok := strings.Cut(filter, "=")
		name = strings.ToLower(strings.TrimSpace(name))
		if !ok || name == "" || strings.TrimSpace(pattern) == "" {
			return nil, fmt.Errorf("invalid --filter %q: expected FIELD=PATTERNS", filter)
		}
		selection.patterns[name] = append(selection.patterns[name], pattern)
	}
	if len(options.Statuses) > 0 {
		selection.patterns["status"] = append(selection.patterns["status"], options.Statuses...)
	}
	return selection, nil
}

func (s *syncSelection) resolve(catalog map[string]FieldDefinition) error {
	s.fields = map[string][]string{}
	for name, patterns := range s.patterns {
		id := name
		if !issueCoreField(name) {
			if _, ok := catalog[name]; !ok {
				var ids []string
				for key, field := range catalog {
					if strings.EqualFold(field.Name, name) || strings.EqualFold(key, name) {
						ids = append(ids, key)
					}
				}
				slices.Sort(ids)
				if len(ids) == 0 {
					return fmt.Errorf("unknown Jira filter field %q", name)
				}
				if len(ids) > 1 {
					return fmt.Errorf("ambiguous Jira filter field %q: use an ID from %v", name, ids)
				}
				id = ids[0]
			}
		}
		s.fields[id] = append(s.fields[id], patterns...)
	}
	return nil
}

func (s *syncSelection) matches(raw json.RawMessage) (bool, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return false, err
	}
	if !s.since.IsZero() {
		var updated *string
		if err := json.Unmarshal(fields["updated"], &updated); err != nil {
			return false, fmt.Errorf("filter updated: %w", err)
		}
		at, err := parseTime(updated, "updated")
		if err != nil {
			return false, err
		}
		if at == nil {
			return false, fmt.Errorf("--since requires Jira updated timestamp")
		}
		if at.Before(s.since) {
			return false, nil
		}
	}
	for id, patterns := range s.fields {
		var value any
		if raw, ok := fields[id]; ok {
			decoder := json.NewDecoder(strings.NewReader(string(raw)))
			decoder.UseNumber()
			if err := decoder.Decode(&value); err != nil {
				return false, err
			}
		}
		projected, err := frontmatterValue(value, false)
		if err != nil {
			return false, fmt.Errorf("filter field %s: %w", id, err)
		}
		values := selectionValues(projected)
		matches, excluded := collections.MatchAny(values, patterns...)
		if !matches || excluded {
			return false, nil
		}
	}
	return true, nil
}

func selectionValues(value any) []string {
	if value == nil {
		return []string{""}
	}
	if values, ok := value.([]any); ok {
		var result []string
		for _, item := range values {
			result = append(result, selectionValues(item)...)
		}
		if len(result) == 0 {
			return []string{""}
		}
		return result
	}
	if value, ok := value.(string); ok {
		return []string{value}
	}
	if object, ok := value.(map[string]any); ok {
		var result []string
		for _, key := range slices.Sorted(maps.Keys(object)) {
			result = append(result, selectionValues(object[key])...)
		}
		return result
	}
	return []string{fmt.Sprint(value)}
}
