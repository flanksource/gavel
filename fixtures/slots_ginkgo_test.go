package fixtures

import (
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/gavel/processruntime"
	"github.com/flanksource/gavel/slots"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Slot-aware fixture backstops", func() {
	It("still cuts off a kind that ignores its context", func() {
		DeferCleanup(DefaultKinds.Replace(stubbornKind{name: "slots-stubborn", hold: time.Second}))
		dir := GinkgoT().TempDir()
		path := filepath.Join(dir, "backstop.fixture.yaml")
		Expect(os.WriteFile(path, []byte("kind: Fixture\nsteps:\n  - {id: hang, kind: slots-stubborn, timeout: 50ms}\n"), 0o600)).To(Succeed())
		runner, err := NewRunner(RunnerOptions{Paths: []string{path}, WorkDir: dir, ProcessRuntime: processruntime.Options{Slots: slots.Limits{Global: 1}}})
		Expect(err).NotTo(HaveOccurred())
		runner.backstopGrace = 150 * time.Millisecond
		tree, err := runner.Run()
		Expect(err).To(HaveOccurred())
		Expect(tree.GetStats().TimedOut).To(Equal(1))
	})
})
