package main

import (
	"github.com/flanksource/gavel/todos/native"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("todos land flags", func() {
	target := TodoTargetOptions{IDs: []string{"abc12345"}}

	DescribeTable("requires exactly one of --merge/--pr", func(opts TodosLandOptions, wantErr string) {
		opts.TodoTargetOptions = target
		_, err := opts.landOptions()
		Expect(err).To(MatchError(ContainSubstring(wantErr)))
	},
		Entry("neither", TodosLandOptions{}, "exactly one of --merge or --pr"),
		Entry("both", TodosLandOptions{Merge: true, PR: true}, "exactly one of --merge or --pr"),
		Entry("--draft on a merge", TodosLandOptions{Merge: true, Draft: true}, "--draft and --base only apply to --pr"),
		Entry("--base on a merge", TodosLandOptions{Merge: true, Base: "origin/main"}, "--draft and --base only apply to --pr"),
	)

	It("maps --pr with --draft and --base onto a PR landing", func() {
		got, err := TodosLandOptions{TodoTargetOptions: target, PR: true, Draft: true, Base: "origin/release"}.landOptions()
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Via).To(Equal(native.LandingPR))
		Expect(got.Draft).To(BeTrue())
		Expect(got.Base).To(Equal("origin/release"))
		Expect(got.Deps.CreatePR).NotTo(BeNil())
	})

	It("maps --merge onto a merge landing", func() {
		got, err := TodosLandOptions{TodoTargetOptions: target, Merge: true}.landOptions()
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Via).To(Equal(native.LandingMerge))
	})

	It("registers land under todos with its flags", func() {
		for _, flag := range []string{"merge", "pr", "draft", "base"} {
			Expect(todosLandCmd.Flags().Lookup(flag)).NotTo(BeNil(), flag)
		}
	})
})
