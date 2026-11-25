---
cwd: "{{.ROOT_DIR}}"
exec: bash
args: ["-c", "go test -v -p 1 {{.packages}} -count=1 {{.flags}}"]
env:
  GAVEL_SLOTS_BINARY: "{{.executablePath}}"
---

## Machine-wide slot verification

| Name | packages | flags | Timeout | Exit Code | CEL Validation |
| --- | --- | --- | --- | --- | --- |
| Slot parsing, disabled behavior and execution budgets | ./slots | -run TestSlots -args -ginkgo.no-color -ginkgo.skip 'Slot subprocess integration' | 10m | 0 | exitCode == 0 && stdout.matches('(?s).*Ran [1-9][0-9]* of.*') |
| Cross-process worktree, repository, crash and timeout integration | ./slots | -run TestSlots -args -ginkgo.no-color -ginkgo.focus 'Slot subprocess integration' | 10m | 0 | exitCode == 0 && stdout.matches('(?s).*Ran [1-9][0-9]* of.*') |
| Real gavel test processes across worktrees and repositories | ./slots/cliintegration | -tags=integration -run TestSlotsCLI -args -ginkgo.no-color | 10m | 0 | exitCode == 0 && stdout.matches('(?s).*Ran [1-9][0-9]* of.*') |
| Layered configuration, generated schema and command admission | ./verify ./processruntime | -timeout=5m -args -ginkgo.no-color | 10m | 0 | exitCode == 0 && stdout.contains('PASS') && !stdout.contains('[no tests to run]') |
| Test package slot admission | ./testrunner | -run TestTestRunner -args -ginkgo.no-color -ginkgo.focus 'Test runner slots' | 10m | 0 | exitCode == 0 && stdout.matches('(?s).*Ran [1-9][0-9]* of.*') |
| Unconfigured Go batching regressions | ./testrunner | -run TestTestRunner -args -ginkgo.no-color -ginkgo.focus 'unified Go runner' | 10m | 0 | exitCode == 0 && stdout.matches('(?s).*Ran [1-9][0-9]* of.*') |
| Test adapter regressions | ./testrunner/runners | -run TestRunners -args -ginkgo.no-color | 10m | 0 | exitCode == 0 && stdout.matches('(?s).*Ran [1-9][0-9]* of.*') |
| Slot-aware fixture hard backstops | ./fixtures | -run TestFixtures -args -ginkgo.no-color -ginkgo.focus 'Slot-aware fixture backstops' | 10m | 0 | exitCode == 0 && stdout.matches('(?s).*Ran [1-9][0-9]* of.*') |
| Fixture worker isolation regressions | ./fixtures | -run TestFixtures -args -ginkgo.no-color -ginkgo.focus 'Fixture runtime isolation' | 10m | 0 | exitCode == 0 && stdout.matches('(?s).*Ran [1-9][0-9]* of.*') |
| Fixture pipe and PTY command admission | ./fixtures/types | -run TestFixtureTypes -args -ginkgo.no-color -ginkgo.focus 'Fixture command slots' | 10m | 0 | exitCode == 0 && stdout.matches('(?s).*Ran [1-9][0-9]* of.*') |
| Linter command admission and timeout accounting | ./lint | -run TestLint -args -ginkgo.no-color -ginkgo.focus 'takes a slot' | 10m | 0 | exitCode == 0 && stdout.matches('(?s).*Ran [1-9][0-9]* of.*') |
| Linter wrapper regression | ./linters | -run TestRunLinterWithTaskUsesResolvedExecutable | 10m | 0 | exitCode == 0 && stdout.contains('--- PASS: TestRunLinterWithTaskUsesResolvedExecutable') |
| Linter shim admission and timeout accounting | ./cmd/gavel | -run TestGavelCLI -args -ginkgo.no-color -ginkgo.focus 'lint shim-exec' | 10m | 0 | exitCode == 0 && stdout.matches('(?s).*Ran [1-9][0-9]* of.*') |
