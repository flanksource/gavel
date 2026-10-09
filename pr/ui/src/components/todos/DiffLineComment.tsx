import { useState } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import { Markdown } from '@flanksource/clicky-ui/data';
import { UiCheck, UiChevronDown, UiChevronUp, UiComment, UiRefresh } from '@flanksource/clicky-ui/icons';
import { RelativeTime } from '../RelativeTime';
import { inputClass } from './format';
import type { LineComment } from './lineComments';

function errorMessage(error: unknown) {
  return error instanceof Error ? error.message : String(error);
}

function Author({ actor }: { actor?: string }) {
  const name = actor?.trim();
  return <span className="font-medium text-foreground">{name || 'Someone'}</span>;
}

/**
 * One review comment pinned under a diff line: author, age, markdown body and a
 * Resolve / Reopen button. A resolved comment collapses to a dimmed one-liner
 * that can be expanded or reopened.
 */
export function DiffLineComment({ comment, onResolve }: {
  comment: LineComment;
  onResolve: (id: string, resolved: boolean) => Promise<void>;
}) {
  const { event, resolved } = comment;
  const [expanded, setExpanded] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const id = event.id;
  if (!id) throw new Error('A line comment needs an id to be resolved');

  async function toggle() {
    setBusy(true);
    setError('');
    try {
      await onResolve(id!, !resolved);
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setBusy(false);
    }
  }

  const collapsed = resolved && !expanded;
  const body = event.body?.trim() ?? '';
  return (
    <div
      data-testid="diff-line-comment"
      data-resolved={resolved}
      className={`mx-2 my-1 rounded-md border border-border bg-card px-3 py-2 text-xs shadow-sm ${resolved ? 'opacity-60' : ''}`}
    >
      <div className="flex items-center gap-2">
        <UiComment className="shrink-0 text-muted-foreground" aria-hidden="true" />
        <Author actor={event.actor} />
        {event.timestamp && <RelativeTime iso={event.timestamp} className="text-muted-foreground" />}
        {resolved && <span className="rounded bg-emerald-500/10 px-1 text-[10px] font-medium text-emerald-700 [[data-theme=dark]_&]:text-emerald-300">Resolved</span>}
        {collapsed && <span className="min-w-0 flex-1 truncate text-muted-foreground">{body.split('\n', 1)[0]}</span>}
        <span className="ml-auto flex shrink-0 items-center gap-1">
          {resolved && (
            <Button
              variant="ghost"
              size="icon"
              type="button"
              aria-label={expanded ? 'Collapse comment' : 'Expand comment'}
              title={expanded ? 'Collapse' : 'Expand'}
              onClick={() => setExpanded(value => !value)}
              className="size-6"
            >
              {expanded ? <UiChevronUp className="text-xs" /> : <UiChevronDown className="text-xs" />}
            </Button>
          )}
          <Button variant="outline" size="sm" type="button" loading={busy} disabled={busy} onClick={() => void toggle()} className="h-6 gap-1 px-2 text-[11px]">
            {resolved ? <UiRefresh className="text-xs" /> : <UiCheck className="text-xs" />}
            {resolved ? 'Reopen' : 'Resolve'}
          </Button>
        </span>
      </div>
      {!collapsed && body && <Markdown text={body} className="mt-1.5 text-xs" />}
      {error && <div role="alert" className="mt-1.5 whitespace-pre-wrap text-red-600">{error}</div>}
    </div>
  );
}

/** The inline composer opened by a diff line's "+" button. */
export function DiffLineComposer({ onSave, onCancel }: {
  onSave: (body: string) => Promise<void>;
  onCancel: () => void;
}) {
  const [text, setText] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const trimmed = text.trim();

  async function save() {
    if (!trimmed || busy) return;
    setBusy(true);
    setError('');
    try {
      await onSave(trimmed);
    } catch (reason) {
      setError(errorMessage(reason));
      setBusy(false);
    }
  }

  return (
    <div data-testid="diff-line-composer" className="mx-2 my-1 space-y-2 rounded-md border border-border bg-card p-2 shadow-sm">
      <textarea
        className={`${inputClass} h-20 resize-y`}
        value={text}
        disabled={busy}
        autoFocus
        onChange={event => setText(event.currentTarget.value)}
        onKeyDown={event => {
          if (event.key === 'Escape') onCancel();
          else if (event.key === 'Enter' && (event.metaKey || event.ctrlKey)) void save();
        }}
        placeholder="Comment on this line…"
        aria-label="Line comment body"
      />
      {error && <div role="alert" className="whitespace-pre-wrap text-xs text-red-600">{error}</div>}
      <div className="flex justify-end gap-2">
        <Button variant="outline" type="button" onClick={onCancel} disabled={busy}>Cancel</Button>
        <Button type="button" onClick={() => void save()} loading={busy} disabled={!trimmed}>Save</Button>
      </div>
    </div>
  );
}
