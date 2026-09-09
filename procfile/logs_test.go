package procfile_test

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	pf "github.com/flanksource/gavel/procfile"
)

// lockedBuffer is a goroutine-safe io.Writer so the spec can read what Stream
// has written so far while Stream is still writing from its own goroutine.
type lockedBuffer struct {
	mu  sync.Mutex
	buf []byte
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buf = append(b.buf, p...)
	return len(p), nil
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return string(b.buf)
}

var _ = Describe("Stream", func() {
	It("follows new log output, prefixed per process, until the context is cancelled", func() {
		root := GinkgoT().TempDir()
		procfilePath := filepath.Join(root, "Procfile")
		Expect(os.WriteFile(procfilePath,
			[]byte("ticker: sh -c 'echo started; for i in 1 2 3 4 5 6 7 8; do echo tick-$i; sleep 0.2; done; sleep 30'\n"),
			0o644)).To(Succeed())

		sup, err := pf.NewSupervisor(pf.Options{Root: root, Procfile: procfilePath})
		Expect(err).NotTo(HaveOccurred())
		Expect(sup.Start()).To(Succeed())
		defer sup.Shutdown()

		buf := &lockedBuffer{}
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)
		go func() { done <- pf.Stream(ctx, root, procfilePath, nil, 5, buf) }()

		// The "ticker |" prefix proves per-process labelling; a later tick proves
		// the loop is following growth, not just the seeded tail.
		Eventually(buf.String, 6*time.Second, 50*time.Millisecond).Should(ContainSubstring("ticker"))
		Eventually(buf.String, 6*time.Second, 50*time.Millisecond).Should(ContainSubstring("tick-7"))

		cancel()
		Eventually(done, 3*time.Second).Should(Receive(BeNil()))
	})
})
