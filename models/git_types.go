package models

import (
	"fmt"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/models/kubernetes"
)

// GitHubReference represents a reference to a GitHub issue or PR
type GitHubReference struct {
	Number int    `json:"number,omitempty"`
	Type   string `json:"type,omitempty"` // "issue" or "pull"
	URL    string `json:"url,omitempty"`
}

// ManagedRepository represents a git repository with multiple worktrees
type ManagedRepository struct {
}

// ScanJob represents a scanning task for a specific file at a specific depth
type ScanJob struct {
	Path        string // Directory path containing the file
	FilePath    string // Specific file to scan (e.g., "go.mod", "Chart.yaml")
	GitURL      string
	Version     string
	Depth       int
	IsLocal     bool
	Parent      string
	ScannerType string // Type of scanner needed ("go", "helm", "npm", etc.)
}

// VersionInstance tracks a specific version of a dependency
type VersionInstance struct {
	Version     string
	Depth       int
	Source      string // Which parent dependency led to this
	GitWorktree string // Path to checked out version
}

// ResolveResult contains the result of version resolution
type ResolveResult struct {
	OriginalAlias   string
	ResolvedVersion string
	ResolvedAt      time.Time
	Error           error
}

// CacheEntry represents a cached git operation result
type CacheEntry struct {
	Value      interface{}
	Timestamp  time.Time
	AccessedAt time.Time
	Error      error
}

// CloneInfo contains information about a managed clone
type CloneInfo struct {
	Path      string
	Version   string
	Depth     int // Clone depth (0 = full, >0 = shallow)
	CreatedAt time.Time
	LastUsed  time.Time
	Hash      string
}

// GitRef is a Git reference (branch, tag) "latest"
type GitRef string

type GitRange struct {
	Start GitRef `json:"start,omitempty"`
	End   GitRef `json:"end,omitempty"`
}

type ReleaseNotesOptions struct {
	IncludeInternalSummary bool         `json:"include_internal_summary,omitempty"`
	IncludeExternalSummary bool         `json:"include_external_summary,omitempty"`
	CommitTypes            []CommitType `json:"commit_types,omitempty"`
	GitRange               `json:",inline"`
}

type ReleaseNotes struct {
	Options      ReleaseNotesOptions `json:"options,omitempty"`
	Commits      CommitAnalyses      `json:"commits,omitempty"`
	Contributors []Author            `json:"contributors,omitempty"`
}

type FileMap struct {
	Path       string
	Scopes     Scopes      `yaml:"scope,omitempty"`
	Language   string      `yaml:"language,omitempty"`
	Tech       Technology  `yaml:"tech,omitempty"`
	Violations []Violation `yaml:"violations,omitempty"`
	// Ignored by .gitignore or arch.yaml ignore settings
	Ignored        bool                       `yaml:"ignored,omitempty"`
	Commits        CommitAnalyses             `yaml:"commits,omitempty"`
	KubernetesRefs []kubernetes.KubernetesRef `yaml:"kubernetes_refs,omitempty" json:"kubernetes_refs,omitempty"`
}

func (f FileMap) Pretty() api.Text {

	t := clicky.Text(f.Path)
	if len(f.Scopes) > 0 {
		t = t.Append(" scopes: ", "text-muted").Append(f.Scopes)
	}
	if len(f.Tech) > 0 {
		t = t.Append(" tech: ", "text-muted").Append(f.Tech)
	}
	if f.Ignored {
		t = t.Append(" [IGNORED]", "text-yellow-600")
	}

	// Display Kubernetes resources if present
	if len(f.KubernetesRefs) > 0 {
		t = t.NewLine().NewLine()
		t = t.Append("🎯 Kubernetes Resources", "font-bold text-blue-600").NewLine()

		for _, ref := range f.KubernetesRefs {
			// Resource kind and API version
			t = t.Append("  ")
			if ref.APIVersion != "" && ref.Kind != "" {
				t = t.Append(ref.APIVersion+" "+ref.Kind, "font-medium text-cyan-600").NewLine()
			} else if ref.Kind != "" {
				t = t.Append(ref.Kind, "font-medium text-cyan-600").NewLine()
			}

			// Name
			if ref.Name != "" {
				t = t.Append("    Name: ", "text-muted").Append(ref.Name).NewLine()
			}

			// Namespace
			if ref.Namespace != "" {
				t = t.Append("    Namespace: ", "text-muted").Append(ref.Namespace).NewLine()
			}

			// Line range
			if ref.StartLine > 0 && ref.EndLine > 0 {
				t = t.Append("    Lines: ", "text-muted").Append(fmt.Sprintf("%d-%d", ref.StartLine, ref.EndLine)).NewLine()
			}

			// Labels
			if len(ref.Labels) > 0 {
				t = t.Append("    Labels:").NewLine()
				for k, v := range ref.Labels {
					t = t.Append("      "+k+": ", "text-muted").Append(v).NewLine()
				}
			}

			// Annotations
			if len(ref.Annotations) > 0 {
				t = t.Append("    Annotations:").NewLine()
				for k, v := range ref.Annotations {
					t = t.Append("      "+k+": ", "text-muted").Append(v).NewLine()
				}
			}

			t = t.NewLine()
		}
	}

	return t

}
