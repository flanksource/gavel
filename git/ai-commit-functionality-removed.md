You are reviewing a git diff for user-visible removals. Return structured output describing functionality removed by this change.

DIFF INPUT:

{{.patch}}

REQUIREMENTS:
- Only include user-visible functionality removed by the diff.
- Include removed commands, flags, config keys, APIs, endpoints, workflows, or behaviors when they are no longer supported.
- Do not include internal refactors, renames, or implementation churn unless they remove externally visible behavior.
- Keep each item concise and specific.
- Return a JSON object with a `functionalityRemoved` array.
- Return `{"functionalityRemoved":[]}` when no user-visible functionality has been removed.
