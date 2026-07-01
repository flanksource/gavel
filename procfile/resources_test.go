package procfile_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	pf "github.com/flanksource/gavel/procfile"
	"github.com/flanksource/gavel/verify"
)

// procByName returns the live state of the named process from the supervisor's
// in-memory snapshot (which carries resource samples, unlike persisted state).
func procByName(sup *pf.Supervisor, name string) (pf.ProcState, bool) {
	for _, p := range sup.State().Processes {
		if p.Name == name {
			return p, true
		}
	}
	return pf.ProcState{}, false
}

var _ = Describe("Supervisor resource tracking", func() {
	It("samples memory and open files for a running process", func() {
		// A busy loop stays resident so the monitor has real RSS/FDs to read.
		sup, _ := newSupervisor("busy: sh -c 'while :; do :; done'\n", verify.ProcfileConfig{})
		defer sup.Shutdown()

		var p pf.ProcState
		// First sample lands one monitor interval (~2s) after start.
		waitFor(8*time.Second, func() bool {
			ps, ok := procByName(sup, "busy")
			if ok && ps.MemoryRSS > 0 {
				p = ps
				return true
			}
			return false
		})
		Expect(p.MemoryRSS).To(BeNumerically(">", uint64(0)))
		// A live process always has open files; OpenFiles is -1 only where the
		// platform cannot report it — never a real 0.
		Expect(p.OpenFiles).ToNot(Equal(0))

		// The per-process tree is wired through from the supervisor's snapshot.
		Expect(p.Tree).ToNot(BeEmpty())
		Expect(p.Tree[0].PID).To(BeNumerically(">", 0))
		Expect(p.Tree[0].MemoryRSS).To(BeNumerically(">", uint64(0)))

		// CPU% needs a second sample to compute a delta; the busy loop pegs a core.
		waitFor(12*time.Second, func() bool {
			ps, _ := procByName(sup, "busy")
			return ps.CPUPercent > 0
		})
	})

	It("kills a process that exceeds its memory limit", func() {
		// 1 byte: any live process exceeds it on the first sample, so the monitor
		// kills it. With the default 'no' restart policy it settles as crashed.
		cfg := verify.ProcfileConfig{Mem: "1"}
		sup, _ := newSupervisor("hog: sh -c 'sleep 60'\n", cfg)
		defer sup.Shutdown()

		waitFor(8*time.Second, func() bool {
			ps, ok := procByName(sup, "hog")
			return ok && (ps.Status == pf.StatusCrashed || ps.Status == pf.StatusExited)
		})
	})
})
