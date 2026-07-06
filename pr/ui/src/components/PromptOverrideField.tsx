import { useCallback, useEffect, useState } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import { PromptEditorDialog } from './settings/PromptEditorDialog';
import type { PromptDetail } from './settings/promptSpec';

// A prompt override is either inline template text or a path to a .prompt file
// (the union the Go PromptOverride marshals). An unset/empty override means the
// built-in default is used.
export type PromptOverrideValue = string | { inline?: string; file?: string };

interface Props {
  value: PromptOverrideValue | undefined;
  onChange: (next: PromptOverrideValue | undefined) => void;
  description?: string;
  /** Stable prompt id (schema x-prompt-id) — the /api/settings/prompts/{id} key. */
  id: string;
  /** Human label for the prompt, shown in the editor dialog title. */
  title: string;
  /** scope=global | project=<name>, scoping the detail request to one layer. */
  scopeQuery: string;
}

type Source = 'default' | 'inline' | 'file';

function sourceOf(value: PromptOverrideValue | undefined): { source: Source; path?: string } {
  if (typeof value === 'string') return value.trim() ? { source: 'inline' } : { source: 'default' };
  if (value && typeof value === 'object') {
    if (value.inline && value.inline.trim()) return { source: 'inline' };
    if (value.file && value.file.trim()) return { source: 'file', path: value.file };
  }
  return { source: 'default' };
}

const BADGE: Record<Source, string> = {
  default: 'bg-muted text-muted-foreground',
  inline: 'bg-blue-100 text-blue-800 dark:bg-blue-900/40 dark:text-blue-200',
  file: 'bg-violet-100 text-violet-800 dark:bg-violet-900/40 dark:text-violet-200',
};

// PromptOverrideField is the settings summary row for one overridable prompt: a
// source badge (default/inline/file), the resolved model, and an Edit button that
// opens the nested spec/body editor. Content is edited only in the dialog (which
// persists server-side); the row keeps the lightweight source switches — Reset to
// default (cleared here, saved with the rest of the form) and, after a dialog
// save, syncing the in-form value so the enclosing Save stays consistent.
export function PromptOverrideField({ value, onChange, description, id, title, scopeQuery }: Props) {
  const [detail, setDetail] = useState<PromptDetail | null>(null);
  const [error, setError] = useState('');
  const [editing, setEditing] = useState(false);

  const load = useCallback(() => {
    setError('');
    fetch(`/api/settings/prompts/${encodeURIComponent(id)}?${scopeQuery}`)
      .then(async (r) => {
        if (!r.ok) throw new Error((await r.text()) || `load failed (${r.status})`);
        return r.json() as Promise<PromptDetail>;
      })
      .then(setDetail)
      .catch((e) => setError(e instanceof Error ? e.message : 'failed to load prompt'));
  }, [id, scopeQuery]);

  useEffect(() => {
    load();
  }, [load]);

  const { source, path } = sourceOf(value);
  const model = (detail?.spec?.model as string | undefined) || 'default';

  function onSaved(next: PromptDetail) {
    setDetail(next);
    onChange(next.source === 'file' ? { file: next.path ?? '' } : { inline: next.raw });
    setEditing(false);
  }

  return (
    <div className="space-y-2">
      {description && <p className="text-xs text-muted-foreground">{description}</p>}

      <div className="flex flex-wrap items-center gap-2">
        <span
          className={`rounded px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wide ${BADGE[source]}`}
        >
          {source}
          {source === 'file' && path ? ` · ${path}` : ''}
        </span>
        <span
          className="rounded bg-muted px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground"
          title="Model"
        >
          {model}
        </span>
        <Button size="sm" variant="secondary" onClick={() => setEditing(true)} disabled={!detail}>
          Edit
        </Button>
        {source !== 'default' && (
          <button
            type="button"
            onClick={() => onChange(undefined)}
            className="text-xs text-muted-foreground underline hover:text-foreground"
          >
            Reset to default
          </button>
        )}
      </div>

      {error && <p className="text-xs text-red-600 dark:text-red-400">{error}</p>}

      {editing && detail && (
        <PromptEditorDialog
          open={editing}
          id={id}
          title={title}
          scopeQuery={scopeQuery}
          detail={detail}
          onClose={() => setEditing(false)}
          onSaved={onSaved}
        />
      )}
    </div>
  );
}
