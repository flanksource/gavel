---
cwd: $GIT_ROOT_DIR
timeout: 10m
---

# Azure DevOps PR support

All service traffic is tested with local HTTP servers; creation pushes to local bare Git repositories. No live PR is created.

| Name | Command | Exit Code | CEL |
| --- | --- | --- | --- |
| Azure client and shared GitHub regressions | go test -count=1 ./azuredevops ./pr/provider ./pr/create ./github ./prwatch | 0 | exitCode == 0 && stdout.contains("gavel/azuredevops") && stdout.contains("gavel/pr/create") |
| Commit push and early unsupported-flag validation | go test -count=1 -v ./commit -run TestCommit -ginkgo.v -ginkgo.focus 'push' | 0 | exitCode == 0 && stdout.contains("Azure commit push validation") |
| Azure status targets and unsupported flags | go test -count=1 -v ./cmd/gavel -run TestGavelCLI -ginkgo.v -ginkgo.focus 'Azure PR status targets' | 0 | exitCode == 0 && stdout.contains("Azure PR status targets") |
| GitHub PR reference regression | go test -count=1 -v ./cmd/gavel -run TestGavelCLI -ginkgo.v -ginkgo.focus 'parsePRRef' | 0 | exitCode == 0 && !stdout.contains("Will run 0 of") |
| Project branch creation API | go test -count=1 -v ./pr/ui -run TestUI -ginkgo.v -ginkgo.focus 'project branch landing' | 0 | exitCode == 0 && stdout.contains("project branch landing") |
| Project Git OpenAPI | go test -count=1 -v ./pr/ui -run TestUI -ginkgo.v -ginkgo.focus 'project git OpenAPI' | 0 | exitCode == 0 && stdout.contains("project git OpenAPI") |
| Required lint gate | make lint | 0 | exitCode == 0 && !stdout.contains("issues:") && !stderr.contains("issues:") |
| Required build gate | make build | 0 | exitCode == 0 |
