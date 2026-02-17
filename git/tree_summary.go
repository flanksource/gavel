package git

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
	. "github.com/flanksource/gavel/models"

	"github.com/samber/lo"
)

type Count struct {
	// Number o lines added
	Adds int `json:"adds,omitempty"`
	// Number of lines deleted
	Dels int `json:"dels,omitempty"`
	// Number of commits
	Commits int `json:"commits,omitempty"`
	// Number of files changed
	Files int `json:"files,omitempty"`
	// Number of commits per scope
	Scopes map[ScopeType]int `json:"scopes,omitempty"`
	// Number of commits per commit type
	CommitTypes map[CommitType]int `json:"commit_types,omitempty"`
	// Number of commits per technology
	Tech map[ScopeTechnology]int `json:"tech,omitempty"`
}

func (c Count) Pretty() api.Text {
	t := clicky.Text("")

	for scope := range c.Scopes {
		t = t.Add(clicky.Badge(string(scope)))
	}

	for tech := range c.Tech {
		t = t.Add(clicky.Badge(string(tech)))
	}

	if c.Adds > 0 {
		t = t.Append(fmt.Sprintf("+%d", c.Adds), "text-green-600")
	}

	if c.Dels > 0 {
		if c.Adds > 0 {
			t = t.Space()
		}
		t = t.Append(fmt.Sprintf("-%d", c.Dels), "text-red-600")
	}

	if c.Commits > 0 {
		if c.Adds > 0 || c.Dels > 0 {
			t = t.Space()
		}
		t = t.Append(fmt.Sprintf("• %d commits", c.Commits), "text-muted")
	}

	if c.Files > 0 {
		t = t.Append(fmt.Sprintf("• %d files", c.Files), "text-muted")
	}

	return t
}

// Add merges another Count into this one
func (c *Count) Add(other Count) {
	c.Adds += other.Adds
	c.Dels += other.Dels
	c.Commits += other.Commits
	c.Files += other.Files

	if c.Scopes == nil {
		c.Scopes = make(map[ScopeType]int)
	}
	for scope, count := range other.Scopes {
		c.Scopes[scope] += count
	}

	if c.CommitTypes == nil {
		c.CommitTypes = make(map[CommitType]int)
	}
	for ctype, count := range other.CommitTypes {
		c.CommitTypes[ctype] += count
	}

	if c.Tech == nil {
		c.Tech = make(map[ScopeTechnology]int)
	}
	for tech, count := range other.Tech {
		c.Tech[tech] += count
	}
}

// Total returns the total number of line changes (adds + dels)
func (c Count) Total() int {
	return c.Adds + c.Dels
}

type PathSummary struct {
	Path        string `json:"path,omitempty"`
	Count       `json:",inline"`
	Children    []PathSummary       `json:"children,omitempty"`
	uniqueFiles map[string]struct{} `json:"-"` // Track unique files for accurate counting
}

func (gs PathSummary) Flatten(parent string) []PathSummary {
	var nodes []PathSummary
	if parent != "" {
		gs.Path = parent + "/" + gs.Path
	}

	for _, child := range gs.Children {
		nodes = append(nodes, child.Flatten(gs.Path)...)
	}
	gs.Children = nil
	nodes = append(nodes, gs)
	return nodes
}

func (gs PathSummary) GetChildren() []api.TreeNode {
	nodes := make([]api.TreeNode, len(gs.Children))
	for i, child := range gs.Children {
		nodes[i] = child
	}
	return nodes
}

func (gs PathSummary) Pretty() api.Text {
	t := clicky.Text("")

	isDir := len(gs.Children) > 0
	isRoot := gs.Path == "." || gs.Path == ""

	// Skip displaying root path
	if !isRoot {
		// Add icon
		if isDir {
			t = t.Add(icons.Folder).Space()
			t = t.Append(gs.Path+"/", "font-bold")
		} else {
			t = t.Add(getFileIcon(gs.Path)).Space()
			t = t.Append(gs.Path, "font-mono")
		}
		t = t.Space()
	}

	// Inline stats with colors
	t = t.Append("(", "text-muted")

	if gs.Adds > 0 {
		t = t.Append(fmt.Sprintf("+%d", gs.Adds), "text-green-600")
	}

	if gs.Dels > 0 {
		if gs.Adds > 0 {
			t = t.Space()
		}
		t = t.Append(fmt.Sprintf("-%d", gs.Dels), "text-red-600")
	}

	if gs.Commits > 1 {
		t = t.Append(fmt.Sprintf(" • %d commits", gs.Commits), "text-muted")
	}

	if isDir && gs.Files > 1 {
		t = t.Append(fmt.Sprintf(" • %d files", gs.Files), "text-muted")
	}

	// Show top commit types
	if len(gs.CommitTypes) > 0 {
		topTypes := lo.Entries(gs.CommitTypes)
		sort.Slice(topTypes, func(i, j int) bool {
			return topTypes[i].Value > topTypes[j].Value
		})

		t = t.Append(" • ", "text-muted")
		for i, entry := range topTypes {
			if i > 0 {
				t = t.Append(",", "text-muted")
			}
			if isDir {
				t = t.Append(fmt.Sprintf("%s:%d", entry.Key, entry.Value))
			} else {
				t = t.Append(string(entry.Key))
			}
			if i >= 2 {
				break
			}
		}
	}

	// Show non-extension-based technologies only
	if len(gs.Tech) > 0 {
		filteredTech := make(map[ScopeTechnology]int)
		for tech, count := range gs.Tech {
			// For directories, show all tech
			// For files, filter out extension-based tech
			if isDir || !isExtensionBasedTech(gs.Path, tech) {
				filteredTech[tech] = count
			}
		}

		if len(filteredTech) > 0 {
			topTech := lo.Entries(filteredTech)
			sort.Slice(topTech, func(i, j int) bool {
				return topTech[i].Value > topTech[j].Value
			})

			t = t.Append(" • ", "text-muted")
			for i, entry := range topTech {
				if i > 0 {
					t = t.Append(",", "text-muted")
				}
				if isDir {
					t = t.Append(fmt.Sprintf("%s:%d", entry.Key, entry.Value))
				} else {
					t = t.Append(string(entry.Key))
				}
				if i >= 2 {
					break
				}
			}
		}
	}

	t = t.Append(")", "text-muted")

	return t
}

// NewGitSummary creates a new GitSummary root node
func NewGitSummary(path string) *PathSummary {
	return &PathSummary{
		Path: path,
		Count: Count{
			Scopes:      make(map[ScopeType]int),
			CommitTypes: make(map[CommitType]int),
			Tech:        make(map[ScopeTechnology]int),
		},
		Children:    []PathSummary{},
		uniqueFiles: make(map[string]struct{}),
	}
}

// ensureChild finds or creates a child node with the given path
func (gs *PathSummary) ensureChild(path string) *PathSummary {
	for i := range gs.Children {
		if gs.Children[i].Path == path {
			return &gs.Children[i]
		}
	}

	child := PathSummary{
		Path: path,
		Count: Count{
			Scopes:      make(map[ScopeType]int),
			CommitTypes: make(map[CommitType]int),
			Tech:        make(map[ScopeTechnology]int),
		},
		Children:    []PathSummary{},
		uniqueFiles: make(map[string]struct{}),
	}
	gs.Children = append(gs.Children, child)
	return &gs.Children[len(gs.Children)-1]
}

// AddFile adds file statistics to the tree, creating directory nodes as needed
func (gs *PathSummary) AddFile(filePath string, changes []CommitChange) {
	if filePath == "" || filePath == "." {
		return
	}

	// Build path segments
	segments := strings.Split(filepath.Clean(filePath), string(filepath.Separator))

	// Walk tree creating nodes as needed
	current := gs
	for i, segment := range segments {
		isFile := i == len(segments)-1

		if isFile {
			// Create file node
			fileNode := current.ensureChild(segment)
			fileNode.Files = 1

			// Aggregate file statistics
			for _, change := range changes {
				if change.File == filePath {
					fileNode.Adds += change.Adds
					fileNode.Dels += change.Dels
					fileNode.Commits++

					for _, ctype := range change.Scope {
						fileNode.Scopes[ctype]++
					}

					for _, tech := range change.Tech {
						fileNode.Tech[tech]++
					}
				}
			}
		} else {
			// Create or get directory node
			current = current.ensureChild(segment)
		}
	}

	// Bubble up statistics from file to all parent directories
	current = gs
	fileStats := Count{}
	for _, change := range changes {
		if change.File == filePath {
			fileStats.Adds += change.Adds
			fileStats.Dels += change.Dels
			fileStats.Commits++

			if len(change.Scope) > 0 {
				if fileStats.Scopes == nil {
					fileStats.Scopes = make(map[ScopeType]int)
				}
				for _, scope := range change.Scope {
					fileStats.Scopes[scope]++
				}
			}

			for _, tech := range change.Tech {
				if fileStats.Tech == nil {
					fileStats.Tech = make(map[ScopeTechnology]int)
				}
				fileStats.Tech[tech]++
			}
		}
	}

	// Add to each parent directory and track unique files
	for i := 0; i < len(segments)-1; i++ {
		current = current.ensureChild(segments[i])
		current.Add(fileStats)
		// Track this file as unique in the parent directory
		current.uniqueFiles[filePath] = struct{}{}
		// Update Files count to reflect unique files
		current.Files = len(current.uniqueFiles)
	}
}

// BuildFromAnalyses builds the entire tree from commit analyses
func (gs *PathSummary) BuildFromAnalyses(analyses CommitAnalyses) {
	// Process each commit analysis
	for _, analysis := range analyses {
		for _, change := range analysis.Changes {
			// Get all changes for this file across all commits
			var fileChanges []CommitChange

			// Enhance change with commit-level metadata
			enhancedChange := change
			if len(enhancedChange.Scope) == 0 {
				enhancedChange.Scope = Scopes{analysis.Scope}
			}

			// Add commit type to the change tracking
			if gs.CommitTypes == nil {
				gs.CommitTypes = make(map[CommitType]int)
			}

			fileChanges = append(fileChanges, enhancedChange)

			// Add this file to the tree
			gs.AddFile(change.File, fileChanges)
		}
	}

	// Sort all children recursively
	gs.sortChildren()

	// Collapse single-child directory chains
	gs.CollapseChains()
}

// sortChildren sorts children by change volume (adds + dels) descending
func (gs *PathSummary) sortChildren() {
	sort.Slice(gs.Children, func(i, j int) bool {
		return gs.Children[i].Total() > gs.Children[j].Total()
	})

	for i := range gs.Children {
		gs.Children[i].sortChildren()
	}
}

// CollapseChains merges single-child directory chains into collapsed paths
func (gs *PathSummary) CollapseChains() {
	// First, recursively collapse all children
	for i := range gs.Children {
		gs.Children[i].CollapseChains()
	}

	// Then collapse this node's chain if applicable
	for i := range gs.Children {
		child := &gs.Children[i]

		// Only collapse if:
		// 1. Child is a directory (has children)
		// 2. Child has exactly one child
		// 3. That single child is also a directory
		for len(child.Children) == 1 && len(child.Children[0].Children) > 0 {
			grandchild := child.Children[0]

			// Merge paths
			child.Path = child.Path + "/" + grandchild.Path

			// Replace children with grandchildren
			child.Children = grandchild.Children

			// Note: Stats are already aggregated, no need to merge them
		}
	}
	if len(gs.Children) == 1 {
		gs = &gs.Children[0]
	}
}

// getFileIcon returns an appropriate icon for the file type
func getFileIcon(path string) icons.Icon {
	ext := filepath.Ext(path)
	switch ext {
	case ".go":
		return icons.Golang
	case ".js", ".jsx":
		return icons.JS
	case ".ts", ".tsx":
		return icons.TS
	case ".py":
		return icons.Python
	case ".java":
		return icons.Java
	case ".md":
		return icons.MD
	case ".yaml", ".yml":
		return icons.Config
	case ".json":
		return icons.File
	case ".sql":
		return icons.DB
	case ".sh", ".bash":
		return icons.File
	case ".dockerfile", ".Dockerfile":
		return icons.File
	case ".tf":
		return icons.Config
	default:
		return icons.File
	}
}

// isExtensionBasedTech returns true if the tech is just derived from file extension
func isExtensionBasedTech(filePath string, tech ScopeTechnology) bool {
	ext := strings.ToLower(filepath.Ext(filePath))

	// Map extensions to technologies that are "obvious" and should be filtered
	extensionTech := map[string]ScopeTechnology{
		".go":   Go,
		".py":   Python,
		".js":   NodeJS,
		".jsx":  NodeJS,
		".ts":   NodeJS,
		".tsx":  NodeJS,
		".java": Java,
		".rb":   Ruby,
		".rs":   Rust,
		".php":  PHP,
		".sql":  SQL,
		".sh":   Shell,
		".bash": Bash,
	}

	return extensionTech[ext] == tech
}
