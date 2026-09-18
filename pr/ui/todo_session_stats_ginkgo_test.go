package ui

import (
	"context"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("terminal question session statistics", func() {
	It("shows ask without a transcript when the latest prompt run is waiting for an answer", func() {
		root := &captaindb.Session{ID: uuid.New(), Source: "gavel"}
		store := fakeCaptainSessionStore{run: root, runs: []captaindb.PromptRun{{
			SessionID: root.ID, State: captaindb.PromptRunStateWaiting,
			ResultJSON: map[string]any{"endStatus": "ask"},
		}}}

		stats, err := captainSessionStats(context.Background(), store, "provider-1", captainSessionResolution{run: root, known: true})

		Expect(err).NotTo(HaveOccurred())
		Expect(stats.State).To(Equal("ask"))
		Expect(stats.InProgress).To(BeFalse())
	})
})
