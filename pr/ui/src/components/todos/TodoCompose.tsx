import { useState } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import { GavelIcon } from '../GavelIcon';
import { inputClass } from './format';

// TodoEditForm edits a todo's title and body. It is fully controlled so the
// parent owns the draft state and can discard it on cancel or todo switch.
export function TodoEditForm({
  title,
  body,
  busy,
  onTitle,
  onBody,
  onSave,
  onCancel,
}: {
  title: string;
  body: string;
  busy: boolean;
  onTitle: (value: string) => void;
  onBody: (value: string) => void;
  onSave: () => void;
  onCancel: () => void;
}) {
  const canSave = title.trim().length > 0 && !busy;
  return (
    <section className="space-y-2 rounded-md border border-border bg-background p-3">
      <label className="block text-xs font-semibold uppercase text-muted-foreground">
        Title
        <input
          className={`${inputClass} mt-1`}
          value={title}
          disabled={busy}
          onChange={e => onTitle(e.currentTarget.value)}
          aria-label="Edit todo title"
        />
      </label>
      <label className="block text-xs font-semibold uppercase text-muted-foreground">
        Body
        <textarea
          className={`${inputClass} mt-1 h-48 resize-y font-mono`}
          value={body}
          disabled={busy}
          onChange={e => onBody(e.currentTarget.value)}
          placeholder="Markdown body"
          aria-label="Edit todo body"
        />
      </label>
      <div className="flex justify-end gap-2">
        <Button variant="outline" onClick={onCancel} disabled={busy}>Cancel</Button>
        <Button onClick={onSave} loading={busy} disabled={!canSave}>Save</Button>
      </div>
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
    <section className="rounded-md border border-border bg-background">
      <div className="flex items-center gap-2 px-3 py-2">
        <GavelIcon name="codicon:comment" className="shrink-0 text-xs text-muted-foreground" />
        <span className="min-w-0 flex-1 truncate text-xs font-semibold uppercase text-muted-foreground">Add comment</span>
      </div>
      <div className="space-y-2 border-t border-border px-3 py-3">
        <textarea
          className={`${inputClass} h-20 resize-y`}
          value={text}
          disabled={busy}
          onChange={e => setText(e.currentTarget.value)}
          placeholder="Leave a comment…"
          aria-label="Comment body"
        />
        <div className="flex justify-end gap-2">
          {closed && (
            <Button variant="outline" onClick={() => submit(true)} loading={busy} disabled={!trimmed}>
              Reopen &amp; comment
            </Button>
          )}
          <Button onClick={() => submit(false)} loading={busy} disabled={!trimmed}>Comment</Button>
        </div>
      </div>
    </section>
  );
}
