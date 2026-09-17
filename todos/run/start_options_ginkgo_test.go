package run

import (
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("run lifecycle options", func() {
	ginkgo.It("forwards ordered preset selectors and explicit-empty state to resolution and dispatch", func() {
		opts := runOptions(Request{Options: Options{Presets: []string{"organization", "review"}, PresetsSet: true}}, nil)
		Expect(opts.Presets).To(Equal([]string{"organization", "review"}))
		Expect(opts.PresetsSet).To(BeTrue())
	})
})
