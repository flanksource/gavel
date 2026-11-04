package comments_test

import (
	"github.com/flanksource/gavel/comments"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestComments(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "Comments") }

var _ = Describe("workflow outcomes", func() {
	user := comments.Actor{Kind: "user", ID: "reviewer"}
	agent := comments.Actor{Kind: "agent", ID: "worker"}
	It("keeps proposed playground resolution separate from a final native PR resolution", func() {
		status, phase, err := comments.HumanReview().Apply("open", "resolve", agent)
		Expect(err).NotTo(HaveOccurred())
		Expect([]string{status, string(phase)}).To(Equal([]string{"resolved", "awaiting_outcome"}))
		status, phase, err = comments.NativePR().Apply("open", "resolve", user)
		Expect(err).NotTo(HaveOccurred())
		Expect([]string{status, string(phase)}).To(Equal([]string{"resolved", "final"}))
	})
	It("requires human closure and an explicit reopen before resolving closed feedback", func() {
		_, _, err := comments.HumanReview().Apply("resolved", "close", agent)
		Expect(err).To(HaveOccurred())
		_, _, err = comments.HumanReview().Apply("closed", "resolve", user)
		Expect(err).To(HaveOccurred())
		status, phase, err := comments.HumanReview().Apply("closed", "reopen", user)
		Expect(err).NotTo(HaveOccurred())
		Expect([]string{status, string(phase)}).To(Equal([]string{"open", "active"}))
	})
})
