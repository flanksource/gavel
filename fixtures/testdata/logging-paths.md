---
exec: bash
args: ["-c", "{{.cmd}}"]
build: "echo build-ok"
---

## Build and variable expansion logging

Exercises code paths that produce V(2)-V(6) log output.
Run with -vvvvvv to see all levels.

| Name | cmd | CEL Validation |
|------|-----|----------------|
| build ran | echo ok | exitCode == 0 |
| expand exec var | echo $GOOS | stdout.trim() == GOOS |
| expand arg var | echo $GIT_ROOT_DIR | stdout.trim() == GIT_ROOT_DIR |
| expand multiple | echo $GOOS/$GOARCH | stdout.contains("/") |

## CWD resolution logging

| Name | cwd | cmd | CEL Validation |
|------|-----|-----|----------------|
| relative cwd | . | pwd | exitCode == 0 |
| var in cwd | $GIT_ROOT_DIR | pwd | stdout.trim() == GIT_ROOT_DIR |
