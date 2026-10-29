package jira

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"

	"github.com/flanksource/clicky/task"
)

func (c *Client) Attachments(ctx context.Context, key string) ([]Attachment, error) {
	if key == "" {
		return nil, fmt.Errorf("Jira issue key is required")
	}
	var issue struct {
		Fields struct {
			Attachments json.RawMessage `json:"attachment"`
		} `json:"fields"`
	}
	if err := c.json(ctx, requestOptions{Path: "/issue/" + url.PathEscape(key), Query: url.Values{"fields": {"attachment"}}}, &issue); err != nil {
		return nil, err
	}
	if len(issue.Fields.Attachments) == 0 || string(issue.Fields.Attachments) == "null" {
		return nil, fmt.Errorf("Jira issue response has no attachment array")
	}
	return ParseAttachments(issue.Fields.Attachments)
}

type UploadOptions struct {
	Key   string
	Files []string
}

func (c *Client) UploadAttachments(ctx context.Context, options UploadOptions) ([]Attachment, error) {
	if options.Key == "" || len(options.Files) == 0 {
		return nil, fmt.Errorf("attachment upload requires an issue key and files")
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for _, path := range options.Files {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		part, err := writer.CreateFormFile("file", filepath.Base(path))
		if err != nil {
			return nil, errors.Join(err, file.Close())
		}
		_, copyErr := io.Copy(part, file)
		if err := errors.Join(copyErr, file.Close()); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	var raw json.RawMessage
	if err := c.json(ctx, requestOptions{Method: "POST", Path: "/issue/" + url.PathEscape(options.Key) + "/attachments", Body: body.Bytes(), Headers: map[string]string{"Content-Type": writer.FormDataContentType(), "X-Atlassian-Token": "no-check"}}, &raw); err != nil {
		return nil, err
	}
	return ParseAttachments(raw)
}

type DownloadOptions struct {
	ETag   string
	ID     string
	Output string
}

type DownloadResult struct {
	SHA256      string
	ETag        string
	NotModified bool
}

func (c *Client) DownloadAttachment(ctx context.Context, options DownloadOptions) (DownloadResult, error) {
	if options.ID == "" || options.Output == "" {
		return DownloadResult{}, fmt.Errorf("attachment download requires an ID and output path")
	}
	return runJiraTask(ctx, "Download Jira attachment "+options.ID, func(ctx context.Context, progress *task.Task) (_ DownloadResult, returnErr error) {
		progress.SetDescription("Downloading " + filepath.Base(options.Output))
		headers := map[string]string{}
		if options.ETag != "" {
			headers["If-None-Match"] = options.ETag
		}
		response, err := c.request(ctx, requestOptions{Path: "/attachment/content/" + url.PathEscape(options.ID), Query: url.Values{"redirect": {"false"}}, Headers: headers, AllowNotModified: options.ETag != ""})
		if err != nil {
			return DownloadResult{}, err
		}
		defer func() { returnErr = errors.Join(returnErr, response.Body.Close()) }()
		if response.StatusCode == 304 {
			progress.SetDescription("Reusing verified attachment " + options.ID)
			return DownloadResult{ETag: options.ETag, NotModified: true}, nil
		}
		file, err := os.OpenFile(options.Output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err != nil {
			return DownloadResult{}, err
		}
		counter := &attachmentProgressWriter{task: progress, name: filepath.Base(options.Output), total: response.ContentLength}
		if counter.total > 0 {
			progress.SetProgress(0, int(counter.total))
		}
		hash := sha256.New()
		written, copyErr := io.Copy(io.MultiWriter(file, counter, hash), response.Body)
		if err := errors.Join(copyErr, file.Close()); err != nil {
			return DownloadResult{}, errors.Join(err, os.Remove(options.Output))
		}
		progress.SetDescription(fmt.Sprintf("Saved %s (%d bytes)", options.Output, written))
		return DownloadResult{SHA256: fmt.Sprintf("%x", hash.Sum(nil)), ETag: response.Header.Get("ETag")}, nil
	})
}

type attachmentProgressWriter struct {
	task    *task.Task
	name    string
	total   int64
	written int64
}

func (w *attachmentProgressWriter) Write(data []byte) (int, error) {
	w.written += int64(len(data))
	if w.total > 0 {
		w.task.SetProgress(int(w.written), int(w.total))
	}
	w.task.SetDescription(fmt.Sprintf("Downloading %s: %d bytes", w.name, w.written))
	return len(data), nil
}

func (c *Client) DeleteAttachment(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("Jira attachment ID is required")
	}
	return c.json(ctx, requestOptions{Method: "DELETE", Path: "/attachment/" + url.PathEscape(id)}, nil)
}
