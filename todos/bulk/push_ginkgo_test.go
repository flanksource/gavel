package bulk

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bulk push to GitHub", func() {
	It("rejects the whole batch when no base URL resolver is wired", func() {
		_, err := Push(PushFlags{}, nil)
		Expect(err).To(MatchError(ContainSubstring("base URL resolver")))
	})

	It("resolves the base URL against each TODO's own workspace and fails that item when it is invalid", func() {
		const ownerDir, requested = "/repos/misconfigured", "https://gavel.example.com"
		var asked [][2]string
		fn, err := Push(PushFlags{BaseURL: requested}, func(dir, got string) (string, error) {
			asked = append(asked, [2]string{dir, got})
			return "", errors.New("invalid base URL")
		})
		Expect(err).NotTo(HaveOccurred())

		todo := todoRef("a1")
		todo.CWD = ownerDir
		_, err = fn(context.Background(), nil, todo)

		Expect(err).To(MatchError(ContainSubstring("invalid base URL")))
		Expect(asked).To(Equal([][2]string{{ownerDir, requested}}))
	})
})
