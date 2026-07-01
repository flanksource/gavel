import { useState } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import { GavelIcon } from '../GavelIcon';
import { inputClass } from './format';

// EditActions is the shared Cancel/Save row for the inline field editors.
function EditActions({ busy, canSave, onSave, onCancel }: {
  busy: boolean;
  canSave: boolean;
  onSave: () => void;
  onCancel: () => void;
}) {
  return (
    <div className="flex justify-end gap-2">
      <Button variant="outline" onClick={onCancel} disabled={busy}>Cancel</Button>
      <Button onClick={onSave} loading={busy} disabled={!canSave}>Save</Button>
    </div>
  );
}

// TodoTitleEditor edits a todo's title inline. Controlled so the parent owns the
// draft and discards it on cancel or todo switch. Enter saves, Escape cancels.
export function TodoTitleEditor({ value, busy, onChange, onSave, onCancel }: {
  value: string;
  busy: boolean;
  onChange: (value: string) => void;
  onSave: () => void;
  onCancel: () => void;
}) {
  return (
    <div className="mt-1 space-y-2">
      <input
        className={inputClass}
        value={value}
        disabled={busy}
        autoFocus
        onChange={e => onChange(e.currentTarget.value)}
        onKeyDown={e => {
          if (e.key === 'Enter') onSave();
          else if (e.key === 'Escape') onCancel();
        }}
        aria-label="Edit todo title"
      />
      <EditActions busy={busy} canSave={value.trim().length > 0 && !busy} onSave={onSave} onCancel={onCancel} />
    </div>
  );
}

// TodoBodyEditor edits a todo's markdown body inline.
export function TodoBodyEditor({ value, busy, onChange, onSave, onCancel }: {
  value: string;
  busy: boolean;
  onChange: (value: string) => void;
  onSave: () => void;
  onCancel: () => void;
}) {
  return (
    <section className="space-y-2 rounded-lg border border-border bg-card p-3 shadow-sm">
      <textarea
        className={`${inputClass} h-48 resize-y font-mono`}
        value={value}
        disabled={busy}
        autoFocus
        onChange={e => onChange(e.currentTarget.value)}
        placeholder="Markdown body"
        aria-label="Edit todo body"
      />
      <EditActions busy={busy} canSave={!busy} onSave={onSave} onCancel={onCancel} />
    </section>
  );
}

// TodoCommentBox composes a comment. When the todo is closed it also offers
// "Reopen & comment", which reopens the todo and posts the comment in one go.
export function TodoCommentBox({
  closed,
  busy,
  onComment,
}: {
  closed: boolean;
  busy: boolean;
  // onComment posts the comment (reopening first when reopen is true) and
  // resolves true on success so the box can clear its draft.
  onComment: (body: string, reopen: boolean) => Promise<boolean>;
}) {
  const [text, setText] = useState('');
  const trimmed = text.trim();

  async function submit(reopen: boolean) {
    if (!trimmed || busy) return;
    if (await onComment(trimmed, reopen)) setText('');
  }

  return (
    <section className="overflow-hidden rounded-lg border border-border bg-card shadow-sm">
      <div className="flex items-center gap-2 border-b border-border bg-muted/30 px-3 py-2.5">
        <span className="inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-md border border-border bg-background text-muted-foreground">
          <GavelIcon name="codicon:comment" className="text-xs" />
        </span>
        <span className="min-w-0 flex-1 truncate text-xs font-semibold uppercase tracking-wide text-muted-foreground">Add comment</span>
      </div>
      <div className="space-y-2 px-3 py-3">
        <textarea
          className={`${inputClass} h-20 resize-y`}
          value={text}
          disabled={busy}
          onChange={e => setText(e.currentTarget.value)}
          onKeyDown={e => {
            if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
              e.preventDefault();
              submit(false);
            }
          }}
          placeholder="Leave a comment…"
          aria-label="Comment body"
        />
        <div className="flex flex-wrap items-center justify-between gap-2">
          <span className="text-[11px] text-muted-foreground">Markdown supported · Cmd/Ctrl+Enter to comment</span>
          <span className="flex items-center gap-2">
            {closed && (
              <Button variant="outline" onClick={() => submit(true)} loading={busy} disabled={!trimmed}>
                Reopen &amp; comment
              </Button>
            )}
            <Button onClick={() => submit(false)} loading={busy} disabled={!trimmed}>Comment</Button>
          </span>
        </div>
      </div>
    </section>
  );
}
