import { useEffect, useMemo, useState } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import { UiBeaker, UiError, UiPass, UiStop, UiWarningTriangle } from '@flanksource/clicky-ui/icons';
import type { TodoSessionDetailResponse } from '../../types';
import { Spinner } from '../../icons/Spinner';
import { ageShort } from '../../utils';
import { formatDuration } from './TodoSessionTimer';
import { TestRunResults } from '../tests/TestRunResults';
import { fetchRunSnapshot, type RunArtifact, type RunSnapshot } from '../tests/types';
import { VerificationAttemptSession } from './VerificationAttemptSession';
import { VerificationStepOutput } from './VerificationStepOutput';
import {
  attemptRunArtifacts,
  attemptOutputSteps,
  attemptTabs,
  defaultAttemptTab,
  verificationAttempts,
  type AttemptOutcome,
  type AttemptTab,
  type VerificationAttempt,
} from './verificationAttempts';

const TAB_LABELS: Record<AttemptTab, string> = { test: 'Test', lint: 'Lint', session: 'Session', output: 'Output' };

/**
 * Every verification attempt this todo has made — the explicit `gavel todos
 * check` runs and the in-loop definition-of-done evaluations — newest first,
 * each with Test / Lint / Session / Output evidence.
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
        <>
          <ul className="divide-y divide-border">
            {set.attempts.map((entry) => (
              <AttemptRow
                key={entry.attempt.promptRunId}
                entry={entry}
                selected={entry.attempt.promptRunId === selected?.attempt.promptRunId}
                onSelect={() => setSelectedId(entry.attempt.promptRunId)}
              />
            ))}
          </ul>
          {selected && <AttemptEvidence key={selected.attempt.promptRunId} dir={dir} todoRef={todoRef} entry={selected} />}
        </>
      )}
    </section>
  );
}

function AttemptRow({ entry, selected, onSelect }: { entry: VerificationAttempt; selected: boolean; onSelect: () => void }) {
  const { attempt } = entry;
  const started = attempt.startedAt || attempt.queuedAt || attempt.createdAt;
  const failing = attemptRunArtifacts(entry).reduce((sum, item) => sum + item.artifact.failed + (item.artifact.lint_violations ?? 0), 0);
  return (
    <li>
      <Button
        variant="ghost"
        type="button"
        onClick={onSelect}
        aria-pressed={selected}
        className={`flex h-auto w-full items-center justify-start gap-2 rounded-none px-3 py-2 text-left text-xs ${selected ? 'bg-muted/60' : ''}`}
      >
        <OutcomeIcon outcome={entry.outcome} />
        <span className="font-medium tabular-nums">#{attempt.ordinal}</span>
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
    </li>
  );
}

function AttemptEvidence({ dir, todoRef, entry }: { dir: string; todoRef: string; entry: VerificationAttempt }) {
  const tabs = attemptTabs(entry);
  const [tab, setTab] = useState<AttemptTab>(() => defaultAttemptTab(entry));
  const active = tabs[tab].available ? tab : defaultAttemptTab(entry);
  const artifacts = attemptRunArtifacts(entry);
  const sessionId = entry.attempt.executionSessionId || entry.attempt.providerSessionId;

  return (
    <div className="border-t border-border">
      <div className="flex flex-nowrap gap-1 overflow-x-auto border-b border-border px-3 pt-2">
        {(Object.keys(TAB_LABELS) as AttemptTab[]).map((key) => {
          const state = tabs[key];
          return (
            <Button
              key={key}
              variant="ghost"
              type="button"
              disabled={!state.available}
              title={state.reason}
              aria-pressed={key === active}
              onClick={() => setTab(key)}
              className={`-mb-px inline-flex h-auto shrink-0 items-center gap-1.5 border-b-2 px-2.5 py-1.5 text-xs font-medium ${
                key === active ? 'border-primary text-foreground' : 'border-transparent text-muted-foreground hover:text-foreground'
              }`}
            >
              {TAB_LABELS[key]}
              {typeof state.count === 'number' && state.count > 0 && (
                <span className="rounded-full border border-border bg-background px-1.5 py-0.5 text-[10px] tabular-nums text-muted-foreground">
                  {state.count}
                </span>
              )}
            </Button>
          );
        })}
      </div>
      {!tabs[active].available ? (
        <p className="px-3 py-4 text-xs text-muted-foreground">{tabs[active].reason}</p>
      ) : active === 'session' ? (
        <VerificationAttemptSession dir={dir} todoRef={todoRef} sessionId={sessionId!} />
      ) : active === 'output' ? (
        <VerificationStepOutput steps={attemptOutputSteps(entry)} checklist={entry.checklist} />
      ) : (
        <div className="space-y-2 p-3">
          {artifacts
            .filter((item) => item.kind === active)
            .map((item) => <AttemptRunPanel key={item.artifact.run_id} dir={dir} artifact={item.artifact} stepName={item.step.name} />)}
        </div>
      )}
    </div>
  );
}

/**
 * One runner step's results: the recorded summary renders immediately, the full
 * tree is fetched from the .gavel store behind it. A missing artifact file says
 * so and leaves the summary standing.
 */
function AttemptRunPanel({ dir, artifact, stepName }: { dir: string; artifact: RunArtifact; stepName?: string }) {
  const [snapshot, setSnapshot] = useState<RunSnapshot | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    let cancelled = false;
    setSnapshot(null);
    setError('');
    fetchRunSnapshot({ dir, runId: artifact.run_id })
      .then((loaded) => {
        if (!cancelled) setSnapshot(loaded);
      })
      .catch((reason) => {
        if (!cancelled) setError(reason instanceof Error ? reason.message : String(reason));
      });
    return () => {
      cancelled = true;
    };
  }, [dir, artifact.run_id]);

  return (
    <section className="rounded border border-border">
      <header className="flex flex-wrap items-center gap-2 border-b border-border px-2.5 py-1.5 text-xs">
        <span className="min-w-0 flex-1 truncate font-medium">{stepName || artifact.kind}</span>
        <RunCounts artifact={artifact} />
      </header>
      {artifact.error && <p className="px-2.5 py-1.5 text-[11px] text-red-600">{artifact.error}</p>}
      {artifact.failures && artifact.failures.length > 0 && !snapshot && (
        <ul className="space-y-0.5 px-2.5 py-1.5 text-[11px]">
          {artifact.failures.map((failure, index) => (
            <li key={`${failure.name}:${index}`} className="truncate text-red-600" title={failure.message}>
              {failure.suite ? `${failure.suite} > ` : ''}
              {failure.name}
            </li>
          ))}
          {artifact.truncated ? <li className="text-muted-foreground">…and {artifact.truncated} more</li> : null}
        </ul>
      )}
      {error ? (
        <p className="px-2.5 py-1.5 text-[11px] text-muted-foreground">Full results unavailable: {error}</p>
      ) : !snapshot ? (
        <p className="flex items-center gap-2 px-2.5 py-1.5 text-[11px] text-muted-foreground">
          <Spinner /> Loading {artifact.run_id}…
        </p>
      ) : (
        <div className="min-h-0">
          <TestRunResults
            snapshot={snapshot}
            done
            runKey={artifact.run_id}
            projectName={projectNameFor(dir)}
            projectDir={dir}
            emptyMessage="This verification run recorded no tests or lint findings."
          />
        </div>
      )}
    </section>
  );
}

function RunCounts({ artifact }: { artifact: RunArtifact }) {
  const parts: string[] = [];
  if (artifact.total > 0) parts.push(`${artifact.passed}/${artifact.total} passed`);
  if (artifact.failed > 0) parts.push(`${artifact.failed} failed`);
  if (artifact.warned > 0) parts.push(`${artifact.warned} warned`);
  if (artifact.skipped > 0) parts.push(`${artifact.skipped} skipped`);
  if (artifact.lint_violations) parts.push(`${artifact.lint_violations} violations`);
  return <span className="shrink-0 text-[11px] tabular-nums text-muted-foreground">{parts.join(' · ') || 'no results'}</span>;
}

function projectNameFor(dir: string): string {
  const trimmed = dir.replace(/\/+$/, '');
  return trimmed.slice(trimmed.lastIndexOf('/') + 1) || 'workspace';
}

function OutcomeIcon({ outcome }: { outcome: AttemptOutcome }) {
  if (outcome === 'running') return <Spinner className="size-3.5 shrink-0" />;
  if (outcome === 'passed') return <UiPass className="shrink-0 text-sm text-emerald-600" />;
  if (outcome === 'cancelled') return <UiStop className="shrink-0 text-sm text-muted-foreground" />;
  if (outcome === 'errored') return <UiWarningTriangle className="shrink-0 text-sm text-amber-600" />;
  return <UiError className="shrink-0 text-sm text-red-600" />;
}

function OutcomePill({ outcome }: { outcome: AttemptOutcome }) {
  const tone =
    outcome === 'passed'
      ? 'border-emerald-500/25 bg-emerald-500/10 text-emerald-700 dark:text-emerald-400'
      : outcome === 'failed'
        ? 'border-red-500/25 bg-red-500/10 text-red-700 dark:text-red-400'
        : outcome === 'errored'
          ? 'border-amber-500/25 bg-amber-500/10 text-amber-700 dark:text-amber-400'
          : 'border-border bg-background text-muted-foreground';
  return <span className={`rounded border px-1.5 py-0.5 text-[10px] uppercase ${tone}`}>{outcome}</span>;
}
