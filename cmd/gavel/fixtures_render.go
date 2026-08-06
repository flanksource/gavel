package main

import (
	"github.com/flanksource/clicky/api"
	clickytask "github.com/flanksource/clicky/task"
)

type fixtureLiveRenderer struct{}

func (fixtureLiveRenderer) RenderLive(tasks []*clickytask.Task) api.Text {
	text := api.Text{}
	for _, task := range tasks {
		rendered := task.Pretty()
		if rendered.IsEmpty() {
			continue
		}
		if !text.IsEmpty() {
			text = text.NewLine()
		}
		text = text.Add(rendered)
	}
	return text
}

func (fixtureLiveRenderer) RenderFinal([]*clickytask.Task) api.Text {
	return api.Text{}
}
