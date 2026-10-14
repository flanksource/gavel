package services_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/flanksource/gavel/fixtures"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func processAlive(pid int) bool { return syscall.Kill(pid, 0) == nil }

var _ = Describe("services kind: exec entries", func() {
	It("starts an exec entry on an allocated port, waits for http readiness and exports API_URL/API_PORT", func() {
		step := newStep("stack", fmt.Sprintf(`{"services": {"api": {
			"exec": "%s --label api --addr 127.0.0.1:{{.port}}",
			"ready": {"http": "http://127.0.0.1:{{.port}}/healthz", "status": 200}
		}}}`, svcBin))

		outcome, err := execute(step)
		Expect(err).NotTo(HaveOccurred())

		api := entryOutputs(outcome, "api")
		port, ok := api["port"].(int)
		Expect(ok).To(BeTrue(), "port is %T", api["port"])
		expectedURL := "http://127.0.0.1:" + strconv.Itoa(port)
		Expect(api).To(HaveKeyWithValue("url", expectedURL))
		Expect(api["pid"]).To(BeNumerically(">", 0))
		Expect(api["log"]).To(HavePrefix(step.TmpDir))
		Expect(outcome.Outputs["api"]).To(Equal(api), "top-level per-name alias")

		resp, err := http.Get(expectedURL + "/")
		Expect(err).NotTo(HaveOccurred())
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		Expect(string(body)).To(Equal("api"))

		Expect(outcome.Exports).To(Equal(map[string]string{"API_URL": expectedURL, "API_PORT": strconv.Itoa(port)}))
		Expect(envValue(step, "API_URL")).To(Equal(expectedURL))
		Expect(step.Scope.Env.Origin("API_URL")).To(Equal("stack"))

		pid := api["pid"].(int)
		Expect(outcome.Cleanup(context.Background())).To(Succeed())
		Expect(processAlive(pid)).To(BeFalse(), "service process group is gone after cleanup")
	})

	It("leaves shell parameters in exec to the shell while expanding $NAME in the ready URL", func() {
		step := newStep("stack", fmt.Sprintf(`{"services": {"api": {
			"exec": "for label in api; do %s --label \"$label\" --addr 127.0.0.1:{{.port}}; done; exit $?",
			"ready": {"http": "http://$SVC_HOST:{{.port}}/healthz", "status": 200}
		}}}`, svcBin))
		Expect(step.Scope.Env.Export("seed", "SVC_HOST", "127.0.0.1", false)).To(Succeed())

		outcome, err := execute(step)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(outcome.Cleanup(context.Background())).To(Succeed()) })

		resp, err := http.Get(entryOutputs(outcome, "api")["url"].(string) + "/")
		Expect(err).NotTo(HaveOccurred())
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		Expect(string(body)).To(Equal("api"), "the shell expanded the loop variable $label")
	})

	It("fails a non-shell value naming an unset variable instead of rendering it empty", func() {
		step := newStep("stack", fmt.Sprintf(`{"services": {"api": {
			"exec": "%s --label api --addr 127.0.0.1:{{.port}}",
			"ready": {"http": "http://$MISSING_HOST:{{.port}}/healthz"}
		}}}`, svcBin))

		_, err := execute(step)
		Expect(err).To(MatchError(ContainSubstring("MISSING_HOST")))
	})

	It("waits for a tcp readiness probe while the service delays listening", func() {
		step := newStep("stack", fmt.Sprintf(`{"services": {"slow": {
			"exec": "%s --label slow --delay 600ms --addr 127.0.0.1:{{.port}}",
			"ready": {"tcp": "127.0.0.1:{{.port}}"},
			"wait": {"interval": "50ms", "timeout": "10s"}
		}}}`, svcBin))

		outcome, err := execute(step)
		Expect(err).NotTo(HaveOccurred())
		log, err := os.ReadFile(entryOutputs(outcome, "slow")["log"].(string))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(log)).To(ContainSubstring("slow listening on"))
	})

	It("fails a readiness timeout with the service log tail and stops the process", func() {
		step := newStep("stack", fmt.Sprintf(`{"services": {"stuck": {
			"exec": "%s --label stuck-widget --delay 1h --addr 127.0.0.1:{{.port}}",
			"ready": {"tcp": "127.0.0.1:{{.port}}"},
			"wait": {"interval": "50ms", "timeout": "500ms"}
		}}}`, svcBin))

		_, err := execute(step)
		Expect(err).To(MatchError(And(
			ContainSubstring(`service "stuck" not ready after 500ms`),
			ContainSubstring("stuck-widget booting"),
		)))
		Eventually(func() string {
			out, _ := os.ReadFile(filepath.Join(step.TmpDir, "services", "stuck.log"))
			return string(out)
		}).Should(ContainSubstring("stuck-widget booting"))
	})

	It("fails fast when the service exits before it is ready", func() {
		step := newStep("stack", `{"services": {"dies": {
			"exec": "echo dying-loudly; exit 3",
			"ready": {"tcp": "127.0.0.1:1"},
			"wait": {"interval": "50ms", "timeout": "30s"}
		}}}`)

		start := time.Now()
		_, err := execute(step)
		Expect(err).To(MatchError(And(ContainSubstring("exited"), ContainSubstring("dying-loudly"))))
		Expect(time.Since(start)).To(BeNumerically("<", 10*time.Second))
	})

	It("tears entries down in reverse declaration order", func() {
		termFile := filepath.Join(GinkgoT().TempDir(), "order.log")
		entry := func(label string) string {
			return fmt.Sprintf(`"%s": {"exec": "%s --label %s --term-file %s --addr 127.0.0.1:{{.port}}",
				"ready": {"tcp": "127.0.0.1:{{.port}}"}}`, label, svcBin, label, termFile)
		}
		step := newStep("stack", `{"services": {`+entry("zeta")+`,`+entry("alpha")+`,`+entry("mid")+`}}`)

		outcome, err := execute(step)
		Expect(err).NotTo(HaveOccurred())
		Expect(outcome.Cleanup(context.Background())).To(Succeed())

		order, err := os.ReadFile(termFile)
		Expect(err).NotTo(HaveOccurred())
		Expect(strings.Fields(string(order))).To(Equal([]string{"mid", "alpha", "zeta"}))
	})

	It("renders later entries against earlier entries' outputs in declaration order", func() {
		step := newStep("stack", fmt.Sprintf(`{"services": {
			"zeta": {"exec": "%[1]s --label zeta --addr 127.0.0.1:{{.port}}", "ready": {"tcp": "127.0.0.1:{{.port}}"}},
			"alpha": {
				"exec": "%[1]s --label alpha --addr 127.0.0.1:{{.port}}",
				"env": {"UPSTREAM": "{{.services.zeta.url}}"},
				"ready": {"exec": "curl -sf http://127.0.0.1:{{.port}}/healthz"},
				"export": {"*": "{{.outputs.env}}", "ALPHA_LOG": "{{.outputs.log}}"}
			}
		}}`, svcBin))

		outcome, err := execute(step)
		Expect(err).NotTo(HaveOccurred())
		zetaURL := entryOutputs(outcome, "zeta")["url"]
		alpha := entryOutputs(outcome, "alpha")
		Expect(alpha["env"]).To(Equal(map[string]any{"UPSTREAM": zetaURL}))
		Expect(envValue(step, "UPSTREAM")).To(Equal(zetaURL))
		Expect(envValue(step, "ALPHA_LOG")).To(Equal(alpha["log"]))

		log, err := os.ReadFile(alpha["log"].(string))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(log)).To(ContainSubstring(fmt.Sprintf("upstream=%q", zetaURL)))
	})

	It("rejects a reference to an entry declared later in the step", func() {
		step := newStep("stack", `{"services": {
			"first": {"exec": "sleep 60", "env": {"X": "{{.services.second.url}}"}, "ready": {"exec": "true"}},
			"second": {"exec": "sleep 60", "ready": {"exec": "true"}}
		}}`)
		_, err := execute(step)
		Expect(err).To(MatchError(ContainSubstring(`entry "first" references services.second, which is not declared before it`)))
	})

	It("reports readiness through Observe and flips it when the service dies", func() {
		step := newStep("stack", fmt.Sprintf(`{"services": {"api": {
			"exec": "%s --label api --addr 127.0.0.1:{{.port}}",
			"ready": {"http": "http://127.0.0.1:{{.port}}/healthz"}
		}}}`, svcBin))
		outcome, err := execute(step)
		Expect(err).NotTo(HaveOccurred())

		observer, ok := servicesKind().(fixtures.Observer)
		Expect(ok).To(BeTrue(), "services implements fixtures.Observer")
		observed, err := observer.Observe(context.Background(), step, outcome.Handle)
		Expect(err).NotTo(HaveOccurred())
		Expect(observed["ready"]).To(BeTrue())
		Expect(entryOutputs(fixtures.StepOutcome{Outputs: observed}, "api")["http"]).To(Equal(map[string]any{"status": 200}))

		Expect(syscall.Kill(-entryOutputs(outcome, "api")["pid"].(int), syscall.SIGKILL)).To(Succeed())
		Eventually(func() any {
			observed, err := observer.Observe(context.Background(), step, outcome.Handle)
			Expect(err).NotTo(HaveOccurred())
			return observed["ready"]
		}).Should(BeFalse())
	})
})
