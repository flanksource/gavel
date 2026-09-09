package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/fixtures"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("test ansi capture output", func() {
	It("atomically replaces the destination with valid capture JSON", func() {
		capture, err := fixtures.CaptureANSI(fixtures.CaptureOptions{
			Width:     20,
			Height:    4,
			Command:   []string{"/bin/sh", "-c", "printf 'captured\\n'"},
			Snapshots: true,
		})
		Expect(err).NotTo(HaveOccurred())

		out := filepath.Join(GinkgoT().TempDir(), "capture.json")
		Expect(os.WriteFile(out, []byte("stale"), 0o644)).To(Succeed())
		Expect(writeANSICapture(out, capture)).To(Succeed())

		data, err := os.ReadFile(out)
		Expect(err).NotTo(HaveOccurred())
		var decoded struct {
			Final fixtures.FinalState `json:"final"`
		}
		Expect(json.Unmarshal(data, &decoded)).To(Succeed())
		Expect(decoded.Final.Screen).To(ContainSubstring("captured"))
		matches, err := filepath.Glob(filepath.Join(filepath.Dir(out), ".capture.json.*.tmp"))
		Expect(err).NotTo(HaveOccurred())
		Expect(matches).To(BeEmpty())
	})

	It("reports the widest unwrapped line without replaying captured events", func() {
		capture, err := fixtures.CaptureANSI(fixtures.CaptureOptions{
			Width:     5,
			Height:    2,
			Command:   []string{"/bin/sh", "-c", "printf '123456789\\n'"},
			Snapshots: true,
		})
		Expect(err).NotTo(HaveOccurred())

		summary := ansiCaptureSummary(capture, "capture.json")
		Expect(summary.MaxLineWidth).To(Equal(9))
		Expect(summary.WidthOverflow).To(BeTrue())
	})
})
