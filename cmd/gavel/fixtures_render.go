package main

import (
	"github.com/flanksource/clicky/api"
	clickytask "github.com/flanksource/clicky/task"
)

type fixtureLiveRenderer struct{}

func (fixtureLiveRenderer) RenderLive(tasks []*clickytask.Task) api.Text {
	text := api.Text{}
	for i, task := range tasks {
		if i > 0 {
			text = text.NewLine()
		}
		text = text.Add(task.Pretty())
	}
	return text
}

func (fixtureLiveRenderer) RenderFinal([]*clickytask.Task) api.Text {
	return api.Text{}
}
