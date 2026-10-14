import { useState } from 'react';
import { Button, Field } from '@flanksource/clicky-ui/components';
import type { TodoItem } from '../../types';
import { inputClass } from './format';

export function todoPRLink(value: string): NonNullable<TodoItem['pr']> | undefined {
  if (!value.trim()) return undefined;
  const url = new URL(value.trim());
  const number = Number(url.pathname.split('/').filter(Boolean).at(-1));
  if (!Number.isSafeInteger(number) || number <= 0) throw new Error('Enter a pull request URL with a positive PR number.');
  return { number, url: url.href.replace(/\/$/, '') };
}

export function TodoPRField({ value, onChange, disabled }: { value: string; onChange: (value: string) => void; disabled?: boolean }) {
  return <Field label="Pull request URL (optional)">
    <input type="url" aria-label="Pull request URL" className={inputClass} value={value} onChange={event => onChange(event.currentTarget.value)} disabled={disabled} />
    <p className="mt-1 text-xs text-muted-foreground">Link a GitHub or Azure DevOps PR. Completed turns push fixes to this PR and verify all its checks and feedback.</p>
  </Field>;
}

export function TodoPRLink({ pr, disabled, onSave }: { pr?: TodoItem['pr']; disabled?: boolean; onSave: (payload: { pr?: TodoItem['pr']; clearPR?: boolean }) => Promise<boolean> }) {
  const [editing, setEditing] = useState(false);
  const [value, setValue] = useState(pr?.url ?? '');
  const [error, setError] = useState('');
  async function save() {
    setError('');
    try {
      const link = todoPRLink(value);
      if (await onSave(link ? { pr: link } : { clearPR: true })) setEditing(false);
    } catch (err) { setError(err instanceof Error ? err.message : 'Invalid PR URL'); }
  }
  if (editing) return <div className="space-y-2">
    <TodoPRField value={value} onChange={setValue} disabled={disabled} />
    {error && <p role="alert" className="text-sm text-destructive">{error}</p>}
    <div className="flex gap-2"><Button disabled={disabled} onClick={save}>Save PR link</Button><Button variant="outline" onClick={() => setEditing(false)}>Cancel</Button></div>
  </div>;
  return <div className="flex flex-wrap items-center gap-2 text-xs">
    {pr && <a href={pr.url} target="_blank" rel="noreferrer" className="text-primary hover:underline">PR #{pr.number}</a>}
    <Button variant="ghost" size="sm" disabled={disabled} onClick={() => { setValue(pr?.url ?? ''); setError(''); setEditing(true); }}>{pr ? 'Edit PR link' : 'Link pull request'}</Button>
    {pr && <span className="text-muted-foreground">Run pushes fixes here and verifies the entire PR.</span>}
  </div>;
}
