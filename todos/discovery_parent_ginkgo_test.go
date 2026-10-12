package todos

import (
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DiscoveryFilters parent selection", func() {
	const (
		parentID      = "11111111-1111-4111-8111-111111111111"
		otherParentID = "22222222-2222-4222-8222-222222222222"
	)
	topLevel := &types.TODO{ID: parentID}
	child := &types.TODO{ID: "33333333-3333-4333-8333-333333333333", ParentID: parentID}
	otherChild := &types.TODO{ID: "44444444-4444-4444-8444-444444444444", ParentID: otherParentID}

	DescribeTable("Matches",
		func(filters DiscoveryFilters, todo *types.TODO, want bool) {
			Expect(filters.Matches(todo)).To(Equal(want))
		},
		Entry("no parent filter keeps a child", DiscoveryFilters{}, child, true),
		Entry("TopLevelOnly keeps a top-level todo", DiscoveryFilters{TopLevelOnly: true}, topLevel, true),
		Entry("TopLevelOnly drops a child", DiscoveryFilters{TopLevelOnly: true}, child, false),
		Entry("ParentID keeps that parent's child", DiscoveryFilters{ParentID: parentID}, child, true),
		Entry("ParentID drops another parent's child", DiscoveryFilters{ParentID: parentID}, otherChild, false),
		Entry("ParentID drops the parent itself", DiscoveryFilters{ParentID: parentID}, topLevel, false),
	)

	DescribeTable("IsEmpty",
		func(filters DiscoveryFilters, want bool) {
			Expect(filters.IsEmpty()).To(Equal(want))
		},
		Entry("no filters", DiscoveryFilters{}, true),
		Entry("TopLevelOnly", DiscoveryFilters{TopLevelOnly: true}, false),
		Entry("ParentID", DiscoveryFilters{ParentID: parentID}, false),
	)
})
