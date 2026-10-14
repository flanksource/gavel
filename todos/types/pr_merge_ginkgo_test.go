package types

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PR links when merging TODOs", func() {
	It("preserves an explicit link and rejects different PRs", func() {
		linked, unlinked, different := &TODO{}, &TODO{}, &TODO{}
		linked.PR = &PR{Number: 27, URL: "https://github.com/acme/service/pull/27"}
		different.PR = &PR{Number: 28, URL: "https://github.com/acme/service/pull/28"}
		pr, err := MergedPR([]*TODO{unlinked, linked})
		Expect(err).NotTo(HaveOccurred())
		Expect(pr).To(Equal(linked.PR))
		_, err = MergedPR([]*TODO{unlinked, linked, different})
		Expect(err).To(MatchError(ContainSubstring("different PRs")))
		Expect(unlinked.PR).To(BeNil())
	})
})
