package status_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/flanksource/gavel/sdk/status"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// shortCache is a cache dir short enough for a unix socket path under it.
func shortCache() string {
	dir := GinkgoT().TempDir()
	Expect(os.MkdirAll(status.Dir(dir), 0o700)).To(Succeed())
	return dir
}

func write(cache string, s status.Status) {
	data, err := json.Marshal(s)
	Expect(err).NotTo(HaveOccurred())
	Expect(os.WriteFile(status.Path(cache, s.PID), data, 0o600)).To(Succeed())
}

func exitedPID() int {
	cmd := exec.Command("true")
	Expect(cmd.Run()).To(Succeed())
	return cmd.Process.Pid
}

var _ = Describe("CacheDir", func() {
	It("uses XDG_CACHE_HOME on every platform", func() {
		GinkgoT().Setenv("XDG_CACHE_HOME", "/srv/cache")
		Expect(status.CacheDir()).To(Equal("/srv/cache/gavel"))
	})
	It("refuses a relative XDG_CACHE_HOME", func() {
		GinkgoT().Setenv("XDG_CACHE_HOME", "relative/cache")
		_, err := status.CacheDir()
		Expect(err).To(MatchError(ContainSubstring(`XDG_CACHE_HOME must be an absolute path, got "relative/cache"`)))
	})
	It("falls back to the OS user cache directory", func() {
		GinkgoT().Setenv("XDG_CACHE_HOME", "")
		cache, err := os.UserCacheDir()
		Expect(err).NotTo(HaveOccurred())
		Expect(status.CacheDir()).To(Equal(filepath.Join(cache, "gavel")))
	})
})

var _ = Describe("List", func() {
	It("returns live processes only, by pid", func() {
		cache, dead := shortCache(), exitedPID()
		write(cache, status.Status{PID: os.Getpid(), Command: "gavel test"})
		write(cache, status.Status{PID: dead, Command: "gavel lint"})
		statuses, err := status.List(cache)
		Expect(err).NotTo(HaveOccurred())
		Expect(statuses).To(HaveLen(1))
		Expect(statuses[0].Command).To(Equal("gavel test"))

		stale, err := status.Stale(cache)
		Expect(err).NotTo(HaveOccurred())
		Expect(stale).To(ConsistOf(status.Path(cache, dead)))
	})
	It("names a corrupt status file", func() {
		cache := shortCache()
		path := status.Path(cache, os.Getpid())
		Expect(os.WriteFile(path, []byte("{not json"), 0o600)).To(Succeed())
		_, err := status.List(cache)
		Expect(err).To(MatchError(ContainSubstring(path)))
	})
	It("is empty before any gavel ran", func() {
		Expect(status.List(filepath.Join(GinkgoT().TempDir(), "missing"))).To(BeEmpty())
	})
})

var _ = Describe("Watch", func() {
	It("follows the file, then the socket's stream, until the process's status goes", func(ctx SpecContext) {
		cache, pid := shortCache(), os.Getpid()
		write(cache, status.Status{PID: pid, Command: "gavel test", Tests: &status.Summary{Total: 4, Pending: 4}})
		watcher, err := status.Watch(ctx, status.WatchOptions{Cache: cache, PID: pid, Interval: 10 * time.Millisecond})
		Expect(err).NotTo(HaveOccurred())
		Expect((<-watcher.Updates()).Tests.Pending).To(Equal(4))

		write(cache, status.Status{PID: pid, Command: "gavel test", Tests: &status.Summary{Total: 4, Passed: 1, Pending: 3}})
		Eventually(watcher.Updates()).Should(Receive(HaveField("Tests.Passed", 1)))

		socket := status.SocketPath(cache, pid)
		stop := serveStream(socket, status.Status{PID: pid, Tests: &status.Summary{Total: 4, Passed: 3, Pending: 1}})
		write(cache, status.Status{PID: pid, Socket: socket, Tests: &status.Summary{Total: 4, Passed: 2, Pending: 2}})
		Eventually(watcher.Updates()).Should(Receive(HaveField("Tests.Passed", 2)))
		Eventually(watcher.Updates()).Should(Receive(HaveField("Tests.Passed", 3)))

		Expect(os.Remove(status.Path(cache, pid))).To(Succeed())
		stop()
		Eventually(watcher.Updates()).Should(BeClosed())
		Expect(watcher.Err()).NotTo(HaveOccurred())
	})
	It("fails when the stream sends an error event", func(ctx SpecContext) {
		cache, pid := shortCache(), os.Getpid()
		socket := status.SocketPath(cache, pid)
		stop := serveFrames(socket, "event: error\ndata: {\"error\":\"load failed\"}\n\n")
		DeferCleanup(stop)
		write(cache, status.Status{PID: pid, Socket: socket})
		watcher, err := status.Watch(ctx, status.WatchOptions{Cache: cache, PID: pid})
		Expect(err).NotTo(HaveOccurred())
		Eventually(watcher.Updates()).Should(BeClosed())
		Expect(watcher.Err()).To(MatchError(ContainSubstring("load failed")))
	})
	It("refuses a process with no status", func(ctx SpecContext) {
		_, err := status.Watch(ctx, status.WatchOptions{Cache: shortCache(), PID: os.Getpid()})
		Expect(errors.Is(err, os.ErrNotExist)).To(BeTrue(), "%v", err)
	})
})

// serveStream serves one status frame on /status/stream, then holds the stream open.
func serveStream(socket string, s status.Status) func() {
	data, err := json.Marshal(s)
	Expect(err).NotTo(HaveOccurred())
	return serveFrames(socket, ": ping\n\n"+fmt.Sprintf("data: %s\n\n", data))
}

func serveFrames(socket, frames string) func() {
	listener, err := net.Listen("unix", socket)
	Expect(err).NotTo(HaveOccurred())
	done := make(chan struct{})
	server := &http.Server{ReadHeaderTimeout: time.Second, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Expect(r.URL.Path).To(Equal("/status/stream"))
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprint(w, frames)
		w.(http.Flusher).Flush()
		select {
		case <-r.Context().Done():
		case <-done:
		}
	})}
	go func() { _ = server.Serve(listener) }()
	return func() {
		close(done)
		_ = server.Shutdown(context.Background())
	}
}
