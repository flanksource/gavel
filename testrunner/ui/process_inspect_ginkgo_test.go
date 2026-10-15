package testui_test

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strconv"

	"github.com/flanksource/clicky/process"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	testui "github.com/flanksource/gavel/testrunner/ui"
)

const loopbackClient = "127.0.0.1:50000"

func inspect(facet string, pid int, remoteAddr string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, "/api/process/"+facet+"?pid="+strconv.Itoa(pid), nil)
	request.RemoteAddr = remoteAddr
	response := httptest.NewRecorder()
	testui.NewServer().Handler().ServeHTTP(response, request)
	return response
}

var _ = Describe("process inspection", func() {
	var child *exec.Cmd

	BeforeEach(func() {
		child = exec.Command("sleep", "30")
		child.Env = []string{"ACME_TOKEN=secret-value", "PATH=" + os.Getenv("PATH")}
		Expect(child.Start()).To(Succeed())
		DeferCleanup(func() { child.Process.Kill() }) //nolint:errcheck
	})

	It("serves a run process's environment to a client on this machine", func() {
		Eventually(func() map[string]string {
			response := inspect("env", child.Process.Pid, loopbackClient)
			Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
			var env map[string]string
			Expect(json.Unmarshal(response.Body.Bytes(), &env)).To(Succeed())
			return env
		}).Should(HaveKeyWithValue("ACME_TOKEN", "secret-value"))
	})

	It("refuses the environment to a client on another host", func() {
		response := inspect("env", child.Process.Pid, "192.0.2.10:50000")
		Expect(response.Code).To(Equal(http.StatusForbidden))
		Expect(response.Body.String()).NotTo(ContainSubstring("secret-value"))
	})

	It("splits open descriptors into files, sockets and listening ports", func() {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		Expect(err).NotTo(HaveOccurred())
		defer listener.Close()
		port := listener.Addr().(*net.TCPAddr).Port

		read := func(facet string) []process.Descriptor {
			response := inspect(facet, os.Getpid(), "192.0.2.10:50000")
			Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
			var descriptors []process.Descriptor
			Expect(json.Unmarshal(response.Body.Bytes(), &descriptors)).To(Succeed())
			return descriptors
		}
		Expect(read("files")).To(HaveEach(HaveField("Port", 0)))
		Expect(read("sockets")).To(ContainElement(HaveField("Port", port)))
		Expect(read("ports")).To(And(
			ContainElement(HaveField("Port", port)),
			HaveEach(Satisfy(process.Descriptor.IsListening)),
		))
	})

	It("refuses a process outside the server's own tree", func() {
		response := inspect("files", 1, loopbackClient)
		Expect(response.Code).To(Equal(http.StatusNotFound))
		Expect(response.Body.String()).To(ContainSubstring("not part of this run"))
	})

	It("rejects an unknown facet and a missing pid", func() {
		Expect(inspect("threads", os.Getpid(), loopbackClient).Code).To(Equal(http.StatusNotFound))
		Expect(inspect("env", 0, loopbackClient).Code).To(Equal(http.StatusBadRequest))
	})
})
