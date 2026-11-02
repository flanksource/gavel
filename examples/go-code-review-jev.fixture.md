# Go code review (Jev)

The [Jev](https://docs.typesafe.ai/api) variant of [`go-code-review.fixture.md`](go-code-review.fixture.md). It runs the same 30-rule checklist without an agent: this step collects the change and a per-package inventory, and Jev judges each rule in [`go-code-review.jev.prompt`](go-code-review.jev.prompt) in a single API call, returning `pass`, `fail` or `na` for each rule. It returns verdicts only, with no `path:line` messages. Use the agent variant when you need to know where a violation is. Needs `TYPESAFE_API_KEY`.

    gavel fixtures examples/go-code-review-jev.fixture.md

The change is the uncommitted Go work (tracked diff plus untracked files); only when the working tree has no Go changes is the last commit reviewed instead.

### command: Jev verdicts for every rule

```bash
set -euo pipefail
prompt="$PWD/go-code-review.jev.prompt"
cd "$GIT_ROOT_DIR"

if [ -n "$(git status --porcelain -- '*.go')" ]; then
  files=$(git status --porcelain -- '*.go' | awk '{print $NF}')
  change() {
    echo "## Uncommitted Go changes"
    git diff HEAD -- '*.go'
    git ls-files --others --exclude-standard -- '*.go' | while read -r f; do
      echo "### New file $f"
      cat -n "$f"
    done
  }
else
  files=$(git show --name-only --format= HEAD -- '*.go')
  change() {
    echo "## Last commit"
    git show HEAD -- '*.go'
  }
fi
[ -n "$files" ] || { echo "no Go changes to review" >&2; exit 1; }

inventory() {
  echo "## Package inventory"
  for dir in $(printf '%s\n' $files | xargs -n1 dirname | sort -u); do
    [ -d "$dir" ] || continue
    echo "### $dir"
    wc -l "$dir"/*.go
    # rg exits 1 when a package has no top-level declarations; anything else is an error.
    rg -n '^(func|type|var|const) ' --glob '*.go' --max-depth 1 "$dir" || [ $? -eq 1 ]
  done
}

out=$({ change; inventory; } | captain prompt run "$prompt" -V change=- --json)
printf '%s' "$out" | jq -e '.structuredOutput.value | type == "object"' > /dev/null \
  || { echo "Jev returned no verdicts: $out" >&2; exit 1; }
printf '%s' "$out" | jq '.structuredOutput.value | {failed: [to_entries[] | select(.value == "fail") | .key], verdicts: .}'
```

- cel: size(json.failed) == 0
