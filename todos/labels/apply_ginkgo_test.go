package labels_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/todos/labels"
)

var _ = Describe("Split", func() {
	It("flattens comma-separated tokens and slice entries alike", func() {
		Expect(labels.Split([]string{"bug,api", "perf"})).To(Equal([]string{"bug", "api", "perf"}))
	})

	It("trims each token and drops blanks", func() {
		Expect(labels.Split([]string{" bug ", "", "  ", "api,"})).To(Equal([]string{"bug", "api"}))
	})

	It("drops duplicates case-insensitively, keeping the first spelling", func() {
		Expect(labels.Split([]string{"Bug", "bug", "BUG"})).To(Equal([]string{"Bug"}))
	})

	It("returns an empty set for no input", func() {
		Expect(labels.Split(nil)).To(BeEmpty())
	})
})

var _ = Describe("Apply", func() {
	existing := []string{"docs", "source:todo", "perf"}

	It("appends additions after the existing labels, in order", func() {
		Expect(labels.Apply(existing, []string{"bug", "api"}, nil)).
			To(Equal([]string{"docs", "source:todo", "perf", "bug", "api"}))
	})

	It("removes only what was named, preserving the rest of the order", func() {
		Expect(labels.Apply(existing, nil, []string{"docs"})).
			To(Equal([]string{"source:todo", "perf"}))
	})

	It("treats a removal of a label that is not there as a no-op", func() {
		Expect(labels.Apply(existing, nil, []string{"ci"})).To(Equal(existing))
	})

	It("never duplicates a label the set already carries", func() {
		Expect(labels.Apply(existing, []string{"PERF"}, nil)).To(Equal(existing))
	})

	It("matches removals case-insensitively", func() {
		Expect(labels.Apply(existing, nil, []string{"DOCS"})).
			To(Equal([]string{"source:todo", "perf"}))
	})

	It("keeps a label named in both add and remove, because removals apply first", func() {
		Expect(labels.Apply(existing, []string{"docs"}, []string{"docs"})).
			To(Equal([]string{"source:todo", "perf", "docs"}))
	})

	It("leaves the caller's slice untouched", func() {
		before := append([]string(nil), existing...)
		labels.Apply(existing, []string{"bug"}, []string{"docs"})
		Expect(existing).To(Equal(before))
	})

	It("returns the additions alone when there were no labels", func() {
		Expect(labels.Apply(nil, []string{"bug"}, nil)).To(Equal([]string{"bug"}))
	})
})

var _ = Describe("Equal", func() {
	It("treats nil and empty as the same set", func() {
		Expect(labels.Equal(nil, []string{})).To(BeTrue())
	})

	It("is order-sensitive, so a reordering counts as a change", func() {
		Expect(labels.Equal([]string{"bug", "api"}, []string{"api", "bug"})).To(BeFalse())
	})

	It("ignores case and surrounding space", func() {
		Expect(labels.Equal([]string{"Bug"}, []string{" bug "})).To(BeTrue())
	})

	It("reports a differing length as unequal", func() {
		Expect(labels.Equal([]string{"bug"}, []string{"bug", "api"})).To(BeFalse())
	})
})
