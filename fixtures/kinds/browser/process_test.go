package browser

import (
	"context"
	"errors"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("managed browser process", func() {
	It("never starts after cancellation", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := startManagedProcess(ctx, "sleep", []string{"10"}, "", nil, nil)
		Expect(errors.Is(err, context.Canceled)).To(BeTrue())
	})

	It("has a PID before returning and reaps its process group", func() {
		process, err := startManagedProcess(context.Background(), "sleep", []string{"10"}, "", map[string]string{"PATH": os.Getenv("PATH")}, nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(process.pid()).To(BeNumerically(">", 0))
		Expect(process.killTree()).To(Succeed())
		Expect(process.waitWithin(time.Second)).To(Succeed())
	})
})
