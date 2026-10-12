package gitstate_test

import (
	"github.com/flanksource/commons/properties"
	"github.com/flanksource/gavel/git/gitstate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// setProperty sets a commons property for one spec; an empty value reads as
// unset, so the accessor's default applies again afterwards.
func setProperty(key, value string) {
	properties.Set(key, value)
	DeferCleanup(properties.Set, key, "")
}

var _ = Describe("status scan git options", func() {
	DescribeTable("follow the git.fsmonitor and git.untrackedCache properties",
		func(fsmonitor, untrackedCache string, expected []string) {
			setProperty(gitstate.PropertyFSMonitor, fsmonitor)
			setProperty(gitstate.PropertyUntrackedCache, untrackedCache)
			Expect(gitstate.PollGitArgs()).To(Equal(expected))
		},
		Entry("unset: no file monitor, untracked cache on", "", "",
			[]string{"--no-optional-locks", "-c", "core.fsmonitor=false", "-c", "core.untrackedCache=true"}),
		Entry("file monitor enabled", "true", "",
			[]string{"--no-optional-locks", "-c", "core.fsmonitor=true", "-c", "core.untrackedCache=true"}),
		Entry("untracked cache disabled", "", "false",
			[]string{"--no-optional-locks", "-c", "core.fsmonitor=false", "-c", "core.untrackedCache=false"}),
	)
})
