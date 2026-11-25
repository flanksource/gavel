package comments_test

import (
	"github.com/flanksource/gavel/comments"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mentions in a comment body", func() {
	agents := []comments.Mentionable{
		{ID: comments.MentionAgent, Name: "agent", Kind: "agent"},
		{ID: "runtime:claude", Name: "claude", Kind: "agent"},
		{ID: "runtime:codex", Name: "codex", Kind: "agent"},
		{ID: "session:live-1", Name: "rename-pass", Kind: "agent"},
	}
	DescribeTable("matches whole @name tokens case-insensitively, each mentionable once",
		func(body string, ids []string) {
			found := []string{}
			for _, m := range comments.MentionedIn(body, agents) {
				found = append(found, m.ID)
			}
			Expect(found).To(Equal(ids))
		},
		Entry("a leading token", "@claude tighten the copy", []string{"runtime:claude"}),
		Entry("a capitalised token", "Over to @Claude", []string{"runtime:claude"}),
		Entry("a token ending a sentence", "Hand this to @rename-pass.", []string{"session:live-1"}),
		Entry("the same token twice", "@claude, then @CLAUDE again", []string{"runtime:claude"}),
		Entry("two agents, in mentionable order", "@codex and @agent", []string{comments.MentionAgent, "runtime:codex"}),
		Entry("a longer word is not the token", "ask @claudes or @agents", []string{}),
		Entry("an email address is not a mention", "mail bob@claude.example", []string{}),
		Entry("an unknown token is prose", "ping @nobody about it", []string{}),
		Entry("a token inside inline code is not a mention", "run `@claude --help` first", []string{}),
		Entry("a token inside a double-backtick span is not a mention", "see ``a ` @codex`` here", []string{}),
		Entry("a token outside an inline code span still is", "@claude, not `@codex`", []string{"runtime:claude"}),
		Entry("a token inside a backtick fence is not a mention", "Steps:\n```sh\n@codex run\n```\ndone", []string{}),
		Entry("a token inside a tilde fence is not a mention", "~~~\n@agent\n~~~", []string{}),
		Entry("a token after a closed fence still is", "```\n@codex\n```\n@agent go", []string{comments.MentionAgent}),
		Entry("a token in a blockquote is not a mention", "> @claude said this\nI agree", []string{}),
		Entry("a token in an indented blockquote is not a mention", "   > quoting @codex", []string{}),
		Entry("a token after a blockquote still is", "> earlier note\n@rename-pass take it", []string{"session:live-1"}),
	)
})
