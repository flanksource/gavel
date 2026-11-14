package clicky

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/flanksource/gavel/comments"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// plain is the rendered text with clicky's non-breaking spaces as spaces.
func plain(item listItem) string {
	return strings.ReplaceAll(item.Pretty().String(), " ", " ")
}

var _ = Describe("comment list rendering", func() {
	const scope = "gavel://project/acme"
	author := comments.Actor{Kind: "user", ID: "local:501", Label: "Ada"}
	agent := comments.Actor{Kind: "agent", ID: "gavel:1", Label: "Gavel agent"}
	updated := time.Now().Add(-90 * time.Minute)

	It("shows a file thread's location, numbered snippet, full body and only set key: values", func() {
		selector, err := json.Marshal(map[string]any{"kind": "source", "version": 1, "view": "source", "patch": comments.PatchAnchor{Version: 1, StartLine: 11, EndLine: 12, Before: []string{"func load() {"}, Lines: []string{"\tdata, _ := read()", "\treturn data"}, After: []string{"}"}}})
		Expect(err).NotTo(HaveOccurred())
		item := listItem{Unread: true, Comment: comments.Comment{
			ID: "0123456789abcdef", TenantID: "tenant-x", Status: "open", Phase: comments.Active, Author: author, Assignee: &agent, UpdatedAt: updated,
			Body:     "Handle the read error instead of discarding it, the caller cannot tell an empty file from a failure.",
			Target:   comments.Target{URI: scope + "/files/pkg/load.go", ScopeURI: scope, Kind: "code", Selector: selector},
			Content:  json.RawMessage(`{"agentRun":{"state":"sent","target":{"label":"review session"}}}`),
			Messages: []comments.Message{{Body: "Fixed in the next commit", Author: agent, CreatedAt: updated}},
		}}
		Expect(plain(item)).To(Equal("open pkg/load.go:11-12 ● 01234567\n" +
			"  10 | func load() {\n" +
			"> 11 | \tdata, _ := read()\n" +
			"> 12 | \treturn data\n" +
			"  13 | }\n" +
			"Handle the read error instead of discarding it, the caller cannot tell an empty file from a failure.\n" +
			"author: Ada · assignee: Gavel agent · agent: sent review session · updated: 1.5h ago\n" +
			"↳ Gavel agent 1.5h ago\n" +
			"  Fixed in the next commit"))
		Expect(item.Row()).To(HaveKey("comment"), "the whole thread is one full-width column")
	})

	It("shows a page thread at its page URL with the element text it anchors to", func() {
		item := listItem{Comment: comments.Comment{
			ID: "fedcba98", Status: "resolved", Phase: comments.AwaitingOutcome, Author: author, UpdatedAt: updated, Body: "Why is this missing?",
			Target:  comments.Target{URI: scope + "/pages/http%3A%2F%2Flocalhost%2Fp", ScopeURI: scope, Kind: "web-page", Selector: json.RawMessage(`{"anchor":"p","locators":[{"kind":"css","selector":"p"},{"kind":"text","text":"Project was not found"}]}`)},
			Content: json.RawMessage(`{"page":{"url":"http://localhost/p"}}`),
		}}
		Expect(plain(item)).To(Equal("resolved http://localhost/p fedcba98\n> Project was not found\nWhy is this missing?\nauthor: Ada · updated: 1.5h ago"))
	})

	It("shows a reply's proposal state and file count under its body", func() {
		proposal, err := json.Marshal(map[string]any{comments.ProposalContentKey: comments.Proposal{State: comments.ProposalRevisionRequested, Files: []comments.ProposalFile{{Path: "a.go"}, {Path: "b.go"}}}})
		Expect(err).NotTo(HaveOccurred())
		item := listItem{Comment: comments.Comment{
			ID: "fedcba98", Status: "open", Phase: comments.Active, Author: author, UpdatedAt: updated, Body: "Rename it",
			Target:   comments.Target{URI: scope + "/other", ScopeURI: scope},
			Messages: []comments.Message{{Body: "Renamed", Author: agent, CreatedAt: updated, Content: proposal}},
		}}
		Expect(plain(item)).To(HaveSuffix("↳ Gavel agent 1.5h ago\n  Renamed\n  proposal: revision_requested · 2 files"))
	})
})
