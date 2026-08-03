package procfile_test

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	pf "github.com/flanksource/gavel/procfile"
)

const sequentialStartProcfile = "storybook: sh -c 'sleep 30'\nsink: sh -c 'sleep 30'\n"

func writeSequentialStartProcfile() (string, string) {
	root := GinkgoT().TempDir()
	path := filepath.Join(root, "Procfile")
	Expect(os.WriteFile(path, []byte(sequentialStartProcfile), 0o644)).To(Succeed())
	return root, path
}

var _ = Describe("Manager named starts", func() {
	It("registers every process so a later named start succeeds", func() {
		root, path := writeSequentialStartProcfile()
		supervisor, err := pf.NewSupervisor(pf.Options{
			Root:       root,
			Procfile:   path,
			StartNames: []string{"storybook"},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(supervisor.Start()).To(Succeed())
		DeferCleanup(supervisor.Shutdown)

		waitFor(3*time.Second, func() bool {
			return statusOf(supervisor, "storybook") == pf.StatusRunning
		})
		Expect(supervisor.State().Processes).To(HaveLen(2))
		Expect(statusOf(supervisor, "sink")).To(Equal(pf.StatusStopped))

		_, err = pf.Start(root, "", []string{"sink"}, "")
		Expect(err).NotTo(HaveOccurred())
		waitFor(3*time.Second, func() bool {
			return statusOf(supervisor, "sink") == pf.StatusRunning
		})
	})

	It("keeps foreground registration limited to the named subset", func() {
		root, path := writeSequentialStartProcfile()
		supervisor, err := pf.NewSupervisor(pf.Options{
			Root:       root,
			Procfile:   path,
			Names:      []string{"storybook"},
			Foreground: true,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(supervisor.Start()).To(Succeed())
		DeferCleanup(supervisor.Shutdown)

		waitFor(3*time.Second, func() bool {
			return statusOf(supervisor, "storybook") == pf.StatusRunning
		})
		Expect(supervisor.State().Processes).To(HaveLen(1))
		Expect(supervisor.State().Processes[0].Name).To(Equal("storybook"))
	})

	It("rejects an unknown initial process", func() {
		root, path := writeSequentialStartProcfile()
		_, err := pf.NewSupervisor(pf.Options{
			Root:       root,
			Procfile:   path,
			StartNames: []string{"missing"},
		})
		Expect(err).To(MatchError(ContainSubstring("missing")))
	})

	It("rejects conflicting process selections", func() {
		root, path := writeSequentialStartProcfile()
		_, err := pf.NewSupervisor(pf.Options{
			Root:       root,
			Procfile:   path,
			Names:      []string{"storybook"},
			StartNames: []string{"sink"},
		})
		Expect(err).To(MatchError(ContainSubstring("cannot combine")))
	})
})
