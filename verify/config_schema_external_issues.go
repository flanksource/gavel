package verify

func externalIssuesSchema() map[string]any {
	return object("External issue provider settings.", map[string]any{
		"jira": object("Jira issue settings.", map[string]any{
			"ignoreFields": stringArray("Field IDs or Jira display names omitted from issue get, list and sync outputs. Matches either value using commons MatchItems: case-insensitive exact, prefix*, *suffix or *contains*. Patterns accumulate uniquely across configuration layers. Empty leaves all fields unchanged. Identity fields id, key, summary, url and project are always retained. Exclusions override list --fields; comment and attachment exclusions omit their counts, and attachment exclusions disable sync downloads and extraction. Explicit comment and attachment operations are unaffected."),
		}),
	})
}
