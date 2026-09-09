package service

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("userShellInvocation", func() {
	It("runs gavel through the user's login and interactive shell", func() {
		Expect(userShellInvocation("/bin/zsh", "/Users/example/go/bin/gavel")).To(Equal([]string{
			"/bin/zsh",
			"-l",
			"-i",
			"-c",
			`exec "$@"`,
			"gavel-system",
			"/Users/example/go/bin/gavel",
			"pr",
			"list",
			"--all",
			"--ui",
			"--menu-bar",
			"--port=0",
			"--persist-port",
		}))
	})
})

var _ = Describe("renderUserUnit", func() {
	It("renders the shell invocation as a user service", func() {
		unit, err := renderUserUnit("/bin/zsh", "/home/example/go/bin/gavel", "/home/example")
		Expect(err).NotTo(HaveOccurred())
		Expect(unit).To(Equal(`[Unit]
Description=Gavel PR UI (pr list --all --ui)
After=default.target

[Service]
Type=simple
WorkingDirectory="/home/example"
ExecStart="/bin/zsh" "-l" "-i" "-c" "exec \"$$@\"" "gavel-system" "/home/example/go/bin/gavel" "pr" "list" "--all" "--ui" "--menu-bar" "--port=0" "--persist-port"
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=default.target
`))
	})
})

var _ = Describe("userShellPath", func() {
	It("uses the current user's configured shell", func() {
		executable, err := os.Executable()
		Expect(err).NotTo(HaveOccurred())
		GinkgoT().Setenv("SHELL", executable)

		Expect(userShellPath()).To(Equal(executable))
	})

	It("fails when the user shell is unavailable", func() {
		GinkgoT().Setenv("SHELL", "")

		_, err := userShellPath()
		Expect(err).To(MatchError("resolve user shell: SHELL is not set"))
	})
})
