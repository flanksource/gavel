package main

import (
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("commit runtime flag presence", func() {
	It("preserves an explicitly empty group model through the actual command binder", func() {
		var captured CommitOptions
		cmd := registerCommitCommand(&cobra.Command{Use: "gavel"}, func(opts CommitOptions) (any, error) {
			captured = opts
			return nil, nil
		})
		Expect(cmd.ParseFlags([]string{"--model=api:sonnet", "--group-model="})).To(Succeed())
		Expect(cmd.RunE(cmd, nil)).To(Succeed())
		opts := buildCommitOptions(captured, "/work/review", verify.GavelConfig{}, nil)
		model, err := opts.Flags.ToModel()
		Expect(err).NotTo(HaveOccurred())
		_, err = api.ResolveSpecLayers(api.ResolveSpecOptions{
			Layers: []api.SpecLayer{
				api.RequestSpecLayer("command", api.Spec{Model: model}),
				api.RequestSpecLayer("group override", api.Spec{Model: opts.GroupModel}),
			},
			Saved: &captainconfig.AIDefaults{DefaultModel: "api:gpt-5"}, RequireModel: true,
		})
		Expect(err).To(MatchError(ContainSubstring("configure")))
	})

	It("keeps an explicit empty effort above saved provider effort through the actual command binder", func() {
		var captured CommitOptions
		cmd := registerCommitCommand(&cobra.Command{Use: "gavel"}, func(opts CommitOptions) (any, error) {
			captured = opts
			return nil, nil
		})
		Expect(cmd.ParseFlags([]string{"--model=api:sonnet", "--effort="})).To(Succeed())
		Expect(cmd.RunE(cmd, nil)).To(Succeed())
		model, err := captured.ModelFlags.ToModel()
		Expect(err).NotTo(HaveOccurred())
		resolved, err := api.ResolveSpecLayers(api.ResolveSpecOptions{
			Layers: []api.SpecLayer{api.RequestSpecLayer("command", api.Spec{Model: model})},
			Saved: &captainconfig.AIDefaults{Providers: map[string]captainconfig.ProviderDefaults{
				"anthropic": {Mode: "api", ReasoningEffort: "high"},
			}}, RequireModel: true,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolved.Spec.Effort).To(BeEmpty())
		Expect(resolved.Provenance["/effort"].Source.Name).To(Equal("command"))
	})

	It("keeps explicit false and an empty fallback above configured values through the actual command binder", func() {
		var captured CommitOptions
		cmd := registerCommitCommand(&cobra.Command{Use: "gavel"}, func(opts CommitOptions) (any, error) {
			captured = opts
			return nil, nil
		})
		Expect(cmd.ParseFlags([]string{"--model=api:sonnet", "--no-cache=false", "--fallback="})).To(Succeed())
		Expect(cmd.RunE(cmd, nil)).To(Succeed())
		model, err := captured.ModelFlags.ToModel()
		Expect(err).NotTo(HaveOccurred())
		resolved, err := api.ResolveSpecLayers(api.ResolveSpecOptions{
			Layers: []api.SpecLayer{
				api.RequestSpecLayer("configured", api.Spec{Model: api.Model{Fallbacks: api.ModelList{{Name: "api:gpt-5"}}}}),
				api.RequestSpecLayer("command", api.Spec{Model: model}),
			},
			Saved: &captainconfig.AIDefaults{NoCache: true}, RequireModel: true,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolved.Spec.NoCache).To(BeFalse())
		Expect(resolved.Spec.Fallbacks).To(BeEmpty())
	})
})
