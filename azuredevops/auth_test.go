// Tests the az CLI login's token cache: a token is reused until it nears expiry, a failed or
// unusable login is reported.
package azuredevops

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("The az CLI login", func() {
	It("reuses an Entra token and refreshes it before it expires", func() {
		now := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
		calls := 0
		a := authenticator{now: func() time.Time { return now }, fetch: func(context.Context) (accessToken, error) {
			calls++
			return accessToken{Token: fmt.Sprintf("fixture-token-%d", calls), Expires: now.Add(5 * time.Minute)}, nil
		}}
		for range 2 {
			token, err := a.cliToken(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(token).To(Equal("fixture-token-1"))
		}
		now = now.Add(4*time.Minute + time.Second)
		token, err := a.cliToken(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(token).To(Equal("fixture-token-2"))
		Expect(calls).To(Equal(2))
	})
	It("surfaces the CLI's authentication failure", func() {
		a := authenticator{now: time.Now, fetch: func(context.Context) (accessToken, error) {
			return accessToken{}, fmt.Errorf("login required")
		}}
		_, err := a.cliToken(context.Background())
		Expect(err).To(MatchError(ContainSubstring("login required")))
	})
	It("rejects empty or already expired credentials", func() {
		for _, token := range []accessToken{{}, {Token: "expired", Expires: time.Now().Add(-time.Minute)}} {
			a := authenticator{now: time.Now, fetch: func(context.Context) (accessToken, error) { return token, nil }}
			_, err := a.cliToken(context.Background())
			Expect(err).To(HaveOccurred())
		}
	})
})
