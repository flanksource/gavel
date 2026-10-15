package traces_test

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/commons-db/cmd/query/traceserve"
	"github.com/flanksource/commons-db/tracing"
	"github.com/flanksource/commons-db/tracing/process"

	"github.com/flanksource/gavel/traces"
)

// served is a host's store served through query, as the test UI mounts it.
type served struct{ server *httptest.Server }

func serveHost(host *traces.Host) served {
	handler, err := traceserve.Handler(traceserve.Options{Runtime: host.Runtime(), Authorize: traceserve.ReadOnly})
	Expect(err).ToNot(HaveOccurred())
	server := httptest.NewServer(handler)
	DeferCleanup(server.Close)
	return served{server: server}
}

func (s served) post(path string) int {
	response, err := s.server.Client().Post(s.server.URL+path, "application/json", nil)
	Expect(err).ToNot(HaveOccurred())
	Expect(response.Body.Close()).To(Succeed())
	return response.StatusCode
}

// records exports ref's window: the records as the stream holds them.
func (s served) records(ref tracing.TraceRef) []map[string]any {
	query := url.Values{"format": {"ndjson"}, "stream": {ref.Stream}}
	response, err := s.server.Client().Get(s.server.URL + "/api/v1/traces/" + ref.Kind + "/export?" + query.Encode())
	Expect(err).ToNot(HaveOccurred())
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(response.Body)
	Expect(err).ToNot(HaveOccurred())
	if response.StatusCode == http.StatusNotFound {
		return nil
	}
	Expect(response.StatusCode).To(Equal(http.StatusOK), string(body))
	var records []map[string]any
	scanner := bufio.NewScanner(strings.NewReader(string(body)))
	scanner.Buffer(make([]byte, 0, 1<<20), 16<<20)
	for scanner.Scan() {
		var record map[string]any
		Expect(json.Unmarshal(scanner.Bytes(), &record)).To(Succeed())
		records = append(records, record)
	}
	return records
}

func byCommand(records []map[string]any) map[string]map[string]any {
	commands := map[string]map[string]any{}
	for _, record := range records {
		commands[record["command"].(string)] = record
	}
	return commands
}

var _ = Describe("process trace", func() {
	It("serves the tree under the root and its processes' samples through query", func() {
		cmd := exec.Command("sh", "-c", "sleep 30 & wait")
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		Expect(cmd.Start()).To(Succeed())
		root := cmd.Process.Pid
		go func() { _ = cmd.Wait() }()
		DeferCleanup(func() { _ = syscall.Kill(-root, syscall.SIGKILL) })

		host := openHost(GinkgoT().TempDir())
		trace, err := host.StartProcessTrace(root)
		Expect(err).ToNot(HaveOccurred())
		DeferCleanup(func() { Expect(trace.Stop(context.Background())).To(Succeed()) })
		refs := trace.Refs()
		Expect(refs.Tree.Kind).To(Equal(process.TreeKind))
		Expect(refs.Samples.Kind).To(Equal(process.SamplesKind))
		store := serveHost(host)

		Eventually(func() map[string]map[string]any { return byCommand(store.records(refs.Tree)) }).WithTimeout(5 * time.Second).
			Should(HaveKeyWithValue("sleep 30", HaveKeyWithValue("ppid", BeEquivalentTo(root))))
		sleep := byCommand(store.records(refs.Tree))["sleep 30"]
		Eventually(func() []map[string]any { return store.records(refs.Samples) }).WithTimeout(5 * time.Second).
			Should(ContainElement(HaveKeyWithValue("key", sleep["key"])))
	})

	It("checkpoints the run's own goroutines onto its root through the session it names", func() {
		host := openHost(GinkgoT().TempDir())
		trace, err := host.StartProcessTrace(os.Getpid())
		Expect(err).ToNot(HaveOccurred())
		DeferCleanup(func() { Expect(trace.Stop(context.Background())).To(Succeed()) })
		store := serveHost(host)

		Expect(store.post("/api/v1/sessions/" + trace.Session() + "/checkpoint")).To(Equal(http.StatusOK))
		Expect(store.records(trace.Refs().Tree)).To(ContainElement(And(
			HaveKeyWithValue("isRoot", true),
			HaveKeyWithValue("stack", ContainSubstring("goroutine ")),
		)))
	})
})
