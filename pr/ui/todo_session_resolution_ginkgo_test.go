package ui

import (
	"context"
	"errors"
	"fmt"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// recordingSessionStore reports which store call the dashboard made, so a spec
// can assert that the transcript hop is delegated to Captain rather than
// re-derived here from a list of candidate rows.
type recordingSessionStore struct {
	run           *captaindb.Session
	transcript    *captaindb.Session
	transcriptErr error
	calls         []string
}

func (s *recordingSessionStore) GetSessionByIdentity(_ context.Context, identity, source, _, _ string) (*captaindb.Session, error) {
	s.calls = append(s.calls, fmt.Sprintf("GetSessionByIdentity(%s,%s)", identity, source))
	if source == "gavel" && s.run != nil {
		return s.run, nil
	}
	return nil, captaindb.ErrSessionNotFound
}

func (s *recordingSessionStore) GetTranscriptSessionByIdentity(_ context.Context, identity string) (*captaindb.Session, error) {
	s.calls = append(s.calls, "GetTranscriptSessionByIdentity("+identity+")")
	if s.transcriptErr != nil {
		return nil, s.transcriptErr
	}
	return s.transcript, nil
}

func (s *recordingSessionStore) GetSessionOverviewByIdentity(context.Context, string) (*captaindb.SessionOverview, error) {
	return nil, captaindb.ErrSessionNotFound
}

func (s *recordingSessionStore) ListTranscriptMessages(context.Context, captaindb.TranscriptPage) ([]captaindb.TranscriptMessage, error) {
	return nil, nil
}

var _ = Describe("resolveCaptainSession", func() {
	const providerSessionID = "0199e1aa-0000-7000-8000-000000000001"

	It("delegates the transcript hop to Captain instead of choosing between candidate rows", func(ctx SpecContext) {
		store := &recordingSessionStore{
			run: &captaindb.Session{
				ID: uuid.New(), Source: "gavel", Provider: "claude", ProviderSessionID: providerSessionID,
			},
			transcript: &captaindb.Session{
				ID: uuid.New(), Source: "claude", ProviderSessionID: providerSessionID,
			},
		}

		resolved, err := resolveCaptainSession(ctx, store, providerSessionID)

		Expect(err).NotTo(HaveOccurred())
		Expect(resolved.known).To(BeTrue())
		Expect(resolved.run.ID).To(Equal(store.run.ID))
		Expect(resolved.transcript.ID).To(Equal(store.transcript.ID))
		Expect(store.calls).To(Equal([]string{
			"GetSessionByIdentity(" + providerSessionID + ",gavel)",
			"GetTranscriptSessionByIdentity(" + providerSessionID + ")",
		}), "the per-source probe loop and the candidate listing are Captain's job now")
	})

	It("reports an admission root whose transcript has not been ingested yet as known but transcript-less", func(ctx SpecContext) {
		store := &recordingSessionStore{
			run: &captaindb.Session{
				ID: uuid.New(), Source: "gavel", ProviderSessionID: providerSessionID,
			},
			transcriptErr: fmt.Errorf("%w: %s", captaindb.ErrSessionNotFound, providerSessionID),
		}

		resolved, err := resolveCaptainSession(ctx, store, providerSessionID)

		Expect(err).NotTo(HaveOccurred())
		Expect(resolved.known).To(BeTrue())
		Expect(resolved.transcript).To(BeNil())
	})

	It("surfaces an ambiguous provider identity instead of picking a row", func(ctx SpecContext) {
		store := &recordingSessionStore{
			transcriptErr: fmt.Errorf("%w: two transcript-bearing sessions", captaindb.ErrSessionConflict),
		}

		_, err := resolveCaptainSession(ctx, store, providerSessionID)

		Expect(errors.Is(err, captaindb.ErrSessionConflict)).To(BeTrue())
	})
})
