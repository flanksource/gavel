package jira

import (
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Jira sync selection", func() {
	It("compares cutoff instants inclusively across timezones and requires updated data", func() {
		selection, err := (SyncOptions{Since: "2026-11-01T00:00:00Z"}).selection(time.Now())
		Expect(err).NotTo(HaveOccurred())
		Expect(selection.resolve(nil)).To(Succeed())
		for _, test := range []struct {
			updated string
			matches bool
		}{
			{"2026-11-01T02:00:00+02:00", true},
			{"2026-10-31T23:59:59Z", false},
			{"2026-11-01T00:00:01Z", true},
		} {
			raw, err := json.Marshal(map[string]string{"updated": test.updated})
			Expect(err).NotTo(HaveOccurred())
			matches, err := selection.matches(raw)
			Expect(err).NotTo(HaveOccurred())
			Expect(matches).To(Equal(test.matches))
		}
		_, err = selection.matches(json.RawMessage(`{"updated":null}`))
		Expect(err).To(HaveOccurred())
	})
	It("validates targets and builds inclusive range queries", func() {
		query, err := syncTargetQuery("APP-10..APP-50")
		Expect(err).NotTo(HaveOccurred())
		Expect(query).To(Equal(`project = "APP" AND key >= "APP-10" AND key <= "APP-50"`))
		for _, target := range []string{"APP-50..APP-10", "APP-1..OPS-2", "APP-1..", "APP OR 1=1"} {
			_, err := syncTargetQuery(target)
			Expect(err).To(HaveOccurred(), target)
		}
	})
	It("resolves a fixed cutoff and rejects invalid durations", func() {
		now := time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)
		cutoff, err := parseSyncSince("7d", now)
		Expect(err).NotTo(HaveOccurred())
		Expect(cutoff).To(Equal(now.Add(-7 * 24 * time.Hour)))
		for _, value := range []string{"-7d", "0h", "nonsense"} {
			_, err := parseSyncSince(value, now)
			Expect(err).To(HaveOccurred())
		}
	})
	It("matches readable lists with exclusion precedence and merges same-field patterns", func() {
		selection, err := (SyncOptions{Filters: []string{"Countries=Ken*", "Countries=Uganda,!Blocked"}, Statuses: []string{"!Done"}}).selection(time.Now())
		Expect(err).NotTo(HaveOccurred())
		catalog := map[string]FieldDefinition{"customfield_1": {ID: "customfield_1", Name: "Countries"}}
		Expect(selection.resolve(catalog)).To(Succeed())
		for _, test := range []struct {
			raw     string
			matches bool
		}{
			{`{"status":{"name":"Open"},"customfield_1":[{"value":"Kenya"}]}`, true},
			{`{"status":{"name":"Done"},"customfield_1":[{"value":"Kenya"}]}`, false},
			{`{"status":{"name":"Open"},"customfield_1":[{"value":"Kenya"},{"value":"Blocked"}]}`, false},
			{`{"status":{"name":"Open"},"customfield_1":[{"value":"Uganda"}]}`, true},
			{`{"status":{"name":"Open"}}`, false},
		} {
			matches, err := selection.matches(json.RawMessage(test.raw))
			Expect(err).NotTo(HaveOccurred())
			Expect(matches).To(Equal(test.matches), test.raw)
		}
	})
	It("rejects unknown and ambiguous field names", func() {
		selection, err := (SyncOptions{Filters: []string{"Team=*"}}).selection(time.Now())
		Expect(err).NotTo(HaveOccurred())
		Expect(selection.resolve(nil)).To(MatchError(ContainSubstring("unknown")))
		Expect(selection.resolve(map[string]FieldDefinition{"a": {Name: "Team"}, "b": {Name: "Team"}})).To(MatchError(ContainSubstring("ambiguous")))
	})
})
