package shims

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Install", func() {
	var shimDir, realDir string
	settings := Settings{Timeout: "30s", CPU: 200, Mem: "2GB", Nice: 10}

	writeFake := func(dir, name string) string {
		path := filepath.Join(dir, name)
		Expect(os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755)).To(Succeed())
		return path
	}

	BeforeEach(func() {
		shimDir = filepath.Join(GinkgoT().TempDir(), "shims")
		realDir = GinkgoT().TempDir()
		GinkgoT().Setenv("PATH", shimDir+string(os.PathListSeparator)+realDir)
	})

	install := func(force bool) (InstallReport, error) {
		return Install(InstallOptions{Dir: shimDir, GavelPath: "/opt/gavel's bin/gavel", Tools: []string{"golangci-lint"}, Settings: settings, Force: force})
	}

	It("writes an executable shim that resolves past itself to the real binary", func() {
		real := writeFake(realDir, "golangci-lint")

		report, err := install(false)
		Expect(err).NotTo(HaveOccurred())

		shim := filepath.Join(shimDir, "golangci-lint")
		Expect(report).To(Equal(InstallReport{Dir: shimDir, Shims: []Result{{
			Tool: "golangci-lint", Shim: shim, Action: "created", Real: real, Active: true,
		}}}))
		content, err := os.ReadFile(shim)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(Equal("#!/bin/sh\n# gavel-shim tool=golangci-lint\n" +
			"GAVEL_SHIM_TIMEOUT='30s' GAVEL_SHIM_CPU='200' GAVEL_SHIM_MEM='2GB' GAVEL_SHIM_NICE=10 " +
			"exec '/opt/gavel'\\''s bin/gavel' lint shim-exec golangci-lint \"$@\"\n"))
		info, err := os.Stat(shim)
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Mode().Perm()).To(Equal(os.FileMode(0o755)))
		Expect(IsShim(shim)).To(BeTrue())
		Expect(IsShim(real)).To(BeFalse())
		Expect(LookPath("golangci-lint")).To(Equal(real))
	})

	It("rewrites an existing shim", func() {
		_, err := install(false)
		Expect(err).NotTo(HaveOccurred())
		report, err := install(false)
		Expect(err).NotTo(HaveOccurred())
		Expect(report.Shims[0].Action).To(Equal("updated"))
	})

	It("refuses to overwrite a file that is not a shim unless forced", func() {
		Expect(os.MkdirAll(shimDir, 0o755)).To(Succeed())
		writeFake(shimDir, "golangci-lint")

		_, err := install(false)
		Expect(err).To(MatchError(ContainSubstring("is not a gavel shim")))

		_, err = install(true)
		Expect(err).NotTo(HaveOccurred())
		Expect(IsShim(filepath.Join(shimDir, "golangci-lint"))).To(BeTrue())
	})

	It("reports a missing real binary and prints a PATH hint when the shim is shadowed", func() {
		writeFake(realDir, "eslint")
		GinkgoT().Setenv("PATH", realDir+string(os.PathListSeparator)+shimDir)

		report, err := Install(InstallOptions{Dir: shimDir, GavelPath: "/gavel", Tools: []string{"eslint", "tsc"}, Settings: settings})
		Expect(err).NotTo(HaveOccurred())
		Expect(report.Shims[0].Active).To(BeFalse())
		Expect(report.Shims[1].Real).To(BeEmpty())
		Expect(report.Shims[1].Note).To(ContainSubstring("tsc is not installed"))
		Expect(report.PathHint).To(Equal("export PATH=\"" + shimDir + "\":$PATH"))
		_, err = LookPath("tsc")
		Expect(err).To(MatchError(ContainSubstring("real tsc not found on PATH (only gavel shims)")))
	})

	It("writes nothing on a dry run", func() {
		report, err := Install(InstallOptions{Dir: shimDir, GavelPath: "/gavel", Tools: []string{"tsc"}, Settings: settings, DryRun: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(report.Shims[0].Action).To(Equal("would be created"))
		Expect(shimDir).NotTo(BeADirectory())
	})

	DescribeTable("rejects invalid settings",
		func(tools []string, s Settings, message string) {
			_, err := Install(InstallOptions{Dir: shimDir, GavelPath: "/gavel", Tools: tools, Settings: s})
			Expect(err).To(MatchError(ContainSubstring(message)))
		},
		Entry("unknown tool", []string{"pylint"}, settings, `unknown tool "pylint"`),
		Entry("bad timeout", []string{"tsc"}, Settings{Timeout: "soon", CPU: 200}, "timeout must be a nonnegative duration"),
		Entry("negative cpu", []string{"tsc"}, Settings{Timeout: "30s", CPU: -1}, "cpu limit must be nonnegative"),
		Entry("bad memory", []string{"tsc"}, Settings{Timeout: "30s", Mem: "lots"}, "mem limit"),
		Entry("nice out of range", []string{"tsc"}, Settings{Timeout: "30s", Nice: 40}, "nice must be between"),
	)

	DescribeTable("SettingsFromEnv",
		func(env map[string]string, want Settings, message string) {
			got, err := SettingsFromEnv(func(key string) string { return env[key] })
			if message != "" {
				Expect(err).To(MatchError(ContainSubstring(message)))
				return
			}
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(want))
		},
		Entry("all set", map[string]string{EnvTimeout: "30s", EnvCPU: "200", EnvMem: "2GB", EnvNice: "10"}, settings, ""),
		Entry("memory is optional", map[string]string{EnvTimeout: "0", EnvCPU: "0", EnvNice: "0"}, Settings{Timeout: "0"}, ""),
		Entry("missing limits fail", map[string]string{EnvTimeout: "30s"}, Settings{}, "missing [GAVEL_SHIM_CPU GAVEL_SHIM_NICE]"),
	)
})
