package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TailFile", func() {
	var path string

	BeforeEach(func() {
		path = filepath.Join(GinkgoT().TempDir(), "out.log")
	})

	write := func(lines int) {
		var b strings.Builder
		for i := 1; i <= lines; i++ {
			fmt.Fprintf(&b, "line %d\n", i)
		}
		Expect(os.WriteFile(path, []byte(b.String()), 0o644)).To(Succeed())
	}

	It("returns nil for a non-positive line count", func() {
		write(5)
		got, err := TailFile(path, 0)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(BeNil())
	})

	It("treats a missing file as empty", func() {
		got, err := TailFile(filepath.Join(GinkgoT().TempDir(), "absent.log"), 10)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(BeNil())
	})

	It("returns all lines when the file has fewer than n", func() {
		write(3)
		got, err := TailFile(path, 10)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal([]string{"line 1", "line 2", "line 3"}))
	})

	It("returns only the last n lines", func() {
		write(100)
		got, err := TailFile(path, 3)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal([]string{"line 98", "line 99", "line 100"}))
	})

	It("reads the tail correctly across multiple 8KiB chunks", func() {
		// 5000 lines of "line N\n" is ~40KB, forcing several chunk reads.
		write(5000)
		got, err := TailFile(path, 2)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal([]string{"line 4999", "line 5000"}))
	})

	It("ignores a trailing newline when counting lines", func() {
		Expect(os.WriteFile(path, []byte("a\nb\n"), 0o644)).To(Succeed())
		got, err := TailFile(path, 5)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal([]string{"a", "b"}))
	})
})

var _ = Describe("ProcessAlive", func() {
	It("returns false for a non-positive pid", func() {
		Expect(ProcessAlive(0)).To(BeFalse())
		Expect(ProcessAlive(-1)).To(BeFalse())
	})

	It("returns true for the current process", func() {
		Expect(ProcessAlive(os.Getpid())).To(BeTrue())
	})

	It("returns false for a pid that has already exited", func() {
		cmd := exec.Command("sh", "-c", "exit 0")
		Expect(cmd.Start()).To(Succeed())
		pid := cmd.Process.Pid
		Expect(cmd.Wait()).To(Succeed())
		Expect(ProcessAlive(pid)).To(BeFalse())
	})
})
