package services_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func gitOutput(dir string, args ...string) string {
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), string(out))
	return strings.TrimSpace(string(out))
}

var _ = Describe("services kind: git entries and exports", func() {
	It("seeds a git repository with one commit holding the declared files", func() {
		step := newStep("stack", `{"services": {"workspace": {
			"git": {"files": {"go.mod": "module example.com/greet\n", "greet/greet_test.go": "package greet\n"}, "commit": true},
			"export": {"WORKSPACE": "{{.outputs.dir}}"}
		}}}`)

		outcome, err := execute(step)
		Expect(err).NotTo(HaveOccurred())
		dir := entryOutputs(outcome, "workspace")["dir"].(string)
		Expect(dir).To(HavePrefix(step.TmpDir))
		Expect(envValue(step, "WORKSPACE")).To(Equal(dir))

		Expect(gitOutput(dir, "rev-list", "--count", "HEAD")).To(Equal("1"))
		Expect(gitOutput(dir, "status", "--porcelain")).To(BeEmpty())
		Expect(strings.Fields(gitOutput(dir, "ls-files"))).To(ConsistOf("go.mod", "greet/greet_test.go"))
		content, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(Equal("module example.com/greet\n"))
	})

	It("rejects two entries of one step exporting the same variable", func() {
		step := newStep("stack", `{"services": {
			"one": {"git": {"files": {"a": "a"}}, "export": {"SHARED": "{{.outputs.dir}}"}},
			"two": {"git": {"files": {"b": "b"}}, "export": {"SHARED": "{{.outputs.dir}}"}}
		}}`)
		_, err := execute(step)
		Expect(err).To(MatchError(ContainSubstring(`entry "two" exports SHARED, already exported by entry "one"`)))
	})

	It("surfaces a collision with a variable an earlier step exported", func() {
		step := newStep("stack", `{"services": {"one": {"git": {"files": {"a": "a"}}, "export": {"WORKSPACE": "{{.outputs.dir}}"}}}}`)
		Expect(step.Scope.Env.Export("earlier", "WORKSPACE", "/elsewhere", false)).To(Succeed())

		_, err := execute(step)
		Expect(err).To(MatchError(ContainSubstring(`WORKSPACE was already exported by step "earlier"`)))
	})

	It("rejects a wildcard export of an output that is not a map", func() {
		step := newStep("stack", `{"services": {"one": {"git": {"files": {"a": "a"}}, "export": {"*": "{{.outputs.dir}}"}}}}`)
		_, err := execute(step)
		Expect(err).To(MatchError(ContainSubstring(`export "*" of entry "one": outputs.dir is a string, not a map`)))
	})

	It("rejects an entry that declares no service type", func() {
		step := newStep("stack", `{"services": {"empty": {"env": {"A": "b"}}}}`)
		_, err := execute(step)
		Expect(err).To(MatchError(ContainSubstring(`entry "empty": exactly one of deps, exec, git or dbtest is required`)))
	})
})

var _ = Describe("services kind: deps entries", func() {
	It("starts a deps catalog postgres hermetically under the step tmp dir", func() {
		if os.Getenv("GAVEL_E2E_DEPS") != "1" {
			Skip("set GAVEL_E2E_DEPS=1 to start a real postgres through deps")
		}
		step := newStep("stack", `{"services": {"postgres": {"deps": "postgres"}}}`)
		outcome, err := execute(step)
		Expect(err).NotTo(HaveOccurred())

		pg := entryOutputs(outcome, "postgres")
		Expect(pg["url"]).To(HavePrefix("postgres"))
		port := pg["port"].(int)
		conn, err := net.Dial("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
		Expect(err).NotTo(HaveOccurred())
		_ = conn.Close()
		Expect(envValue(step, "POSTGRES_URL")).To(Equal(pg["url"]))
		Expect(envValue(step, "POSTGRES_PORT")).To(Equal(strconv.Itoa(port)))

		Expect(outcome.Cleanup(context.Background())).To(Succeed())
		_, err = net.Dial("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
		Expect(err).To(HaveOccurred(), fmt.Sprintf("postgres still listening on %d after cleanup", port))
	})
})
