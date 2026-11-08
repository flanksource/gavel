package editor

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/comments"
	commentapi "github.com/flanksource/gavel/comments/clicky"
)

type threadMessageOptions struct {
	Comment comments.Comment
	// Path is the file the thread is on, as the editor shows it.
	Path  string
	Agent comments.Actor
}

// threadMessage is what a live session receives for a thread: what to do, the
// `gavel comments` commands that reply and resolve as the agent it was
// acknowledged as (GAVEL_COMMENT_ACTOR), and the thread itself.
func threadMessage(options threadMessageOptions) string {
	agent, _ := json.Marshal(options.Agent)
	id := options.Comment.ID
	commands := strings.Join([]string{
		"```sh",
		fmt.Sprintf("export %s=%s", commentapi.ActorEnv, commentapi.EncodeActor(options.Agent)),
		fmt.Sprintf("gavel comments get %s --json   # the thread and its current version", id),
		fmt.Sprintf("gavel comments reply %s --expectedVersion <version> --bodyMarkdown '<what you changed and how you verified it>'", id),
		fmt.Sprintf("gavel comments resolve %s --expectedVersion <version> --summary '<one-line summary>'", id),
		"```",
	}, "\n")
	return strings.Join([]string{
		fmt.Sprintf("Gavel comment thread %s on %s, sent to this session from `gavel edit`. It is already acknowledged for you as %s: do not acknowledge it again. Implement the scoped change it asks for and verify it, then reply with what you changed and how you verified it, and resolve the thread, using the gavel comments CLI (the same commands are `gavel mcp` tools). Each write needs the thread's current version, so read it first:", id, options.Path, agent),
		commands,
		threadMarkdown(options.Comment, options.Path),
	}, "\n\n")
}

// threadMarkdown renders the thread: where it is anchored, the root, then each reply.
func threadMarkdown(comment comments.Comment, path string) string {
	var selector CommentSelector
	_ = json.Unmarshal(comment.Target.Selector, &selector)
	lines := []string{"## " + path + anchorLabel(selector)}
	if selector.Patch != nil && len(selector.Patch.Lines) > 0 {
		lines = append(lines, "```", strings.Join(selector.Patch.Lines, "\n"), "```")
	}
	lines = append(lines, messageMarkdown(comment.Author, comment.Body))
	for _, message := range comment.Messages {
		lines = append(lines, messageMarkdown(message.Author, message.Body))
	}
	return strings.Join(lines, "\n\n")
}

func anchorLabel(selector CommentSelector) string {
	switch {
	case selector.Patch != nil && selector.Patch.StartLine == selector.Patch.EndLine:
		return fmt.Sprintf(" (line %d, %s view)", selector.Patch.StartLine, selector.View)
	case selector.Patch != nil:
		return fmt.Sprintf(" (lines %d-%d, %s view)", selector.Patch.StartLine, selector.Patch.EndLine, selector.View)
	case selector.Kind == "dom":
		return fmt.Sprintf(" (element %s, %s view)", selector.Anchor, selector.View)
	default:
		return " (whole file)"
	}
}

func messageMarkdown(author comments.Actor, body string) string {
	name := author.Label
	if name == "" {
		name = author.ID
	}
	return fmt.Sprintf("**%s** (%s):\n\n%s", name, author.Kind, body)
}
