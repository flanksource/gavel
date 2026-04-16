You are planning a series of git commits from a staged change set.

Each file below is already staged. Group the files into a small sequence of cohesive commits.

CHANGED FILES:
{{range .changes}}
- path: {{.Path}}
  status: {{.Status}}
  {{- if .PreviousPath}}
  previous_path: {{.PreviousPath}}
  {{- end}}
  additions: {{.Adds}}
  deletions: {{.Dels}}
{{end}}

REQUIREMENTS:
- Every file must appear exactly once across the output groups.
- Prefer keeping tiny related edits and rename-style changes together.
- Split broader unrelated work into a few sizable commits instead of one giant commit.
- Groups must be file-based only. Do not propose partial-file or hunk-level splits.
- Preserve a sensible commit order when one group logically builds on another.
- Use short labels that describe the theme of each group.
{{- if .max}}
- Create at most {{.max}} groups. Merge smaller related changes to stay within this limit.
{{- end}}

Return JSON with a top-level "groups" array. Each element has "label" (string) and "files" (string array).
