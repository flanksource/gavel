package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/formatters/excel"
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
	JQL    string `flag:"jql" required:"true" help:"Jira JQL query"`
	Limit  int    `flag:"limit" help:"Maximum number of issues"`
	After  string `flag:"after" help:"Pagination cursor"`
	Fields string `flag:"fields" help:"Comma-separated Jira fields to include"`
}

type externalIssueGet struct {
	externalIssueRef
	Full bool `flag:"full" help:"Render the complete description, attachments and comments"`
}

type externalIssueSync struct {
	IDs         []string `args:"true" clicky:"cli-file-read" help:"Jira issue keys or IDs, or @issues.xlsx:[Sheet!]Column / @issues.csv:Column"`
	From        string   `flag:"from" help:"Resync a saved Markdown file in place using its frontmatter"`
	Force       bool     `flag:"force" help:"Refresh downloads and extraction without reusing cached artifacts"`
	SyncDir     string   `flag:"sync-dir" help:"Root directory for <project-id>/<issue-key>.md files (default .gavel/jira)"`
	Attachments *bool    `flag:"attachments" help:"Download attachments and link to local files"`
	Extract     *bool    `flag:"extract" help:"Download attachments and include PDF/DOCX Markdown with images in attachments/cas (requires uvx)"`
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

func Register(rootCmd *cobra.Command) {
	excel.Use()
	root := &cobra.Command{Use: "external-issues", Short: "Work with issues in external systems"}
	provider := &cobra.Command{Use: "jira", Short: "Work with Jira through its REST API"}
	root.AddCommand(provider)
	rootCmd.AddCommand(root)
	var site, profile string
	var noCache bool
	provider.PersistentFlags().StringVar(&site, "site", "", "Jira site URL or Atlassian site prefix")
	provider.PersistentFlags().StringVar(&profile, "profile", "", "TWG credential profile")
	provider.PersistentFlags().BoolVar(&noCache, "no-cache", false, "Rebuild the Jira field metadata cache")
	client := func(ctx context.Context) (*jira.Client, error) {
		return jira.NewClient(ctx, jira.ClientOptions{Site: site, Profile: profile})
	}
	store := func() jira.CatalogStore { return jira.CatalogStore{Dir: os.Getenv("GAVEL_JIRA_CACHE_DIR")} }
	syncCmd := clicky.AddNamedCommandWithContext("sync", provider, externalIssueSync{}, func(ctx context.Context, opts externalIssueSync) (jira.SyncResults, error) {
		options := jira.SyncOptions{Keys: opts.IDs, Dir: opts.SyncDir, From: opts.From, Force: opts.Force, NoCache: noCache, Store: store()}
		resolved, selectedSite, err := jira.ResolveSyncFile(options, site)
		if err != nil {
			return nil, err
		}
		options = resolved
		if opts.Attachments != nil {
			options.Attachments = *opts.Attachments
		}
		if opts.Extract != nil {
			options.Extract = *opts.Extract
		}
		if options.Extract && opts.Attachments != nil && !*opts.Attachments {
			return nil, fmt.Errorf("extraction requires attachments; disable extraction as well")
		}
		if err := options.Validate(); err != nil {
			return nil, err
		}
		c, err := jira.NewClient(ctx, jira.ClientOptions{Site: selectedSite, Profile: profile, Credentials: jira.Credentials{CloudID: options.CloudID}})
		if err != nil {
			return nil, err
		}
		return c.Sync(ctx, options)
	})
	syncCmd.Use = "sync <id> [id...]"
	syncCmd.Short = "Sync Jira issues to local Markdown files"
	syncCmd.Long = "Sync Jira issues to local Markdown files. Read issue keys or IDs from an Excel column with @file.xlsx:Column, @file.xlsx:Sheet!Column, or @file.xlsx:#2!Column. Column is the header text, and worksheet positions start at one. Quote the entire argument when it contains spaces or !. Empty cells are skipped."
	syncCmd.Example = `  # Read the Issue Key column from the first worksheet
  gavel external-issues jira sync '@issues.xlsx:Issue Key'

  # Select a named worksheet or the second worksheet
  gavel external-issues jira sync '@issues.xlsx:Sprint Backlog!Issue Key'
  gavel external-issues jira sync '@issues.xlsx:#2!Issue Key'

  # Use a workbook path with spaces and download attachments
  gavel external-issues jira sync '@./exports/Sprint issues.xlsx:Issue Key' --attachments

  # Extract PDF/DOCX attachments into Markdown under a custom directory
  gavel external-issues jira sync '@issues.xlsx:Sprint Backlog!Issue Key' --extract --sync-dir ./issues

  # Combine an Excel column with another workbook and an explicit issue key
  gavel external-issues jira sync '@issues.xlsx:Issue Key' '@follow-up.xlsx:Issue Key' EXAMPLE-99

  # Read a CSV column instead
  gavel external-issues jira sync '@issues.csv:Issue Key'`
	clicky.AddNamedCommandWithContext("get", provider, externalIssueGet{}, func(ctx context.Context, opts externalIssueGet) (jira.Issue, error) {
		id, err := opts.one()
		if err != nil {
			return jira.Issue{}, err
		}
		c, err := client(ctx)
		if err != nil {
			return jira.Issue{}, err
		}
		return c.Get(ctx, jira.GetOptions{Key: id, Full: opts.Full, NoCache: noCache, Store: store()})
	})
	clicky.AddNamedCommandWithContext("list", provider, externalIssueList{}, func(ctx context.Context, opts externalIssueList) ([]jira.Issue, error) {
		c, err := client(ctx)
		if err != nil {
			return nil, err
		}
		var fields []string
		if opts.Fields != "" {
			fields = strings.Split(opts.Fields, ",")
		}
		return c.List(ctx, jira.SearchOptions{JQL: opts.JQL, Limit: opts.Limit, After: opts.After, Fields: fields, NoCache: noCache, Store: store()})
	})
	comments := &cobra.Command{Use: "comment", Short: "Read or create issue comments"}
	provider.AddCommand(comments)
	clicky.AddNamedCommandWithContext("list", comments, externalIssueRef{}, func(ctx context.Context, opts externalIssueRef) (jira.Comments, error) {
		id, err := opts.one()
		if err != nil {
			return nil, err
		}
		c, err := client(ctx)
		if err != nil {
			return nil, err
		}
		return c.Comments(ctx, id)
	})
	clicky.AddNamedCommandWithContext("create", comments, externalCommentCreate{}, func(ctx context.Context, opts externalCommentCreate) (any, error) {
		id, err := opts.one()
		if err != nil {
			return nil, err
		}
		c, err := client(ctx)
		if err != nil {
			return nil, err
		}
		return c.CreateComment(ctx, jira.CommentOptions{Key: id, Body: opts.Body})
	})
	attachments := &cobra.Command{Use: "attachment", Short: "Manage issue attachments"}
	provider.AddCommand(attachments)
	clicky.AddNamedCommandWithContext("list", attachments, externalIssueRef{}, func(ctx context.Context, opts externalIssueRef) ([]jira.Attachment, error) {
		id, err := opts.one()
		if err != nil {
			return nil, err
		}
		c, err := client(ctx)
		if err != nil {
			return nil, err
		}
		return c.Attachments(ctx, id)
	})
	clicky.AddNamedCommandWithContext("upload", attachments, externalAttachmentUpload{}, func(ctx context.Context, opts externalAttachmentUpload) ([]jira.Attachment, error) {
		id, err := opts.one()
		if err != nil {
			return nil, err
		}
		c, err := client(ctx)
		if err != nil {
			return nil, err
		}
		return c.UploadAttachments(ctx, jira.UploadOptions{Key: id, Files: opts.Files})
	})
	clicky.AddNamedCommandWithContext("delete", attachments, externalAttachmentDelete{}, func(ctx context.Context, opts externalAttachmentDelete) (any, error) {
		id, err := opts.one()
		if err != nil {
			return nil, err
		}
		if !opts.Yes {
			return nil, fmt.Errorf("attachment deletion is permanent; pass --yes to confirm")
		}
		c, err := client(ctx)
		if err != nil {
			return nil, err
		}
		return nil, c.DeleteAttachment(ctx, id)
	})
	clicky.AddNamedCommandWithContext("download", attachments, externalAttachmentDownload{}, func(ctx context.Context, opts externalAttachmentDownload) (any, error) {
		id, err := opts.one()
		if err != nil {
			return nil, err
		}
		c, err := client(ctx)
		if err != nil {
			return nil, err
		}
		_, err = c.DownloadAttachment(ctx, jira.DownloadOptions{ID: id, Output: opts.Output})
		return nil, err
	})
}
