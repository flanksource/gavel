import { useState } from 'react';
import { AnsiHtml } from '@flanksource/clicky-ui/data';
import { UiCheck, UiCircleOutline, UiCircleX, UiClose, UiWarningTriangle } from '@flanksource/clicky-ui/icons';
import { Spinner } from '../icons/Spinner';
import { Elapsed } from './Elapsed';

export type CommitQueueEntryStatus = 'pending' | 'running' | 'success' | 'failed' | 'warning' | 'canceled';
export type CommitQueueAction = 'commit' | 'open-pr';

export interface CommitQueueEntry {
  id: string;
  action: CommitQueueAction;
  files: string[];
  status: CommitQueueEntryStatus;
  startedAt?: string;
  endedAt?: string;
  exitCode?: number;
  output?: string;
  error?: string;
}

export interface CommitQueueStatus {
  runId?: string;
  href?: string;
  running: boolean;
  entries?: CommitQueueEntry[];
}

export function isCommitQueueEntryPending(entry: CommitQueueEntry) {
  return entry.status === 'pending' || entry.status === 'running';
}

// commitQueueLockedFiles are the paths already claimed by a group that has not
// committed yet: they stay out of the next selection so two groups can never
// commit the same file.
export function commitQueueLockedFiles(queue: CommitQueueStatus | undefined) {
  const locked = new Map<string, number>();
  (queue?.entries ?? []).forEach((entry, index) => {
    if (!isCommitQueueEntryPending(entry)) return;
    entry.files.forEach(file => locked.set(file, index + 1));
  });
  return locked;
}

interface Props {
  queue: CommitQueueStatus;
  onCancel: (id: string) => void;
}

export function CommitQueuePanel({ queue, onCancel }: Props) {
  const [expanded, setExpanded] = useState<string>('');
  const entries = queue.entries ?? [];
  if (entries.length === 0) return null;
  const waiting = entries.filter(entry => entry.status === 'pending').length;

  return (
    <section className="shrink-0 border-b border-border bg-muted/30" aria-label="Commit queue" aria-live="polite">
      <div className="flex items-center gap-2 px-4 py-2 text-xs">
        <span className="font-medium text-foreground">Commit queue</span>
        <span className="text-muted-foreground">
          {entries.length} group{entries.length === 1 ? '' : 's'}{waiting > 0 ? `, ${waiting} waiting` : ''}
        </span>
        {queue.href && (
          <a href={queue.href} className="ml-auto font-medium text-primary hover:underline">Details</a>
        )}
      </div>
      <ol className="divide-y divide-border border-t border-border">
        {entries.map((entry, index) => (
          <CommitQueueRow
            key={entry.id}
            entry={entry}
            position={index + 1}
            expanded={expanded === entry.id}
            onToggle={() => setExpanded(current => current === entry.id ? '' : entry.id)}
            onCancel={() => onCancel(entry.id)}
          />
        ))}
      </ol>
    </section>
  );
}

function CommitQueueRow({ entry, position, expanded, onToggle, onCancel }: {
  entry: CommitQueueEntry;
  position: number;
  expanded: boolean;
  onToggle: () => void;
  onCancel: () => void;
}) {
  const output = entry.output?.trimEnd();
  const summary = entry.files.length === 1 ? entry.files[0] : `${entry.files.length} files`;
  const actionLabel = entry.action === 'open-pr' ? 'Open PR' : 'Commit';
  return (
    <li>
      <div className="flex items-center gap-2 px-4 py-1.5 text-xs">
        <span className="w-4 shrink-0 text-right font-mono text-muted-foreground">{position}</span>
        <CommitQueueIcon status={entry.status} />
        <span className="shrink-0 font-medium text-foreground">{actionLabel}</span>
        <button
          type="button"
          onClick={onToggle}
          aria-expanded={expanded}
          className="min-w-0 flex-1 truncate text-left font-mono text-foreground hover:underline"
          title={entry.files.join('\n')}
        >
          {summary}
        </button>
        <span className={statusTone(entry.status)}>{entry.status}</span>
        <Elapsed startedAt={entry.startedAt} endedAt={entry.endedAt} running={entry.status === 'running'} />
        {entry.exitCode !== undefined && entry.exitCode !== 0 && (
          <span className="font-mono text-muted-foreground">Exit {entry.exitCode}</span>
        )}
        {isCommitQueueEntryPending(entry) && (
          <button
            type="button"
            onClick={onCancel}
            aria-label={`Cancel ${entry.action === 'open-pr' ? 'open PR' : 'commit'} group ${position}`}
            title={entry.status === 'running' ? 'Stop this commit' : 'Remove from the queue'}
            className="inline-flex size-5 shrink-0 items-center justify-center rounded text-muted-foreground hover:bg-accent hover:text-accent-foreground"
          >
            <UiClose className="size-3.5" />
          </button>
        )}
      </div>
      {entry.error && <div className="px-4 pb-1.5 pl-10 text-[11px] text-red-600">{entry.error}</div>}
      {expanded && (
        output
          ? <AnsiHtml as="pre" text={output} className="max-h-40 overflow-auto border-t border-border bg-black px-4 py-2 text-[11px] leading-relaxed whitespace-pre-wrap text-gray-100" />
          : <div className="border-t border-border px-4 py-2 text-[11px] text-muted-foreground">
              {entry.status === 'pending' ? `Waiting for ${entry.files.length === 1 ? 'the running commit' : 'earlier groups'}…` : 'No command output.'}
            </div>
      )}
    </li>
  );
}

function CommitQueueIcon({ status }: { status: CommitQueueEntryStatus }) {
  if (status === 'running') return <Spinner />;
  if (status === 'success') return <UiCheck className="size-3.5 shrink-0 text-green-600" />;
  if (status === 'failed') return <UiWarningTriangle className="size-3.5 shrink-0 text-red-600" />;
  if (status === 'canceled') return <UiCircleX className="size-3.5 shrink-0 text-muted-foreground" />;
  return <UiCircleOutline className="size-3.5 shrink-0 text-muted-foreground" />;
}

function statusTone(status: CommitQueueEntryStatus) {
  if (status === 'failed') return 'text-red-600';
  if (status === 'success') return 'text-green-600';
  if (status === 'running') return 'text-primary';
  return 'text-muted-foreground';
}
