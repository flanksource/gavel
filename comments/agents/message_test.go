package agents

import (
	"encoding/json"
	"maps"
	"slices"
	"strings"

	"github.com/flanksource/gavel/comments"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("threadMarkdown", func() {
	const (
		scope   = "gavel://project/acme"
		baseURL = "https://comments.example.test"
	)
	author := comments.Actor{Kind: "user", ID: "user:reviewer", Label: "Reviewer"}
	fence := "```"

	It("carries a playground comment's page, component, source, locators, html and screenshot link", func() {
		c := comments.Comment{
			ID: "c-play", Body: "Make the button primary", Author: author,
			Target: comments.Target{URI: scope + "/pages/home", ScopeURI: scope, Kind: "playground-page", Label: "home"},
			Content: json.RawMessage(`{"playground":{"page":"home","element":{"pageUrl":"http://localhost:5801/home","componentName":"SaveButton","source":"src/pages/Home.tsx:42",` +
				`"html":"<button class=\"btn\">Save</button>","locators":[{"kind":"css","selector":"main > button"},{"kind":"role","tag":"button","role":"button","name":"Save"}]}}}`),
			Attachments: []comments.Attachment{{ID: "a-1", Name: "screenshot.png", MediaType: "image/png"}},
		}
		Expect(threadMarkdown(c, baseURL)).To(Equal(strings.Join([]string{
			"## http://localhost:5801/home (whole target)",
			"Element: `SaveButton` at `src/pages/Home.tsx:42`",
			"Locators:\n\n" + fence + "json\n" + `{"kind":"css","selector":"main > button"}` + "\n" + `{"kind":"role","tag":"button","role":"button","name":"Save"}` + "\n" + fence,
			"Element HTML:\n\n" + fence + "html\n" + `<button class="btn">Save</button>` + "\n" + fence,
			"Attachments:\n\n- [screenshot.png](https://comments.example.test/api/v1/comments/c-play/attachments/a-1) (image/png)",
			"**Reviewer** (user):\n\nMake the button primary",
		}, "\n\n")))
	})

	It("carries an extension web-page comment's page title, anchor, locators and html trimmed to the html limit", func() {
		html := strings.Repeat("<li>item</li>", 400)
		selector := `{"kind":"dom","anchor":"#plans > li","locators":[{"kind":"css","selector":"#plans > li"}],"html":"` + html + `"}`
		c := comments.Comment{
			ID: "c-web", Body: "Align the plans", Author: author,
			Target:  comments.Target{URI: scope + "/pages/pricing", ScopeURI: scope, Kind: "web-page", Selector: json.RawMessage(selector)},
			Content: json.RawMessage(`{"page":{"url":"https://example.com/pricing","title":"Pricing - Example"}}`),
		}
		Expect(threadMarkdown(c, "")).To(Equal(strings.Join([]string{
			"## https://example.com/pricing (element #plans > li)",
			"Page title: Pricing - Example",
			"Locators:\n\n" + fence + "json\n" + `{"kind":"css","selector":"#plans > li"}` + "\n" + fence,
			"Element HTML:\n\n" + fence + "html\n" + html[:elementHTMLLimit] + "\n<!-- truncated to 4096 bytes -->\n" + fence,
			"**Reviewer** (user):\n\nAlign the plans",
		}, "\n\n")))
	})

	It("renders a file comment with its lines, and attachment links as paths without a base URL", func() {
		selector, err := json.Marshal(map[string]any{"kind": "patch", "patch": comments.PatchAnchor{Version: 1, StartLine: 3, EndLine: 4, Lines: []string{"a := 1", "b := 2"}}})
		Expect(err).NotTo(HaveOccurred())
		c := comments.Comment{
			ID: "c-file", Body: "Name these", Author: author,
			Target:      comments.Target{URI: scope + "/files/src/app.go", ScopeURI: scope, Kind: "file", Selector: selector},
			Attachments: []comments.Attachment{{ID: "a-2", Name: "trace.txt", MediaType: "text/plain"}},
			Messages:    []comments.Message{{Body: "On it", Author: comments.Actor{Kind: "agent", ID: "claude:live-1"}}},
		}
		Expect(threadMarkdown(c, "")).To(Equal(strings.Join([]string{
			"## src/app.go (lines 3-4)",
			fence + "\na := 1\nb := 2\n" + fence,
			"Attachments:\n\n- [trace.txt](/api/v1/comments/c-file/attachments/a-2) (text/plain)",
			"**Reviewer** (user):\n\nName these",
			"**claude:live-1** (agent):\n\nOn it",
		}, "\n\n")))
	})

	It("renders a reply's earlier proposal with its state, files and reviewer feedback", func() {
		agent := comments.Actor{Kind: "agent", ID: "gavel:1", Label: "Gavel agent"}
		proposal, err := json.Marshal(map[string]any{comments.ProposalContentKey: comments.Proposal{
			State: comments.ProposalRevisionRequested, Feedback: "Keep the old name as an alias",
			Files: []comments.ProposalFile{{Path: "src/app.go", Additions: 3, Deletions: 1}},
		}})
		Expect(err).NotTo(HaveOccurred())
		c := comments.Comment{
			ID: "c-prop", Body: "Rename load", Author: author,
			Target:   comments.Target{URI: scope + "/files/src/app.go", ScopeURI: scope},
			Messages: []comments.Message{{ID: "m-1", Body: "Renamed load to read", Author: agent, Content: proposal}},
		}
		Expect(threadMarkdown(c, "")).To(Equal(strings.Join([]string{
			"## src/app.go (whole target)",
			"**Reviewer** (user):\n\nRename load",
			"**Gavel agent** (agent):\n\nRenamed load to read",
			"Proposed patch (revision_requested): `src/app.go` (+3 -1)\n\nReviewer feedback: Keep the old name as an alias",
		}, "\n\n")))
	})
})

var _ = Describe("edits policy", func() {
	const (
		scope = "gavel://project/acme"
		dir   = "/work/acme"
	)
	agent := comments.Actor{Kind: "agent", ID: "claude:live-1", Label: "claude"}
	thread := comments.Comment{
		ID: "c-1", Version: 3, Body: "Rename load", Author: comments.Actor{Kind: "user", ID: "user:reviewer"},
		Target: comments.Target{URI: scope + "/files/src/app.go", ScopeURI: scope},
	}

	DescribeTable("a live session's commands",
		func(edits string, present, absent []string) {
			message, err := threadMessage([]comments.Comment{thread}, messageOptions{Agent: agent, Edits: edits, Dir: dir})
			Expect(err).NotTo(HaveOccurred())
			for _, text := range present {
				Expect(message).To(ContainSubstring(text))
			}
			for _, text := range absent {
				Expect(message).NotTo(ContainSubstring(text))
			}
		},
		Entry("direct mode replies and resolves", comments.EditsDirect,
			[]string{"gavel comments reply c-1 --expectedVersion <version> --bodyMarkdown", "gavel comments resolve c-1"},
			[]string{"--edits", "Do not create or modify"}),
		Entry("propose mode only proposes edits, writing no file and never resolving", comments.EditsPropose,
			[]string{"Do not create or modify any file — not even scratch copies", "Read the files under /work/acme", `gavel comments reply c-1 --expectedVersion <version> --bodyMarkdown '<what the change does and why>' --edits '[{"path":"<path>","oldString":"<exact text>","newString":"<replacement>"}]'`, "must match the file exactly once"},
			[]string{"gavel comments resolve", ".tmp", "--patch", "diff -u"}),
	)

	DescribeTable("a managed session's thread tools",
		func(edits string, names []string) {
			tools := threadTools(nil, []comments.Comment{thread}, agent, edits)
			got := make([]string, 0, len(tools))
			for _, tool := range tools {
				got = append(got, tool.Name)
			}
			Expect(got).To(Equal(names))
		},
		Entry("direct mode resolves", comments.EditsDirect, []string{"list_threads", "reply_to_thread", "resolve_thread"}),
		Entry("propose mode proposes instead of resolving", comments.EditsPropose, []string{"list_threads", "reply_to_thread", "propose_change"}),
	)

	It("tells a propose-mode managed session to answer with propose_change, not by editing files", func() {
		message, err := sessionMessage([]comments.Comment{thread}, messageOptions{Edits: comments.EditsPropose, Dir: dir})
		Expect(err).NotTo(HaveOccurred())
		Expect(message).To(ContainSubstring("## Proposal only\n\nDo not create or modify any file — not even scratch copies. Read the files under /work/acme to understand them, then propose the change with propose_change;"))
		Expect(message).NotTo(ContainSubstring(".tmp"))
	})

	It("gives propose_change an edits schema shaped like FileEdit and decodes its edits", func(ctx SpecContext) {
		tools := threadTools(nil, []comments.Comment{thread}, agent, comments.EditsPropose)
		schema := tools[2].InputSchema
		edits := schema["properties"].(map[string]any)["edits"].(map[string]any)
		item := edits["items"].(map[string]any)
		Expect([]any{schema["required"], slices.Sorted(maps.Keys(item["properties"].(map[string]any))), item["required"], item["additionalProperties"]}).
			To(Equal([]any{[]string{"body", "id"}, []string{"content", "delete", "newString", "oldString", "path", "replaceAll"}, []string{"path"}, false}))

		decoded, err := fileEdits([]any{map[string]any{"path": "a.go", "oldString": "x", "newString": "y"}, map[string]any{"path": "b.go", "delete": true}})
		Expect(err).NotTo(HaveOccurred())
		Expect(decoded).To(Equal([]comments.FileEdit{{Path: "a.go", OldString: "x", NewString: "y"}, {Path: "b.go", Delete: true}}))
		_, err = fileEdits([]any{map[string]any{"path": "a.go", "old": "x"}})
		Expect(err).To(MatchError(ContainSubstring(`unknown field "old"`)))
		_, err = tools[2].Handler(ctx, map[string]any{"id": "c-1", "body": "Fix"})
		Expect(err).To(MatchError(ContainSubstring("needs edits or a patch, exactly one")))
	})
})
