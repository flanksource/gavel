package jira

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/externalissues/syncdoc"
	"github.com/flanksource/gavel/internal/safefs"
)

const DefaultSyncDir = ".gavel/jira"

type SyncOptions struct {
	CloudID     string
	From        string
	Output      string
	Force       bool
	Keys        []string
	Dir         string
	Attachments bool
	Extract     bool
	NoCache     bool
	Store       CatalogStore
}

type SyncResult struct {
	artifacts   []string
	Key         string   `json:"key"`
	Path        string   `json:"path"`
	Attachments []string `json:"attachments,omitempty"`
	Extracted   []string `json:"extracted,omitempty"`
	Images      []string `json:"images,omitempty"`
}

func (r SyncResult) Pretty() api.Text {
	return clicky.Text(r.Key).Space().Add(clicky.Link(r.Path).AddText(r.Path))
}

type SyncResults []SyncResult

func (results SyncResults) Pretty() api.Text {
	var text api.Text
	for i, result := range results {
		if i > 0 {
			text = text.NewLine()
		}
		text = text.Add(result.Pretty())
	}
	return text
}

func (options SyncOptions) Validate() error {
	if options.Output != "" && len(options.Keys) != 1 {
		return fmt.Errorf("in-place sync requires exactly one issue")
	}
	if len(options.Keys) == 0 {
		return fmt.Errorf("Jira sync requires at least one issue key or ID")
	}
	for _, key := range options.Keys {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("Jira sync issue keys must not be empty")
		}
	}
	return nil
}

func (c *Client) Sync(ctx context.Context, options SyncOptions) (SyncResults, error) {
	if options.From != "" && options.Output == "" {
		resolved, _, err := ResolveSyncFile(options, c.Site)
		if err != nil {
			return nil, err
		}
		options = resolved
	}
	if err := options.Validate(); err != nil {
		return nil, err
	}
	if options.Dir == "" {
		options.Dir = DefaultSyncDir
	}
	indexDir := options.Dir
	if options.Output != "" {
		indexDir = filepath.Dir(options.Output)
	}
	files, err := indexIssueFiles(indexDir)
	if err != nil {
		return nil, err
	}
	return runJiraTask(ctx, "Sync Jira issues", func(ctx context.Context, progress *task.Task) (SyncResults, error) {
		results := make(SyncResults, 0, len(options.Keys))
		progress.SetProgress(0, len(options.Keys))
		for _, key := range options.Keys {
			progress.SetDescription(fmt.Sprintf("[%d/%d] Syncing %s", len(results)+1, len(options.Keys), key))
			result, err := c.syncIssue(ctx, syncIssueOptions{SyncOptions: options, Key: strings.TrimSpace(key), Files: files})
			if err != nil {
				progress.SetDescription(fmt.Sprintf("Stopped at %s; %d/%d issues saved", key, len(results), len(options.Keys)))
				return results, fmt.Errorf("sync Jira issue %s (%d/%d): %w", key, len(results)+1, len(options.Keys), err)
			}
			results = append(results, result)
			files[result.Key] = result.Path
			progress.SetProgress(len(results), len(options.Keys))
			progress.Infof("Saved %s to %s", result.Key, result.Path)
			options.NoCache = false
		}
		progress.SetDescription(fmt.Sprintf("Saved %d issues", len(results)))
		return results, nil
	})
}

type syncIssueOptions struct {
	SyncOptions
	Key   string
	Files map[string]string
}

func (c *Client) syncIssue(ctx context.Context, options syncIssueOptions) (result SyncResult, returnErr error) {
	return runJiraTask(ctx, "Sync "+options.Key, func(ctx context.Context, progress *task.Task) (result SyncResult, returnErr error) {
		progress.SetProgress(0, 3)
		progress.SetDescription("Loading issue, fields and comments")
		issue, err := c.Get(ctx, GetOptions{Key: options.Key, Full: true, NoCache: options.NoCache, Store: options.Store})
		if err != nil {
			return result, err
		}
		if err := validateSyncIssue(issue); err != nil {
			return result, err
		}
		progress.SetProgress(1, 3)
		if options.Dir == "" {
			options.Dir = DefaultSyncDir
		}
		projectDir := issue.Project.ID
		filename := issue.Key + ".md"
		if options.Output != "" {
			options.Dir = filepath.Dir(options.Output)
			projectDir = "."
			filename = filepath.Base(options.Output)
		}
		if err := os.MkdirAll(options.Dir, 0o700); err != nil {
			return result, err
		}
		base, err := os.OpenRoot(options.Dir)
		if err != nil {
			return result, err
		}
		defer func() { returnErr = errors.Join(returnErr, base.Close()) }()
		if err := base.MkdirAll(projectDir, 0o700); err != nil {
			return result, err
		}
		root, err := base.OpenRoot(projectDir)
		if err != nil {
			return result, err
		}
		defer func() { returnErr = errors.Join(returnErr, root.Close()) }()
		name := ".sync-" + rand.Text()
		if err := root.Mkdir(name, 0o700); err != nil {
			return result, err
		}
		dir := filepath.Join(options.Dir, projectDir)
		stage := filepath.Join(dir, name)
		defer func() {
			returnErr = errors.Join(returnErr, safefs.SafeRemove(stage, safefs.RemoveContext{AllowedRoot: dir}))
		}()
		result = SyncResult{Key: issue.Key, Path: filepath.Join(dir, filename)}
		doc, err := syncdoc.Read(root, filename)
		if err != nil {
			return result, err
		}
		if err := prepareSyncMetadata(doc, issue, options.SyncOptions); err != nil {
			return result, err
		}
		doc.Metadata.Source.CloudID = c.credentials.CloudID
		var extracted string
		if options.Attachments || options.Extract {
			progress.SetDescription(fmt.Sprintf("Processing %d attachments", len(issue.Attachments)))
			extracted, err = c.syncAttachments(ctx, syncAttachmentOptions{Issue: &issue, Stage: stage, Result: &result, Extract: options.Extract, Root: root, Document: doc, Force: options.Force})
			if err != nil {
				return SyncResult{}, err
			}
		}
		progress.SetProgress(2, 3)
		progress.SetDescription("Rendering and saving Markdown")
		links := localLinkOptions{Site: c.Site, Path: result.Path, Files: options.Files}
		view, err := issue.Render(RenderOptions{Full: true, ResolveIssueURL: links.resolve})
		if err != nil {
			return SyncResult{}, err
		}
		finishSyncMetadata(doc, root)
		data, err := doc.Markdown(strings.TrimSpace(view.Markdown()) + extracted)
		if err != nil {
			return result, err
		}
		if err := os.WriteFile(filepath.Join(stage, "issue.md"), data, 0o600); err != nil {
			return SyncResult{}, err
		}
		if err := ctx.Err(); err != nil {
			return SyncResult{}, err
		}
		if err := doc.CheckUnchanged(root, filename); err != nil {
			return result, err
		}
		if err := publishSyncFiles(root, stage, result); err != nil {
			return SyncResult{}, err
		}
		progress.SetProgress(3, 3)
		progress.SetDescription("Saved " + result.Path)
		return result, nil
	})
}

func validateSyncIssue(issue Issue) error {
	if issue.Project == nil || !syncComponent(issue.Project.ID) || !syncComponent(issue.Key) {
		return fmt.Errorf("Jira sync requires a safe issue key and project ID for %q", issue.Key)
	}
	ids := make(map[string]bool)
	for _, attachment := range issue.Attachments {
		if !syncComponent(attachment.ID) || ids[attachment.ID] {
			return fmt.Errorf("issue %s has an invalid or duplicate attachment ID %q", issue.Key, attachment.ID)
		}
		ids[attachment.ID] = true
		if !syncComponent(attachment.Name) {
			return fmt.Errorf("issue %s attachment %s has an unsafe filename %q", issue.Key, attachment.ID, attachment.Name)
		}
	}
	return nil
}

func syncComponent(value string) bool {
	return value != "" && value != "." && value != ".." && filepath.IsLocal(value) && !strings.ContainsAny(value, "/\\") && !strings.ContainsFunc(value, unicode.IsControl)
}

type syncAttachmentOptions struct {
	Root     *os.Root
	Document *syncdoc.Document
	Force    bool
	Issue    *Issue
	Stage    string
	Result   *SyncResult
	Extract  bool
}

func (c *Client) syncAttachments(ctx context.Context, options syncAttachmentOptions) (string, error) {
	return runJiraTask(ctx, "Attachments "+options.Issue.Key, func(ctx context.Context, progress *task.Task) (string, error) {
		var extracted strings.Builder
		progress.SetProgress(0, len(options.Issue.Attachments))
		for index := range options.Issue.Attachments {
			attachment := &options.Issue.Attachments[index]
			progress.SetDescription(fmt.Sprintf("[%d/%d] Downloading %s", index+1, len(options.Issue.Attachments), attachment.Name))
			relative := filepath.Join(options.Issue.Key+".attachments", attachment.ID, attachment.Name)
			path := filepath.Join(options.Stage, relative)
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				return "", err
			}
			record := options.Document.Metadata.Attachments[attachment.ID]
			if err := c.syncDownload(ctx, syncDownloadOptions{Attachment: attachment, Record: record, Root: options.Root, Stage: options.Stage, Relative: relative, Force: options.Force}); err != nil {
				return "", fmt.Errorf("issue %s attachment %s (%s): %w", options.Issue.Key, attachment.ID, attachment.Name, err)
			}
			attachment.URL = (&url.URL{Path: filepath.ToSlash(relative)}).String()
			options.Result.Attachments = append(options.Result.Attachments, filepath.Join(filepath.Dir(options.Result.Path), relative))
			if !options.Extract || documentKind(*attachment) == "" {
				progress.SetProgress(index+1, len(options.Issue.Attachments))
				continue
			}
			progress.SetDescription(fmt.Sprintf("[%d/%d] Extracting %s", index+1, len(options.Issue.Attachments), attachment.Name))
			body, err := syncExtraction(ctx, syncExtractionOptions{Attachment: *attachment, Record: record, Root: options.Root, Stage: options.Stage, Path: path, Force: options.Force, Result: options.Result})
			if err != nil {
				return "", fmt.Errorf("issue %s attachment %s (%s): %w", options.Issue.Key, attachment.ID, attachment.Name, err)
			}
			heading := clicky.Heading(3, (Attachment{Attachment: *attachment}).PrettyShort()).Markdown()
			extracted.WriteString("\n\n" + heading + "\n\n" + strings.TrimSpace(body.Markdown))
			for _, image := range body.Images {
				path := filepath.Join(filepath.Dir(options.Result.Path), filepath.FromSlash(image))
				if !slices.Contains(options.Result.Images, path) {
					options.Result.Images = append(options.Result.Images, path)
				}
			}
			options.Result.Extracted = append(options.Result.Extracted, attachment.ID)
			progress.SetProgress(index+1, len(options.Issue.Attachments))
		}
		progress.SetDescription(fmt.Sprintf("Saved %d attachments; extracted %d documents", len(options.Result.Attachments), len(options.Result.Extracted)))
		if extracted.Len() == 0 {
			return "", nil
		}
		return "\n\n" + clicky.Heading(2, clicky.Text("Extracted attachments")).Markdown() + extracted.String(), nil
	})
}

func publishSyncFiles(root *os.Root, stage string, result SyncResult) error {
	seen := map[string]bool{}
	for _, path := range append(append(append([]string{}, result.Attachments...), result.Images...), result.artifacts...) {
		if seen[path] {
			continue
		}
		seen[path] = true
		relative, err := filepath.Rel(filepath.Dir(result.Path), path)
		if err != nil {
			return err
		}
		if err := root.MkdirAll(filepath.Dir(relative), 0o700); err != nil {
			return err
		}
		if err := root.Rename(filepath.Join(filepath.Base(stage), relative), relative); err != nil {
			return err
		}
	}
	return root.Rename(filepath.Join(filepath.Base(stage), "issue.md"), filepath.Base(result.Path))
}
