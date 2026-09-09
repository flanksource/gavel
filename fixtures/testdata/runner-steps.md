# Runner step fixtures

Demonstrates the `yaml test` / `yaml lint` fence kinds. The block body is
unmarshalled directly onto `gavel test` / `gavel lint` options. Run with:

    gavel fixtures fixtures/testdata/runner-steps.md

## Run the sample tests

```yaml test
paths: [./fixtures/testdata/runstep_sample]
framework: [go]
show-passed: true
test-timeout: 60s
```

## Lint the sample

```yaml lint
linters: [golangci-lint]
files: [./fixtures/testdata/runstep_sample]
timeout: 60s
```
