---
cwd: /Users/moshe/go/src/github.com/flanksource/clicky-ui
---

# Structured lint launch verification

These checks validate the local linked workspaces. The API suite exercises persisted admission, strict root and child schemas, caller isolation, cancellation, recovery, public prompt redaction and native execution. Store contract tests run against JSON and SQLite.

| Name | CWD | Command | CEL |
| --- | --- | --- | --- |
| Launch lifecycle and both result adapters | /Users/moshe/go/src/github.com/flanksource/gavel | go test ./launch ./resultstore -race -count=1 | exitCode == 0 |
| Native configuration schemas | /Users/moshe/go/src/github.com/flanksource/gavel | go test ./cmd/gavel -run '^TestGavelCLI$' -ginkgo.focus='launch native option schemas' -count=1 | exitCode == 0 |
| Private Git worktree patches | /Users/moshe/go/src/github.com/flanksource/commons-db | go test ./shell -run '^TestShell$' -ginkgo.focus='supplied workspace patches' -count=1 | exitCode == 0 |
| Frozen native prompt preparation | /Users/moshe/go/src/github.com/flanksource/captain | go test ./pkg/cli -run '^TestSessionGetMulti$' -ginkgo.focus='prepared prompt inputs' -count=1 | exitCode == 0 |
| Source review persistence and prompt recovery | /Users/moshe/go/src/github.com/flanksource/clicky-ui/apps/playground | node node_modules/vitest/vitest.mjs run plugins/reviewable-source.test.ts plugins/reviewable-store.test.ts plugins/prompt-runs-server.test.ts src/prompt-runs/RunPromptDialog.test.tsx --pool forks --maxWorkers 1 --testTimeout 15000 --hookTimeout 15000 | exitCode == 0 |
