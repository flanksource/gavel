package main

import (
	"context"

	"github.com/flanksource/gavel/comments"
	commentapi "github.com/flanksource/gavel/comments/clicky"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("comments CLI registration", func() {
	It("exposes structured creation, native publication, agent delivery, routes, lint and dispatch commands", func() {
		for _, path := range [][]string{{"comments", "create"}, {"comments", "publish"}, {"comments", "dispatch"}, {"comments", "deliver"}, {"comments", "routes", "set"}, {"comments", "routes", "list"}, {"comments", "routes", "delete"}, {"comment-lint", "preview"}, {"comment-lint", "run"}} {
			command, remaining, err := rootCmd.Find(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(remaining).To(BeEmpty())
			Expect(command.Name()).To(Equal(path[len(path)-1]))
		}
		create, _, err := rootCmd.Find([]string{"comments", "create"})
		Expect(err).NotTo(HaveOccurred())
		Expect(create.Flags().Lookup("target")).NotTo(BeNil())
		publish, _, err := rootCmd.Find([]string{"comments", "publish"})
		Expect(err).NotTo(HaveOccurred())
		Expect(publish.Flags().Lookup("expectedVersion").Value.Type()).To(Equal("int64"))
	})

	Describe("deliver --fresh", func() {
		targetOf := func(args ...string) (comments.DeliverTarget, error) {
			var flags deliverTargetFlags
			cmd := &cobra.Command{Use: "deliver"}
			flags.register(cmd)
			Expect(cmd.ParseFlags(args)).To(Succeed())
			return flags.target()
		}
		It("starts a new session for a prompt run even with a group", func() {
			target, err := targetOf("--prompt", "editor-comments", "--group", "page:home", "--fresh")
			Expect(err).NotTo(HaveOccurred())
			Expect(target).To(Equal(comments.DeliverTarget{Kind: "run", Prompt: "editor-comments", GroupKey: "page:home", Fresh: true}))
		})
		It("is rejected with --session", func() {
			_, err := targetOf("--session", "live-1", "--fresh")
			Expect(err).To(MatchError(ContainSubstring("--fresh")))
		})
	})

	Describe("GAVEL_COMMENT_ACTOR", func() {
		agent := comments.Actor{Kind: "agent", ID: "claude:session-1", Label: "claude-opus"}
		run := func(value string) (*cobra.Command, error) {
			GinkgoT().Setenv(commentapi.ActorEnv, value)
			reply, _, err := rootCmd.Find([]string{"comments", "reply"})
			Expect(err).NotTo(HaveOccurred())
			reply.SetContext(context.Background())
			return reply, reply.Parent().PersistentPreRunE(reply, nil)
		}

		It("makes a comments command act as the asserted agent", func() {
			reply, err := run(commentapi.EncodeActor(agent))
			Expect(err).NotTo(HaveOccurred())
			actor, ok := commentapi.ActorFrom(reply.Context())
			Expect([]any{actor, ok}).To(Equal([]any{agent, true}))
		})
		It("fails the command on a malformed assertion", func() {
			_, err := run("not-an-actor")
			Expect(err).To(MatchError(ContainSubstring(commentapi.ActorEnv)))
		})
	})
})
