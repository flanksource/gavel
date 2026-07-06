import { useState } from 'react';
import { Modal, Button, SegmentedControl } from '@flanksource/clicky-ui/components';
import { SpecRuntimeEditor, type AISpecRuntimeValue } from '@flanksource/clicky-ui/ai';
import { runtimeValueToPayload, specToRuntimeValue, type PromptDetail } from './promptSpec';

// A .prompt document is model / prompt / workspace / permissions / environment /
// cli. The verify section is a gavel fixture edited elsewhere, so it is left out.
// SpecSectionId is not exported from the barrel, but these string literals are
// assignable to the sections prop's element type.
const SECTIONS = ['model', 'prompt', 'workspace', 'permissions', 'environment', 'cli'] as const;

const SOURCE_OPTIONS: { id: 'inline' | 'file'; label: string }[] = [
  { id: 'inline', label: 'Inline' },
  { id: 'file', label: 'File' },
];

interface Props {
  open: boolean;
  id: string;
  title: string;
  scopeQuery: string;
  detail: PromptDetail;
  onClose: () => void;
  onSaved: (next: PromptDetail) => void;
}

// PromptEditorDialog is the nested dialog that edits one prompt as a full spec +
// body with the clicky SpecRuntimeEditor. Save serializes back to a .prompt on
// the server (inline into .gavel.yaml or into the referenced file) and returns
// the refreshed detail; Cancel discards. It is hosted in clicky's own Modal so it
// participates in the modal stack (z-index + escape) above the settings dialog.
export function PromptEditorDialog({ open, id, title, scopeQuery, detail, onClose, onSaved }: Props) {
  const [value, setValue] = useState<AISpecRuntimeValue>(() =>
    specToRuntimeValue(detail.spec, detail.body),
  );
  const [source, setSource] = useState<'inline' | 'file'>(detail.source === 'file' ? 'file' : 'inline');
  const [path, setPath] = useState(detail.path ?? '');
  const [error, setError] = useState('');
  const [saving, setSaving] = useState(false);

  if (!open) return null;

  async function save() {
    setSaving(true);
    setError('');
    const { spec, body } = runtimeValueToPayload(value);
    try {
      const res = await fetch(`/api/settings/prompts/${encodeURIComponent(id)}?${scopeQuery}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          source,
          path: source === 'file' ? path : undefined,
          spec,
          body,
          baseRaw: detail.raw,
        }),
      });
      if (!res.ok) throw new Error((await res.text()) || `save failed (${res.status})`);
      onSaved((await res.json()) as PromptDetail);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'save failed');
    } finally {
      setSaving(false);
    }
  }

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={`Edit prompt · ${title}`}
      size="2xl"
      footer={
        <div className="flex items-center justify-between gap-3">
          <span className="whitespace-pre-wrap text-sm text-red-600 dark:text-red-400">{error}</span>
          <div className="flex gap-2">
            <Button variant="ghost" onClick={onClose} disabled={saving}>
              Cancel
            </Button>
            <Button onClick={save} disabled={saving || (source === 'file' && path.trim() === '')}>
              {saving ? 'Saving…' : 'Save'}
            </Button>
          </div>
        </div>
      }
    >
      <div className="space-y-3">
        <div className="flex items-center gap-2">
          <span className="text-xs font-medium text-muted-foreground">Save to</span>
          <SegmentedControl
            size="sm"
            value={source}
            options={SOURCE_OPTIONS}
            onChange={setSource}
            aria-label="Save location"
          />
          {source === 'file' && (
            <input
              type="text"
              className="flex-1 rounded-md border border-border bg-background p-1.5 font-mono text-xs"
              value={path}
              placeholder="./prompts/my-prompt.prompt"
              spellCheck={false}
              onChange={(e) => setPath(e.target.value)}
            />
          )}
        </div>

        <SpecRuntimeEditor value={value} onChange={setValue} sections={SECTIONS} />
      </div>
    </Modal>
  );
}
