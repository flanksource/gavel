#!/usr/bin/env bash
#
# Write a gavel crash envelope for a run that died before serialising results.
#
#   gavel-crash-stub.sh <exit-code> <log-file> <out-json>
#
# The envelope is the same shape gavel itself writes (report.ResultFile):
# {"error": ..., "exit_code": N, "log_tail": ...}. It exists only for the cases
# gavel cannot cover itself — a panic, an OOM kill, a signal — where the process
# never reaches its own output pipeline.
#
# The log tail is embedded via jq so control characters (tabs from `go mod
# tidy`, backslashes from Windows paths, quotes from error messages) are
# escaped. Hand-rolled quoting here previously produced invalid JSON that made
# every downstream reader report "no results" instead of the crash reason.

set -euo pipefail

if [ "$#" -ne 3 ]; then
  echo "usage: $(basename "$0") <exit-code> <log-file> <out-json>" >&2
  exit 2
fi

code="$1"
log_file="$2"
out_json="$3"

case "$code" in
  '' | *[!0-9]*)
    echo "$(basename "$0"): exit code must be numeric, got '$code'" >&2
    exit 2
    ;;
esac

error="gavel exited $code before writing results"

if ! command -v jq >/dev/null 2>&1; then
  # Without jq there is no safe way to embed arbitrary log text, so the tail is
  # omitted rather than risking a malformed document. The message is fixed text
  # and needs no escaping. gavel.log is uploaded with the artifact regardless.
  echo "$(basename "$0"): jq not found; omitting log_tail from ${out_json}" >&2
  printf '{"error":"%s","exit_code":%s}\n' "$error" "$code" > "$out_json"
  exit 0
fi

tail_file="$(mktemp)"
trap 'rm -f "$tail_file"' EXIT
tail -n 200 "$log_file" 2>/dev/null > "$tail_file" || true

jq -n \
  --arg error "$error" \
  --argjson exit_code "$code" \
  --rawfile log_tail "$tail_file" \
  '{error: $error, exit_code: $exit_code, log_tail: $log_tail}' > "$out_json"
