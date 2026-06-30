import { useState, useEffect, useMemo, Suspense } from 'react';
import { Modal, Button, JsonSchemaForm } from '@flanksource/clicky-ui/components';
import type { JsonSchemaObject, JsonSchemaProperty, PostExtension } from '@flanksource/clicky-ui/components';
import { GavelIcon } from './GavelIcon';
import { PromptOverrideField, type PromptOverrideValue } from './PromptOverrideField';

// SettingsScope selects which .gavel.yaml the dialog edits: the user's global
// ~/.gavel.yaml (navbar) or a registered workspace's root .gavel.yaml (sidebar).
export type SettingsScope =
  | { kind: 'global' }
  | { kind: 'project'; project: string; label: string };

interface Props {
  open: boolean;
  onClose: () => void;
  scope: SettingsScope;
}

// PromptDescriptor mirrors prompts.Prompt from the Go registry: the embedded
// default and metadata for one overridable AI prompt, keyed by x-prompt-id.
interface PromptDescriptor {
  id: string;
  title: string;
  description: string;
  configPath: string;
  default: string;
}

// The schema and prompt registry are the same for every scope, so fetch each
// once and reuse.
let schemaCache: JsonSchemaObject | null = null;
let promptsCache: Record<string, PromptDescriptor> | null = null;

function scopeQuery(scope: SettingsScope): string {
  return scope.kind === 'global'
    ? 'scope=global'
    : `project=${encodeURIComponent(scope.project)}`;
}

function scopeTitle(scope: SettingsScope): string {
  return scope.kind === 'global' ? 'Global settings' : `${scope.label} settings`;
}

// promptIdOf returns a schema node's x-prompt-id extension, marking it an
// overridable prompt linked to a registry descriptor.
function promptIdOf(node: JsonSchemaProperty | undefined): string | undefined {
  const id = node?.['x-prompt-id'];
  return typeof id === 'string' ? id : undefined;
}

// coercePromptNodes walks the schema and forces every prompt-override node (one
// carrying x-prompt-id) to object type, so the form resolves it as a section the
// post-extension replaces with PromptOverrideField instead of rendering the bare
// string|object union. The Go schema stays tool-agnostic; this hint is client-side.
function coercePromptNodes(node: JsonSchemaProperty): void {
  if (promptIdOf(node) && node.properties) node.type = 'object';
  if (node.properties) {
    for (const child of Object.values(node.properties)) coercePromptNodes(child);
  }
}

function decorateSchema(schema: JsonSchemaObject): JsonSchemaObject {
  const clone: JsonSchemaObject = structuredClone(schema);
  coercePromptNodes(clone);
  return clone;
}

// promptPost replaces any prompt-override field's value node with the rich
// PromptOverrideField (segmented Inline/File source + default display), keyed by
// the schema's x-prompt-id → registry default.
function promptPost(registry: Record<string, PromptDescriptor>): PostExtension {
  return (field, nodes) => {
    const id = promptIdOf(field.schema);
    if (!id) return nodes;
    const desc = registry[id];
    return {
      label: nodes.label,
      value: (
        <PromptOverrideField
          value={field.value as PromptOverrideValue | undefined}
          onChange={next => field.onChange(next)}
          defaultText={desc?.default ?? ''}
          description={desc?.description}
        />
      ),
    };
  };
}

// SettingsDialog edits one .gavel.yaml file as a schema-driven form. It loads and
// saves a single layer (never the merged view) so editing the project file does
// not bake in global values. Overridable AI prompts render with a segmented
// Inline/File editor that shows the built-in default. Note: saving rewrites the
// file from the form, so hand-written comments are not preserved.
export function SettingsDialog({ open, onClose, scope }: Props) {
  const [schema, setSchema] = useState<JsonSchemaObject | null>(schemaCache);
  const [registry, setRegistry] = useState<Record<string, PromptDescriptor> | null>(promptsCache);
  const [value, setValue] = useState<Record<string, unknown>>({});
  const [path, setPath] = useState('');
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [saved, setSaved] = useState(false);

  const query = useMemo(() => scopeQuery(scope), [scope]);
  const post = useMemo<PostExtension[]>(() => (registry ? [promptPost(registry)] : []), [registry]);

  useEffect(() => {
    if (!open || schemaCache) return;
    fetch('/api/settings/schema')
      .then(r => r.json())
      .then((s: JsonSchemaObject) => { schemaCache = decorateSchema(s); setSchema(schemaCache); })
      .catch(e => setError(e?.message || 'failed to load schema'));
  }, [open]);

  useEffect(() => {
    if (!open || promptsCache) return;
    fetch('/api/settings/prompts')
      .then(r => r.json())
      .then((list: PromptDescriptor[]) => {
        promptsCache = Object.fromEntries((list ?? []).map(p => [p.id, p]));
        setRegistry(promptsCache);
      })
      .catch(e => setError(e?.message || 'failed to load prompt registry'));
  }, [open]);

  useEffect(() => {
    if (!open) return;
    setError('');
    setSaved(false);
    setLoading(true);
    fetch(`/api/settings/gavel?${query}`)
      .then(async r => {
        if (!r.ok) throw new Error((await r.text()) || `HTTP ${r.status}`);
        return r.json();
      })
      .then((resp) => {
        setValue((resp.config as Record<string, unknown>) ?? {});
        setPath(resp.path || '');
      })
      .catch(e => setError(e?.message || 'failed to load config'))
      .finally(() => setLoading(false));
  }, [open, query]);

  if (!open) return null;

  async function save() {
    setSaving(true);
    setError('');
    try {
      const res = await fetch(`/api/settings/gavel?${query}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(value),
      });
      if (!res.ok) {
        setError((await res.text()) || `HTTP ${res.status}`);
        setSaving(false);
        return;
      }
      setSaved(true);
    } catch (e: any) {
      setError(e?.message || 'save failed');
      setSaving(false);
      return;
    }
    setSaving(false);
  }

  const ready = schema && registry;

  return (
    <Modal
      open
      onClose={onClose}
      title={scopeTitle(scope)}
      size="xl"
      footer={
        <div className="flex items-center justify-between gap-2">
          <span className="truncate font-mono text-xs text-muted-foreground" title={path}>{path}</span>
          <div className="flex items-center gap-2">
            {saved && <span className="text-xs text-green-600 dark:text-green-400">Saved</span>}
            <Button variant="outline" onClick={onClose}>Close</Button>
            <Button onClick={save} loading={saving} disabled={loading || !ready}>Save</Button>
          </div>
        </div>
      }
    >
      <div className="max-h-[70vh] space-y-3 overflow-y-auto pr-1">
        {error && <div className="text-sm text-destructive">{error}</div>}
        {loading || !ready ? (
          <div className="flex items-center gap-2 py-8 text-sm text-muted-foreground">
            <GavelIcon name="svg-spinners:ring-resize" /> Loading…
          </div>
        ) : (
          <Suspense fallback={<div className="py-8 text-sm text-muted-foreground">Loading editor…</div>}>
            <JsonSchemaForm
              schema={schema}
              value={value}
              onChange={(next) => { setValue(next); setSaved(false); }}
              post={post}
              size="sm"
            />
          </Suspense>
        )}
      </div>
    </Modal>
  );
}
