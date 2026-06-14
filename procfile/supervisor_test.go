package procfile_test

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	pf "github.com/flanksource/gavel/procfile"
	"github.com/flanksource/gavel/utils"
	"github.com/flanksource/gavel/verify"
)

// newSupervisor writes a Procfile into a fresh temp root and returns a started
// supervisor plus its state dir. The caller is responsible for Shutdown.
func newSupervisor(procfile string, cfg verify.ProcfileConfig) (*pf.Supervisor, string) {
	root := GinkgoT().TempDir()
	Expect(os.WriteFile(filepath.Join(root, "Procfile"), []byte(procfile), 0o644)).To(Succeed())
	dir, err := pf.StateDir(root)
	Expect(err).NotTo(HaveOccurred())

	sup, err := pf.NewSupervisor(pf.Options{
		Root:     root,
		Procfile: filepath.Join(root, "Procfile"),
		Config:   cfg,
	})
	Expect(err).NotTo(HaveOccurred())
	Expect(sup.Start()).To(Succeed())
	return sup, dir
}

// waitFor polls cond up to timeout, failing the spec if it never holds.
func waitFor(timeout time.Duration, cond func() bool) {
	EventuallyWithOffset(1, cond, timeout, 25*time.Millisecond).Should(BeTrue())
}

var _ = Describe("Supervisor", func() {
	It("starts processes, records pids and captures their output", func() {
		sup, dir := newSupervisor("ticker: sh -c 'echo hello-from-ticker; sleep 30'\n", verify.ProcfileConfig{})
		defer sup.Shutdown()

		var pid int
		waitFor(3*time.Second, func() bool {
			st, _ := pf.ReadState(dir)
			p, ok := st.Process("ticker")
			if ok && p.Status == pf.StatusRunning && p.PID > 0 {
				pid = p.PID
				return true
			}
			return false
		})
		Expect(utils.ProcessAlive(pid)).To(BeTrue())

		waitFor(3*time.Second, func() bool {
			out, _ := utils.TailFile(pf.LogPath(dir, "ticker"), 10)
			for _, line := range out {
				if line == "hello-from-ticker" {
					return true
				}
			}
			return false
		})

		sup.Shutdown()
		waitFor(3*time.Second, func() bool { return !utils.ProcessAlive(pid) })
	})

	It("injects env vars and leaves shell-internal expansion to the process", func() {
		// PORT comes from config env; $((1+1)) is shell arithmetic that must NOT
		// be pre-expanded by the supervisor.
		cfg := verify.ProcfileConfig{Env: map[string]string{"PORT": "9999"}}
		sup, dir := newSupervisor("envtest: sh -c 'echo port=$PORT count=$((1+1)); sleep 30'\n", cfg)
		defer sup.Shutdown()

		waitFor(3*time.Second, func() bool {
			out, _ := utils.TailFile(pf.LogPath(dir, "envtest"), 10)
			for _, line := range out {
				if line == "port=9999 count=2" {
					return true
				}
			}
			return false
		})
	})

	It("injects the captured login-shell environment into supervised processes", func() {
		// The fake shell exports only INJECTED_MARKER (no PATH), so the child
		// keeps its real PATH and `sh` still resolves while the captured var
		// flows through as the lowest-priority overlay layer.
		restore := pf.SetShellForHome(func(string) string {
			return writeFakeShell("INJECTED_MARKER=fromloginshell\n")
		})
		defer restore()

		sup, dir := newSupervisor("probe: sh -c 'echo marker=$INJECTED_MARKER; sleep 30'\n", verify.ProcfileConfig{})
		defer sup.Shutdown()

		waitFor(3*time.Second, func() bool {
			out, _ := utils.TailFile(pf.LogPath(dir, "probe"), 10)
			for _, line := range out {
				if line == "marker=fromloginshell" {
					return true
				}
			}
			return false
		})
	})

	It("does not restart a process under the default 'no' policy", func() {
		sup, dir := newSupervisor("once: sh -c 'exit 1'\n", verify.ProcfileConfig{})
		defer sup.Shutdown()

		waitFor(3*time.Second, func() bool {
			st, _ := pf.ReadState(dir)
			p, ok := st.Process("once")
			return ok && p.Status == pf.StatusCrashed
		})
		st, _ := pf.ReadState(dir)
		p, _ := st.Process("once")
		Expect(p.Restarts).To(Equal(0))
	})

	It("restarts a failing process up to maxRestarts under on-failure", func() {
		cfg := verify.ProcfileConfig{RestartPolicy: pf.RestartOnFailure, MaxRestarts: 2}
		sup, dir := newSupervisor("flaky: sh -c 'exit 1'\n", cfg)
		defer sup.Shutdown()

		// Restarts honour a 500ms backoff, so 2 retries take ~1.5s.
		waitFor(8*time.Second, func() bool {
			st, _ := pf.ReadState(dir)
			p, ok := st.Process("flaky")
			return ok && p.Status == pf.StatusCrashed && p.Restarts == 2
		})
	})

	It("stops every process and cleans up on Shutdown", func() {
		sup, dir := newSupervisor("a: sh -c 'sleep 30'\nb: sh -c 'sleep 30'\n", verify.ProcfileConfig{})

		var pids []int
		waitFor(3*time.Second, func() bool {
			st, _ := pf.ReadState(dir)
			pids = nil
			for _, p := range st.Processes {
				if p.Status == pf.StatusRunning && p.PID > 0 {
					pids = append(pids, p.PID)
				}
			}
			return len(pids) == 2
		})

		sup.Shutdown()

		for _, pid := range pids {
			pidCopy := pid
			waitFor(3*time.Second, func() bool { return !utils.ProcessAlive(pidCopy) })
		}
		Expect(pf.StatePath(dir)).NotTo(BeAnExistingFile())
		Expect(pf.SupervisorPidPath(dir)).NotTo(BeAnExistingFile())
	})
})
