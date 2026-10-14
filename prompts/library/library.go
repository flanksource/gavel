// Package library ships gavel's runtime-discovered prompts. Every *.prompt here
// is offered to Captain as a read-only prompt source, so adding one needs no Go:
// it is run by bare name (POST /api/v1/prompt/run?id=<name>) and declares its
// own setup, permissions, verification and post-run steps in frontmatter.
package library

import "embed"

//go:embed *.prompt
var FS embed.FS

// Label names the source in Captain's prompt listings.
const Label = "gavel"
