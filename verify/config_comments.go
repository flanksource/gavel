package verify

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/flanksource/captain/pkg/api"
)

// DefaultCommentAgentPrompt is the captain prompt a comment agent runs when its
// entry names none: the editor's managed comment-working prompt.
const DefaultCommentAgentPrompt = "editor-comments"

// CommentsConfig configures comment threads on the files under this
// .gavel.yaml's directory.
type CommentsConfig struct {
	Agents CommentAgentsConfig `yaml:"agents,omitempty" json:"agents,omitempty"`
}

// CommentAgentsConfig is which agent works the comment threads on pages under
// this .gavel.yaml's directory: Default for every page beneath it, Pages for
// single files, keyed by their path relative to the directory. The nearest
// .gavel.yaml in the config chain wins, and a page entry wins over its
// directory's Default.
type CommentAgentsConfig struct {
	Default *CommentAgent           `yaml:"default,omitempty" json:"default,omitempty"`
	Pages   map[string]CommentAgent `yaml:"pages,omitempty" json:"pages,omitempty"`
}

// CommentAgent is the captain run a new human comment thread is handed to.
// Enabled false opts its pages out of an agent inherited from further up.
type CommentAgent struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
	// Assign makes the agent the thread's assignee as soon as it is created.
	Assign bool `yaml:"assign,omitempty" json:"assign,omitempty"`
	// Prompt is the captain prompt the run starts; DefaultCommentAgentPrompt when empty.
	Prompt string `yaml:"prompt,omitempty" json:"prompt,omitempty"`
	// Presets replace the prompt frontmatter's runtime presets when set.
	Presets []string `yaml:"presets,omitempty" json:"presets,omitempty"`
	// Spec is layered over the prompt frontmatter's runtime spec.
	Spec *api.Spec `yaml:"spec,omitempty" json:"spec,omitempty"`
	// Variables feed the prompt template below the host's own.
	Variables map[string]any `yaml:"variables,omitempty" json:"variables,omitempty"`
	// Instructions are appended to the thread the session is given.
	Instructions string `yaml:"instructions,omitempty" json:"instructions,omitempty"`
}

// PromptOrDefault is the prompt the agent's run starts.
func (a CommentAgent) PromptOrDefault() string {
	if a.Prompt == "" {
		return DefaultCommentAgentPrompt
	}
	return a.Prompt
}

// Validate rejects a page key that does not name a file beneath the
// .gavel.yaml's directory, so an entry can never route threads elsewhere.
func (c CommentAgentsConfig) Validate() error {
	for key := range c.Pages {
		if !IsCommentPageKey(key) {
			return fmt.Errorf("comments.agents.pages: %q must be a relative file path beneath the .gavel.yaml directory", key)
		}
	}
	return nil
}

// IsCommentPageKey is whether key is a clean, relative, slash-separated path
// beneath its directory.
func IsCommentPageKey(key string) bool {
	return key != "" && key != "." && path.Clean(key) == key && !path.IsAbs(key) && key != ".." && !strings.HasPrefix(key, "../") && !strings.Contains(key, "\\")
}

var commentAgentMu sync.Mutex

// SaveCommentAgent makes agent the comment agent of directory's own .gavel.yaml
// (its default when page is empty, else its entry for page), or removes that
// entry when agent is nil, creating the file when absent.
func SaveCommentAgent(directory, page string, agent *CommentAgent) error {
	if page != "" && !IsCommentPageKey(page) {
		return fmt.Errorf("comment agent page %q must be a relative file path beneath %s", page, directory)
	}
	commentAgentMu.Lock()
	defer commentAgentMu.Unlock()
	cfg, err := LoadSingleGavelConfig(filepath.Join(directory, GavelConfigFileName))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	configured := &cfg.Comments.Agents
	switch {
	case page == "":
		configured.Default = agent
	case agent == nil:
		delete(configured.Pages, page)
	default:
		if configured.Pages == nil {
			configured.Pages = map[string]CommentAgent{}
		}
		configured.Pages[page] = *agent
	}
	return SaveGavelConfig(directory, cfg)
}

func commentsSchema(specSchema map[string]any) map[string]any {
	agent := object("A captain run that new human comment threads are handed to.", map[string]any{
		"enabled":      boolProp("Hand new human threads on these pages to the agent. false opts the pages out of an agent inherited from a .gavel.yaml further up."),
		"assign":       boolProp("Make the agent the thread's assignee as soon as the thread is created."),
		"prompt":       stringWithDefault("Captain prompt the run starts.", DefaultCommentAgentPrompt),
		"presets":      stringArray("Runtime presets replacing the prompt frontmatter's presets."),
		"spec":         specNodeSchema(specSchema, "Runtime spec layered over the prompt frontmatter's.", "", "Model the run uses; the prompt's when unset."),
		"variables":    mapObject("Prompt template variables, below the ones the host sets.", map[string]any{}),
		"instructions": stringProp("Instructions appended to the thread the session is given."),
	})
	return object("Comment thread settings for the files beneath this .gavel.yaml's directory.", map[string]any{
		"agents": object("Which agent works new comment threads. The nearest .gavel.yaml wins, and a page entry wins over its directory's default.", map[string]any{
			"default": agent,
			"pages":   mapObject("Per-file agents, keyed by the file's path relative to this .gavel.yaml's directory.", agent),
		}),
	})
}
