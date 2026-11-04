---
cwd: $GIT_ROOT_DIR
---

# Comments service verification

| Name | Command | Exit Code |
|---|---|---|
| Comments unit and isolated database integration | go run ./cmd/gavel test ginkgo ./comments --concurrency 1 -- -p=1 -race | 0 |
| GitHub HTTP contracts | go run ./cmd/gavel test ginkgo ./github --recursive=false --concurrency 1 -- -p=1 -run '^TestGitHubSuite$' --focus 'comments publication' | 0 |
| Azure HTTP contracts | go run ./cmd/gavel test ginkgo ./azuredevops --recursive=false --concurrency 1 -- -p=1 -run '^TestAzureDevOps$' --focus 'comments publication' | 0 |
| Gavel project authorization | go run ./cmd/gavel test ginkgo ./pr/ui --recursive=false --concurrency 1 -- -p=1 -run '^TestUI$' --focus 'Gavel comments authorization' | 0 |
| Gavel entity catalog | go run ./cmd/gavel test ginkgo ./pr/ui --recursive=false --concurrency 1 -- -p=1 -run '^TestUI$' --focus 'entity catalog' | 0 |
| Gavel CLI registration | go run ./cmd/gavel test ginkgo ./cmd/gavel --recursive=false --concurrency 1 -- -p=1 -run '^TestGavelCLI$' --focus 'comments CLI registration' | 0 |
| Owning repository build | make build | 0 |
| Strict repository lint | golangci-lint run | 0 |
| Published dependency compatibility | env GOWORK=off go test -p=1 ./comments/... | 0 |

These checks execute serially, including Go package execution. The strict lint and published-dependency rows intentionally retain the full delivery gates; known pre-existing lint findings and unpublished dependency changes must remain visible as failures rather than being suppressed.
