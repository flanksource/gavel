---
exec: bash
args: ["-c", "{{.cmd}}"]
---

## Shell-style $VAR expansion

Tests that $VAR and ${VAR} syntax works in exec, args, and cwd fields.

| Name | cmd | CEL Validation |
|------|-----|----------------|
| $GOOS is set | echo $GOOS | stdout.trim() == GOOS |
| $GOARCH is set | echo $GOARCH | stdout.trim() == GOARCH |
| ${GIT_ROOT_DIR} braces | echo ${GIT_ROOT_DIR} | stdout.trim() == GIT_ROOT_DIR |
| $ROOT_DIR matches git root | echo $ROOT_DIR | stdout.trim() == GIT_ROOT_DIR |
| unknown $VAR passthrough | bash -c 'echo $HOME' | stdout.trim() != "" |
| multiple vars | echo $GOOS-$GOARCH | stdout.trim() == GOOS + "-" + GOARCH |

## CWD with $VAR

| Name | cwd | cmd | CEL Validation |
|------|-----|-----|----------------|
| cwd uses $GIT_ROOT_DIR | $GIT_ROOT_DIR | pwd | stdout.trim() == GIT_ROOT_DIR |
