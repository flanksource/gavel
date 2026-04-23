You are a commit message generator. Analyze the diff below and produce a Conventional Commit message.

DIFF INPUT:

{{.patch}}

REQUIREMENTS:
- type: one of feat|fix|perf|refactor|test|docs|build|ci|chore|revert
- scope (optional): the area of the codebase affected, e.g. db|api|fe|crd|chart|docker|kubernetes|terraform|aws
- subject: imperative mood, no trailing period, <=100 characters
- body (optional): a concise explanation of why and impact; include "BREAKING CHANGE: ..." if behavior changes or APIs are removed; reference issues like "Refs #123" or "Fixes #123" if applicable
- Explain intention and effect; do not restate the code.
