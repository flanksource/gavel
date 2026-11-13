package linters

import (
	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Lint source excerpts", func() {
	It("shows numbered YAML at normal verbosity without rereading the file", func() {
		original := clicky.Flags.LevelCount
		DeferCleanup(func() { clicky.Flags.LevelCount = original })
		clicky.Flags.LevelCount = 0
		v := models.Violation{File: "unavailable.yaml", Line: 2, Column: 3, Message: models.StringPtr("unknown cleanup"), Severity: models.SeverityError,
			Excerpt: models.NewSourceExcerpt([]byte("step:\n  cleanup: true\n"), 2, 3)}
		Expect((&violationNode{v: &v}).Pretty().String()).To(ContainSubstring("> 2 |   cleanup: true\n    |   ^"))
	})
})
