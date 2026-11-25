package main

import (
	"context"

	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("commit target flag", func() {
	It("accepts a target separately from the push mode and file arguments", func() {
		const target = "zim/main:feat/new-topic"
		var files []string
		cmd := registerCommitCommand(&cobra.Command{Use: "gavel"}, func(opts CommitOptions) (any, error) {
			files = opts.Args
			Expect(opts.Push).To(Equal("branch"))
			Expect(buildCommitOptions(opts, "/repo", verify.GavelConfig{}, nil).Target).To(Equal(target))
			return nil, nil
		})
		Expect(cmd.ParseFlags([]string{"--push=branch", "--target", target, "feature.go"})).To(Succeed())
		Expect(cmd.Flags().GetString("target")).To(Equal(target))
		cmd.SetContext(context.Background())
		Expect(cmd.RunE(cmd, cmd.Flags().Args())).To(Succeed())
		Expect(files).To(Equal([]string{"feature.go"}))
	})
})
