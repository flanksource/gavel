package runtime

import (
	"encoding/json"
	"time"

	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("providerEvents", func() {
	It("carries each event's structured payload to the provider event", func() {
		id := uuid.MustParse("0123abcd-0000-4000-8000-000000000001")
		at := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
		payload := json.RawMessage(`{"via":"merge","targetBranch":"main","landedSha":"abc1234"}`)

		events := providerEvents([]native.Event{{ID: id, Kind: "run_landed", Actor: "tester", Payload: payload, CreatedAt: at}})

		Expect(events).To(Equal([]types.ProviderEvent{{
			ID: id.String(), ShortID: "0123abcd", Kind: "run_landed", Actor: "tester",
			Timestamp: at, Title: "run_landed", Payload: payload,
		}}))
	})
})
