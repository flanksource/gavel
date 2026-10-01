package runtime

import (
	"context"
	"fmt"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
)

// externalAnswer is what Captain recorded after the ask a candidate parked on:
// the answer message, and the Captain turn that answered it, if any.
type externalAnswer struct {
	MessageID      uuid.UUID
	MessageText    string
	AnswerRecorded bool
	// TurnInProgress is an origin=captain prompt run on the session, created
	// after the park and sending the answer's text, that has not finished.
	TurnInProgress bool
	TurnState      string
	TurnText       string
	TurnError      string
	TurnResult     []byte
	// ReplyIngested is true once the transcript holds the turn's reply: the last
	// message after the answer (and before any later user message) is an
	// assistant message carrying text or a StructuredOutput call, rather than a
	// tool call still waiting on its result.
	ReplyIngested bool
	Envelope      []byte
	FinalText     string
}

// turnFinished reports whether the Captain turn can be settled now. A turn that
// succeeded is classified by its reply, so it waits for that reply to be
// ingested; one that failed or was cancelled may have no reply at all.
func (a externalAnswer) turnFinished() bool {
	if a.TurnState == "" || a.TurnInProgress {
		return false
	}
	return a.TurnState != string(captaindb.PromptRunStateSucceeded) || a.ReplyIngested
}

// externalAnswerQuery reads, for one parked run, everything Captain recorded
// after gavel parked it. Every ordering uses the database's own clock
// (recorded_at, created_at) or transcript sequence, never a provider timestamp.
//
// The answer is the first user text message ingested after gavel parked the run
// that follows an asking StructuredOutput on the transcript. Gavel's own resume
// sends its answer as a user message too; that message is recognised by the
// ask_answered event gavel recorded when it admitted the resume, and skipped.
// The turn's messages are the ones after the answer and before the next user
// text message.
//
// The turn is bound to the answer by what it sent, not by when either row was
// written: Captain records a turn's prompt run when the turn ends, and the
// monitor may ingest the answer message before or after that. The turn is the
// earliest origin=captain run created after the park whose prompt is the
// answer's text.
const externalAnswerQuery = `
WITH answer AS (
  SELECT message.id, message.sequence, message.recorded_at, reply.text
  FROM captain_messages message
  CROSS JOIN LATERAL (
    SELECT string_agg(part->>'text', E'\n\n' ORDER BY ordinal) AS text
    FROM jsonb_array_elements(CASE WHEN jsonb_typeof(message.parts) = 'array' THEN message.parts ELSE '[]'::jsonb END) WITH ORDINALITY AS parts(part, ordinal)
    WHERE part->>'type' = 'text' AND btrim(part->>'text') <> ''
  ) reply
  WHERE message.session_id = @session AND message.sequence >= 0 AND message.role = 'user'
    AND message.recorded_at > @parked
    AND reply.text IS NOT NULL
    AND EXISTS (
      SELECT 1 FROM captain_messages ask
      CROSS JOIN LATERAL jsonb_array_elements(CASE WHEN jsonb_typeof(ask.parts) = 'array' THEN ask.parts ELSE '[]'::jsonb END) part
      WHERE ask.session_id = @session AND ask.role = 'assistant'
        AND ask.sequence >= 0 AND ask.sequence < message.sequence
        AND part->>'type' = 'dynamic-tool' AND part->>'toolName' = 'StructuredOutput'
        AND part->'input'->>'endStatus' = 'ask')
    AND NOT EXISTS (
      SELECT 1 FROM todo_issue_events event
      WHERE event.issue_id = @issue AND event.kind = @answered AND event.source = 'gavel'
        AND event.payload->>'promptRunId' = @run
        AND event.body = '**Answer:** ' || btrim(reply.text))
  ORDER BY message.sequence LIMIT 1
), bounds AS (
  SELECT answer.*, COALESCE((
    SELECT min(next.sequence) FROM captain_messages next
    WHERE next.session_id = @session AND next.role = 'user' AND next.sequence > answer.sequence
      AND EXISTS (SELECT 1 FROM jsonb_array_elements(CASE WHEN jsonb_typeof(next.parts) = 'array' THEN next.parts ELSE '[]'::jsonb END) part
                  WHERE part->>'type' = 'text' AND btrim(part->>'text') <> '')
  ), 9223372036854775807) AS next_sequence
  FROM answer
)
SELECT bounds.id AS message_id, bounds.text AS message_text,
  EXISTS (
    SELECT 1 FROM todo_issue_events event
    WHERE event.source = @origin AND event.source_id = 'ask-answer:' || @run || ':' || bounds.id::text
  ) AS answer_recorded,
  EXISTS (
    SELECT 1 FROM captain_prompt_runs run
    WHERE run.session_id = @session AND run.origin = @origin AND run.created_at > @parked
      AND btrim(run.prompt_markdown) = btrim(bounds.text)
      AND run.state IN ('pending', 'running', 'waiting')
  ) AS turn_in_progress,
  COALESCE(turn.state, '') AS turn_state, COALESCE(turn.result_text, '') AS turn_text,
  COALESCE(turn.error, '') AS turn_error, turn.result_json AS turn_result,
  COALESCE(last.role = 'assistant' AND (last.has_text OR last.has_output), false) AS reply_ingested,
  envelope.input AS envelope, COALESCE(final.text, '') AS final_text
FROM bounds
LEFT JOIN LATERAL (
  SELECT run.state::text AS state, run.result_text, run.error, run.result_json
  FROM captain_prompt_runs run
  WHERE run.session_id = @session AND run.origin = @origin AND run.created_at > @parked
    AND btrim(run.prompt_markdown) = btrim(bounds.text)
    AND run.state IN ('succeeded', 'failed', 'cancelled')
  ORDER BY run.created_at, run.id LIMIT 1
) turn ON true
LEFT JOIN LATERAL (
  SELECT message.role,
    EXISTS (SELECT 1 FROM jsonb_array_elements(CASE WHEN jsonb_typeof(message.parts) = 'array' THEN message.parts ELSE '[]'::jsonb END) part
            WHERE part->>'type' = 'text' AND btrim(part->>'text') <> '') AS has_text,
    EXISTS (SELECT 1 FROM jsonb_array_elements(CASE WHEN jsonb_typeof(message.parts) = 'array' THEN message.parts ELSE '[]'::jsonb END) part
            WHERE part->>'type' = 'dynamic-tool' AND part->>'toolName' = 'StructuredOutput') AS has_output
  FROM captain_messages message
  WHERE message.session_id = @session AND message.sequence > bounds.sequence
    AND message.sequence < bounds.next_sequence
  ORDER BY message.sequence DESC LIMIT 1
) last ON true
LEFT JOIN LATERAL (
  SELECT part->'input' AS input
  FROM captain_messages message
  CROSS JOIN LATERAL jsonb_array_elements(CASE WHEN jsonb_typeof(message.parts) = 'array' THEN message.parts ELSE '[]'::jsonb END) WITH ORDINALITY AS parts(part, ordinal)
  WHERE message.session_id = @session AND message.role = 'assistant'
    AND message.sequence > bounds.sequence AND message.sequence < bounds.next_sequence
    AND part->>'type' = 'dynamic-tool' AND part->>'toolName' = 'StructuredOutput'
    AND jsonb_typeof(part->'input') = 'object'
  ORDER BY message.sequence DESC, ordinal DESC LIMIT 1
) envelope ON true
LEFT JOIN LATERAL (
  SELECT string_agg(part->>'text', E'\n\n' ORDER BY ordinal) AS text
  FROM (
    SELECT message.parts FROM captain_messages message
    WHERE message.session_id = @session AND message.role = 'assistant'
      AND message.sequence > bounds.sequence AND message.sequence < bounds.next_sequence
      AND EXISTS (SELECT 1 FROM jsonb_array_elements(CASE WHEN jsonb_typeof(message.parts) = 'array' THEN message.parts ELSE '[]'::jsonb END) part
                  WHERE part->>'type' = 'text' AND btrim(part->>'text') <> '')
    ORDER BY message.sequence DESC LIMIT 1
  ) reply
  CROSS JOIN LATERAL jsonb_array_elements(CASE WHEN jsonb_typeof(reply.parts) = 'array' THEN reply.parts ELSE '[]'::jsonb END) WITH ORDINALITY AS parts(part, ordinal)
  WHERE part->>'type' = 'text' AND btrim(part->>'text') <> ''
) final ON true`

// detectExternalAnswer reads what Captain recorded after a candidate parked. It
// returns nil when no answer has been ingested yet.
func (p *Provider) detectExternalAnswer(ctx context.Context, candidate askCandidate) (*externalAnswer, error) {
	var rows []externalAnswer
	if err := p.db.WithContext(ctx).Raw(externalAnswerQuery, map[string]any{
		"session": candidate.ExecutionSessionID, "parked": candidate.ParkedAt, "issue": candidate.IssueID,
		"answered": native.EventAskAnswered, "run": candidate.PromptRunID.String(), "origin": externalAnswerSource,
	}).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("read Captain activity after the ask: %w", err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}
