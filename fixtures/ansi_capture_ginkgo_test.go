package fixtures

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ANSI screen capture", func() {
	It("scrolls output within the configured viewport height", func() {
		screen := newANSIScreen(8, 3)
		screen.Write([]byte("one\ntwo\nthree\nfour"))

		Expect(screen.String()).To(Equal("two\nthree\nfour"))
	})

	It("continues control sequences split across reads", func() {
		screen := newANSIScreen(20, 4)
		screen.Write([]byte("old\n"))
		screen.Write([]byte("\x1b[1"))
		screen.Write([]byte("A\x1b[2"))
		screen.Write([]byte("Knew\x1b]ignored"))
		screen.Write([]byte(" title\x1b"))
		screen.Write([]byte("\\"))

		Expect(screen.String()).To(Equal("new\n"))
	})

	It("bounds every periodic snapshot to the configured viewport", func() {
		const (
			width  = 20
			height = 4
			lines  = 120
		)
		capture, err := CaptureANSI(CaptureOptions{
			Width:            width,
			Height:           height,
			SnapshotInterval: 2 * time.Millisecond,
			Snapshots:        true,
			Command: []string{
				"/bin/sh", "-c",
				fmt.Sprintf("i=0; while [ $i -lt %d ]; do printf 'line-%%03d\\n' \"$i\"; sleep 0.002; i=$((i+1)); done", lines),
			},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(capture.Snapshots)).To(BeNumerically(">", 1))

		for _, snapshot := range capture.Snapshots {
			Expect(strings.Count(snapshot.Screen, "\n") + 1).To(BeNumerically("<=", height))
			Expect(len(snapshot.Screen)).To(BeNumerically("<=", height*(width+1)))
		}
		Expect(capture.Final.Screen).To(ContainSubstring(fmt.Sprintf("line-%03d", lines-1)))
		Expect(capture.Final.Screen).NotTo(ContainSubstring("line-000"))
		Expect(capture.MaxLineWidth()).To(Equal(len("line-000")))
	})
})
