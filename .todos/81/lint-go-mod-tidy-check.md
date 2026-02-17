---
build: git fetch origin && git checkout pr/fix-terminal
priority: high
status: pending
last_run: 2026-02-13T08:49:13.831514+02:00
attempts: 1
---

# Lint / Go Mod Tidy Check

PR #81 (`pr/fix-terminal`)
Job: https://github.com/flanksource/clicky/actions/runs/21948453289/job/63392268240

## Logs

```
2026-02-12T13:26:11.0067073Z go: downloading github.com/ugorji/go/codec v1.2.12
2026-02-12T13:26:11.0282127Z go: downloading github.com/gosimple/unidecode v1.0.1
2026-02-12T13:26:11.0628127Z go: downloading github.com/ghodss/yaml v1.0.0
2026-02-12T13:26:11.0689145Z go: downloading github.com/tidwall/sjson v1.2.5
2026-02-12T13:26:11.0705527Z go: downloading k8s.io/klog/v2 v2.130.1
2026-02-12T13:26:11.0742873Z go: downloading sigs.k8s.io/randfill v1.0.0
2026-02-12T13:26:11.0813962Z go: downloading github.com/gogo/protobuf v1.3.2
2026-02-12T13:26:11.0825923Z go: downloading sigs.k8s.io/structured-merge-diff/v6 v6.3.0
2026-02-12T13:26:11.0907641Z go: downloading gopkg.in/sourcemap.v1 v1.0.5
2026-02-12T13:26:11.1293285Z go: downloading github.com/shirou/gopsutil/v3 v3.24.5
2026-02-12T13:26:11.1324451Z go: downloading github.com/google/pprof v0.0.0-20250403155104-27863c87afa6
2026-02-12T13:26:11.2030094Z go: downloading github.com/gobwas/glob v0.2.3
2026-02-12T13:26:11.2224625Z go: downloading github.com/yuin/gopher-lua v1.1.1
2026-02-12T13:26:11.2522555Z go: downloading gopkg.in/yaml.v2 v2.4.0
2026-02-12T13:26:11.2734322Z go: downloading layeh.com/gopher-json v0.0.0-20201124131017-552bb3c4c3bf
2026-02-12T13:26:11.2851111Z go: downloading github.com/cert-manager/cert-manager v1.16.1
2026-02-12T13:26:11.3684734Z go: downloading github.com/robfig/cron/v3 v3.0.1
2026-02-12T13:26:11.3690790Z go: downloading github.com/bmatcuk/doublestar/v4 v4.8.1
2026-02-12T13:26:11.3832314Z go: downloading github.com/tidwall/match v1.1.1
2026-02-12T13:26:11.3944857Z go: downloading github.com/distribution/reference v0.6.0
2026-02-12T13:26:11.3954305Z go: downloading github.com/jeremywohl/flatten v0.0.0-20180923035001-588fe0d4c603
2026-02-12T13:26:11.4085616Z go: downloading github.com/sirupsen/logrus v1.9.3
2026-02-12T13:26:11.4291802Z go: downloading k8s.io/client-go v0.34.1
2026-02-12T13:26:11.4931960Z go: downloading k8s.io/utils v0.0.0-20251002143259-bc988d571ff4
2026-02-12T13:26:11.5335241Z go: downloading sigs.k8s.io/json v0.0.0-20241014173422-cfa47c3a1cc8
2026-02-12T13:26:11.5579195Z go: downloading github.com/fxamacker/cbor/v2 v2.9.0
2026-02-12T13:26:11.5838319Z go: downloading github.com/jmespath/go-jmespath/internal/testify v1.5.1
2026-02-12T13:26:11.6069140Z go: downloading gopkg.in/inf.v0 v0.9.1
2026-02-12T13:26:11.6202051Z go: downloading github.com/json-iterator/go v1.1.12
2026-02-12T13:26:11.6662739Z go: downloading github.com/tklauser/go-sysconf v0.3.12
2026-02-12T13:26:11.6894910Z go: downloading golang.org/x/mod v0.29.0
2026-02-12T13:26:11.7287557Z go: downloading github.com/prashantv/gostub v1.1.0
2026-02-12T13:26:11.7567427Z go: downloading github.com/opencontainers/go-digest v1.0.0
2026-02-12T13:26:11.7789642Z go: downloading github.com/x448/float16 v0.8.4
2026-02-12T13:26:11.7913258Z go: downloading github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd
2026-02-12T13:26:11.8059770Z go: downloading github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee
2026-02-12T13:26:11.8213415Z go: downloading github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0
2026-02-12T13:26:11.8384734Z go: downloading github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c
2026-02-12T13:26:11.8571726Z go: downloading github.com/shoenig/go-m1cpu v0.1.6
2026-02-12T13:26:11.8707817Z go: downloading github.com/yusufpapurcu/wmi v1.2.4
2026-02-12T13:26:11.8838240Z go: downloading github.com/tklauser/numcpus v0.6.1
2026-02-12T13:26:11.9036489Z go: downloading github.com/shoenig/test v0.6.4
2026-02-12T13:26:11.9375956Z go: downloading github.com/go-ole/go-ole v1.2.6
2026-02-12T13:26:12.0012428Z go: downloading sigs.k8s.io/gateway-api v1.1.0
2026-02-12T13:26:12.0015581Z go: downloading k8s.io/apiextensions-apiserver v0.31.1
2026-02-12T13:26:12.2904898Z go mod tidy resulted in changes:
2026-02-12T13:26:12.2928139Z diff --git a/go.mod b/go.mod
2026-02-12T13:26:12.2928558Z index 94e7328..10fc557 100644
2026-02-12T13:26:12.2928797Z --- a/go.mod
2026-02-12T13:26:12.2928999Z +++ b/go.mod
2026-02-12T13:26:12.2929191Z @@ -7,6 +7,7 @@ require (
2026-02-12T13:26:12.2929453Z  	github.com/alecthomas/chroma/v2 v2.20.0
2026-02-12T13:26:12.2929798Z  	github.com/charmbracelet/bubbletea v1.3.10
2026-02-12T13:26:12.2930111Z  	github.com/charmbracelet/lipgloss v1.1.0
2026-02-12T13:26:12.2930388Z +	github.com/creack/pty v1.1.24
2026-02-12T13:26:12.2930737Z  	github.com/flanksource/commons v1.42.3
2026-02-12T13:26:12.2931209Z  	github.com/flanksource/gomplate/v3 v3.24.60
2026-02-12T13:26:12.2931558Z  	github.com/go-xmlfmt/xmlfmt v1.1.3
2026-02-12T13:26:12.2931808Z @@ -30,6 +31,7 @@ require (
2026-02-12T13:26:12.2932049Z  	github.com/xuri/excelize/v2 v2.9.1
2026-02-12T13:26:12.2932307Z  	golang.org/x/crypto v0.43.0
2026-02-12T13:26:12.2932548Z  	golang.org/x/sync v0.17.0
2026-02-12T13:26:12.2932762Z +	golang.org/x/sys v0.37.0
2026-02-12T13:26:12.2933131Z  	golang.org/x/term v0.36.0
2026-02-12T13:26:12.2933545Z  	golang.org/x/text v0.30.0
2026-02-12T13:26:12.2933933Z  	golang.org/x/time v0.13.0
2026-02-12T13:26:12.2934295Z @@ -50,7 +52,6 @@ require (
2026-02-12T13:26:12.2934773Z  	github.com/charmbracelet/x/ansi v0.10.1 // indirect
2026-02-12T13:26:12.2935349Z  	github.com/charmbracelet/x/cellbuf v0.0.13-0.20250311204145-2c3ea96c31dd // indirect
2026-02-12T13:26:12.2935813Z  	github.com/charmbracelet/x/term v0.2.1 // indirect
2026-02-12T13:26:12.2936124Z -	github.com/creack/pty v1.1.24 // indirect
2026-02-12T13:26:12.2936546Z  	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
2026-02-12T13:26:12.2937221Z  	github.com/deckarep/golang-set/v2 v2.6.0 // indirect
2026-02-12T13:26:12.2937595Z  	github.com/distribution/reference v0.6.0 // indirect
2026-02-12T13:26:12.2937879Z @@ -158,7 +159,6 @@ require (
2026-02-12T13:26:12.2938148Z  	golang.org/x/mod v0.29.0 // indirect
2026-02-12T13:26:12.2938407Z  	golang.org/x/net v0.46.0 // indirect
2026-02-12T13:26:12.2938674Z  	golang.org/x/oauth2 v0.32.0 // indirect
2026-02-12T13:26:12.2938932Z -	golang.org/x/sys v0.37.0 // indirect
2026-02-12T13:26:12.2939193Z  	golang.org/x/tools v0.38.0 // indirect
2026-02-12T13:26:12.2939672Z  	google.golang.org/genproto/googleapis/api v0.0.0-20250825161204-c5933d9347a5 // indirect
2026-02-12T13:26:12.2940278Z  	google.golang.org/genproto/googleapis/rpc v0.0.0-20251002232023-7c0ddcbb5797 // indirect
2026-02-12T13:26:12.2940722Z Please run 'go mod tidy' and commit the changes
2026-02-12T13:26:12.2954727Z ##[error]Process completed with exit code 1.
2026-02-12T13:26:12.3061953Z Post job cleanup.
2026-02-12T13:26:12.4022901Z [command]/usr/bin/git version
2026-02-12T13:26:12.4060325Z git version 2.52.0
2026-02-12T13:26:12.4112627Z Temporarily overriding HOME='/home/runner/work/_temp/1ffb0279-851b-4f02-81f9-ad3a58bc0514' before making global git config changes
2026-02-12T13:26:12.4114035Z Adding repository directory to the temporary git global config as a safe directory
2026-02-12T13:26:12.4117880Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/clicky/clicky
2026-02-12T13:26:12.4153848Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-12T13:26:12.4186022Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-12T13:26:12.4415913Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-12T13:26:12.4436178Z http.https://github.com/.extraheader
2026-02-12T13:26:12.4448491Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-02-12T13:26:12.4479120Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-12T13:26:12.4703050Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-12T13:26:12.4733791Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-12T13:26:12.5061635Z Cleaning up orphan processes
```

## Workflow Steps

```yaml
- name: Checkout code
  uses: actions/checkout@v4
- name: Set up Go
  uses: actions/setup-go@v5
  with:
    cache: true
    go-version: ${{ env.GO_VERSION }}
- name: Check go mod tidy
  run: |
    go mod tidy
    if [ -n "$(git status --porcelain)" ]; then
      echo "go mod tidy resulted in changes:"
      git diff
      echo "Please run 'go mod tidy' and commit the changes"
      exit 1
    fi
```
