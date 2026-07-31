package main

import (
	"fmt"
	"strings"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos"
)

type todoTextOptions struct {
	WorkDir string
	Flag    string
	Value   string
}

func resolveTodoText(opts todoTextOptions) (string, error) {
	ref, err := fixtures.ResolveFileRef(opts.WorkDir, opts.Value)
	if err != nil {
		return "", fmt.Errorf("resolve %s: %w", opts.Flag, err)
	}
	if ref.IsFile {
		return strings.TrimSpace(ref.Contents), nil
	}
	return strings.TrimSpace(ref.Raw), nil
}

type todoCreateContentOptions struct {
	Body            string
	BodySet         bool
	Plan            string
	PlanSet         bool
	Verification    string
	VerificationSet bool
}

type todoCreateContent struct {
	Body         string
	Plan         string
	Verification string
}

func resolveTodoCreateContent(workDir string, opts todoCreateContentOptions) (todoCreateContent, error) {
	var content todoCreateContent
	var err error
	if opts.BodySet {
		content.Body, err = resolveTodoText(todoTextOptions{WorkDir: workDir, Flag: "--body", Value: opts.Body})
		if err != nil {
			return todoCreateContent{}, err
		}
	}
	if opts.PlanSet {
		content.Plan, err = resolveTodoText(todoTextOptions{WorkDir: workDir, Flag: "--plan", Value: opts.Plan})
		if err != nil {
			return todoCreateContent{}, err
		}
		if content.Plan == "" {
			return todoCreateContent{}, fmt.Errorf("--plan cannot be empty")
		}
	}
	if opts.VerificationSet {
		content.Verification, err = resolveTodoText(todoTextOptions{WorkDir: workDir, Flag: "--verification", Value: opts.Verification})
		if err != nil {
			return todoCreateContent{}, err
		}
		if content.Verification == "" {
			return todoCreateContent{}, fmt.Errorf("--verification cannot be empty")
		}
	}
	var bodyVerification string
	content.Body, bodyVerification, _ = todos.SplitVerificationFixture(content.Body)
	content.Verification = todos.CombineVerificationFixtures(content.Verification, bodyVerification)
	return content, nil
}
