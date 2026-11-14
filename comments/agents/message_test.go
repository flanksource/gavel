package agents

import (
	"encoding/json"
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
})
