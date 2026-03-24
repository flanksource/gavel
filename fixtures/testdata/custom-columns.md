---
exec: bash
args: ["-c", "curl {{.flags}} {{.baseUrl}}{{.path}}"]
baseUrl: https://httpbin.flanksource.com
flags: "-s"
---

## Exec templated from custom columns

Frontmatter defines the command template with `{{.col}}` placeholders.
Frontmatter metadata provides defaults; per-row column values override them.

| Name | path | CEL Validation |
|------|------|----------------|
| get endpoint | /get | json.url.contains("httpbin") |
| get ip | /ip | json.origin != "" |

## Override defaults per row

| Name | flags | path | CEL Validation |
|------|-------|------|----------------|
| get with headers | -sI | /get | stdout.contains("HTTP") |
