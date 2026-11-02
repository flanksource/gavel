package testrunner

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	osexec "os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/commons-db/cmd/query/traceserve"
	"github.com/flanksource/commons-db/tracing"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/traces"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const hangPackage = "./testrunner/testdata/hang"

// traceStore is a run's store served through query, as the UI reads it.
type traceStore struct{ server *httptest.Server }

func serveStore(host *traces.Host) traceStore {
	handler, err := traceserve.Handler(traceserve.Options{Runtime: host.Runtime(), Authorize: traceserve.ReadOnly})
	Expect(err).NotTo(HaveOccurred())
	server := httptest.NewServer(handler)
	DeferCleanup(server.Close)
	return traceStore{server: server}
}

// records exports ref's stream.
func (s traceStore) records(ref tracing.TraceRef) []map[string]any {
	response, err := s.server.Client().Get(s.server.URL + "/api/v1/traces/" + ref.Kind + "/export?" + url.Values{"format": {"ndjson"}, "stream": {ref.Stream}}.Encode())
	Expect(err).NotTo(HaveOccurred())
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return nil
	}
	Expect(response.StatusCode).To(Equal(http.StatusOK))
	var records []map[string]any
	scanner := bufio.NewScanner(response.Body)
	scanner.Buffer(make([]byte, 0, 1<<20), 16<<20)
	for scanner.Scan() {
		var record map[string]any
		Expect(json.Unmarshal(scanner.Bytes(), &record)).To(Succeed())
		records = append(records, record)
	}
	return records
}

func (s traceStore) checkpoint(session, body string) int {
	response, err := s.server.Client().Post(s.server.URL+"/api/v1/sessions/"+session+"/checkpoint", "application/json", strings.NewReader(body))
	Expect(err).NotTo(HaveOccurred())
	Expect(response.Body.Close()).To(Succeed())
	return response.StatusCode
}

// hangBinary is the tree node of the hang package's Go test binary: the process named hang.test,
// not one whose command only mentions it, as the linker's `-o .../hang.test` does.
func (s traceStore) hangBinary(trace *traces.ProcessTrace) map[string]any {
	for _, node := range s.records(trace.Refs().Tree) {
		if node["name"] == "hang.test" {
			return node
		}
	}
	return nil
}

func withField(records []map[string]any, field string, value any) []map[string]any {
	var matched []map[string]any
	for _, record := range records {
		if record[field] == value {
			matched = append(matched, record)
		}
	}
	return matched
}

func findTest(results parsers.TestSuiteResults, name string) *parsers.Test {
	for ri := range results {
		for ti := range results[ri].Tests {
			if results[ri].Tests[ti].Name == name {
				return &results[ri].Tests[ti]
			}
		}
	}
	return nil
}

func openHost(runID string) *traces.Host {
	host, err := traces.Open(GinkgoT().TempDir(), runID)
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() { Expect(host.Close(context.Background())).To(Succeed()) })
	return host
}

var _ = Describe("--test-timeout in a go test batch", Serial, func() {
	It("bounds each package, not the batch as a whole", func() {
		if runtime.GOOS == "windows" {
			Skip("needs SIGQUIT")
		}
		workDir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(workDir, "go.mod"), []byte("module example.com/slow\n\ngo 1.22\n"), 0o600)).To(Succeed())
		write := func(pkg, body string) {
			Expect(os.MkdirAll(filepath.Join(workDir, pkg), 0o755)).To(Succeed())
			src := "package " + pkg + "\n\nimport (\n\t\"testing\"\n\t\"time\"\n)\n\n" + body + "\n"
			Expect(os.WriteFile(filepath.Join(workDir, pkg, pkg+"_test.go"), []byte(src), 0o600)).To(Succeed())
		}
		// Three packages of 2s each run one after another: 6s together, past
		// the 5s deadline, each well within it.
		for _, pkg := range []string{"a", "b", "c"} {
			write(pkg, "func TestSlow_"+pkg+"(t *testing.T) { time.Sleep(2 * time.Second) }")
		}
		write("hang", "func TestHang(t *testing.T) { <-time.After(time.Hour) }")
		warm := osexec.Command("go", "test", "-count=1", "-run", "^$", "./...")
		warm.Dir = workDir
		Expect(warm.CombinedOutput()).Error().NotTo(HaveOccurred())
		GinkgoT().Setenv("GOFLAGS", "-p=1")

		o := &TestOrchestrator{
			RunOptions: RunOptions{WorkDir: workDir, TestTimeout: 5 * time.Second},
			registry:   DefaultRegistry(workDir),
		}
		results, err := o.detectAndRun([]Framework{parsers.GoTest}, []string{"./a", "./b", "./c", "./hang"}, nil)
		Expect(err).NotTo(HaveOccurred())

		for _, pkg := range []string{"a", "b", "c"} {
			slow := findTest(results, "TestSlow_"+pkg)
			Expect(slow).NotTo(BeNil(), pkg)
			Expect(slow.Failed).To(BeFalse(), "%s ran within its own deadline: %s", pkg, slow.Message)
			Expect(slow.TimedOut).To(BeFalse(), pkg)
		}
		hang := findTest(results, "TestHang")
		Expect(hang).NotTo(BeNil())
		Expect(hang.TimedOut).To(BeTrue())
		Expect(hang.Attempts[len(hang.Attempts)-1].StackTrace).To(ContainSubstring("hang.TestHang"))
	})
})

var _ = Describe("a Go test package that times out", Ordered, func() {
	var workDir string

	BeforeAll(func() {
		if runtime.GOOS == "windows" {
			Skip("goroutine dumps need SIGQUIT")
		}
		var err error
		workDir, err = filepath.Abs("..")
		Expect(err).NotTo(HaveOccurred())
		// Compile once up front, so the deadline below is spent running.
		warm := osexec.Command("go", "test", "-count=1", "-run", "^$", hangPackage)
		warm.Dir = workDir
		Expect(warm.CombinedOutput()).Error().NotTo(HaveOccurred())
	})

	It("gets its goroutines dumped onto the attempt and into stacks on its binary's key, its peaks read from samples", func() {
		host := openHost("run-timeout")
		trace, err := host.StartProcessTrace(traces.ProcessTraceOptions{RootPID: os.Getpid()})
		Expect(err).NotTo(HaveOccurred())
		store := serveStore(host)

		o := &TestOrchestrator{
			RunOptions: RunOptions{WorkDir: workDir, TestTimeout: 6 * time.Second, ProcessTrace: trace},
			registry:   DefaultRegistry(workDir),
		}
		results, err := o.detectAndRun([]Framework{parsers.GoTest}, []string{hangPackage}, nil)
		Expect(err).NotTo(HaveOccurred())

		hang := findTest(results, "TestHang")
		Expect(hang).NotTo(BeNil())
		Expect(hang.TimedOut).To(BeTrue())
		attempt := hang.Attempts[len(hang.Attempts)-1]
		Expect(attempt.StackTrace).To(And(HavePrefix("SIGQUIT: quit"), ContainSubstring("hang.TestHang")))

		binary := store.hangBinary(trace)
		Expect(binary).NotTo(BeNil())
		dumped := withField(store.records(trace.Refs().Stacks), "reason", PackageTimeoutReason(hangPackage))
		Expect(dumped).To(HaveEach(HaveKeyWithValue("key", binary["key"])))
		Expect(dumped).To(ContainElement(HaveKeyWithValue("frames", ContainElement(HaveKeyWithValue("func", ContainSubstring("hang.TestHang"))))))
		Expect(attempt.GoroutineCt).To(Equal(len(dumped)), "goroutine_count is the record count of the timeout checkpoint")

		samples := withField(store.records(trace.Refs().Samples), "pid", float64(attempt.PID))
		Expect(samples).NotTo(BeEmpty(), "go test ran for its whole deadline, through several ticks")
		var peakRSS float64
		for _, sample := range samples {
			peakRSS = max(peakRSS, sample["rssBytes"].(float64))
		}
		Expect(attempt.ProcessKey).To(Equal(samples[0]["key"]))
		Expect(float64(attempt.RSS)).To(Equal(peakRSS), "the attempt's peak RSS is its samples' peak")
		Expect(trace.Stop(context.Background())).To(Succeed())
	})

	It("dumps a running Go test binary over sigquit through the stacks session, refusing gavel itself", func() {
		host := openHost("run-dump")
		dumper := NewSigquitDumper()
		Expect(host.RegisterDumper(dumper)).To(Succeed())
		trace, err := host.StartProcessTrace(traces.ProcessTraceOptions{RootPID: os.Getpid()})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(trace.Stop(context.Background())).To(Succeed()) })
		store := serveStore(host)

		o := &TestOrchestrator{
			RunOptions: RunOptions{WorkDir: workDir, TestTimeout: time.Minute, ProcessTrace: trace, SigquitDumper: dumper},
			registry:   DefaultRegistry(workDir),
		}
		ran := make(chan parsers.TestSuiteResults, 1)
		go func() {
			defer GinkgoRecover()
			results, err := o.detectAndRun([]Framework{parsers.GoTest}, []string{hangPackage}, nil)
			Expect(err).NotTo(HaveOccurred())
			ran <- results
		}()
		Eventually(func() map[string]any { return store.hangBinary(trace) }).WithTimeout(30 * time.Second).
			Should(HaveKeyWithValue("stackTransports", ContainElement(map[string]any{"transport": TransportSigquit, "destructive": true})))
		key := store.hangBinary(trace)["key"].(string)

		session := trace.Sessions().Stacks
		Expect(store.checkpoint(session, `{"pid": `+strconv.Itoa(os.Getpid())+`, "transport": "sigquit"}`)).
			To(Equal(http.StatusBadRequest), "gavel itself is not a go test binary of the batch")
		Expect(store.checkpoint(session, `{"key": "`+key+`", "transport": "sigquit", "reason": "manual"}`)).To(Equal(http.StatusOK))
		Expect(withField(store.records(trace.Refs().Stacks), "reason", "manual")).To(And(
			HaveEach(And(HaveKeyWithValue("key", key), HaveKeyWithValue("transport", TransportSigquit))),
			ContainElement(HaveKeyWithValue("frames", ContainElement(HaveKeyWithValue("func", ContainSubstring("hang.TestHang"))))),
		))
		Eventually(ran).WithTimeout(30 * time.Second).Should(Receive())
	})
})
