package ui

import (
	"net/http/httptest"
	"time"

	"github.com/flanksource/gavel/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PR age exports", func() {
	DescribeTable("filters export results by update age", func(age string, expected []int) {
		now := time.Now()
		server := &Server{prs: github.PRSearchResults{
			{Number: 1, UpdatedAt: now.Add(-time.Hour)},
			{Number: 2, UpdatedAt: now.Add(-48 * time.Hour)},
			{Number: 3, UpdatedAt: now.Add(-14 * 24 * time.Hour)},
			{Number: 4, UpdatedAt: now.Add(-60 * 24 * time.Hour)},
			{Number: 5},
		}}
		req, ok := parseRouteRequest(httptest.NewRequest("GET", "/prs.json?age="+age, nil))
		Expect(ok).To(BeTrue())
		report, err := server.buildExportReport(req)
		Expect(err).NotTo(HaveOccurred())
		Expect(report.PRs).To(WithTransform(func(nodes []*PRViewNode) []int {
			var numbers []int
			for _, node := range nodes {
				numbers = append(numbers, node.Number)
			}
			return numbers
		}, Equal(expected)))
		Expect(renderPRFilters(req.PRFilters)).To(Equal("age=" + age))
	},
		Entry("under one day", "day", []int{1}),
		Entry("one to seven days", "week", []int{2}),
		Entry("seven to thirty days", "month", []int{3}),
		Entry("older", "older", []int{4}),
		Entry("missing update", "unknown", []int{5}),
		Entry("excluded age", "-day", []int{2, 3, 4, 5}),
		Entry("multiple ages", "day,week", []int{1, 2}),
	)
})
