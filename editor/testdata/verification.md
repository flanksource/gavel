---
cwd: $GIT_ROOT_DIR
timeout: 300s
---

# Browser editor verification

| Name | Command | Exit Code | CEL Validation |
| --- | --- | --- | --- |
| Command interface | .bin/gavel edit --help | 0 | stdout.contains("no-open") && stdout.contains("schema") && stdout.contains("port") |
| Required input | .bin/gavel edit | 1 | (stdout + stderr).contains("at least one") |
| Backend contracts | .bin/gavel test ./editor ./linters/json -- --focus 'Editor' | 0 | (stdout + stderr).contains("passed") |
| Store boundaries | .bin/gavel test ./editor -- --focus 'File store' | 0 | (stdout + stderr).contains("passed") |
| Autosave ordering | pnpm --dir editor/ui test | 0 | (stdout + stderr).contains("passed") |
| Frontend types | pnpm --dir editor/ui run check | 0 | exitCode == 0 |
| Patch anchors | .bin/gavel test ./comments --recursive=false -- --focus 'Patch anchors' | 0 | (stdout + stderr).contains("passed") |
| Comment persistence | .bin/gavel test ./editor --recursive=false -- --focus 'Editor comment' | 0 | (stdout + stderr).contains("passed") |
| Shared annotation controls | pnpm --dir ../clicky-ui --filter @flanksource/clicky-ui exec vitest run src/comments/AnnotationSurface.test.tsx src/monaco/DocumentEditor.test.tsx | 0 | (stdout + stderr).contains("passed") |
