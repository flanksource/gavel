package main

import (
	"github.com/flanksource/clicky/api"
	clickytask "github.com/flanksource/clicky/task"
	flanksourceContext "github.com/flanksource/commons/context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type fixtureRenderTaskResult struct {
	text api.Text
}

func (r fixtureRenderTaskResult) PrettyShort() api.Textable {
	return r.text
}

var _ = Describe("Fixture live renderer", func() {
	It("clears the final frame", func() {
		var renderer clickytask.LiveRenderer = fixtureLiveRenderer{}
		Expect(renderer.RenderFinal(nil).IsEmpty()).To(BeTrue())
	})

	It("separates only visible fixture results", func() {
		clickytask.ClearTasks()
		clickytask.SetNoRender(true)
		DeferCleanup(func() {
			clickytask.ClearTasks()
			clickytask.SetNoRender(false)
		})

		hidden := clickytask.StartTask("hidden", func(_ flanksourceContext.Context, _ *clickytask.Task) (fixtureRenderTaskResult, error) {
			return fixtureRenderTaskResult{}, nil
		})
		visible := clickytask.StartTask("visible", func(_ flanksourceContext.Context, _ *clickytask.Task) (fixtureRenderTaskResult, error) {
			return fixtureRenderTaskResult{text: api.Text{}.Append("visible")}, nil
		})
		_, err := hidden.GetResult()
		Expect(err).NotTo(HaveOccurred())
		_, err = visible.GetResult()
		Expect(err).NotTo(HaveOccurred())

		rendered := fixtureLiveRenderer{}.RenderLive([]*clickytask.Task{hidden.Task, visible.Task, hidden.Task})

		Expect(rendered.String()).To(Equal("visible"))
	})
})
