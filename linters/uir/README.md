# UIR architecture lint

Install a UIR version providing `uir lint` on PATH and put `.uir.yaml` at the project root. Gavel detects that file and runs `uir lint --config <project>/.uir.yaml --format json --no-progress`. Run `gavel lint uir` or `gavel lint --linters uir`; the results use Gavel's normal snapshots, baselines, summaries and UI.

UIR incrementally indexes the project's Go modules, then evaluates SQL rules against pinned snapshots. Every row is a violation. A rule requires `id`, `query` and a Go template `message`; severity defaults to `error`, with `warning` and `info` also supported. Additional UIR flags can be supplied through the existing linter runtime arguments.

```yaml
version: 1
lint:
  include_tests: false
  rules:
    - id: parameter-count
      message: "{{.name}} has {{.params}} parameters"
      query: |
        SELECT symbol.name AS name, symbol.metrics.param_count AS params
        FROM func:* WHERE symbol.metrics.param_count > 3
```

The adapter analyzes the whole project before Gavel applies file scopes. A `.uir.yaml` scope reports the owning project, and nested configurations run independently. Fileless `uir/coverage` warnings remain visible through file filtering. Error findings fail the command; warnings and info alone succeed. Malformed reports, missing binaries, inconsistent exit statuses and execution errors fail loudly while retaining findings from a valid partial report. Autofixing is unsupported.
