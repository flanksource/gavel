---
daemon: go run ./examples/sample-app --addr :{{.port}}
---

# Smoke test

Starts the bundled [`examples/sample-app`](sample-app) in the background — gavel
picks a free `{{.port}}`, waits for it to listen, and stops it afterwards — then
probes it. This is the shape you'd use to smoke-test a real service after a
deploy; swap the `daemon:` command and paths for your own app.

    gavel fixtures examples/smoke-test.fixture.md

## Fast smoke tests

```yaml test
paths: [./examples/sample-app]
framework: [go]
extra-args: ["-run", "Smoke"]
test-timeout: 60s
```

## Health endpoint responds

### command: health

```bash
curl -fsS "http://localhost:{{.port}}/health"
```

- cel: json.status == "ok"
