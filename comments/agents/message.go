package agents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strconv"
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
	Quote    *comments.TextQuote   `json:"quote,omitempty"`
	Region   *comments.Region      `json:"region,omitempty"`
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

// messageOptions shape a delivery's message: the agent the threads were
// acknowledged as (live sessions), the base URL attachment links are made
// absolute against, the edits policy, the project directory a propose-mode
// agent must not modify, and a managed run's extra instructions.
type messageOptions struct {
	Agent        comments.Actor
	BaseURL      string
	Edits        string
	Dir          string
	Instructions string
}

func (o messageOptions) proposes() bool { return o.Edits == comments.EditsPropose }

// threadMessage is what a live session receives for threads (one scope's):
// what to do, the `gavel comments` commands that reply and resolve (or, in
// propose mode, propose) as the agent they were acknowledged as
// (GAVEL_COMMENT_ACTOR), and the threads.
func threadMessage(threads []comments.Comment, o messageOptions) (string, error) {
	body, err := threadsMarkdown(threads, o.BaseURL)
	if err != nil {
		return "", err
	}
	identity, err := json.Marshal(o.Agent)
	if err != nil {
		return "", err
	}
	task := "implement the scoped change it asks for and verify it, then reply with what you changed and how you verified it, and resolve the thread,"
	if o.proposes() {
		task = fmt.Sprintf("%s A person reviews the proposal and accepts it, rejects it or asks for revisions, so do not resolve the thread. Propose", proposeRule(o.Dir, "`gavel comments reply <id> --expectedVersion <v> --bodyMarkdown '...' --edits '<json>'`"))
	}
	const tail = "using the gavel comments CLI (the same commands are `gavel mcp` tools). Each write needs the thread's current version, so read it first:"
	id := "<thread id>"
	intro := fmt.Sprintf("Gavel comment %s in %s, sent to this session by gavel. They are already acknowledged for you as %s: do not acknowledge them again. For each thread, %s %s", threadIDs(threads), threads[0].Target.ScopeURI, identity, task, tail)
	if len(threads) == 1 {
		tc, err := readContext(threads[0])
		if err != nil {
			return "", err
		}
		id = threads[0].ID
		intro = fmt.Sprintf("Gavel comment thread %s on %s, sent to this session by gavel. It is already acknowledged for you as %s: do not acknowledge it again. For this thread, %s %s", id, location(threads[0], tc), identity, task, tail)
	}
	return strings.Join([]string{intro, threadCommands(id, o), body}, "\n\n"), nil
}

// threadCommands are the `gavel comments` commands a live session runs for
// thread id: reply and resolve, or in propose mode reply with edits.
func threadCommands(id string, o messageOptions) string {
	lines := []string{
		"```sh",
		fmt.Sprintf("export %s=%s", commentapi.ActorEnv, commentapi.EncodeActor(o.Agent)),
		fmt.Sprintf("gavel comments get %s --json   # the thread and its current version", id),
	}
	if o.proposes() {
		lines = append(lines,
			fmt.Sprintf("gavel comments reply %s --expectedVersion <version> --bodyMarkdown '<what the change does and why>' --edits '[{\"path\":\"<path>\",\"oldString\":\"<exact text>\",\"newString\":\"<replacement>\"}]'", id),
			"# each edit: {path, oldString, newString, replaceAll?} | {path, content} to write a whole file | {path, delete: true}",
		)
	} else {
		lines = append(lines,
			fmt.Sprintf("gavel comments reply %s --expectedVersion <version> --bodyMarkdown '<what you changed and how you verified it>'", id),
			fmt.Sprintf("gavel comments resolve %s --expectedVersion <version> --summary '<one-line summary>'", id),
		)
	}
	return strings.Join(append(lines, "```"), "\n")
}

// sessionMessage is what a managed session is given: the threads, in propose
// mode the proposal-only rule, and the run's extra instructions.
func sessionMessage(threads []comments.Comment, o messageOptions) (string, error) {
	message, err := threadsMarkdown(threads, o.BaseURL)
	if err != nil {
		return "", err
	}
	if o.proposes() {
		message += "\n\n## Proposal only\n\n" + proposeRule(o.Dir, "propose_change") + " A person accepts, rejects or asks for revisions, so never resolve the thread yourself.\n"
	}
	if instructions := strings.TrimSpace(o.Instructions); instructions != "" {
		message += "\n\n## Additional instructions\n\n" + instructions + "\n"
	}
	return message, nil
}

// readyMessage is what a session started ahead of its threads is given: the
// page it is for, its edits policy, that the threads follow as later messages,
// and the run's extra instructions.
func readyMessage(target comments.Target, o messageOptions) string {
	policy := "Edits policy: direct. For each thread, implement the scoped change it asks for, verify it, reply with what you changed and how you verified it, and resolve the thread."
	if o.proposes() {
		policy = "Edits policy: propose. " + proposeRule(o.Dir, "propose_change") + " A person accepts, rejects or asks for revisions, so never resolve a thread yourself."
	}
	message := fmt.Sprintf("Gavel started this session for comment threads on the page %s in %s. There are none yet: gavel delivers each comment thread on the page to this session later, as its own message with the tools to answer it. Do not start any work until then; reply only that you are ready.\n\n%s\n", target.URI, target.ScopeURI, policy)
	if instructions := strings.TrimSpace(o.Instructions); instructions != "" {
		message += "\n## Additional instructions\n\n" + instructions + "\n"
	}
	return message
}

// proposeRule is the propose-mode rule: gavel builds the commit from the
// agent's edits, sent with via, so the agent writes no file at all.
func proposeRule(dir, via string) string {
	return fmt.Sprintf("Do not create or modify any file — not even scratch copies. Read the files under %s to understand them, then propose the change with %s; gavel commits it onto HEAD itself. Edit paths are relative to the repository root, and each oldString must match the file exactly once (include surrounding lines to make it unique, or set replaceAll).", dir, via)
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

// threadsMarkdown renders one thread as threadMarkdown does, and several each
// under a "# Thread <id>" heading.
func threadsMarkdown(threads []comments.Comment, baseURL string) (string, error) {
	if len(threads) == 1 {
		return threadMarkdown(threads[0], baseURL)
	}
	sections := make([]string, 0, len(threads))
	for _, thread := range threads {
		markdown, err := threadMarkdown(thread, baseURL)
		if err != nil {
			return "", err
		}
		sections = append(sections, "# Thread "+thread.ID+"\n\n"+markdown)
	}
	return strings.Join(sections, "\n\n"), nil
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
	if quote := tc.selector.Quote; quote != nil {
		sections = append(sections, quotedText(quote.Exact))
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
		proposal, err := proposalMarkdown(message)
		if err != nil {
			return "", fmt.Errorf("thread %s reply %s: %w", comment.ID, message.ID, err)
		}
		if proposal != "" {
			sections = append(sections, proposal)
		}
	}
	return strings.Join(sections, "\n\n"), nil
}

// proposalMarkdown is the patch a reply proposed: its state, the files it
// changes and the reviewer's feedback, so a revising agent sees what was
// decided; empty for a reply with no proposal.
func proposalMarkdown(m comments.Message) (string, error) {
	var content map[string]json.RawMessage
	if err := decodeOptional(m.Content, &content); err != nil {
		return "", fmt.Errorf("content: %w", err)
	}
	raw, ok := content[comments.ProposalContentKey]
	if !ok {
		return "", nil
	}
	var p comments.Proposal
	if err := json.Unmarshal(raw, &p); err != nil {
		return "", fmt.Errorf("content.%s: %w", comments.ProposalContentKey, err)
	}
	files := make([]string, 0, len(p.Files))
	for _, f := range p.Files {
		files = append(files, fmt.Sprintf("`%s` (+%d -%d)", f.Path, f.Additions, f.Deletions))
	}
	out := fmt.Sprintf("Proposed patch (%s): %s", p.State, strings.Join(files, ", "))
	if p.Feedback != "" {
		out += "\n\nReviewer feedback: " + p.Feedback
	}
	return out, nil
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
	case selector.Region != nil:
		r := selector.Region
		return fmt.Sprintf(" (region x %s, y %s, %s wide, %s high%s)", percent(r.X), percent(r.Y), percent(r.W), percent(r.H), view)
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

// percent renders a [0,1] fraction as a percentage to two decimals.
func percent(fraction float64) string {
	return strconv.FormatFloat(math.Round(fraction*10000)/100, 'f', -1, 64) + "%"
}

// quotedText is the exact text a comment selected: inline when it is one
// plain line, fenced otherwise.
func quotedText(exact string) string {
	if strings.ContainsAny(exact, "\n`") {
		return "Quoted text:\n\n```\n" + exact + "\n```"
	}
	return "Quoted text: `" + exact + "`"
}

func messageMarkdown(author comments.Actor, body string) string {
	name := author.Label
	if name == "" {
		name = author.ID
	}
	return fmt.Sprintf("**%s** (%s):\n\n%s", name, author.Kind, body)
}
