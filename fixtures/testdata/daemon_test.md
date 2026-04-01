---
daemon: python3 -m http.server {{.port}}
exec: curl
args: [-sf]
---

## HTTP Server Tests

| name             | args                                  | expected       |
|------------------|---------------------------------------|----------------|
| root returns 200 | http://localhost:{{.port}}/            | exitCode == 0  |
| 404 fails        | http://localhost:{{.port}}/nonexistent | exitCode != 0  |
