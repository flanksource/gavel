package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	cexec "github.com/flanksource/clicky/exec"
)

type Client struct {
	Site    string
	Profile string
	run     cexec.WrapperFunc
}

func (c Client) Get(ctx context.Context, issue string) (json.RawMessage, error) {
	if strings.TrimSpace(issue) == "" {
		return nil, fmt.Errorf("issue key is required")
	}
	data, err := c.Call(ctx, "jira", "workitem", "get", issue, "--full")
	if err != nil {
		return nil, err
	}
	var issues []json.RawMessage
	if err := json.Unmarshal(data, &issues); err != nil {
		return nil, fmt.Errorf("decode Jira issue %q: %w", issue, err)
	}
	if len(issues) != 1 {
		return nil, fmt.Errorf("get Jira issue %q: expected one issue, received %d", issue, len(issues))
	}
	return issues[0], nil
}

func (c Client) Call(ctx context.Context, args ...string) (json.RawMessage, error) {
	command := strings.Join(args[:min(len(args), 3)], " ")
	run := c.run
	if run == nil {
		run = cexec.NewExec("twg").WithoutShell().AsWrapper()
	}
	argv := make([]any, 0, len(args)+9)
	argv = append(argv, cexec.WithContext(ctx))
	for _, arg := range args {
		argv = append(argv, arg)
	}
	argv = append(argv, "--output", "json", "--output-summary", "none")
	if c.Site != "" {
		argv = append(argv, "--site", c.Site)
	}
	if c.Profile != "" {
		argv = append(argv, "--profile", c.Profile)
	}
	result, err := run(argv...)
	if err != nil {
		if result != nil {
			return nil, fmt.Errorf("twg %s: %w: %s", command, err, strings.TrimSpace(result.Stderr+result.Stdout))
		}
		return nil, fmt.Errorf("twg %s: %w", command, err)
	}
	if result == nil {
		return nil, fmt.Errorf("twg %s: missing process result", command)
	}
	var envelope struct {
		Data  json.RawMessage `json:"data"`
		Error json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal([]byte(result.Stdout), &envelope); err != nil {
		return nil, fmt.Errorf("twg %s: invalid JSON response: %w", command, err)
	}
	if len(envelope.Error) > 0 && string(envelope.Error) != "null" {
		return nil, fmt.Errorf("twg %s: %s", command, envelope.Error)
	}
	if len(envelope.Data) == 0 {
		return nil, fmt.Errorf("twg %s: response has no data", command)
	}
	return envelope.Data, nil
}
