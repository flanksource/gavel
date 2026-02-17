---
build: git fetch origin && git checkout pr/fix-terminal
priority: high
status: pending
last_run: 2026-02-13T08:49:13.832629+02:00
attempts: 1
---

# Test / Test (ubuntu-latest, Go 1.25)

PR #81 (`pr/fix-terminal`)
Job: https://github.com/flanksource/clicky/actions/runs/21948453261/job/63392268193

## Logs

```
2026-02-12T13:28:21.8109079Z === RUN   TestOpenAPIValidator_isValidResponseCode/99
2026-02-12T13:28:21.8109219Z === RUN   TestOpenAPIValidator_isValidResponseCode/1234
2026-02-12T13:28:21.8109353Z === RUN   TestOpenAPIValidator_isValidResponseCode/abc
2026-02-12T13:28:21.8109486Z === RUN   TestOpenAPIValidator_isValidResponseCode/#00
2026-02-12T13:28:21.8109639Z --- PASS: TestOpenAPIValidator_isValidResponseCode (0.00s)
2026-02-12T13:28:21.8109860Z     --- PASS: TestOpenAPIValidator_isValidResponseCode/200 (0.00s)
2026-02-12T13:28:21.8110085Z     --- PASS: TestOpenAPIValidator_isValidResponseCode/404 (0.00s)
2026-02-12T13:28:21.8110299Z     --- PASS: TestOpenAPIValidator_isValidResponseCode/500 (0.00s)
2026-02-12T13:28:21.8110516Z     --- PASS: TestOpenAPIValidator_isValidResponseCode/2XX (0.00s)
2026-02-12T13:28:21.8110845Z     --- PASS: TestOpenAPIValidator_isValidResponseCode/4xx (0.00s)
2026-02-12T13:28:21.8111088Z     --- PASS: TestOpenAPIValidator_isValidResponseCode/default (0.00s)
2026-02-12T13:28:21.8111304Z     --- PASS: TestOpenAPIValidator_isValidResponseCode/999 (0.00s)
2026-02-12T13:28:21.8111612Z     --- PASS: TestOpenAPIValidator_isValidResponseCode/99 (0.00s)
2026-02-12T13:28:21.8111850Z     --- PASS: TestOpenAPIValidator_isValidResponseCode/1234 (0.00s)
2026-02-12T13:28:21.8112070Z     --- PASS: TestOpenAPIValidator_isValidResponseCode/abc (0.00s)
2026-02-12T13:28:21.8112325Z     --- PASS: TestOpenAPIValidator_isValidResponseCode/#00 (0.00s)
2026-02-12T13:28:21.8112515Z === RUN   TestOpenAPIValidator_isValidOperationID
2026-02-12T13:28:21.8112683Z === RUN   TestOpenAPIValidator_isValidOperationID/getUserById
2026-02-12T13:28:21.8112832Z === RUN   TestOpenAPIValidator_isValidOperationID/user_create
2026-02-12T13:28:21.8112982Z === RUN   TestOpenAPIValidator_isValidOperationID/list123
2026-02-12T13:28:21.8113250Z === RUN   TestOpenAPIValidator_isValidOperationID/valid_operation_ID
2026-02-12T13:28:21.8113503Z === RUN   TestOpenAPIValidator_isValidOperationID/#00
2026-02-12T13:28:21.8113662Z === RUN   TestOpenAPIValidator_isValidOperationID/user-create
2026-02-12T13:28:21.8113825Z === RUN   TestOpenAPIValidator_isValidOperationID/user_create#01
2026-02-12T13:28:21.8114049Z === RUN   TestOpenAPIValidator_isValidOperationID/user@create
2026-02-12T13:28:21.8114251Z --- PASS: TestOpenAPIValidator_isValidOperationID (0.00s)
2026-02-12T13:28:21.8114504Z     --- PASS: TestOpenAPIValidator_isValidOperationID/getUserById (0.00s)
2026-02-12T13:28:21.8114781Z     --- PASS: TestOpenAPIValidator_isValidOperationID/user_create (0.00s)
2026-02-12T13:28:21.8115227Z     --- PASS: TestOpenAPIValidator_isValidOperationID/list123 (0.00s)
2026-02-12T13:28:21.8115646Z     --- PASS: TestOpenAPIValidator_isValidOperationID/valid_operation_ID (0.00s)
2026-02-12T13:28:21.8115966Z     --- PASS: TestOpenAPIValidator_isValidOperationID/#00 (0.00s)
2026-02-12T13:28:21.8116219Z     --- PASS: TestOpenAPIValidator_isValidOperationID/user-create (0.00s)
2026-02-12T13:28:21.8116537Z     --- PASS: TestOpenAPIValidator_isValidOperationID/user_create#01 (0.00s)
2026-02-12T13:28:21.8116836Z     --- PASS: TestOpenAPIValidator_isValidOperationID/user@create (0.00s)
2026-02-12T13:28:21.8116935Z === RUN   TestValidateRPCService
2026-02-12T13:28:21.8117052Z === RUN   TestValidateRPCService/nil_service
2026-02-12T13:28:21.8117165Z === RUN   TestValidateRPCService/missing_name
2026-02-12T13:28:21.8117329Z === RUN   TestValidateRPCService/missing_version
2026-02-12T13:28:21.8117514Z === RUN   TestValidateRPCService/no_operations
2026-02-12T13:28:21.8117631Z === RUN   TestValidateRPCService/valid_service
2026-02-12T13:28:21.8117762Z === RUN   TestValidateRPCService/invalid_operation
2026-02-12T13:28:21.8117878Z --- PASS: TestValidateRPCService (0.00s)
2026-02-12T13:28:21.8118072Z     --- PASS: TestValidateRPCService/nil_service (0.00s)
2026-02-12T13:28:21.8118261Z     --- PASS: TestValidateRPCService/missing_name (0.00s)
2026-02-12T13:28:21.8118458Z     --- PASS: TestValidateRPCService/missing_version (0.00s)
2026-02-12T13:28:21.8118648Z     --- PASS: TestValidateRPCService/no_operations (0.00s)
2026-02-12T13:28:21.8118836Z     --- PASS: TestValidateRPCService/valid_service (0.00s)
2026-02-12T13:28:21.8119038Z     --- PASS: TestValidateRPCService/invalid_operation (0.00s)
2026-02-12T13:28:21.8119137Z === RUN   TestValidateRPCOperation
2026-02-12T13:28:21.8119263Z === RUN   TestValidateRPCOperation/nil_operation
2026-02-12T13:28:21.8119381Z === RUN   TestValidateRPCOperation/missing_name
2026-02-12T13:28:21.8119519Z === RUN   TestValidateRPCOperation/missing_description
2026-02-12T13:28:21.8119655Z === RUN   TestValidateRPCOperation/invalid_HTTP_method
2026-02-12T13:28:21.8119783Z === RUN   TestValidateRPCOperation/invalid_path_format
2026-02-12T13:28:21.8119914Z === RUN   TestValidateRPCOperation/valid_operation
2026-02-12T13:28:21.8120133Z === RUN   TestValidateRPCOperation/valid_operation_with_case_insensitive_method
2026-02-12T13:28:21.8120247Z --- PASS: TestValidateRPCOperation (0.00s)
2026-02-12T13:28:21.8120453Z     --- PASS: TestValidateRPCOperation/nil_operation (0.00s)
2026-02-12T13:28:21.8120646Z     --- PASS: TestValidateRPCOperation/missing_name (0.00s)
2026-02-12T13:28:21.8120872Z     --- PASS: TestValidateRPCOperation/missing_description (0.00s)
2026-02-12T13:28:21.8121089Z     --- PASS: TestValidateRPCOperation/invalid_HTTP_method (0.00s)
2026-02-12T13:28:21.8121302Z     --- PASS: TestValidateRPCOperation/invalid_path_format (0.00s)
2026-02-12T13:28:21.8121510Z     --- PASS: TestValidateRPCOperation/valid_operation (0.00s)
2026-02-12T13:28:21.8121843Z     --- PASS: TestValidateRPCOperation/valid_operation_with_case_insensitive_method (0.00s)
2026-02-12T13:28:21.8121916Z PASS
2026-02-12T13:28:21.8122044Z ok  	github.com/flanksource/clicky/rpc	4.307s
2026-02-12T13:28:21.8122195Z ?   	github.com/flanksource/clicky/shutdown	[no test files]
2026-02-12T13:28:21.8122342Z FAIL	github.com/flanksource/clicky/task [build failed]
2026-02-12T13:28:21.8122431Z === RUN   TestLineProcessor
2026-02-12T13:28:21.8122636Z Running Suite: LineProcessor Suite - /home/runner/work/clicky/clicky/text
2026-02-12T13:28:21.8122755Z =========================================================================
2026-02-12T13:28:21.8123152Z Random Seed: 1770902891
2026-02-12T13:28:21.8123165Z 
2026-02-12T13:28:21.8123331Z Will run 78 of 78 specs
2026-02-12T13:28:21.8127291Z ••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••
2026-02-12T13:28:21.8127467Z 
2026-02-12T13:28:21.8127742Z Ran 78 of 78 Specs in 0.309 seconds
2026-02-12T13:28:21.8128333Z SUCCESS! -- 78 Passed | 0 Failed | 0 Pending | 0 Skipped
2026-02-12T13:28:21.8128474Z --- PASS: TestLineProcessor (0.31s)
2026-02-12T13:28:21.8128542Z PASS
2026-02-12T13:28:21.8128668Z ok  	github.com/flanksource/clicky/text	0.367s
2026-02-12T13:28:21.8128739Z FAIL
2026-02-12T13:28:21.9654227Z make: *** [Makefile:12: test] Error 1
2026-02-12T13:28:21.9662486Z ##[error]Process completed with exit code 2.
2026-02-12T13:28:21.9764803Z Post job cleanup.
2026-02-12T13:28:22.0690190Z [command]/usr/bin/git version
2026-02-12T13:28:22.0726192Z git version 2.52.0
2026-02-12T13:28:22.0775486Z Temporarily overriding HOME='/home/runner/work/_temp/5b8eb6f7-2f29-42b0-9bdd-fea364ff7d6a' before making global git config changes
2026-02-12T13:28:22.0777081Z Adding repository directory to the temporary git global config as a safe directory
2026-02-12T13:28:22.0781654Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/clicky/clicky
2026-02-12T13:28:22.0815247Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-12T13:28:22.0846932Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-12T13:28:22.1070971Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-12T13:28:22.1090583Z http.https://github.com/.extraheader
2026-02-12T13:28:22.1102832Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-02-12T13:28:22.1132063Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-12T13:28:22.1349982Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-12T13:28:22.1378819Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-12T13:28:22.1698049Z Cleaning up orphan processes
```

## Workflow Definition

```yaml
name: Test

on:
  push:
    branches: [main, master, develop]
  pull_request:
    types: [opened, synchronize, reopened]
  workflow_dispatch:

env:
  GO_VERSION: "1.25"

jobs:
  test:
    name: Test (${{ matrix.os }}, Go ${{ matrix.go-version }})
    runs-on: ${{ matrix.os }}
    strategy:
      fail-fast: false
      matrix:
        os: [ubuntu-latest]
        go-version: ["1.25"]

    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}
          cache: true

      - name: Install dependencies
        run: |
          go mod download
          go mod tidy

      - name: Verify dependencies
        run: go mod verify

      - name: Run tests
        run: make test
        env:
          CGO_ENABLED: 1

  build:
    name: Build
    runs-on: ubuntu-latest
    needs: test

    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Build binary
        run: make build

      - name: Test binary
        run: |
          ./clicky --help
          ./clicky --version || echo "Version command may not be implemented yet"

      - name: Upload binary artifact
        uses: actions/upload-artifact@v4
        with:
          name: clicky-linux-amd64
          path: clicky
```
