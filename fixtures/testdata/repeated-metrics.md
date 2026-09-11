---
exec: bash
args:
  - -c
  - |
    state="/tmp/gavel-repeated-metrics-$PPID-{{ .key }}"
    sample=$(( $(cat "$state" 2>/dev/null || echo 0) + 1 ))
    printf '{"schemaVersion":"captain.observation/v1","execution":{"state":"completed"},"metrics":{"durationMs":{"state":"known","value":%s,"unit":"ms"}}}\n' "$sample"
    if [ "$sample" -ge "{{ .samplecount }}" ]; then rm -f "$state"; else printf '%s\n' "$sample" > "$state"; fi
repeat: 3
timeout: 30s
metrics:
  - name: mean_value
    extract: json.metrics.durationMs.value
    unit: ms
    aggregate: mean
    direction: none
    threshold:
      min: 1
      max: 10
  - name: median_value
    extract: json.metrics.durationMs.value
    unit: ms
    aggregate: median
    direction: none
  - name: min_value
    extract: json.metrics.durationMs.value
    unit: ms
    aggregate: min
    direction: none
  - name: max_value
    extract: json.metrics.durationMs.value
    unit: ms
    aggregate: max
    direction: none
  - name: p95_value
    extract: json.metrics.durationMs.value
    unit: ms
    aggregate: p95
    direction: none
  - name: lower_regression
    extract: json.metrics.durationMs.value
    unit: ms
    aggregate: median
    direction: lower
    baseline: API baseline
    threshold:
      regressionPercent: 60
  - name: higher_regression
    extract: json.metrics.durationMs.value
    unit: ms
    aggregate: median
    direction: higher
    baseline: API baseline
    threshold:
      regressionPercent: 60
---

# Generic repeated metrics

| Name | key | sampleCount | Repeat | CEL Validation |
|---|---|---:|---:|---|
| API baseline | baseline | 3 | | json.schemaVersion == "captain.observation/v1" && json.execution.state == "completed" |
| CLI candidate | candidate | 5 | 5 | |

### command: Command frontmatter repeat

```yaml
repeat: 2
metrics:
  - name: command_value
    extract: json.metrics.durationMs.value
    unit: ms
```

```bash
printf '{"schemaVersion":"captain.observation/v1","execution":{"state":"completed"},"metrics":{"durationMs":{"state":"known","value":7,"unit":"ms"}}}\n'
```
