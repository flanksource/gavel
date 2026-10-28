package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/externalissues/jira"
	"github.com/spf13/cobra"
)

type externalIssueRef struct {
	IDs []string `args:"true" required:"true"`
}

func (o externalIssueRef) one() (string, error) {
	if len(o.IDs) != 1 || strings.TrimSpace(o.IDs[0]) == "" {
		return "", fmt.Errorf("exactly one Jira issue key or attachment ID is required")
	}
	return o.IDs[0], nil
}

type externalIssueList struct {
	JQL   string `flag:"jql" required:"true" help:"Jira JQL query"`
	Limit int    `flag:"limit" help:"Maximum number of issues"`
	After string `flag:"after" help:"Pagination cursor"`
}

type externalCommentCreate struct {
	externalIssueRef
	Body string `flag:"body" required:"true" help:"Comment in Markdown"`
}

type externalAttachmentUpload struct {
	externalIssueRef
	Files []string `flag:"file" required:"true" help:"File paths to upload"`
}

type externalAttachmentDelete struct {
	externalIssueRef
	Yes bool `flag:"yes" help:"Confirm permanent deletion"`
}

type externalAttachmentDownload struct {
	externalIssueRef
	Output string `flag:"output" required:"true" help:"Destination file path"`
}

func decodeExternalJSON(raw json.RawMessage) (any, error) {
	if len(raw) > 0 && raw[0] == '[' {
		var items []json.RawMessage
		if err := json.Unmarshal(raw, &items); err != nil {
			return nil, fmt.Errorf("decode Jira response: %w", err)
		}
		return items, nil
	}
	var result any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("decode Jira response: %w", err)
	}
	return result, nil
}

func Register(rootCmd *cobra.Command) {
	root := &cobra.Command{Use: "external-issues", Short: "Work with issues in external systems"}
	provider := &cobra.Command{Use: "jira", Short: "Work with Jira issues using twg"}
	root.AddCommand(provider)
	rootCmd.AddCommand(root)
	var site, profile string
	provider.PersistentFlags().StringVar(&site, "site", "", "Atlassian site")
	provider.PersistentFlags().StringVar(&profile, "profile", "", "twg authentication profile")
	client := func() jira.Client { return jira.Client{Site: site, Profile: profile} }
	clicky.AddNamedCommandWithContext("get", provider, externalIssueRef{}, func(ctx context.Context, opts externalIssueRef) (jira.Issue, error) {
		id, err := opts.one()
		if err != nil {
			return jira.Issue{}, err
		}
		raw, err := client().Get(ctx, id)
		if err != nil {
			return jira.Issue{}, err
		}
		return jira.ParseIssue(raw)
	})
	clicky.AddNamedCommandWithContext("list", provider, externalIssueList{}, func(ctx context.Context, opts externalIssueList) (any, error) {
		args := []string{"jira", "workitem", "query", "--jql", opts.JQL}
		if opts.Limit > 0 {
			args = append(args, "--limit", strconv.Itoa(opts.Limit))
		}
		if opts.After != "" {
			args = append(args, "--after", opts.After)
		}
		raw, err := client().Call(ctx, args...)
		if err != nil {
			return nil, err
		}
		return jira.ParseList(raw)
	})
	comments := &cobra.Command{Use: "comment", Short: "Read or create issue comments"}
	provider.AddCommand(comments)
	clicky.AddNamedCommandWithContext("list", comments, externalIssueRef{}, func(ctx context.Context, opts externalIssueRef) (any, error) {
		id, err := opts.one()
		if err != nil {
			return nil, err
		}
		raw, err := client().Get(ctx, id)
		if err != nil {
			return nil, err
		}
		return jira.ParseComments(raw)
	})
	clicky.AddNamedCommandWithContext("create", comments, externalCommentCreate{}, func(ctx context.Context, opts externalCommentCreate) (any, error) {
		id, err := opts.one()
		if err != nil {
			return nil, err
		}
		raw, err := client().Call(ctx, "jira", "workitem", "comment", "create", "--issue-id", id, "--body", opts.Body, "--body-format", "markdown")
		if err != nil {
			return nil, err
		}
		return decodeExternalJSON(raw)
	})
	attachments := &cobra.Command{Use: "attachment", Short: "Manage issue attachments"}
	provider.AddCommand(attachments)
	clicky.AddNamedCommandWithContext("list", attachments, externalIssueRef{}, func(ctx context.Context, opts externalIssueRef) (any, error) {
		id, err := opts.one()
		if err != nil {
			return nil, err
		}
		raw, err := client().Call(ctx, "jira", "workitem", "attachment", "query", "--issue-id", id)
		if err != nil {
			return nil, err
		}
		return jira.ParseAttachments(raw)
	})
	clicky.AddNamedCommandWithContext("upload", attachments, externalAttachmentUpload{}, func(ctx context.Context, opts externalAttachmentUpload) (any, error) {
		id, err := opts.one()
		if err != nil {
			return nil, err
		}
		args := []string{"jira", "workitem", "attachment", "upload", "--issue-id", id}
		for _, file := range opts.Files {
			args = append(args, "--file", file)
		}
		raw, err := client().Call(ctx, args...)
		if err != nil {
			return nil, err
		}
		return decodeExternalJSON(raw)
	})
	clicky.AddNamedCommandWithContext("delete", attachments, externalAttachmentDelete{}, func(ctx context.Context, opts externalAttachmentDelete) (any, error) {
		id, err := opts.one()
		if err != nil {
			return nil, err
		}
		if !opts.Yes {
			return nil, fmt.Errorf("attachment deletion is permanent; pass --yes to confirm")
		}
		raw, err := client().Call(ctx, "jira", "workitem", "attachment", "delete", "--attachment-id", id, "--yes")
		if err != nil {
			return nil, err
		}
		return decodeExternalJSON(raw)
	})
	clicky.AddNamedCommandWithContext("download", attachments, externalAttachmentDownload{}, func(ctx context.Context, opts externalAttachmentDownload) (any, error) {
		id, err := opts.one()
		if err != nil {
			return nil, err
		}
		_, err = client().Call(ctx, "jira", "workitem", "attachment", "download", "--attachment-id", id, "--output-path", opts.Output)
		return nil, err
	})
}
