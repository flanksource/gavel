You are reviewing a git diff for backward compatibility risks. Return structured output describing compatibility issues introduced by this change.

DIFF INPUT:

{{.patch}}

REQUIREMENTS:
- Include breaking changes, required migrations, config changes, upgrade risks, or other backward compatibility issues introduced by the diff.
- Focus on risks a caller, operator, or downstream consumer must know before taking this change.
- Do not include ordinary refactors or low-risk internal cleanup.
- Keep each item concise and specific.
- Return a JSON object with a `compatibilityIssues` array.
- Return `{"compatibilityIssues":[]}` when there are no compatibility issues.
