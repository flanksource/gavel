# UIR architecture lint

Install a UIR version providing `.uir.d` configuration discovery and `_expires` timestamps on PATH. Gavel detects home, ancestor and project `.uir.yaml` files and `.uir.d/` YAML fragments, then runs `uir lint --format json --no-progress` from the project directory. UIR owns validation and merging; Gavel does not import its library. Run `gavel lint uir`; the results use Gavel's normal snapshots, baselines, summaries and UI.

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

The adapter analyzes the whole project before Gavel applies file scopes. A `.uir.yaml` or `.uir.d/` fragment scope reports the owning project, and nested configurations run independently with UIR's inherited rules. Fileless `uir/coverage` warnings remain visible through file filtering. Error findings fail the command; warnings and info alone succeed. Malformed reports, missing binaries, inconsistent exit statuses and execution errors fail loudly while retaining findings from a valid partial report. Autofixing is unsupported.

To keep native rules in `.gavel.yaml`, put the complete UIR document under `lint.linters.uir.config`. Gavel writes it atomically to a unique project `.uir.d/gavel-*.yaml` fragment, sets `_expires` to the linter deadline (or five minutes ahead when no deadline exists), and removes its own fragment after execution. On a later discovery UIR deletes expired leftovers and does not add their configuration. Unexpired fragments load normally. Dry-run creates no fragment, and authored files remain untouched. See [the UIR integration guide](../../docs/content/docs/linter-uir.mdx) for a runnable configuration example.

```yaml
lint:
  linters:
    uir:
      config:
        version: 1
        lint:
          rules:
            - id: parameter-count
              query: SELECT symbol.name AS name FROM func:* WHERE symbol.metrics.param_count > 4
              message: '{{.name}} has more than four parameters'
              severity: warning
      runtime:
        timeout: 5m
```
