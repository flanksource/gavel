package utils

import (
	"net"
	"os/exec"
	"strconv"
	"syscall"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("parseListenPorts", func() {
	It("extracts ports from IPv4, IPv6 and wildcard addresses", func() {
		out := "p123\nf3\nn*:3000\nf5\nn127.0.0.1:8080\nf7\nn[::1]:5000\n"
		Expect(parseListenPorts(out)).To(Equal([]int{3000, 5000, 8080}))
	})

	It("deduplicates the same port reported on multiple fds", func() {
		out := "p123\nf8\nn*:49152\nf9\nn*:49152\n"
		Expect(parseListenPorts(out)).To(Equal([]int{49152}))
	})

	It("ignores established connections and malformed addresses", func() {
		out := "n127.0.0.1:8080->127.0.0.1:55000\nn*:notaport\nnno-colon-here\nn*:\n"
		Expect(parseListenPorts(out)).To(BeNil())
	})

	It("returns nil for empty output", func() {
		Expect(parseListenPorts("")).To(BeNil())
	})
})

var _ = Describe("ListeningPorts", func() {
	It("returns nil for a non-positive pgid", func() {
		got, err := ListeningPorts(0)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(BeNil())
	})

	It("detects a port opened by the current process group", func() {
		if _, err := exec.LookPath("lsof"); err != nil {
			Skip("lsof not installed")
		}
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		Expect(err).NotTo(HaveOccurred())
		defer ln.Close()
		want := ln.Addr().(*net.TCPAddr).Port

		// The test binary's listeners live in its own process group; lsof -g of
		// that pgid must surface the port we just bound.
		pgid, err := syscall.Getpgid(syscall.Getpid())
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() []int {
			ports, perr := ListeningPorts(pgid)
			Expect(perr).NotTo(HaveOccurred())
			return ports
		}, "3s", "100ms").Should(ContainElement(want), "expected to detect bound port "+strconv.Itoa(want))
	})
})
