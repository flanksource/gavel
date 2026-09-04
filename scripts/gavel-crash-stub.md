---
cwd: $GIT_ROOT_DIR
---

# gavel-crash-stub.sh

The action's last-resort crash envelope (`scripts/gavel-crash-stub.sh`) is the only
producer of gavel result JSON that is not written by Go. It must stay valid for any
log tail — the previous hand-rolled `printf` + `sed` version emitted raw tabs inside a
JSON string literal, which made every reader report "no results" instead of the crash.

### command: log tail with control characters stays valid JSON

The tail below carries the four bytes that break naive escaping: a tab, a backslash,
a double quote and a newline.

```yaml
exitCode: 0
```

```bash
set -euo pipefail
d="$(mktemp -d)"
trap 'rm -rf "$d"' EXIT
printf 'go: updates to go.mod needed; to update it:\n\tgo mod tidy "C:\\path"\n' > "$d/gavel.log"

scripts/gavel-crash-stub.sh 1 "$d/gavel.log" "$d/out.json"

jq -e . "$d/out.json" > /dev/null
jq -e '.exit_code == 1' "$d/out.json" > /dev/null
jq -e '.error | startswith("gavel exited 1")' "$d/out.json" > /dev/null
jq -e '.log_tail | contains("\tgo mod tidy \"C:\\path\"\n")' "$d/out.json" > /dev/null
```

### command: missing log file still produces a readable envelope

```yaml
exitCode: 0
```

```bash
set -euo pipefail
d="$(mktemp -d)"
trap 'rm -rf "$d"' EXIT

scripts/gavel-crash-stub.sh 137 "$d/absent.log" "$d/out.json"

jq -e '.exit_code == 137' "$d/out.json" > /dev/null
jq -e '.log_tail == ""' "$d/out.json" > /dev/null
```

### command: a non-numeric exit code is rejected

A quoted-but-empty `$code` used to become `{"exit_code":}`. Fail loudly instead.

```yaml
exitCode: 2
```

```bash
d="$(mktemp -d)"
trap 'rm -rf "$d"' EXIT
scripts/gavel-crash-stub.sh "" "$d/absent.log" "$d/out.json"
```
