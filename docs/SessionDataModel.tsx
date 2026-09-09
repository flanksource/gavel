import React from 'react';
import { Page, Section, BoxNode, COLORS } from '@flanksource/facet';

/**
 * Session ERD — target end state.
 *
 * Build:
 *   cd docs && facet html SessionDataModel.tsx -o ../.tmp/facet-out
 *   cd docs && facet pdf  SessionDataModel.tsx -o ../.tmp/facet-out
 *
 * Columns are limited to keys, foreign keys and status/state. Payload, text and
 * measurement columns are omitted.
 *
 * Uses BoxNode without Diagram/Arrow: arrows are drawn by react-xarrows against
 * a live DOM, so a <Diagram> renders empty under SSR. Relationships are encoded
 * inline on each FK field instead.
 */

const CAPTAIN = COLORS.primary;
const VIEW = COLORS.accent;
const GAVEL = '#7c3aed';

type KeyKind = 'PK' | 'FK' | 'UK' | undefined;

const KEY_COLOR: Record<'PK' | 'FK' | 'UK', string> = {
  PK: COLORS.pk,
  FK: COLORS.fk,
  UK: COLORS.muted,
};

function Field({
  name,
  type,
  keyKind,
  to,
  nullable,
}: {
  name: string;
  type: string;
  keyKind?: KeyKind;
  to?: string;
  nullable?: boolean;
}) {
  return (
    <div className="flex items-baseline gap-1 text-[6.5pt] leading-[10pt] border-b border-dotted border-gray-200 py-[1px]">
      <span
        className="inline-block w-[16px] shrink-0 text-center rounded text-white font-bold text-[5pt]"
        style={{ backgroundColor: keyKind ? KEY_COLOR[keyKind] : 'transparent' }}
      >
        {keyKind ?? ''}
      </span>
      <span className="font-mono font-semibold">{name}</span>
      <span className="text-gray-500">{type}</span>
      {nullable && <span className="text-gray-400 italic">null</span>}
      {to && (
        <span className="ml-auto font-mono" style={{ color: COLORS.fk }}>
          →{to}
        </span>
      )}
    </div>
  );
}

/** Row count and on-disk size, or "derived" for views. */
function Stats({ rows, size }: { rows?: string; size: string }) {
  return (
    <div className="flex items-baseline justify-between text-[6pt] font-mono pb-1 mb-1 border-b border-gray-300">
      <span className="text-gray-600">{rows ? `${rows} rows` : 'derived'}</span>
      <span className="font-bold" style={{ color: COLORS.accent }}>
        {size}
      </span>
    </div>
  );
}

/**
 * Observed value distributions, one entry per sampled column, as
 * "column: value n · value n". Ids and timestamps are excluded.
 */
function Sample({ lines }: { lines: string[] }) {
  return (
    <div className="mt-1 pt-1 border-t border-gray-300">
      <div className="text-[5.5pt] font-bold uppercase tracking-wide text-gray-400 mb-[1px]">
        observed
      </div>
      {lines.map((line) => {
        const [column, values] = line.split(': ');
        return (
          <div key={line} className="text-[5.5pt] leading-[8pt] font-mono">
            <span className="font-semibold text-gray-700">{column}</span>{' '}
            <span className="text-gray-500">{values}</span>
          </div>
        );
      })}
    </div>
  );
}

function Entity({
  name,
  color,
  width,
  rows,
  size,
  sample,
  note,
  children,
}: {
  name: string;
  color: string;
  width: string;
  rows?: string;
  size: string;
  sample?: string[];
  note?: string;
  children: React.ReactNode;
}) {
  return (
    <BoxNode
      title={name}
      headerColor={color}
      bodyColor="#ffffff"
      borderColor={color}
      minWidth={width}
      compact
    >
      <Stats rows={rows} size={size} />
      <div className="flex flex-col">{children}</div>
      {sample && <Sample lines={sample} />}
      {note && <div className="text-[6pt] italic text-gray-500 pt-1 leading-[8pt]">{note}</div>}
    </BoxNode>
  );
}

function Row({ children }: { children: React.ReactNode }) {
  return <div className="flex items-start justify-start gap-3 flex-wrap mb-3">{children}</div>;
}

export default function SessionDataModel() {
  return (
    <Page title="Session ERD" product="captain + gavel">
      <Section title="captain — sessions and transcript">
        <Row>
          <Entity
            name="captain_sessions"
            color={CAPTAIN}
            width="250px"
            rows="6,337"
            size="358 MB"
            sample={[
              'source: claude 3,267 · codex 2,433 · captain 567 · gavel 99',
              'provider: (empty) 5,149 · multi-model 294 · openai 179 · codex-cmux 153 · +12 more',
              'lifecycle_status: created 5,236 · succeeded 624 · failed 186 · running 164 · partial 156',
              'activity_state: idle 6,359 · ask 5 · working 2',
              'health_state: healthy 6,366 — stalled and zombie never observed',
            ]}
            note="316 MB of that is main heap — see the sizing note. Parent/root cascade: deleting a root removes the whole agent subtree."
          >
            <Field name="id" type="uuid" keyKind="PK" />
            <Field name="provider_session_id" type="text" keyKind="UK" nullable />
            <Field name="source" type="text" keyKind="UK" />
            <Field name="provider" type="text" keyKind="UK" />
            <Field name="host_id" type="text" keyKind="UK" />
            <Field name="parent_session_id" type="uuid" keyKind="FK" to="self" nullable />
            <Field name="root_session_id" type="uuid" keyKind="FK" to="self" nullable />
            <Field name="lifecycle_status" type="enum" />
            <Field name="activity_state" type="enum" />
            <Field name="health_state" type="enum" />
          </Entity>

          <Entity
            name="captain_messages"
            color={CAPTAIN}
            width="256px"
            rows="648,793"
            size="1057 MB"
            sample={[
              'role: assistant 476,727 · user 171,823 · system 865',
              'sequence: 1..5,228 (avg 374)',
            ]}
            note="402 MB heap + 487 MB toast + 168 MB indexes. Two durable identity keys: (session_id, sequence) and (session_id, provider_message_id)."
          >
            <Field name="id" type="uuid" keyKind="PK" />
            <Field name="session_id" type="uuid" keyKind="FK" to="sessions" />
            <Field name="turn_id" type="uuid" keyKind="FK" to="turns" nullable />
            <Field name="model_call_id" type="uuid" keyKind="FK" to="model_calls" nullable />
            <Field name="provider_message_id" type="text" keyKind="UK" nullable />
            <Field name="sequence" type="bigint" keyKind="UK" />
            <Field name="source_line" type="bigint" nullable />
            <Field name="role" type="text" />
          </Entity>

          <Entity
            name="captain_turns"
            color={CAPTAIN}
            width="222px"
            rows="12,826"
            size="8.4 MB"
            sample={[
              'status: ended 12,836 — open never observed at rest',
              'turn_index: 1..95 (avg 7)',
            ]}
            note="at most one open turn per session"
          >
            <Field name="id" type="uuid" keyKind="PK" />
            <Field name="session_id" type="uuid" keyKind="FK" to="sessions" />
            <Field name="provider_turn_id" type="text" keyKind="UK" nullable />
            <Field name="turn_index" type="integer" keyKind="UK" />
            <Field name="status" type="enum" />
          </Entity>
        </Row>

        <Row>
          <Entity
            name="captain_model_calls"
            color={CAPTAIN}
            width="240px"
            rows="12,826"
            size="9.2 MB"
            sample={['status: succeeded 12,836 — pending and failed never observed at rest']}
            note="turn_id is mandatory, so prompt-run costs are always reachable via turns too"
          >
            <Field name="id" type="uuid" keyKind="PK" />
            <Field name="turn_id" type="uuid" keyKind="FK" to="turns" />
            <Field name="prompt_run_id" type="uuid" keyKind="FK" to="prompt_runs" nullable />
            <Field name="iteration_id" type="uuid" keyKind="FK" to="iterations" nullable />
            <Field name="call_index" type="integer" keyKind="UK" />
            <Field name="status" type="enum" />
          </Entity>

          <Entity name="captain_events" color={CAPTAIN} width="222px" rows="0" size="72 kB">
            <Field name="id" type="uuid" keyKind="PK" />
            <Field name="session_id" type="uuid" keyKind="FK" to="sessions" />
            <Field name="turn_id" type="uuid" keyKind="FK" to="turns" nullable />
            <Field name="prompt_run_id" type="uuid" keyKind="FK" to="prompt_runs" nullable />
            <Field name="model_call_id" type="uuid" keyKind="FK" to="model_calls" nullable />
            <Field name="parent_event_id" type="uuid" keyKind="FK" to="self" nullable />
            <Field name="event_key" type="text" keyKind="UK" nullable />
            <Field name="stream / sequence" type="text" keyKind="UK" />
            <Field name="kind / scope" type="text" />
          </Entity>

          <Entity
            name="captain_turn_requests"
            color={CAPTAIN}
            width="222px"
            rows="0"
            size="48 kB"
            note="questions, tool approvals, plan-exit approvals"
          >
            <Field name="id" type="uuid" keyKind="PK" />
            <Field name="session_id" type="uuid" keyKind="FK" to="sessions" />
            <Field name="turn_id" type="uuid" keyKind="FK" to="turns" nullable />
            <Field name="prompt_run_id" type="uuid" keyKind="FK" to="prompt_runs" nullable />
            <Field name="plan_id" type="uuid" keyKind="FK" to="plans" nullable />
            <Field name="kind" type="enum" />
            <Field name="state" type="enum" />
            <Field name="idempotency_key" type="text" keyKind="UK" nullable />
          </Entity>
        </Row>

        <Row>
          <Entity
            name="captain_session_processes"
            color={CAPTAIN}
            width="238px"
            rows="4,129"
            size="3.5 MB"
            sample={['status: exited 4,200 · sleeping 10']}
            note="at most one live process per session (unique where ended_at is null)"
          >
            <Field name="id" type="uuid" keyKind="PK" />
            <Field name="session_id" type="uuid" keyKind="FK" to="sessions" />
            <Field name="host_id / boot_id" type="text" keyKind="UK" />
            <Field name="pid" type="bigint" keyKind="UK" />
            <Field name="process_started_at" type="tstz" keyKind="UK" />
            <Field name="status" type="text" />
            <Field name="ended_at" type="tstz" nullable />
          </Entity>

          <Entity
            name="captain_session_sources"
            color={CAPTAIN}
            width="230px"
            rows="3,793"
            size="3.6 MB"
            sample={['source_kind: claude 2,266 · codex 1,532']}
            note="a transcript file maps to exactly one session"
          >
            <Field name="id" type="uuid" keyKind="PK" />
            <Field name="session_id" type="uuid" keyKind="FK" to="sessions" />
            <Field name="source_kind" type="text" keyKind="UK" />
            <Field name="path" type="text" keyKind="UK" />
          </Entity>

          <Entity name="captain_artifacts" color={CAPTAIN} width="222px" rows="0" size="56 kB">
            <Field name="id" type="uuid" keyKind="PK" />
            <Field name="session_id" type="uuid" keyKind="FK" to="sessions" />
            <Field name="turn_id" type="uuid" keyKind="FK" to="turns" nullable />
            <Field name="prompt_run_id" type="uuid" keyKind="FK" to="prompt_runs" nullable />
            <Field name="kind" type="text" />
          </Entity>
        </Row>
      </Section>

      <Section title="captain — prompt runs and plans">
        <Row>
          <Entity
            name="captain_prompt_runs"
            color={CAPTAIN}
            width="262px"
            rows="850"
            size="2.0 MB"
            sample={[
              'phase: finished 786 · generate 44 · queued 13 · verify 7',
              'state: succeeded 625 · failed 204 · pending 13 · running 5 · waiting 3',
            ]}
            note="three FKs into sessions; one in-flight run per root_session_id"
          >
            <Field name="id" type="uuid" keyKind="PK" />
            <Field name="session_id" type="uuid" keyKind="FK" to="sessions" />
            <Field name="root_session_id" type="uuid" keyKind="FK" to="sessions" />
            <Field name="execution_session_id" type="uuid" keyKind="FK" to="sessions" nullable />
            <Field name="parent_run_id" type="uuid" keyKind="FK" to="self" nullable />
            <Field name="input_plan_id" type="uuid" keyKind="FK" to="plans" nullable />
            <Field name="phase" type="enum" />
            <Field name="state" type="enum" />
            <Field name="admission_key" type="text" keyKind="UK" nullable />
          </Entity>

          <Entity
            name="captain_prompt_run_iterations"
            color={CAPTAIN}
            width="234px"
            rows="0"
            size="40 kB"
            note="carries a redundant (prompt_run_id, id) unique as a composite-FK target"
          >
            <Field name="id" type="uuid" keyKind="PK" />
            <Field name="prompt_run_id" type="uuid" keyKind="FK" to="prompt_runs" />
            <Field name="iteration" type="integer" keyKind="UK" />
            <Field name="state" type="enum" />
          </Entity>

          <Entity
            name="captain_plans + revisions"
            color={CAPTAIN}
            width="240px"
            rows="20 + 21"
            size="128 + 296 kB"
            sample={['approval_state: approved 12 · pending 9']}
            note="approval_state lives on the plan, not the revision; a revision carries only its ordinal and body"
          >
            <Field name="id" type="uuid" keyKind="PK" />
            <Field name="source_session_id" type="uuid" keyKind="FK" to="sessions" nullable />
            <Field name="source_prompt_run_id" type="uuid" keyKind="FK" to="prompt_runs" nullable />
            <Field name="approved_revision_id" type="uuid" keyKind="FK" to="revisions" nullable />
            <Field name="approval_state" type="enum" />
            <Field name="revision.plan_id" type="uuid" keyKind="FK" to="plans" />
            <Field name="revision.revision" type="integer" keyKind="UK" />
          </Entity>
        </Row>
      </Section>

      <Section title="gavel — todos, and how they bind to captain">
        <Row>
          <Entity name="todo_workspaces" color={GAVEL} width="196px" rows="14" size="80 kB">
            <Field name="id" type="uuid" keyKind="PK" />
            <Field name="repo_key" type="text" keyKind="UK" nullable />
          </Entity>

          <Entity
            name="todo_workspace_paths"
            color={GAVEL}
            width="212px"
            rows="14"
            size="96 kB"
            sample={['is_primary: true 14 — every workspace currently has exactly one path']}
            note="at most one primary path per workspace"
          >
            <Field name="workspace_id" type="uuid" keyKind="PK" to="workspaces" />
            <Field name="path" type="text" keyKind="PK" />
            <Field name="is_primary" type="boolean" />
          </Entity>

          <Entity
            name="todo_issues"
            color={GAVEL}
            width="256px"
            rows="824"
            size="4.5 MB"
            sample={[
              'status: closed 488 · open 221 · verified 69 · draft 38 · cancelled 8',
              'priority: medium 753 · high 50 · low 21 — critical never used',
            ]}
            note="the two pointer FKs are composite on (id, …) so a run or plan must already belong to this issue"
          >
            <Field name="id" type="uuid" keyKind="PK" />
            <Field name="workspace_id" type="uuid" keyKind="FK" to="workspaces" />
            <Field name="status" type="text" />
            <Field name="priority" type="text" />
            <Field name="active_prompt_run_id" type="uuid" keyKind="FK" to="issue_prompt_runs" nullable />
            <Field name="selected_plan_id" type="uuid" keyKind="FK" to="issue_plans" nullable />
          </Entity>
        </Row>

        <Row>
          <Entity
            name="todo_issue_prompt_runs"
            color={GAVEL}
            width="268px"
            rows="97"
            size="104 kB"
            sample={['step_kind: plan 51 · verify 28 · run 21']}
            note="prompt_run_id is globally unique — a captain run belongs to exactly one issue. ON DELETE RESTRICT keeps execution history from vanishing through a session cascade."
          >
            <Field name="issue_id" type="uuid" keyKind="PK" to="todo_issues" />
            <Field name="prompt_run_id" type="uuid" keyKind="PK" to="captain_prompt_runs" />
            <Field name="step_kind" type="text" keyKind="UK" />
            <Field name="ordinal" type="integer" keyKind="UK" />
          </Entity>

          <Entity
            name="todo_issue_plans"
            color={GAVEL}
            width="238px"
            rows="18"
            size="72 kB"
            note="plan_id is NOT globally unique — one captain plan may bind to several issues"
          >
            <Field name="issue_id" type="uuid" keyKind="PK" to="todo_issues" />
            <Field name="plan_id" type="uuid" keyKind="PK" to="captain_plans" />
            <Field name="ordinal" type="integer" keyKind="UK" />
          </Entity>

          <Entity
            name="todo_issue_events"
            color={GAVEL}
            width="230px"
            rows="6,155"
            size="5.8 MB"
            sample={[
              'kind: comment 1,435 · label_added 894 · label_removed 773 · issue_created 718 · +14 more',
              'sequence: 1..62 (avg 6)',
            ]}
            note="append-only; (source, source_id) is the import and projection idempotency key"
          >
            <Field name="id" type="uuid" keyKind="PK" />
            <Field name="issue_id" type="uuid" keyKind="FK" to="todo_issues" />
            <Field name="sequence" type="bigint" keyKind="UK" />
            <Field name="kind" type="text" />
            <Field name="source" type="text" keyKind="UK" />
            <Field name="source_id" type="text" keyKind="UK" nullable />
          </Entity>
        </Row>

        <Row>
          <Entity
            name="todo_issue_aliases"
            color={GAVEL}
            width="210px"
            rows="718"
            size="264 kB"
          >
            <Field name="workspace_id" type="uuid" keyKind="PK" to="workspaces" />
            <Field name="alias" type="text" keyKind="PK" />
            <Field name="issue_id" type="uuid" keyKind="FK" to="todo_issues" />
          </Entity>

          <Entity
            name="todo_issue_relationships"
            color={GAVEL}
            width="248px"
            rows="70"
            size="80 kB"
            sample={['relation: depends_on 61 · related_to 9']}
            note="depends_on is directed; related_to is canonicalised issue_id < target so it is stored once"
          >
            <Field name="workspace_id" type="uuid" keyKind="PK" to="workspaces" />
            <Field name="issue_id" type="uuid" keyKind="PK" to="todo_issues" />
            <Field name="target_issue_id" type="uuid" keyKind="PK" to="todo_issues" />
            <Field name="relation" type="text" keyKind="PK" />
          </Entity>
        </Row>
      </Section>

      <Section title="views">
        <Row>
          <Entity
            name="captain_session_overview"
            color={VIEW}
            width="236px"
            size="view"
            note="1:1 with captain_sessions, via ten lateral joins. agent_count counts all descendants + 1."
          >
            <Field name="session_id" type="uuid" keyKind="FK" to="sessions" />
            <Field name="thread_id" type="uuid" />
            <Field name="process_active" type="boolean" />
            <Field name="lifecycle / activity / health" type="enum" />
          </Entity>

          <Entity
            name="captain_session_transcript"
            color={VIEW}
            width="228px"
            size="view"
            note="1:1 with captain_messages, left-joined to model_calls"
          >
            <Field name="id" type="uuid" keyKind="FK" to="messages" />
            <Field name="session_id" type="uuid" keyKind="FK" to="sessions" />
            <Field name="sequence" type="bigint" keyKind="UK" />
            <Field name="model_call_status" type="enum" nullable />
          </Entity>

          <Entity
            name="captain_session_agents"
            color={VIEW}
            width="216px"
            size="view"
            note="a view over captain_sessions, 1:1. child_count counts DIRECT children only."
          >
            <Field name="session_id" type="uuid" keyKind="FK" to="sessions" />
            <Field name="thread_id" type="uuid" />
            <Field name="is_root" type="boolean" />
          </Entity>
        </Row>

        <Row>
          <Entity
            name="captain_session_costs"
            color={VIEW}
            width="248px"
            size="view"
            note="N rows per session. Synthetic text id: session:model:backend:effort:currency."
          >
            <Field name="id" type="text" keyKind="PK" />
            <Field name="session_id" type="uuid" keyKind="FK" to="sessions" />
          </Entity>

          <Entity name="captain_prompt_run_overview" color={VIEW} width="228px" size="view">
            <Field name="id" type="uuid" keyKind="FK" to="prompt_runs" />
            <Field name="latest_iteration_state" type="text" />
          </Entity>

          <Entity
            name="todo_issue_runtime"
            color={VIEW}
            width="216px"
            size="view"
            note="one row per issue; execution_state derived, never cached"
          >
            <Field name="issue_id" type="uuid" keyKind="FK" to="todo_issues" />
            <Field name="execution_state" type="text" />
          </Entity>
        </Row>

        <Row>
          <Entity
            name="todo_issue_plan_revision_details"
            color={VIEW}
            width="290px"
            size="view"
            note="exposes durable captain revision bodies without copying plan content into todo tables"
          >
            <Field name="issue_id" type="uuid" keyKind="FK" to="todo_issues" />
            <Field name="plan_id" type="uuid" keyKind="FK" to="captain_plans" />
            <Field name="revision_id" type="uuid" keyKind="FK" to="captain_plan_revisions" />
            <Field name="approval_state" type="text" nullable />
            <Field name="selected / approved" type="boolean" />
          </Entity>
        </Row>
      </Section>

      <Section title="notes">
        <ul className="text-[7pt] leading-relaxed list-disc pl-5">
          <li>
            <strong>Sizing and sampling.</strong> Row counts are exact where cheap to count,
            otherwise <code>reltuples</code> estimates; sizes are <code>pg_total_relation_size</code>{' '}
            (heap + toast + indexes). The <em>observed</em> block on each entity is the actual value
            distribution for its enum and classifier columns — ids, timestamps and free text are
            excluded. Measured against the local <code>gavel</code> database on 2026-07-20, 1482 MB
            total. Not production figures.
          </li>
          <li>
            <strong>Most sessions never start.</strong> 5,236 of 6,337 sit at{' '}
            <code>lifecycle_status = created</code> and 5,149 have an empty <code>provider</code>, so
            roughly 82% are admission rows that never became a run. Anything paging or counting
            sessions should expect that shape.
          </li>
          <li>
            <strong>Role coverage is not a rounding error.</strong> 171,823 of 648,793 messages are{' '}
            <code>user</code> and 865 are <code>system</code> — 27% of the transcript. A read path
            that returns only assistant messages drops more than a quarter of the corpus.
          </li>
          <li>
            <strong>Terminal states dominate at rest.</strong> Turns are 100% <code>ended</code> and
            model calls 100% <code>succeeded</code>; <code>health_state</code> is uniformly{' '}
            <code>healthy</code>, with <code>stalled</code> and <code>zombie</code> never observed.
            The in-flight and unhealthy branches are therefore untested by this data.
          </li>
          <li>
            <strong>captain_messages dominates</strong> at 71% of the database. Its toast segment
            (487 MB) exceeds its heap (402 MB), because <code>parts</code> and <code>raw</code> hold
            the full message payload.
          </li>
          <li>
            <strong>captain_sessions carries heap bloat.</strong> 6,337 live rows occupy a 316 MB main
            heap — roughly 50 kB per live row, against only 30 MB of toast. It has taken 220,843
            updates (158,566 HOT) from state and heartbeat writes. Autovacuum is keeping up with dead
            tuples (527 outstanding) but only reclaims space for reuse; the file never shrinks without{' '}
            <code>VACUUM FULL</code> or <code>pg_repack</code>.
          </li>
          <li>
            <strong>Ownership boundary.</strong> gavel binds to captain at exactly two points —{' '}
            <code>todo_issue_prompt_runs.prompt_run_id</code> and <code>todo_issue_plans.plan_id</code>,
            both <code>ON DELETE RESTRICT</code>. Everything else gavel needs is reached by traversing
            captain from those two.
          </li>
          <li>
            <strong>Provider identity is the thread bridge.</strong> One provider session yields a
            gavel-orchestration row and a Claude/Codex transcript row in <code>captain_sessions</code>;
            they share <code>provider_session_id</code>. That match is a soft join key with no
            constraint behind it — the only unenforced edge in the model.
          </li>
          <li>
            <strong>PK</strong> primary key · <strong>FK</strong> foreign key · <strong>UK</strong>{' '}
            unique-constraint member (composite unless noted).
          </li>
        </ul>
      </Section>
    </Page>
  );
}
