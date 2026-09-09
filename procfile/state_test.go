package procfile_test

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	pf "github.com/flanksource/gavel/procfile"
)

var _ = Describe("State", func() {
	var root, dir string

	BeforeEach(func() {
		root = GinkgoT().TempDir()
		var err error
		dir, err = pf.StateDir(root)
		Expect(err).NotTo(HaveOccurred())
	})

	It("creates .gavel/proc under the root", func() {
		Expect(dir).To(Equal(filepath.Join(root, ".gavel", "proc")))
		info, err := os.Stat(dir)
		Expect(err).NotTo(HaveOccurred())
		Expect(info.IsDir()).To(BeTrue())
	})

	It("round-trips a state through write and read", func() {
		started := time.Now().Truncate(time.Second).UTC()
		code := 1
		want := pf.State{
			Root:          root,
			Procfile:      filepath.Join(root, "Procfile"),
			SupervisorPID: os.Getpid(),
			Started:       &started,
			Processes: []pf.ProcState{
				{Name: "web", Command: "serve", PID: 1234, Status: pf.StatusRunning, Started: &started, LogFile: pf.LogPath(dir, "web")},
				{Name: "once", Command: "do", Status: pf.StatusCrashed, Restarts: 3, ExitCode: &code, LogFile: pf.LogPath(dir, "once")},
			},
		}

		Expect(pf.WriteState(dir, want)).To(Succeed())
		got, err := pf.ReadState(dir)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(want))
	})

	It("reports a missing state file as a zero, not-running state", func() {
		got, err := pf.ReadState(dir)
		Expect(err).NotTo(HaveOccurred())
		Expect(got.SupervisorPID).To(BeZero())
		Expect(got.Running()).To(BeFalse())
	})

	It("considers the current process alive via Running()", func() {
		Expect(pf.WriteState(dir, pf.State{SupervisorPID: os.Getpid()})).To(Succeed())
		got, err := pf.ReadState(dir)
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Running()).To(BeTrue())
	})

	It("looks up a process by name", func() {
		Expect(pf.WriteState(dir, pf.State{Processes: []pf.ProcState{{Name: "web"}}})).To(Succeed())
		got, _ := pf.ReadState(dir)
		p, ok := got.Process("web")
		Expect(ok).To(BeTrue())
		Expect(p.Name).To(Equal("web"))
		_, ok = got.Process("absent")
		Expect(ok).To(BeFalse())
	})

	It("cleans state and pid files but keeps logs", func() {
		Expect(pf.WriteState(dir, pf.State{SupervisorPID: 1})).To(Succeed())
		Expect(pf.WritePid(pf.SupervisorPidPath(dir), 1)).To(Succeed())
		Expect(pf.WritePid(pf.PidPath(dir, "web"), 2)).To(Succeed())
		Expect(os.WriteFile(pf.LogPath(dir, "web"), []byte("log line\n"), 0o644)).To(Succeed())

		Expect(pf.Clean(dir)).To(Succeed())

		Expect(pf.StatePath(dir)).NotTo(BeAnExistingFile())
		Expect(pf.SupervisorPidPath(dir)).NotTo(BeAnExistingFile())
		Expect(pf.PidPath(dir, "web")).NotTo(BeAnExistingFile())
		Expect(pf.LogPath(dir, "web")).To(BeAnExistingFile())
	})

	It("round-trips a pid file", func() {
		path := pf.PidPath(dir, "worker")
		Expect(pf.WritePid(path, 4321)).To(Succeed())
		pid, err := pf.ReadPid(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(pid).To(Equal(4321))
	})
})
