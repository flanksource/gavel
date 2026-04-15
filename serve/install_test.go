//go:build linux

package serve

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("renderUnit", func() {
	It("renders all placeholders from options", func() {
		unit, err := renderUnit(InstallOptions{
			Port:       2222,
			Host:       "0.0.0.0",
			User:       "gavel",
			DataDir:    "/var/lib/gavel",
			BinaryPath: "/usr/local/bin/gavel",
		})
		Expect(err).ToNot(HaveOccurred())

		expectedFragments := []string{
			"Description=Gavel SSH git-push backend",
			"After=network.target",
			"User=gavel",
			"Group=gavel",
			"WorkingDirectory=/var/lib/gavel",
			"ExecStart=/usr/local/bin/gavel ssh serve --host 0.0.0.0 --port 2222 --host-key /var/lib/gavel/ssh_host_key --repo-dir /var/lib/gavel/repos",
			"Restart=on-failure",
			"WantedBy=multi-user.target",
		}
		for _, fragment := range expectedFragments {
			Expect(unit).To(ContainSubstring(fragment))
		}

		// Hardening directives must be absent: they break gavel's subprocess
		// workload (go build, go test, $HOME module cache, real /tmp).
		forbidden := []string{
			"NoNewPrivileges",
			"ProtectSystem",
			"ProtectHome",
			"ReadWritePaths",
			"PrivateTmp",
		}
		for _, line := range forbidden {
			Expect(unit).NotTo(ContainSubstring(line),
				"unit must not contain %s — it blocks go subprocesses", line)
		}
	})

	It("substitutes custom user, port and paths", func() {
		unit, err := renderUnit(InstallOptions{
			Port:       9000,
			Host:       "127.0.0.1",
			User:       "ci",
			DataDir:    "/opt/gavel",
			BinaryPath: "/opt/gavel/bin/gavel",
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(unit).To(ContainSubstring("User=ci"))
		Expect(unit).To(ContainSubstring("Group=ci"))
		Expect(unit).To(ContainSubstring("--host 127.0.0.1 --port 9000"))
		Expect(unit).To(ContainSubstring("--host-key /opt/gavel/ssh_host_key"))
		Expect(unit).To(ContainSubstring("--repo-dir /opt/gavel/repos"))
		Expect(unit).To(ContainSubstring("ExecStart=/opt/gavel/bin/gavel ssh serve"))
	})
})

var _ = Describe("writeUnit", func() {
	var unitPath string

	BeforeEach(func() {
		unitPath = filepath.Join(GinkgoT().TempDir(), "gavel-ssh.service")
	})

	It("writes the file when it does not exist", func() {
		Expect(writeUnit(unitPath, "hello", false)).To(Succeed())
		data, err := os.ReadFile(unitPath)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(data)).To(Equal("hello"))
	})

	It("refuses to overwrite without force", func() {
		Expect(os.WriteFile(unitPath, []byte("existing"), 0o644)).To(Succeed())
		err := writeUnit(unitPath, "new", false)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("already exists"))
		data, _ := os.ReadFile(unitPath)
		Expect(string(data)).To(Equal("existing"))
	})

	It("overwrites with force", func() {
		Expect(os.WriteFile(unitPath, []byte("existing"), 0o644)).To(Succeed())
		Expect(writeUnit(unitPath, "new", true)).To(Succeed())
		data, _ := os.ReadFile(unitPath)
		Expect(string(data)).To(Equal("new"))
	})

	It("creates parent directories if missing", func() {
		nested := filepath.Join(GinkgoT().TempDir(), "a", "b", "gavel-ssh.service")
		Expect(writeUnit(nested, "hi", false)).To(Succeed())
		data, err := os.ReadFile(nested)
		Expect(err).ToNot(HaveOccurred())
		Expect(strings.TrimSpace(string(data))).To(Equal("hi"))
	})
})

var _ = Describe("InstallOptions.applyDefaults", func() {
	It("fills in defaults when fields are zero", func() {
		opts := InstallOptions{}
		Expect(opts.applyDefaults()).To(Succeed())
		Expect(opts.Port).To(Equal(2222))
		Expect(opts.Host).To(Equal("0.0.0.0"))
		Expect(opts.User).To(Equal("gavel"))
		Expect(opts.UnitPath).To(Equal("/etc/systemd/system/gavel-ssh.service"))
		Expect(opts.DataDir).To(Equal("/var/lib/gavel"))
		Expect(opts.BinaryPath).ToNot(BeEmpty())
	})

	It("preserves explicitly set values", func() {
		opts := InstallOptions{
			Port:       9000,
			Host:       "127.0.0.1",
			User:       "ci",
			UnitPath:   "/tmp/unit.service",
			DataDir:    "/opt/gavel",
			BinaryPath: "/opt/gavel/bin/gavel",
		}
		Expect(opts.applyDefaults()).To(Succeed())
		Expect(opts.Port).To(Equal(9000))
		Expect(opts.Host).To(Equal("127.0.0.1"))
		Expect(opts.User).To(Equal("ci"))
		Expect(opts.UnitPath).To(Equal("/tmp/unit.service"))
		Expect(opts.DataDir).To(Equal("/opt/gavel"))
		Expect(opts.BinaryPath).To(Equal("/opt/gavel/bin/gavel"))
	})
})
