package jira

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	osexec "os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/clicky/exec"
)

func convertDocumentMetafiles(ctx context.Context, dir string) (returnErr error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer func() { returnErr = errors.Join(returnErr, root.Close()) }()
	return fs.WalkDir(root.FS(), ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || (strings.ToLower(filepath.Ext(path)) != ".emf" && strings.ToLower(filepath.Ext(path)) != ".wmf") {
			return nil
		}
		data, err := root.ReadFile(path)
		if err != nil {
			return err
		}
		return convertDocumentMetafile(ctx, metafileOptions{Data: data, Source: path, Output: filepath.Join(dir, filepath.FromSlash(path)+".svg")})
	})
}

type metafileOptions struct {
	Data   []byte
	Source string
	Output string
}

func convertDocumentMetafile(ctx context.Context, options metafileOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	npm, err := osexec.LookPath("npm")
	if err != nil {
		return fmt.Errorf("convert metafile %q requires Node.js and npm: %w", options.Source, err)
	}
	file, err := os.OpenFile(options.Output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	result := exec.NewExec(npm, "exec", "--yes", "--package=emf-converter@"+extractionTools("docx")["emf-converter"], "--", "emf-converter", "-", "--format", "svg", "--out", options.Output).
		WithoutShell().WithStdin(bytes.NewReader(options.Data)).WithContext(ctx).WithTimeout(5 * time.Minute).WithCaptureLimit(16 * 1024).Run().Result()
	if result.Error != nil {
		return fmt.Errorf("convert metafile %q with emf-converter: %w: %s", options.Source, result.Error, strings.TrimSpace(result.Stderr))
	}
	data, err := os.ReadFile(options.Output)
	if err != nil {
		return err
	}
	if err := validateDocumentSVG(data); err != nil {
		return fmt.Errorf("convert metafile %q: %w", options.Source, err)
	}
	return nil
}
