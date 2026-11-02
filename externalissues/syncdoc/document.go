package syncdoc

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/gavel/externalissues"
	"gopkg.in/yaml.v3"
)

var ErrArtifactHashMismatch = errors.New("artifact hash mismatch")

type Artifact struct {
	Extra        map[string]yaml.Node `yaml:",inline"`
	Path         string               `yaml:"path"`
	SHA256       string               `yaml:"sha256"`
	MIMEType     string               `yaml:"mime_type,omitempty"`
	ETag         string               `yaml:"etag,omitempty"`
	DownloadedAt time.Time            `yaml:"downloaded_at,omitempty"`
}
type Source struct {
	Extra      map[string]yaml.Node `yaml:",inline"`
	Provider   string               `yaml:"provider"`
	Site       string               `yaml:"site"`
	CloudID    string               `yaml:"cloud_id,omitempty"`
	ID         string               `yaml:"id,omitempty"`
	Key        string               `yaml:"key"`
	URL        string               `yaml:"url,omitempty"`
	Title      string               `yaml:"title,omitempty"`
	ProjectID  string               `yaml:"project_id,omitempty"`
	ProjectKey string               `yaml:"project_key,omitempty"`
	UpdatedAt  *time.Time           `yaml:"updated_at,omitempty"`
}
type Settings struct {
	Extra       map[string]yaml.Node `yaml:",inline"`
	SyncedAt    time.Time            `yaml:"synced_at"`
	Attachments bool                 `yaml:"attachments"`
	Extract     bool                 `yaml:"extract"`
}
type Extraction struct {
	Extra               map[string]yaml.Node `yaml:",inline"`
	SourceSHA256        string               `yaml:"source_sha256"`
	PipelineFingerprint string               `yaml:"pipeline_fingerprint"`
	Tools               map[string]string    `yaml:"tools"`
	ExtractedAt         time.Time            `yaml:"extracted_at"`
	Markdown            Artifact             `yaml:"markdown"`
	Images              []Artifact           `yaml:"images,omitempty"`
}
type Enrichment struct {
	SourceSHA256 string               `yaml:"source_sha256,omitempty"`
	Stale        bool                 `yaml:"stale"`
	Result       yaml.Node            `yaml:"result,omitempty"`
	Artifacts    []Artifact           `yaml:"artifacts,omitempty"`
	Extra        map[string]yaml.Node `yaml:",inline"`
}
type Attachment struct {
	Name        string                `yaml:"name"`
	URL         string                `yaml:"url"`
	MIMEType    string                `yaml:"mime_type,omitempty"`
	Size        int64                 `yaml:"size"`
	CreatedAt   *time.Time            `yaml:"created_at,omitempty"`
	Present     bool                  `yaml:"present"`
	Local       *Artifact             `yaml:"local,omitempty"`
	Extraction  *Extraction           `yaml:"extraction,omitempty"`
	Enrichments map[string]Enrichment `yaml:"enrichments,omitempty"`
	Extra       map[string]yaml.Node  `yaml:",inline"`
}

// Issue is the issue's own fields as structured data, rewritten from the provider on every sync.
type Issue struct {
	Type           string                        `yaml:"type,omitempty"`
	Status         string                        `yaml:"status,omitempty"`
	StatusCategory externalissues.StatusCategory `yaml:"status_category,omitempty"`
	Priority       string                        `yaml:"priority,omitempty"`
	Assignee       string                        `yaml:"assignee,omitempty"`
	Reporter       string                        `yaml:"reporter,omitempty"`
	Project        string                        `yaml:"project,omitempty"`
	Created        *time.Time                    `yaml:"created,omitempty"`
	Updated        *time.Time                    `yaml:"updated,omitempty"`
	Resolved       *time.Time                    `yaml:"resolved,omitempty"`
	Resolution     string                        `yaml:"resolution,omitempty"`
	Due            string                        `yaml:"due,omitempty"`
	Labels         []string                      `yaml:"labels,omitempty,flow"`
	Components     []string                      `yaml:"components,omitempty,flow"`
	Fields         map[string]any                `yaml:"fields,omitempty"`
}
type Metadata struct {
	Version     int                    `yaml:"version"`
	Source      Source                 `yaml:"source"`
	Issue       *Issue                 `yaml:"issue,omitempty"`
	Sync        Settings               `yaml:"sync"`
	Attachments map[string]*Attachment `yaml:"attachments"`
	Extra       map[string]yaml.Node   `yaml:",inline"`
}
type Document struct {
	Exists   bool                 `yaml:"-"`
	Metadata Metadata             `yaml:"gavel"`
	Extra    map[string]yaml.Node `yaml:",inline"`
	Original []byte               `yaml:"-"`
}

func Read(root *os.Root, path string) (*Document, error) {
	data, err := root.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &Document{}, nil
	}
	if err != nil {
		return nil, err
	}
	d := &Document{Original: data, Exists: true}
	normalized := bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	if !bytes.HasPrefix(normalized, []byte("---\n")) {
		return d, nil
	}
	lines := strings.Split(string(normalized), "\n")
	end := 1
	for end < len(lines) && strings.TrimSpace(lines[end]) != "---" && strings.TrimSpace(lines[end]) != "..." {
		end++
	}
	if end == len(lines) {
		return nil, fmt.Errorf("%s: unterminated frontmatter", path)
	}
	header := []byte(strings.Join(lines[1:end], "\n"))
	if err := yaml.Unmarshal(header, d); err != nil {
		return nil, fmt.Errorf("%s frontmatter: %w", path, err)
	}
	d.Original = data
	var envelope map[string]yaml.Node
	if err := yaml.Unmarshal(header, &envelope); err != nil {
		return nil, err
	}
	_, managed := envelope["gavel"]
	if managed && d.Metadata.Version != 1 {
		return nil, fmt.Errorf("unsupported sync metadata version %d", d.Metadata.Version)
	}
	for id, record := range d.Metadata.Attachments {
		if record == nil || strings.TrimSpace(id) == "" {
			return nil, fmt.Errorf("invalid attachment record %q", id)
		}
		var artifacts []Artifact
		if record.Local != nil {
			artifacts = append(artifacts, *record.Local)
		}
		if record.Extraction != nil {
			artifacts = append(artifacts, record.Extraction.Markdown)
			artifacts = append(artifacts, record.Extraction.Images...)
		}
		for _, enrichment := range record.Enrichments {
			artifacts = append(artifacts, enrichment.Artifacts...)
		}
		for _, artifact := range artifacts {
			if !filepath.IsLocal(artifact.Path) || strings.Contains(artifact.Path, "\\") {
				return nil, fmt.Errorf("attachment %s: unsafe artifact path %q", id, artifact.Path)
			}
		}
	}
	return d, nil
}

func (d *Document) Markdown(body string) ([]byte, error) {
	header, err := yaml.Marshal(d)
	if err != nil {
		return nil, err
	}
	return []byte("---\n" + string(header) + "---\n\n" + strings.TrimSpace(body) + "\n"), nil
}

func Hash(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }

func ReadArtifact(root *os.Root, artifact Artifact) ([]byte, error) {
	if !filepath.IsLocal(artifact.Path) || strings.Contains(artifact.Path, "\\") {
		return nil, fmt.Errorf("unsafe artifact path %q", artifact.Path)
	}
	data, err := root.ReadFile(artifact.Path)
	if err != nil {
		return nil, err
	}
	if Hash(data) != artifact.SHA256 {
		return nil, fmt.Errorf("%w: %s", ErrArtifactHashMismatch, artifact.Path)
	}
	return data, nil
}

func (d *Document) CheckUnchanged(root *os.Root, path string) error {
	data, err := root.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) && !d.Exists {
		return nil
	}
	if err != nil {
		return err
	}
	if !d.Exists || !bytes.Equal(data, d.Original) {
		return fmt.Errorf("sync document changed during sync: %s", path)
	}
	return nil
}
