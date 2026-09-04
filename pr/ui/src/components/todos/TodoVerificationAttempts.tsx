import { useEffect, useMemo, useState } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import { VerificationResults } from '@flanksource/clicky-ui/data';
import { UiBeaker, UiError, UiPass, UiStop, UiWarningTriangle } from '@flanksource/clicky-ui/icons';
import type { TodoSessionDetailResponse } from '../../types';
import { Spinner } from '../../icons/Spinner';
import { ageShort } from '../../utils';
import { formatDuration } from './TodoSessionTimer';
import { VerificationAttemptSession } from './VerificationAttemptSession';
import {
  verificationAttempts,
  type AttemptOutcome,
  type VerificationAttemptEntry,
} from './verificationReport';

type EvidenceTab = 'results' | 'session';

/**
 * Every verification attempt this todo has made — the explicit `gavel todos
 * check` runs and the in-loop definition-of-done evaluations — newest first,
 * each rendered through the shared VerificationResults (the same TestRunner the
 * captain webapp uses), with the agent session behind it when one was recorded.
 */
export function TodoVerificationAttempts({
  dir,
  todoRef,
  detail,
  error,
}: {
  dir: string;
  todoRef: string;
  detail: TodoSessionDetailResponse | null;
  error?: string;
}) {
  const set = useMemo(() => verificationAttempts(detail), [detail]);
  const [selectedId, setSelectedId] = useState<string | null>(null);

  // Follow the newest attempt until a row is picked, and recover if the selected
  // attempt disappears (a different todo, a pruned run).
  const known = set.attempts.some((entry) => entry.attempt.promptRunId === selectedId);
  useEffect(() => {
    if (!known) setSelectedId(set.attempts[0]?.attempt.promptRunId ?? null);
  }, [known, set.attempts]);

  const selected = set.attempts.find((entry) => entry.attempt.promptRunId === selectedId) ?? set.attempts[0] ?? null;

  return (
    <section className="overflow-hidden rounded-lg border border-border bg-card shadow-sm">
      <header className="flex flex-wrap items-center gap-2 border-b border-border bg-muted/30 px-3 py-2.5">
        <span className="inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-md border border-border bg-background text-muted-foreground">
          <UiBeaker className="text-xs" />
        </span>
        <span className="min-w-0 flex-1 truncate text-xs font-semibold uppercase text-muted-foreground">
          Verification Attempts
        </span>
        {set.attempts.length > 0 && <OutcomePill outcome={set.attempts[0].outcome} />}
        <span className="text-[11px] tabular-nums text-muted-foreground">{set.attempts.length}</span>
      </header>

      {error && (
        <p role="alert" className="border-b border-border px-3 py-2 text-xs text-red-600">
          {error}
        </p>
      )}
      {set.malformed.length > 0 && (
        <div role="alert" className="border-b border-border bg-red-500/10 px-3 py-2 text-xs text-red-700 dark:text-red-300">
          {set.malformed.length} attempt(s) could not be read:
          <ul className="mt-0.5 list-disc pl-4">
            {set.malformed.map((bad) => (
              <li key={bad.promptRunId}>
                #{bad.ordinal} — {bad.reason}
              </li>
            ))}
          </ul>
        </div>
      )}

      {!detail && !error ? (
        <p className="flex items-center gap-2 px-3 py-4 text-xs text-muted-foreground">
          <Spinner /> Loading attempts…
        </p>
      ) : set.attempts.length === 0 ? (
        <p className="px-3 py-4 text-xs text-muted-foreground">
          No verification has run yet. Use “Run verification” above to check the definition of done.
        </p>
      ) : (
        <ul className="divide-y divide-border">
          {set.attempts.map((entry) => {
            const isSelected = entry.attempt.promptRunId === selected?.attempt.promptRunId;
            return (
              <li key={entry.attempt.promptRunId}>
                <AttemptRow
                  entry={entry}
                  selected={isSelected}
                  onSelect={() => setSelectedId(entry.attempt.promptRunId)}
                />
                {isSelected && <AttemptEvidence dir={dir} todoRef={todoRef} entry={entry} />}
              </li>
            );
          })}
        </ul>
      )}
    </section>
  );
}

function AttemptRow({ entry, selected, onSelect }: { entry: VerificationAttemptEntry; selected: boolean; onSelect: () => void }) {
  const { attempt, report } = entry;
  const started = attempt.startedAt || attempt.queuedAt || attempt.createdAt;
  const failing = (report?.summary.failed ?? 0) + (report?.summary.timedout ?? 0);
  return (
    <Button
      variant="ghost"
      type="button"
      onClick={onSelect}
      aria-pressed={selected}
      className={`flex h-auto w-full items-center justify-start gap-2 rounded-none px-3 py-2 text-left text-xs ${selected ? 'bg-muted/60' : ''}`}
    >
      <OutcomeIcon outcome={entry.outcome} />
      <span className={`${selected ? 'font-semibold' : 'font-medium'} tabular-nums`}>#{attempt.ordinal}</span>
      <span className="rounded border border-border px-1.5 py-0.5 text-[10px] uppercase text-muted-foreground">{attempt.step}</span>
      {started && <span className="text-[11px] text-muted-foreground">{ageShort(started)}</span>}
      {typeof attempt.durationMs === 'number' && attempt.durationMs > 0 && (
        <span className="text-[11px] tabular-nums text-muted-foreground">{formatDuration(attempt.durationMs)}</span>
      )}
      <span className="min-w-0 flex-1 truncate text-[11px] text-muted-foreground">
        {[attempt.provider, attempt.model].filter(Boolean).join(' · ')}
      </span>
      {failing > 0 && <span className="shrink-0 text-[11px] tabular-nums text-red-600">{failing} failing</span>}
      {attempt.error && (
        <span className="max-w-40 shrink-0 truncate text-[11px] text-red-600" title={attempt.error}>
          {attempt.error}
        </span>
      )}
    </Button>
  );
}

function AttemptEvidence({ dir, todoRef, entry }: { dir: string; todoRef: string; entry: VerificationAttemptEntry }) {
  const sessionId = entry.attempt.executionSessionId || entry.attempt.providerSessionId;
  const [tab, setTab] = useState<EvidenceTab>('results');
  const active = tab === 'session' && sessionId ? 'session' : 'results';

  return (
    <div className="border-t border-border">
      {sessionId && (
        <div className="flex flex-nowrap gap-1 border-b border-border px-3 pt-2">
          <EvidenceTabButton label="Results" isActive={active === 'results'} onClick={() => setTab('results')} />
          <EvidenceTabButton label="Session" isActive={active === 'session'} onClick={() => setTab('session')} />
        </div>
      )}
      {active === 'session' && sessionId ? (
        <VerificationAttemptSession dir={dir} todoRef={todoRef} sessionId={sessionId} />
      ) : (
        <VerificationResults
          report={entry.report}
          emptyText="This attempt recorded no verification results."
        />
      )}
    </div>
  );
}

function EvidenceTabButton({ label, isActive, onClick }: { label: string; isActive: boolean; onClick: () => void }) {
  return (
    <Button
      variant="ghost"
      type="button"
      aria-pressed={isActive}
      onClick={onClick}
      className={`-mb-px inline-flex h-auto shrink-0 items-center gap-1.5 border-b-2 px-2.5 py-1.5 text-xs font-medium ${
        isActive ? 'border-primary text-foreground' : 'border-transparent text-muted-foreground hover:text-foreground'
      }`}
    >
      {label}
    </Button>
  );
}

function OutcomeIcon({ outcome }: { outcome: AttemptOutcome }) {
  switch (outcome) {
    case 'running':
      return <Spinner className="size-3.5 shrink-0" />;
    case 'passed':
      return <UiPass className="shrink-0 text-sm text-emerald-600" />;
    case 'cancelled':
      return <UiStop className="shrink-0 text-sm text-muted-foreground" />;
    case 'failed':
      return <UiError className="shrink-0 text-sm text-red-600" />;
    default:
      // errored, warned, skipped, timed_out, queued.
      return <UiWarningTriangle className="shrink-0 text-sm text-amber-600" />;
  }
}

function OutcomePill({ outcome }: { outcome: AttemptOutcome }) {
  const tone =
    outcome === 'passed'
      ? 'border-emerald-500/25 bg-emerald-500/10 text-emerald-700 dark:text-emerald-400'
      : outcome === 'failed' || outcome === 'timed_out'
        ? 'border-red-500/25 bg-red-500/10 text-red-700 dark:text-red-400'
        : outcome === 'running' || outcome === 'queued' || outcome === 'cancelled'
          ? 'border-border bg-background text-muted-foreground'
          : 'border-amber-500/25 bg-amber-500/10 text-amber-700 dark:text-amber-400';
  return <span className={`rounded border px-1.5 py-0.5 text-[10px] uppercase ${tone}`}>{outcome}</span>;
}
