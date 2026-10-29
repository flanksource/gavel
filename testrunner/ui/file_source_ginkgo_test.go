package testui_test

import (
	"bytes"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GET /api/file", func() {
	const source = "package main\n\nfunc main() {}\n"
	var workDir, outside string

	BeforeEach(func() {
		workDir = GinkgoT().TempDir()
		outside = GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(workDir, "cmd"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workDir, "cmd", "main.go"), []byte(source), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("secret"), 0o644)).To(Succeed())
		Expect(os.Symlink(filepath.Join(outside, "secret.txt"), filepath.Join(workDir, "link.txt"))).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workDir, "big.txt"), bytes.Repeat([]byte("a"), 2<<20+1), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workDir, "blob.bin"), []byte("text\x00more"), 0o644)).To(Succeed())
	})

	fetch := func(srv *testui.Server, dir, path string) (int, string, string) {
		query := url.Values{"work_dir": {dir}, "path": {path}}
		response := get(srv.Handler(), "/api/file?"+query.Encode())
		return response.Code, response.Header().Get("Content-Type"), response.Body.String()
	}

	lintServer := func() *testui.Server {
		srv := testui.NewServer()
		srv.SetLintResults([]*linters.LinterResult{{Linter: "eslint", WorkDir: workDir}})
		return srv
	}

	It("serves a file under a lint work dir as utf-8 text", func() {
		code, contentType, body := fetch(lintServer(), workDir, "cmd/main.go")
		Expect([]any{code, contentType, body}).To(Equal([]any{http.StatusOK, "text/plain; charset=utf-8", source}))
	})

	It("serves a virtual file from its lint result's source map, and still misses an unmapped one", func() {
		const xml = "<Transaction>\n  <Math/>\n</Transaction>\n"
		srv := testui.NewServer()
		srv.SetLintResults([]*linters.LinterResult{{Linter: "oipa-transaction", WorkDir: workDir, Sources: map[string]linters.SourceFile{
			"rules/pc/Acme/pl/Gold/tr/Fee/XMLData.xml": {URL: "https://mc.example.com/entity/transaction/tx-1/xml", Content: xml},
		}}})

		code, contentType, body := fetch(srv, workDir, "rules/pc/Acme/pl/Gold/tr/Fee/XMLData.xml")
		Expect([]any{code, contentType, body}).To(Equal([]any{http.StatusOK, "text/plain; charset=utf-8", xml}))
		code, _, _ = fetch(srv, workDir, "rules/pc/Acme/pl/Gold/tr/Other/XMLData.xml")
		Expect(code).To(Equal(http.StatusNotFound))
	})

	It("serves a file under a replayed snapshot's test work dir", func() {
		srv := testui.NewServer()
		srv.LoadSnapshot(testui.Snapshot{Tests: []parsers.Test{{Name: "TestMain", WorkDir: workDir, File: "cmd/main.go"}}})
		code, _, body := fetch(srv, workDir, "cmd/main.go")
		Expect([]any{code, body}).To(Equal([]any{http.StatusOK, source}))
	})

	DescribeTable("rejects unsafe or unreadable requests",
		func(dir func() string, path func() string, want int) {
			code, _, body := fetch(lintServer(), dir(), path())
			Expect(code).To(Equal(want), body)
		},
		Entry("work dir not in the snapshot", func() string { return outside }, func() string { return "secret.txt" }, http.StatusBadRequest),
		Entry("../ escape", func() string { return workDir }, func() string { return "../" + filepath.Base(outside) + "/secret.txt" }, http.StatusBadRequest),
		Entry("absolute path", func() string { return workDir }, func() string { return filepath.Join(workDir, "cmd", "main.go") }, http.StatusBadRequest),
		Entry("symlink escaping the work dir", func() string { return workDir }, func() string { return "link.txt" }, http.StatusBadRequest),
		Entry("missing file", func() string { return workDir }, func() string { return "cmd/missing.go" }, http.StatusNotFound),
		Entry("file over 2 MiB", func() string { return workDir }, func() string { return "big.txt" }, http.StatusRequestEntityTooLarge),
		Entry("binary content", func() string { return workDir }, func() string { return "blob.bin" }, http.StatusUnsupportedMediaType),
	)
})
