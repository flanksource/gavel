---
exec: bash
args: ["-c", "printf '%s\\n' '{{.message}}'"]
message: ready
timeout: 30s
---

# CLI checks

| Name | message | Exit Code | CEL Validation |
| --- | --- | --- | --- |
| Default message | | 0 | stdout.trim() == "ready" |
| Row override | done | 0 | stdout.trim() == "done" |

## JSON contract

```bash
printf '%s\n' '{"status":"ready","count":2}'
```

- cel: json.status == "ready" && json.count == 2.0
