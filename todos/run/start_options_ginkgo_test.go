package run

import (
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("run lifecycle options", func() {
	ginkgo.DescribeTable("forwards the runtime profile selector to resolution and dispatch",
		func(profile string) {
			opts := runOptions(Request{Options: Options{RuntimeProfile: profile}}, nil)
			Expect(opts.RuntimeProfile).To(Equal(profile))
		},
		ginkgo.Entry("explicit selection", "Review profile"),
		ginkgo.Entry("unset selection", ""),
	)
})
