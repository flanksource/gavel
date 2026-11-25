// Package library ships gavel's runtime-discovered prompts. Every *.prompt here
// is offered to Captain as a read-only prompt source, so adding one needs no Go:
// it is run by bare name (POST /api/v1/prompt/run?id=<name>) and declares its
// own setup, permissions, verification and post-run steps in frontmatter.
package library

import (
	"embed"

	"github.com/flanksource/gavel/prompts"
)

//go:embed *.prompt
var FS embed.FS

// Label names the source in Captain's prompt listings.
const Label = "gavel"

const ChecklistTemplateKey = "gavel.checklist.template"

//go:embed acceptance-checklist.prompt
var ChecklistDefault string

func Prompts() []prompts.Prompt {
	return []prompts.Prompt{{ID: prompts.TodosVerify, Title: "Acceptance checklist", Description: "Instructions for grading acceptance criteria; the verdict schema is fixed.", ConfigPath: prompts.TodosVerify, Default: ChecklistDefault, UsedBy: []string{"gavel todos check", "gavel fixtures"}}}
}
