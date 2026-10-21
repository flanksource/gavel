package vitedev

import (
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const entry = "/src/index.tsx"

// serveOnFreePort starts body on a loopback port and returns an App on it.
func serveOnFreePort(body string) App {
	GinkgoHelper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	Expect(err).NotTo(HaveOccurred())
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	server.Listener = listener
	server.Start()
	DeferCleanup(server.Close)
	return App{Name: "example UI", Dir: "example/ui", Port: listener.Addr().(*net.TCPAddr).Port, Entry: entry}
}

func sourceDir(files ...string) string {
	GinkgoHelper()
	dir := GinkgoT().TempDir()
	for _, name := range files {
		Expect(os.WriteFile(filepath.Join(dir, name), []byte("{}"), 0o600)).To(Succeed())
	}
	return dir
}

var _ = Describe("vite dev server", func() {
	It("reuses a dev server already serving the app's entry, without starting another", func() {
		app := serveOnFreePort(`<script type="module" src="` + entry + `"></script>`)

		Expect(Ensure(app, "/nonexistent")).To(Succeed())
	})

	It("refuses a port held by a server that is not the app", func() {
		app := serveOnFreePort("some other service")

		Expect(Ensure(app, "/nonexistent")).To(MatchError(
			"localhost:" + strconv.Itoa(app.Port) + " is in use but is not the example UI Vite dev server — free the port and retry"))
	})

	It("takes a source dir with package.json and vite.config.ts, and nothing less", func() {
		app := App{Name: "example UI", Dir: "no/such/ui"}
		valid := sourceDir("package.json", "vite.config.ts")

		Expect(ResolveDir(app, valid)).To(Equal(valid))
		_, err := ResolveDir(app, sourceDir("package.json"))
		Expect(err).To(MatchError(ContainSubstring("--dev: example UI source not found")))
	})

	It("answers an unreachable dev server with a 502 naming the app", func() {
		app := App{Name: "example UI", Dir: "example/ui", Port: closedPort(), Entry: entry}
		rec := httptest.NewRecorder()

		Proxy(app).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		Expect(rec.Code).To(Equal(http.StatusBadGateway))
		Expect(rec.Body.String()).To(ContainSubstring("the example UI vite dev server is unreachable"))
	})
})

func closedPort() int {
	GinkgoHelper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	Expect(err).NotTo(HaveOccurred())
	port := listener.Addr().(*net.TCPAddr).Port
	Expect(listener.Close()).To(Succeed())
	return port
}
