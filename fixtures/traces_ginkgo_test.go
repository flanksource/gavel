package fixtures_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/clicky/task"

	"github.com/flanksource/gavel/blob"
	"github.com/flanksource/gavel/fixtures"
	_ "github.com/flanksource/gavel/fixtures/kinds/services"
	_ "github.com/flanksource/gavel/fixtures/types"
	"github.com/flanksource/gavel/testrunner/parsers"
)

// traceServer answers every path with its own name, so a trace's requests
// say which fixture made them.
func traceServer() *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, r.URL.Path)
	}))
	DeferCleanup(server.Close)
	return server
}

// runTraced writes body as fixture file name in its own directory and runs it
// through a Runner, returning each leaf result by its fixture name.
func runTraced(name, body string, configure func(*fixtures.RunnerOptions)) map[string]fixtures.FixtureResult {
	dir := GinkgoT().TempDir()
	path := filepath.Join(dir, name)
	Expect(os.WriteFile(path, []byte(strings.TrimSpace(body)+"\n"), 0o644)).To(Succeed())
	options := fixtures.RunnerOptions{Paths: []string{path}, WorkDir: dir}
	if configure != nil {
		configure(&options)
	}
	runner, err := fixtures.NewRunner(options)
	Expect(err).ToNot(HaveOccurred())
	tree, _ := runner.Run()
	Expect(tree).ToNot(BeNil())
	results := map[string]fixtures.FixtureResult{}
	tree.Walk(func(node *fixtures.FixtureNode) {
		if node.Test != nil && node.Results != nil {
			results[node.Test.Name] = *node.Results
		}
	})
	return results
}

func requestPaths(result fixtures.FixtureResult) []string {
	Expect(result.Traces).To(HaveLen(1))
	requests, ok := result.Traces[0].Summary["requests"].([]any)
	Expect(ok).To(BeTrue(), "requests is %T", result.Traces[0].Summary["requests"])
	var paths []string
	for _, request := range requests {
		paths = append(paths, request.(map[string]any)["path"].(string))
	}
	return paths
}

var _ = Describe("traces: arming", func() {
	It("traces the HTTP calls a services step's child makes through a test-scope http_proxy", func() {
		server := traceServer()
		results := runTraced("services.fixture.yaml", fmt.Sprintf(`
kind: Fixture
steps:
  - id: stack
    kind: services
    traces: [{kind: http_proxy}]
    cel: http_proxy.entries > 0
    services:
      caller:
        exec: curl -sS -o /dev/null %s/from-service && touch called && exec sleep 60
        ready: {exec: test -f called}
`, server.URL), nil)

		result := results["stack"]
		Expect(result.Status).To(Equal(task.StatusPASS), result.Error)
		Expect(result.Traces).To(ConsistOf(HaveField("Count", 1)))
		Expect(requestPaths(result)).To(Equal([]string{"/from-service"}))
	})

	It("keeps two concurrent rows' test-scope captures apart", func() {
		server := traceServer()
		results := runTraced("rows.fixture.md", fmt.Sprintf(`
# Rows

| Name  | Command                                                  | Traces     | CEL                      |
|-------|----------------------------------------------------------|------------|--------------------------|
| alpha | curl -sS -o /dev/null %[1]s/alpha && sleep 1             | http_proxy | http_proxy.entries == 1  |
| beta  | curl -sS -o /dev/null %[1]s/beta && sleep 1              | http_proxy | http_proxy.entries == 1  |
`, server.URL), func(options *fixtures.RunnerOptions) { options.MaxWorkers = 2 })

		Expect(results["alpha"].Status).To(Equal(task.StatusPASS), results["alpha"].Error)
		Expect(results["beta"].Status).To(Equal(task.StatusPASS), results["beta"].Error)
		Expect(requestPaths(results["alpha"])).To(Equal([]string{"/alpha"}))
		Expect(requestPaths(results["beta"])).To(Equal([]string{"/beta"}))
		Expect(results["alpha"].Traces[0].Stream).ToNot(Equal(results["beta"].Traces[0].Stream))
	})

	It("renders a test-scope trace's templated params against the fixture's env when it starts", func() {
		server := traceServer()
		results := runTraced("templated.fixture.md", fmt.Sprintf(`
---
env:
  TRACE_PATH: /keep
traces:
  - kind: http_proxy
    params: {paths: ["{{.env.TRACE_PATH}}"]}
---

# Templated

| Name     | Command                                                                  | CEL                     |
|----------|--------------------------------------------------------------------------|-------------------------|
| filtered | curl -sS -o /dev/null %[1]s/drop && curl -sS -o /dev/null %[1]s/keep     | http_proxy.entries == 1 |
`, server.URL), nil)

		Expect(results["filtered"].Status).To(Equal(task.StatusPASS), results["filtered"].Error)
		Expect(requestPaths(results["filtered"])).To(Equal([]string{"/keep"}))
	})

	It("fails a fixture whose trace's require does not hold, though its own assertions passed", func() {
		results := runTraced("require.fixture.md", `
---
traces:
  - kind: http_proxy
    require: count > 0
---

# Require

| Name   | Command    | CEL           |
|--------|------------|---------------|
| silent | echo quiet | exitCode == 0 |
`, nil)

		Expect(results["silent"].Status).To(Equal(task.StatusFAIL))
		Expect(results["silent"].Error).To(ContainSubstring("require count > 0 is false"))
		Expect(results["silent"].Traces).To(ConsistOf(HaveField("Count", 0)))
	})

	It("arms the --traces default only for fixtures that declare no traces of their own", func() {
		results := runTraced("optout.fixture.md", `
# Opt out

| Name     | Command  | Traces | CEL           |
|----------|----------|--------|---------------|
| default  | echo one |        | exitCode == 0 |
| optedout | echo two | none   | exitCode == 0 |
`, func(options *fixtures.RunnerOptions) {
			defaults, err := fixtures.ParseTraces("http_proxy")
			Expect(err).ToNot(HaveOccurred())
			options.DefaultTraces = defaults
		})

		Expect(results["default"].Traces).To(ConsistOf(HaveField("Kind", "http_proxy")))
		Expect(results["optedout"].Traces).To(BeEmpty())
	})

	It("copies the trace store a runner opened itself to TraceArchive once the run ends", func() {
		bucket := blob.Local(GinkgoT().TempDir())
		results := runTraced("archive.fixture.md", `
# Archive

| Name  | Command  | Traces     | CEL           |
|-------|----------|------------|---------------|
| armed | echo one | http_proxy | exitCode == 0 |
`, func(options *fixtures.RunnerOptions) { options.TraceArchive = bucket })

		Expect(results["armed"].Traces).To(ConsistOf(HaveField("Kind", "http_proxy")))
		objects, err := bucket.List(context.Background(), "traces/")
		Expect(err).ToNot(HaveOccurred())
		Expect(objects).ToNot(BeEmpty(), "the closed store's files are under traces/<runID>/")
	})

	It("records a command run under a PTY when a terminal trace is armed, and publishes its window on the progress tree", func() {
		var last fixtures.ExecutionSnapshot
		sink := func(options *fixtures.RunnerOptions) {
			options.ProgressSink = func(_ context.Context, snapshot fixtures.ExecutionSnapshot) error {
				last = snapshot
				return nil
			}
		}
		results := runTraced("terminal.fixture.md", `
# Terminal

| Name  | Command                  | Exit Code | Traces   | CEL                                                          |
|-------|--------------------------|-----------|----------|--------------------------------------------------------------|
| shell | printf 'hello\n'; exit 3 | 3         | terminal | terminal.exit_code == 3 && terminal.final.contains("hello") |
`, sink)

		result := results["shell"]
		Expect(result.Status).To(Equal(task.StatusPASS), result.Error)
		Expect(result.Traces).To(HaveLen(1))
		Expect(result.Traces[0].Summary).To(And(HaveKeyWithValue("exit_code", BeEquivalentTo(3)), HaveKeyWithValue("final", ContainSubstring("hello"))))

		var published []parsers.TraceResult
		var visit func(*fixtures.ExecutionNode)
		visit = func(node *fixtures.ExecutionNode) {
			if node.Name == "shell" {
				published = node.Traces
			}
			for _, child := range node.Children {
				visit(child)
			}
		}
		visit(last.Root)
		Expect(published).To(Equal(result.Traces))
	})
})
