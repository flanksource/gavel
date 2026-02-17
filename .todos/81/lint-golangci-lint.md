---
build: git fetch origin && git checkout pr/fix-terminal
priority: high
status: pending
last_run: 2026-02-13T08:49:13.832227+02:00
attempts: 1
---

# Lint / golangci-lint

PR #81 (`pr/fix-terminal`)
Job: https://github.com/flanksource/clicky/actions/runs/21948453289/job/63392268266

## Logs

```
2026-02-12T13:26:07.5269144Z GOPROXY='https://proxy.golang.org,direct'
2026-02-12T13:26:07.5269709Z GOROOT='/opt/hostedtoolcache/go/1.25.6/x64'
2026-02-12T13:26:07.5270198Z GOSUMDB='sum.golang.org'
2026-02-12T13:26:07.5270582Z GOTELEMETRY='local'
2026-02-12T13:26:07.5271050Z GOTELEMETRYDIR='/home/runner/.config/go/telemetry'
2026-02-12T13:26:07.5271562Z GOTMPDIR=''
2026-02-12T13:26:07.5271885Z GOTOOLCHAIN='auto'
2026-02-12T13:26:07.5272397Z GOTOOLDIR='/opt/hostedtoolcache/go/1.25.6/x64/pkg/tool/linux_amd64'
2026-02-12T13:26:07.5272983Z GOVCS=''
2026-02-12T13:26:07.5273539Z GOVERSION='go1.25.6'
2026-02-12T13:26:07.5273884Z GOWORK=''
2026-02-12T13:26:07.5274191Z PKG_CONFIG='pkg-config'
2026-02-12T13:26:07.5274415Z 
2026-02-12T13:26:07.5274883Z ##[endgroup]
2026-02-12T13:26:07.5403579Z ##[group]Run golangci/golangci-lint-action@v9
2026-02-12T13:26:07.5403873Z with:
2026-02-12T13:26:07.5404044Z   version: v2.4.0
2026-02-12T13:26:07.5404282Z   args: --timeout=10m --tests=false --verbose
2026-02-12T13:26:07.5404547Z   skip-cache: true
2026-02-12T13:26:07.5404740Z   install-mode: binary
2026-02-12T13:26:07.5404928Z   install-only: false
2026-02-12T13:26:07.5405263Z   github-token: ***
2026-02-12T13:26:07.5405452Z   verify: true
2026-02-12T13:26:07.5405640Z   only-new-issues: false
2026-02-12T13:26:07.5405849Z   skip-save-cache: false
2026-02-12T13:26:07.5406067Z   cache-invalidation-interval: 7
2026-02-12T13:26:07.5406307Z   problem-matchers: false
2026-02-12T13:26:07.5406503Z env:
2026-02-12T13:26:07.5406677Z   GO_VERSION: 1.25
2026-02-12T13:26:07.5406859Z ##[endgroup]
2026-02-12T13:26:07.6812204Z ##[group]Restore cache
2026-02-12T13:26:07.6814842Z Skipping cache restoration
2026-02-12T13:26:07.6817277Z (node:2245) [DEP0040] DeprecationWarning: The `punycode` module is deprecated. Please use a userland alternative instead.
2026-02-12T13:26:07.6818076Z (Use `node --trace-deprecation ...` to show where the warning was created)
2026-02-12T13:26:07.6819081Z ##[endgroup]
2026-02-12T13:26:07.6819471Z ##[group]Install
2026-02-12T13:26:07.6821499Z Finding needed golangci-lint version...
2026-02-12T13:26:07.6824600Z Installation mode: binary
2026-02-12T13:26:07.6826028Z Installing golangci-lint binary v2.4.0...
2026-02-12T13:26:07.6827566Z Downloading binary https://github.com/golangci/golangci-lint/releases/download/v2.4.0/golangci-lint-2.4.0-linux-amd64.tar.gz ...
2026-02-12T13:26:08.1320739Z [command]/usr/bin/tar xz --overwrite --warning=no-unknown-keyword --overwrite -C /home/runner -f /home/runner/work/_temp/439a8136-b4eb-48e2-9cb6-8efb0f9e0c62
2026-02-12T13:26:08.3764765Z Installed golangci-lint into /home/runner/golangci-lint-2.4.0-linux-amd64/golangci-lint in 694ms
2026-02-12T13:26:08.3768897Z ##[endgroup]
2026-02-12T13:26:08.3773087Z ##[group]run golangci-lint
2026-02-12T13:26:08.3779578Z Running [/home/runner/golangci-lint-2.4.0-linux-amd64/golangci-lint config path] in [/home/runner/work/clicky/clicky] ...
2026-02-12T13:26:08.4849895Z Running [/home/runner/golangci-lint-2.4.0-linux-amd64/golangci-lint run  --timeout=10m --tests=false --verbose] in [/home/runner/work/clicky/clicky] ...
2026-02-12T13:28:04.8128784Z ##[error]task/manager_output.go:105:24: Error return value of `tm.stdoutWriter.Close` is not checked (errcheck)
2026-02-12T13:28:04.8139388Z 		tm.stdoutWriter.Close()
2026-02-12T13:28:04.8139943Z 		                     ^
2026-02-12T13:28:04.8141429Z ##[error]task/manager_output.go:108:24: Error return value of `tm.stderrWriter.Close` is not checked (errcheck)
2026-02-12T13:28:04.8143455Z 		tm.stderrWriter.Close()
2026-02-12T13:28:04.8143962Z 		                     ^
2026-02-12T13:28:04.8145222Z ##[error]task/manager_output.go:117:24: Error return value of `tm.stdoutReader.Close` is not checked (errcheck)
2026-02-12T13:28:04.8146701Z 		tm.stdoutReader.Close()
2026-02-12T13:28:04.8147220Z 		                     ^
2026-02-12T13:28:04.8148611Z ##[error]task/manager_output.go:120:24: Error return value of `tm.stderrReader.Close` is not checked (errcheck)
2026-02-12T13:28:04.8150162Z 		tm.stderrReader.Close()
2026-02-12T13:28:04.8150646Z 		                     ^
2026-02-12T13:28:04.8150978Z 4 issues:
2026-02-12T13:28:04.8151259Z * errcheck: 4
2026-02-12T13:28:04.8151433Z 
2026-02-12T13:28:04.8151958Z level=info msg="golangci-lint has version 2.4.0 built with go1.25.0 from 43d03392 on 2025-08-13T23:36:29Z"
2026-02-12T13:28:04.8153465Z level=info msg="[config_reader] Config search paths: [./ /home/runner/work/clicky/clicky /home/runner/work/clicky /home/runner/work /home/runner /home /]"
2026-02-12T13:28:04.8154792Z level=info msg="[config_reader] Module name \"github.com/flanksource/clicky\""
2026-02-12T13:28:04.8155617Z level=info msg="maxprocs: Leaving GOMAXPROCS=4: CPU quota undefined"
2026-02-12T13:28:04.8157241Z level=info msg="[goenv] Read go env for 3.265287ms: map[string]string{\"GOCACHE\":\"/home/runner/.cache/go-build\", \"GOROOT\":\"/opt/hostedtoolcache/go/1.25.6/x64\"}"
2026-02-12T13:28:04.8158953Z level=info msg="[lintersdb] Active 5 linters: [errcheck govet ineffassign staticcheck unused]"
2026-02-12T13:28:04.8160349Z level=info msg="[loader] Go packages loading at mode 8767 (compiled_files|deps|exports_file|name|types_sizes|files|imports) took 1m36.574125968s"
2026-02-12T13:28:04.8161638Z level=info msg="[runner/filename_unadjuster] Pre-built 0 adjustments in 26.606716ms"
2026-02-12T13:28:04.8164122Z level=info msg="[linters_context/goanalysis] analyzers took 55.483528379s with top 10 stages: buildir: 46.694188207s, inspect: 1.600141368s, printf: 960.951656ms, ctrlflow: 851.864867ms, fact_deprecated: 827.164872ms, fact_purity: 750.80548ms, nilness: 662.938463ms, SA5012: 438.081587ms, typedness: 301.80511ms, S1038: 258.667759ms"
2026-02-12T13:28:04.8168680Z level=info msg="[runner] Processors filtering stat (in/out): exclusion_paths: 4/4, max_same_issues: 4/4, source_code: 4/4, severity-rules: 4/4, cgo: 4/4, diff: 4/4, path_shortener: 4/4, path_prettifier: 4/4, invalid_issue: 4/4, max_from_linter: 4/4, sort_results: 4/4, generated_file_filter: 4/4, exclusion_rules: 4/4, nolint_filter: 4/4, fixer: 4/4, uniq_by_line: 4/4, max_per_file_from_linter: 4/4, path_absoluter: 4/4, filename_unadjuster: 4/4, path_relativity: 4/4"
2026-02-12T13:28:04.8174885Z level=info msg="[runner] processing took 378.986µs with stages: nolint_filter: 297.335µs, source_code: 32.139µs, generated_file_filter: 28.864µs, path_relativity: 3.957µs, uniq_by_line: 3.606µs, max_same_issues: 3.397µs, path_shortener: 1.854µs, sort_results: 1.762µs, max_from_linter: 881ns, exclusion_paths: 722ns, cgo: 651ns, invalid_issue: 591ns, fixer: 591ns, path_absoluter: 571ns, max_per_file_from_linter: 452ns, filename_unadjuster: 411ns, diff: 351ns, path_prettifier: 341ns, exclusion_rules: 330ns, severity-rules: 180ns"
2026-02-12T13:28:04.8178234Z level=info msg="[runner] linters took 19.529933368s with stages: goanalysis_metalinter: 19.529465524s"
2026-02-12T13:28:04.8179330Z level=info msg="File cache stats: 1 entries of total size 2.9KiB"
2026-02-12T13:28:04.8180005Z level=info msg="Memory: 1147 samples, avg is 166.7MB, max is 1340.0MB"
2026-02-12T13:28:04.8180602Z level=info msg="Execution took 1m56.134386317s"
2026-02-12T13:28:04.8180804Z 
2026-02-12T13:28:04.8189573Z ##[error]issues found
2026-02-12T13:28:04.8190348Z Ran golangci-lint in 116324ms
2026-02-12T13:28:04.8190760Z ##[endgroup]
2026-02-12T13:28:04.8290329Z Post job cleanup.
2026-02-12T13:28:04.9765580Z Skipping cache saving
2026-02-12T13:28:04.9769040Z (node:6855) [DEP0040] DeprecationWarning: The `punycode` module is deprecated. Please use a userland alternative instead.
2026-02-12T13:28:04.9770657Z (Use `node --trace-deprecation ...` to show where the warning was created)
2026-02-12T13:28:04.9934229Z Post job cleanup.
2026-02-12T13:28:05.0898753Z [command]/usr/bin/git version
2026-02-12T13:28:05.0935230Z git version 2.52.0
2026-02-12T13:28:05.0987350Z Temporarily overriding HOME='/home/runner/work/_temp/02f86369-3145-4dc5-91a2-7b73461d55e6' before making global git config changes
2026-02-12T13:28:05.0988956Z Adding repository directory to the temporary git global config as a safe directory
2026-02-12T13:28:05.0993840Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/clicky/clicky
2026-02-12T13:28:05.1030159Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-12T13:28:05.1063361Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-12T13:28:05.1327687Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-12T13:28:05.1353847Z http.https://github.com/.extraheader
2026-02-12T13:28:05.1367540Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-02-12T13:28:05.1403061Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-12T13:28:05.1656875Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-12T13:28:05.1695735Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-12T13:28:05.2073379Z Cleaning up orphan processes
```

## Workflow Steps

```yaml
- name: Checkout code
  uses: actions/checkout@v4
  with:
    fetch-depth: 0
- name: Set up Go
  uses: actions/setup-go@v5
  with:
    cache: true
    go-version: ${{ env.GO_VERSION }}
- name: golangci-lint
  uses: golangci/golangci-lint-action@v9
  with:
    args: --timeout=10m --tests=false --verbose
    skip-cache: true
    version: v2.4.0
```
