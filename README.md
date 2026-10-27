# Gavel

Gavel discovers, runs, and reports tests, linters, fixtures, and development workflows across Go and TypeScript projects.

## Install

```bash
go install github.com/flanksource/gavel/cmd/gavel@latest
```

Build from a source checkout with `make build`; the binary is written to `.bin/gavel`.

## Quick start

```bash
gavel test
gavel lint
gavel test --lint --ui
gavel fixtures examples/smoke-test.fixture.md
```

## Documentation

Browse [gavel.flanksource.com/docs](https://gavel.flanksource.com/docs) or read the [documentation source](docs/content/docs/getting-started.mdx).

- [CLI usage](docs/content/docs/cli-usage.mdx) and [GitHub Action](docs/content/docs/github-action.mdx)
- [Testing](docs/content/docs/test-execution.mdx), [linting](docs/content/docs/lint-execution.mdx), and [reporting](docs/content/docs/reporting.mdx)
- [Fixtures](docs/content/docs/fixtures-overview.mdx) and [worked service guides](docs/content/docs/fixture-kind-services.mdx)
- [TODO workflows](docs/content/docs/todo-engine.mdx), [PR-linked TODOs](docs/content/docs/pr-todos.mdx), and [pull requests](docs/content/docs/pull-requests.mdx)
- [Configuration](docs/content/docs/configuration-reference.mdx), [JSON Schema](gavel.schema.json), and [annotated example](gavel.yaml.example)
- [Agent skills](docs/content/docs/agent-skills.mdx)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for architecture, build/test instructions, PostgreSQL integration tests, and the documentation app. Coding agents should read [AGENTS.md](AGENTS.md).

## License

See [LICENSE](LICENSE).
