package clicky

import (
	"errors"

	"github.com/flanksource/gavel/comments"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("localUser", func() {
	const uid = 501
	git := func(name string, set bool, err error) func(string) (string, bool, error) {
		return func(key string) (string, bool, error) {
			Expect(key).To(Equal("user.name"))
			return name, set, err
		}
	}

	It("labels the OS account with git's user.name", func() {
		actor, err := localUser(uid, git("Ada Lovelace", true, nil))
		Expect(err).NotTo(HaveOccurred())
		Expect(actor).To(Equal(comments.Actor{Kind: "user", ID: "local:501", Label: "Ada Lovelace"}))
	})
	It("labels it You while git has no user.name", func() {
		actor, err := localUser(uid, git("", false, nil))
		Expect(err).NotTo(HaveOccurred())
		Expect(actor).To(Equal(comments.Actor{Kind: "user", ID: "local:501", Label: "You"}))
	})
	It("fails when git cannot be read", func() {
		_, err := localUser(uid, git("", false, errors.New("git: executable file not found")))
		Expect(err).To(MatchError(ContainSubstring("git identity")))
	})
})
