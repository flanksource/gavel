package azuredevops

import (
	"context"
	"encoding/base64"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("Azure authentication", func() {
	It("uses a configured PAT without calling Azure CLI", func() {
		a := authenticator{pat: "fixture-pat", fetch: func(context.Context) (accessToken, error) {
			Fail("PAT authentication must not invoke Azure CLI")
			return accessToken{}, nil
		}}
		header, err := a.authorization(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(header).To(Equal("Basic " + base64.StdEncoding.EncodeToString([]byte(":fixture-pat"))))
	})
	It("reuses an Entra token and refreshes before expiration", func() {
		now := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
		calls := 0
		a := authenticator{now: func() time.Time { return now }, fetch: func(context.Context) (accessToken, error) {
			calls++
			return accessToken{Token: fmt.Sprintf("fixture-token-%d", calls), Expires: now.Add(5 * time.Minute)}, nil
		}}
		for range 2 {
			header, err := a.authorization(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(header).To(Equal("Bearer fixture-token-1"))
		}
		now = now.Add(4*time.Minute + time.Second)
		header, err := a.authorization(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(header).To(Equal("Bearer fixture-token-2"))
		Expect(calls).To(Equal(2))
	})
	It("surfaces CLI authentication failure", func() {
		a := authenticator{now: time.Now, fetch: func(context.Context) (accessToken, error) {
			return accessToken{}, fmt.Errorf("login required")
		}}
		_, err := a.authorization(context.Background())
		Expect(err).To(MatchError(ContainSubstring("login required")))
	})
	It("rejects empty or already expired credentials", func() {
		for _, token := range []accessToken{{}, {Token: "expired", Expires: time.Now().Add(-time.Minute)}} {
			a := authenticator{now: time.Now, fetch: func(context.Context) (accessToken, error) { return token, nil }}
			_, err := a.authorization(context.Background())
			Expect(err).To(HaveOccurred())
		}
	})
})
