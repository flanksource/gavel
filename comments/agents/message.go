package agents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/flanksource/gavel/comments"
	commentapi "github.com/flanksource/gavel/comments/clicky"
)

// elementHTMLLimit bounds the element HTML a message carries.
const elementHTMLLimit = 4096

// threadSelector is the part of any host's selector a message describes.
type threadSelector struct {
	Kind     string                `json:"kind"`
	Anchor   string                `json:"anchor,omitempty"`
	View     string                `json:"view,omitempty"`
	Patch    *comments.PatchAnchor `json:"patch,omitempty"`
	HTML     string                `json:"html,omitempty"`
	Locators []json.RawMessage     `json:"locators,omitempty"`
}

// elementContext is the element a host captured with a comment
// (content.playground.element in the playground).
type elementContext struct {
	PageURL       string            `json:"pageUrl,omitempty"`
	ComponentName string            `json:"componentName,omitempty"`
	Source        string            `json:"source,omitempty"`
	HTML          string            `json:"html,omitempty"`
	Locators      []json.RawMessage `json:"locators,omitempty"`
}

// threadContext is what a message reads from a thread's selector and content.
// Every field is optional; a known field of the wrong shape is an error.
type threadContext struct {
	selector threadSelector
	page     struct {
		URL   string `json:"url,omitempty"`
		Title string `json:"title,omitempty"`
	}
	element elementContext
}

func readContext(c comments.Comment) (threadContext, error) {
	var tc threadContext
	var content struct {
		Page       json.RawMessage `json:"page"`
		Playground struct {
			Element elementContext `json:"element"`
		} `json:"playground"`
	}
	if err := decodeOptional(c.Target.Selector, &tc.selector); err != nil {
		return tc, fmt.Errorf("thread %s selector: %w", c.ID, err)
	}
	if err := decodeOptional(c.Content, &content); err != nil {
		return tc, fmt.Errorf("thread %s content: %w", c.ID, err)
	}
	if err := decodeOptional(content.Page, &tc.page); err != nil {
		return tc, fmt.Errorf("thread %s content.page: %w", c.ID, err)
	}
	tc.element = content.Playground.Element
	return tc, nil
}

func decodeOptional(raw json.RawMessage, out any) error {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	return json.Unmarshal(raw, out)
}

// threadMessage is what a live session receives for a thread: what to do, the
// `gavel comments` commands that reply and resolve as the agent it was
// acknowledged as (GAVEL_COMMENT_ACTOR), and the thread itself.
func threadMessage(comment comments.Comment, agent comments.Actor, baseURL string) (string, error) {
	tc, err := readContext(comment)
	if err != nil {
		return "", err
	}
	thread, err := threadMarkdown(comment, baseURL)
	if err != nil {
		return "", err
	}
	identity, err := json.Marshal(agent)
	if err != nil {
		return "", err
	}
	id := comment.ID
	commands := strings.Join([]string{
		"```sh",
		fmt.Sprintf("export %s=%s", commentapi.ActorEnv, commentapi.EncodeActor(agent)),
		fmt.Sprintf("gavel comments get %s --json   # the thread and its current version", id),
		fmt.Sprintf("gavel comments reply %s --expectedVersion <version> --bodyMarkdown '<what you changed and how you verified it>'", id),
		fmt.Sprintf("gavel comments resolve %s --expectedVersion <version> --summary '<one-line summary>'", id),
		"```",
	}, "\n")
	return strings.Join([]string{
		fmt.Sprintf("Gavel comment thread %s on %s, sent to this session by gavel. It is already acknowledged for you as %s: do not acknowledge it again. Implement the scoped change it asks for and verify it, then reply with what you changed and how you verified it, and resolve the thread, using the gavel comments CLI (the same commands are `gavel mcp` tools). Each write needs the thread's current version, so read it first:", id, location(comment, tc), identity),
		commands,
		thread,
	}, "\n\n"), nil
}

// sessionMessage is the first message of a managed session: the thread plus the run's extra instructions.
func sessionMessage(comment comments.Comment, instructions, baseURL string) (string, error) {
	message, err := threadMarkdown(comment, baseURL)
	if err != nil {
		return "", err
	}
	if instructions = strings.TrimSpace(instructions); instructions != "" {
		message += "\n\n## Additional instructions\n\n" + instructions + "\n"
	}
	return message, nil
}

// location is where the thread is: the page it was made on, a file under its
// scope, or the target's label and URI.
func location(comment comments.Comment, tc threadContext) string {
	t := comment.Target
	switch file, isFile := strings.CutPrefix(t.URI, t.ScopeURI+"/files/"); {
	case tc.page.URL != "":
		return tc.page.URL
	case tc.element.PageURL != "":
		return tc.element.PageURL
	case isFile:
		return file
	case t.Label != "":
		return t.Label + " (" + t.URI + ")"
	default:
		return t.URI
	}
}

// threadMarkdown renders the thread: where it is anchored, the element it is
// on, its attachments, the root, then each reply.
func threadMarkdown(comment comments.Comment, baseURL string) (string, error) {
	tc, err := readContext(comment)
	if err != nil {
		return "", err
	}
	sections := []string{"## " + location(comment, tc) + anchorLabel(tc.selector)}
	if patch := tc.selector.Patch; patch != nil && len(patch.Lines) > 0 {
		sections = append(sections, "```\n"+strings.Join(patch.Lines, "\n")+"\n```")
	}
	element, err := elementSections(tc)
	if err != nil {
		return "", fmt.Errorf("thread %s element: %w", comment.ID, err)
	}
	sections = append(sections, element...)
	if len(comment.Attachments) > 0 {
		sections = append(sections, attachmentLinks(comment, baseURL))
	}
	sections = append(sections, messageMarkdown(comment.Author, comment.Body))
	for _, message := range comment.Messages {
		sections = append(sections, messageMarkdown(message.Author, message.Body))
	}
	return strings.Join(sections, "\n\n"), nil
}

// elementSections describe the element the thread is on: the page title, its
// component and source, the locators that find it again, and its HTML.
func elementSections(tc threadContext) ([]string, error) {
	var sections []string
	if tc.page.Title != "" {
		sections = append(sections, "Page title: "+tc.page.Title)
	}
	if e := tc.element; e.ComponentName != "" || e.Source != "" {
		sections = append(sections, elementLine(e))
	}
	locators := tc.selector.Locators
	if len(locators) == 0 {
		locators = tc.element.Locators
	}
	if len(locators) > 0 {
		lines := make([]string, 0, len(locators))
		for _, locator := range locators {
			var compact bytes.Buffer
			if err := json.Compact(&compact, locator); err != nil {
				return nil, err
			}
			lines = append(lines, compact.String())
		}
		sections = append(sections, "Locators:\n\n```json\n"+strings.Join(lines, "\n")+"\n```")
	}
	html := tc.selector.HTML
	if html == "" {
		html = tc.element.HTML
	}
	if html != "" {
		sections = append(sections, "Element HTML:\n\n```html\n"+trimHTML(html)+"\n```")
	}
	return sections, nil
}

func elementLine(e elementContext) string {
	line := "Element"
	if e.ComponentName != "" {
		line += ": `" + e.ComponentName + "`"
	}
	switch {
	case strings.Contains(e.Source, "\n"):
		line += " at:\n\n```\n" + e.Source + "\n```"
	case e.Source != "":
		line += " at `" + e.Source + "`"
	}
	return line
}

func trimHTML(html string) string {
	if len(html) <= elementHTMLLimit {
		return html
	}
	cut := elementHTMLLimit
	for cut > 0 && !utf8.RuneStart(html[cut]) {
		cut--
	}
	return html[:cut] + fmt.Sprintf("\n<!-- truncated to %d bytes -->", elementHTMLLimit)
}

// attachmentLinks lists the thread's attachments, linked through the comments
// API: absolute against baseURL when the host gives one, else as paths.
func attachmentLinks(comment comments.Comment, baseURL string) string {
	lines := []string{"Attachments:", ""}
	for _, a := range comment.Attachments {
		link := strings.TrimSuffix(baseURL, "/") + "/api/v1/comments/" + url.PathEscape(comment.ID) + "/attachments/" + url.PathEscape(a.ID)
		lines = append(lines, fmt.Sprintf("- [%s](%s) (%s)", a.Name, link, a.MediaType))
	}
	return strings.Join(lines, "\n")
}

func anchorLabel(selector threadSelector) string {
	view := ""
	if selector.View != "" {
		view = ", " + selector.View + " view"
	}
	switch {
	case selector.Patch != nil && selector.Patch.StartLine == selector.Patch.EndLine:
		return fmt.Sprintf(" (line %d%s)", selector.Patch.StartLine, view)
	case selector.Patch != nil:
		return fmt.Sprintf(" (lines %d-%d%s)", selector.Patch.StartLine, selector.Patch.EndLine, view)
	case selector.Kind == "dom":
		return fmt.Sprintf(" (element %s%s)", selector.Anchor, view)
	default:
		return " (whole target)"
	}
}

func messageMarkdown(author comments.Actor, body string) string {
	name := author.Label
	if name == "" {
		name = author.ID
	}
	return fmt.Sprintf("**%s** (%s):\n\n%s", name, author.Kind, body)
}
