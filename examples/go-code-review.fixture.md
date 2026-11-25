---
ai:
  model: agent:haiku:low
  maxTokens: 4000
  criteriaSection: Checklist
---

# Go code review

A fast, low-context AI review of the current Go change. It covers the structural and design problems that linters cannot see: command layering, package and file shape, prefix clusters, parallel abstractions, silent fallbacks, workarounds and test intent. Each checklist item is scored pass/fail on its own.

    gavel fixtures examples/go-code-review.fixture.md

```prompt
Work with as little context as possible:

- The change under review is the uncommitted Go work: read only `git status --short -- '*.go'` and `git diff HEAD -- '*.go'`. Untracked (`??`) files are new: read them whole, since they are all added lines. Earlier commits are out of scope. Only when the working tree has no Go changes, review `git show HEAD -- '*.go'` instead.
- For each directory the change touches, run only `wc -l <dir>/*.go`. When a prefix or duplication criterion needs symbol names, add `rg -n '^func |^type |^var |^const ' <dir>`. Open a whole file only when a hunk is ambiguous without it.
- Judge only added and changed lines. Pre-existing code is out of scope.
- Do not grade anything golangci-lint, go vet, gofmt, errcheck, funlen, nestif, unparam, betterleaks or the race detector would report.
- Ignore generated files (`dist/`, `*.pb.go`, `zz_generated*`).
- When a criterion does not apply to this change (for example, no `cmd/` files were touched), it passes with the message `n/a`.
- On a fail, write the message as `path:line — evidence → fix`, in 25 words or fewer.
```

## Checklist

### Command layering

- [ ] **CMD-1** No `cobra.Command` or `AddCommand` is defined outside `cmd/` or a `main` package.
- [ ] **CMD-2** Commands are thin. `RunE` only parses flags, builds one options struct, calls one provider/service/store function and renders the result. It has no loops over domain data, DB/HTTP/exec calls, business branching or result aggregation.
- [ ] **CMD-3** No domain types, view types or conversion/formatting helpers are declared in `cmd/`. They live beside the provider.
- [ ] **CMD-4** Flags are forwarded as an options struct, never read from package-level flag vars deep inside logic.

### Package and file shape

- [ ] **PKG-1** No new prefixed or tiny (<150 lines) file is added to a package where no file is ≥1000 lines. New code is folded into the existing file it belongs with.
- [ ] **PKG-2** Where ≥3 files share a prefix (`session_parse.go`, `session_store.go`, …) and one of them is ≥1000 lines, the change moves toward a `session/` subpackage rather than adding another `session_*.go`.
- [ ] **PKG-3** No touched file grows past 1000 lines (hard limit 2000) without being split by domain responsibility into a subpackage.
- [ ] **PKG-4** No package gains a cluster of ≥4 top-level funcs, vars or types sharing a prefix (`todoRunStart`, `todoRunStop`, …). Use a type with methods or a subpackage.
- [ ] **PKG-5** No new registry, provider, client, wrapper or interface parallels an existing one. The change extends the shared seam.
- [ ] **PKG-6** New names reuse the package's existing domain vocabulary (no `Job` beside `Task`, no `Fetch` beside `Get`).
- [ ] **PKG-7** No hardcoded list of kinds, names or IDs duplicates an existing catalog or enum.

### API design

- [ ] **API-1** There are no `Foo` + `FooWithX`/`FooV2` variants. One function takes an options struct.
- [ ] **API-2** No deprecated, legacy or compat shim is kept for callers in the same module.
- [ ] **API-3** No function takes more than 2 same-typed positional params or bool flag params. Use an options struct instead.
- [ ] **API-4** No third-party internal types leak through exported signatures.

### Correctness intent

- [ ] **ERR-1** No silent fallback. Invalid input or missing config fails loudly, and no zero, empty or default value is returned to "keep working".
- [ ] **ERR-2** No old code path is kept as a fallback for the new one.
- [ ] **ERR-3** Clicky task callbacks return errors, with no `panic`, `Fatal` or `os.Exit`, and log via `t.Infof`/`t.Errorf`, not `fmt`.
- [ ] **WA-1** Every workaround carries the `// WORKAROUND(reason):` + `Correct fix:` + `Ref:` block, and none hides behind innocuous helper names (`safeGet`, `tryParse`, `withFallback`).
- [ ] **WA-2** There is no `TODO fix later`, commented-out code, skipped test or hardcoded expected value.
- [ ] **WA-3** A fix lands in the shared wrapper or seam, not as a guard patched into one call site.

### Function shape

- [ ] **FN-1** No new single-call-site helper is extracted without a test or readability need.
- [ ] **FN-2** No new function near-duplicates an existing one.
- [ ] **FN-3** Comments state caveats or why, never narrate what the code does.

### Concurrency and output

- [ ] **CONC-1** Work orchestration uses `clicky.StartTask`/`StartGroup`, not raw `go func` + `sync.WaitGroup` or result channels.
- [ ] **OUT-1** User-facing CLI output returns clicky `Pretty`/`api.Text`, not `fmt.Print*` or string concatenation, and the format is chosen only at the output boundary.

### Tests

- [ ] **TEST-1** New Go tests use Ginkgo/Gomega, not `testing.T` asserts or testify.
- [ ] **TEST-2** Assertions compare against independent expectations: no trivial asserts, no reuse of the code-under-test's output as the oracle, no unexplained literals.
- [ ] **TEST-3** Each `It`/`Describe` description matches what its final `Expect` verifies.
- [ ] **TEST-4** Pure unit tests touch no DB or network, and core dependencies (SQL, redis) are not mocked where an integration test belongs.
