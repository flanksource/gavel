package procfile_test

import (
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	pf "github.com/flanksource/gavel/procfile"
)

func TestProcfile(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Procfile Suite")
}

// restoreShell resets the login-shell resolver after the suite.
var restoreShell func()

// By default every supervisor created in this suite resolves its login shell to
// `true`, which prints nothing — so LoadUserEnv contributes an empty layer and
// existing specs behave as if no shell env were captured (and no real,
// possibly-slow interactive shell is spawned). Specs that exercise env
// injection override the resolver locally.
var _ = BeforeSuite(func() {
	truePath, err := exec.LookPath("true")
	Expect(err).NotTo(HaveOccurred())
	restoreShell = pf.SetShellForHome(func(string) string { return truePath })
})

var _ = AfterSuite(func() {
	if restoreShell != nil {
		restoreShell()
	}
})
