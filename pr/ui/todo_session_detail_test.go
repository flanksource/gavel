package ui

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/session"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeProviderThreadStore struct {
	sessions []captaindb.SessionOverview
	messages []captaindb.TranscriptMessage
	owners   map[uuid.UUID]*captaindb.Session
}

func (f fakeProviderThreadStore) GetSessionByIdentity(_ context.Context, identity, _, _, _ string) (*captaindb.Session, error) {
	owner := f.owners[uuid.MustParse(identity)]
	if owner == nil {
		return nil, captaindb.ErrSessionNotFound
	}
	return owner, nil
}

func (f fakeProviderThreadStore) ListThreadSessionOverviews(context.Context, uuid.UUID) ([]captaindb.SessionOverview, error) {
	return f.sessions, nil
}

func (f fakeProviderThreadStore) ListThreadTranscriptMessages(context.Context, uuid.UUID) ([]captaindb.TranscriptMessage, error) {
	return f.messages, nil
}

func (fakeProviderThreadStore) ListThreadTurns(context.Context, uuid.UUID) ([]captaindb.SessionTurn, error) {
	return nil, nil
}

func (fakeProviderThreadStore) ListThreadAgents(context.Context, uuid.UUID) ([]captaindb.SessionAgent, error) {
	return nil, nil
}

func (fakeProviderThreadStore) ListThreadCosts(context.Context, uuid.UUID) ([]captaindb.SessionCost, error) {
	return nil, nil
}

func TestLoadProviderThreadIncludesTranscriptSnapshot(t *testing.T) {
	rootID := uuid.New()
	messageID := uuid.New()
	now := time.Now().UTC()
	parts, err := json.Marshal([]session.Part{{Type: session.PartText, Text: "parallel result"}})
	require.NoError(t, err)
	providerSessionID := "provider-parallel"
	thread, err := loadProviderThread(t.Context(), fakeProviderThreadStore{
		sessions: []captaindb.SessionOverview{{ID: rootID, Source: "codex", ProviderSessionID: &providerSessionID}},
		messages: []captaindb.TranscriptMessage{{ID: messageID, SessionID: rootID, Role: "assistant", Parts: parts, OccurredAt: &now}},
		owners:   map[uuid.UUID]*captaindb.Session{rootID: {ID: rootID, Source: "codex", ProviderSessionID: providerSessionID}},
	}, rootID)
	require.NoError(t, err)
	require.Len(t, thread.Messages, 1)
	assert.Equal(t, "parallel result", thread.Messages[0].Parts[0].Text)
	assert.Equal(t, providerSessionID, thread.Messages[0].Provenance.SessionID)
}

func TestSelectLegacyThreadCandidateUsesTheOnlyTranscriptBearingSession(t *testing.T) {
	providerID := "30fd5794-b1b3-4007-8de3-741d7d53ae4f"
	selected, diagnostics, conflict := selectLegacyThreadCandidate(providerID, []captaindb.SessionOverview{
		{ID: uuid.New(), ProviderSessionID: &providerID, Source: "gavel", HostID: "local"},
		{ID: uuid.New(), ProviderSessionID: &providerID, Source: "claude", HostID: "local"},
		{ID: uuid.New(), ProviderSessionID: &providerID, Source: "claude", HostID: "MacBook-Pro.local", MessageCount: 79, TurnCount: 8},
	})

	require.NotNil(t, selected)
	assert.EqualValues(t, 79, selected.MessageCount)
	assert.False(t, conflict)
	require.Len(t, diagnostics, 1)
	assert.Equal(t, "legacy_session_identity_resolved", diagnostics[0].Code)
}

func TestSelectLegacyThreadCandidateRejectsMultipleTranscripts(t *testing.T) {
	providerID := "provider-session"
	selected, diagnostics, conflict := selectLegacyThreadCandidate(providerID, []captaindb.SessionOverview{
		{ID: uuid.New(), ProviderSessionID: &providerID, Source: "claude", MessageCount: 1},
		{ID: uuid.New(), ProviderSessionID: &providerID, Source: "claude", TurnCount: 1},
	})

	assert.Nil(t, selected)
	assert.True(t, conflict)
	require.Len(t, diagnostics, 1)
	assert.Equal(t, "ambiguous_transcript_sessions", diagnostics[0].Code)
}

func TestProjectThreadStatusKeepsAttemptFailureIndependent(t *testing.T) {
	status := projectThreadStatus([]captaindb.SessionOverview{{
		ID: uuid.New(), Source: "claude", LifecycleStatus: string(captaindb.SessionLifecycleCreated), MessageCount: 4,
	}})
	assert.Equal(t, "idle", status)

	status = projectThreadStatus([]captaindb.SessionOverview{{
		ID: uuid.New(), Source: "claude", LifecycleStatus: string(captaindb.SessionLifecycleCreated), ProcessActive: true,
	}})
	assert.Equal(t, "working", status)
}
