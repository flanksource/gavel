package jira

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("site field catalog", func() {
	It("reuses fresh snapshots and replaces expired or forced snapshots", func() {
		now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		store := CatalogStore{Dir: GinkgoT().TempDir(), Now: func() time.Time { return now }}
		scope := CatalogScope{Site: "site", Account: "account"}
		calls := 0
		fetch := func(context.Context) ([]FieldDefinition, error) {
			calls++
			return []FieldDefinition{{ID: "field", Name: "Visible"}}, nil
		}
		for range 2 {
			_, err := store.Resolve(context.Background(), scope, false, fetch)
			Expect(err).NotTo(HaveOccurred())
		}
		Expect(calls).To(Equal(1))
		now = now.Add(25 * time.Hour)
		_, err := store.Resolve(context.Background(), scope, false, fetch)
		Expect(err).NotTo(HaveOccurred())
		Expect(calls).To(Equal(2))
		_, err = store.Resolve(context.Background(), scope, true, fetch)
		Expect(err).NotTo(HaveOccurred())
		Expect(calls).To(Equal(3))
	})

	It("isolates sites and accounts and propagates refresh errors", func() {
		store := CatalogStore{Dir: GinkgoT().TempDir()}
		fetch := func(context.Context) ([]FieldDefinition, error) {
			return []FieldDefinition{{ID: "field", Name: "Visible"}}, nil
		}
		scope := CatalogScope{Site: "site-a", Account: "account-a"}
		_, err := store.Resolve(context.Background(), scope, false, fetch)
		Expect(err).NotTo(HaveOccurred())
		failed := func(context.Context) ([]FieldDefinition, error) { return nil, errors.New("fetch failed") }
		for _, other := range []CatalogScope{{Site: "site-b", Account: "account-a"}, {Site: "site-a", Account: "account-b"}} {
			_, err := store.Resolve(context.Background(), other, false, failed)
			Expect(err).To(MatchError(ContainSubstring("fetch failed")))
		}
		_, err = store.Resolve(context.Background(), scope, true, failed)
		Expect(err).To(MatchError(ContainSubstring("fetch failed")))
	})

	It("serializes concurrent cache fills", func() {
		store := CatalogStore{Dir: GinkgoT().TempDir()}
		var calls atomic.Int32
		var group sync.WaitGroup
		results := make(chan error, 8)
		for range 8 {
			group.Add(1)
			go func() {
				defer group.Done()
				_, err := store.Resolve(context.Background(), CatalogScope{Site: "site", Account: "account"}, false, func(context.Context) ([]FieldDefinition, error) {
					calls.Add(1)
					return []FieldDefinition{{ID: "field", Name: "Visible"}}, nil
				})
				results <- err
			}()
		}
		group.Wait()
		close(results)
		for err := range results {
			Expect(err).NotTo(HaveOccurred())
		}
		Expect(calls.Load()).To(Equal(int32(1)))
	})

	It("reports corrupt snapshots and allows an explicit rebuild", func() {
		store := CatalogStore{Dir: GinkgoT().TempDir()}
		scope := CatalogScope{Site: "site", Account: "account"}
		fetch := func(context.Context) ([]FieldDefinition, error) {
			return []FieldDefinition{{ID: "field", Name: "Visible"}}, nil
		}
		_, err := store.Resolve(context.Background(), scope, false, fetch)
		Expect(err).NotTo(HaveOccurred())
		files, err := filepath.Glob(filepath.Join(store.Dir, "*.json"))
		Expect(err).NotTo(HaveOccurred())
		Expect(files).To(HaveLen(1))
		Expect(os.WriteFile(files[0], []byte("{"), 0o600)).To(Succeed())
		_, err = store.Resolve(context.Background(), scope, false, fetch)
		Expect(err).To(MatchError(ContainSubstring("decode Jira catalog")))
		_, err = store.Resolve(context.Background(), scope, true, fetch)
		Expect(err).NotTo(HaveOccurred())
	})
})
