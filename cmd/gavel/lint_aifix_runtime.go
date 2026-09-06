package main

import (
	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/captainconfig"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/gavel/ai/aifix"
	"github.com/flanksource/gavel/linters"
)

type lintFixRuntime struct {
	Prompt  aifix.ResolveOptions
	Runtime captaincli.AIRuntimeOptions
	Saved   captainconfig.Config
}

func (o lintFixRuntime) Resolve(results []*linters.LinterResult) (captaincli.AIRuntimeResolved, error) {
	o.Prompt.Results = results
	layers, err := aifix.Layers(o.Prompt)
	if err != nil {
		return captaincli.AIRuntimeResolved{}, err
	}
	return buildAIFixRequest(aiFixRequestOptions{Runtime: o.Runtime, Layers: layers, Saved: o.Saved, Dir: o.Prompt.Dir})
}

func (o lintFixRuntime) BuildRequest(results []*linters.LinterResult) (captainai.Request, error) {
	resolved, err := o.Resolve(results)
	return resolved.Request, err
}
