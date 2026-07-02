import { useState, useEffect, useMemo, Suspense } from 'react';
import { Modal, Button, JsonSchemaForm, Tabs } from '@flanksource/clicky-ui/components';
import type {
  JsonSchemaObject,
  JsonSchemaProperty,
  PostExtension,
  PreExtension,
  TabItem,
} from '@flanksource/clicky-ui/components';
import { UiFolder } from '@flanksource/clicky-ui/icons';
import type { Project } from '../types';
import { GavelIcon } from './GavelIcon';
import { sectionIcon } from '../icons/settings';
import { useProjectRegistration, ProjectFields } from './ProjectForm';
import { PromptOverrideField, type PromptOverrideValue } from './PromptOverrideField';

// SettingsScope selects what the dialog edits: the user's global ~/.gavel.yaml
// (navbar) or one registered workspace. Project scope also edits the workspace
// registration (directory/repos) via the Project tab, so it carries the Project.
export type SettingsScope =
  | { kind: 'global' }
  | { kind: 'project'; project: Project };

interface Props {
  open: boolean;
  onClose: () => void;
  scope: SettingsScope;
  /** Repos offered in the Project tab's picker (project scope only). */
  repoOptions: string[];
  /** Called after the workspace registration is saved or deleted. */
  onSaved: () => void;
}

// The Project tab edits the workspace registration rather than a .gavel.yaml
// section, so it is handled specially (not part of TABS / the schema form).
const PROJECT_TAB = 'project';

// A settings tab groups one or more top-level .gavel.yaml sections under a human
// label + icon; the form for a tab is the schema scoped to just those sections.
interface TabDef {
  id: string;
  label: string;
  sections: string[];
}

const TABS: TabDef[] = [
  { id: 'review', label: 'AI Review', sections: ['verify'] },
  { id: 'commit', label: 'Commit', sections: ['commit'] },
  { id: 'lint', label: 'Linting', sections: ['lint', 'secrets'] },
  { id: 'tests', label: 'Tests', sections: ['fixtures'] },
  { id: 'todos', label: 'Todos', sections: ['todos', 'checks'] },
  { id: 'processes', label: 'Processes', sections: ['procfile'] },
  { id: 'hooks', label: 'Hooks', sections: ['pre', 'post', 'ssh'] },
];

// Human labels for the top-level sections (used as in-form section headings) and
// for a handful of nested keys the automatic humanizer would render awkwardly.
const SECTION_TITLES: Record<string, string> = {
  verify: 'AI code review',
  commit: 'Commit',
  lint: 'Lint rules',
  secrets: 'Secret scanning',
  fixtures: 'Fixture tests',
  todos: 'Todo runs',
  checks: 'Post-run checks',
  procfile: 'Processes',
  pre: 'Pre-hooks',
  post: 'Post-hooks',
  ssh: 'SSH hook',
};

const FIELD_TITLES: Record<string, string> = {
  mem: 'Memory limit',
  cpu: 'CPU limit',
  gitignore: 'Ignore globs',
  precommit: 'Pre-commit gate',
  linkedDeps: 'Linked dependencies',
  allow: 'Allowlist',
};

const ACRONYMS = new Set(['ssh', 'ai', 'pr', 'cpu', 'id', 'url', 'ci', 'yaml']);

// humanizeKey turns a config key (camelCase / snake / kebab) into a sentence-case
// label: "maxIterations" → "Max iterations", "prContentPrompt" → "PR content prompt".
function humanizeKey(key: string): string {
  const words = key
    .replace(/([a-z0-9])([A-Z])/g, '$1 $2')
    .replace(/[_.-]+/g, ' ')
    .trim()
    .split(/\s+/)
    .filter(Boolean);
  if (words.length === 0) return key;
  return words
    .map((w, i) => {
      const lower = w.toLowerCase();
      if (ACRONYMS.has(lower)) return lower.toUpperCase();
      return i === 0 ? lower.charAt(0).toUpperCase() + lower.slice(1) : lower;
    })
    .join(' ');
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
    : `project=${encodeURIComponent(scope.project.name)}`;
}

function scopeTitle(scope: SettingsScope): string {
  return scope.kind === 'global' ? 'Global settings' : `${scope.project.name} settings`;
}

// promptIdOf returns a schema node's x-prompt-id extension, marking it an
// overridable prompt linked to a registry descriptor.
function promptIdOf(node: JsonSchemaProperty | undefined): string | undefined {
  const id = node?.['x-prompt-id'];
  return typeof id === 'string' ? id : undefined;
}

// decorateNode walks a schema node, forcing prompt-override nodes to object type
// (so the form resolves them as a section the post-extension replaces) and
// stamping a human-readable `title` on every property that lacks one.
function decorateNode(node: JsonSchemaProperty, topLevel: boolean): void {
  if (promptIdOf(node) && node.properties) node.type = 'object';
  if (node.properties) {
    for (const [key, child] of Object.entries(node.properties)) {
      if (child.title == null) {
        child.title =
          (topLevel ? SECTION_TITLES[key] : undefined) ?? FIELD_TITLES[key] ?? humanizeKey(key);
      }
      decorateNode(child, false);
    }
  }
  if (node.items) decorateNode(node.items, false);
  if (node.additionalProperties && typeof node.additionalProperties === 'object') {
    decorateNode(node.additionalProperties, false);
  }
  for (const key of ['oneOf', 'anyOf', 'allOf'] as const) {
    const subs = node[key];
    if (Array.isArray(subs)) for (const sub of subs) decorateNode(sub as JsonSchemaProperty, false);
  }
}

function decorateSchema(schema: JsonSchemaObject): JsonSchemaObject {
  const clone: JsonSchemaObject = structuredClone(schema);
  decorateNode(clone, true);
  return clone;
}

// sectionSchema scopes the full schema to one tab's sections so each tab renders
// only its slice. The full value is still passed to the form (it edits keys in
// place), so switching tabs never drops another tab's settings.
function sectionSchema(schema: JsonSchemaObject, sections: string[]): JsonSchemaObject {
  const properties: Record<string, JsonSchemaProperty> = {};
  for (const key of sections) {
    const prop = schema.properties?.[key];
    if (prop) properties[key] = prop;
  }
  return { type: 'object', properties, 'x-order': sections };
}

// sectionIconPre adds each top-level section's glyph to its form heading. It
// keys off the field name, so nested fields that reuse a section name (e.g.
// commit.lint) inherit the matching icon too, which reads sensibly.
const sectionIconPre: PreExtension = (field) => {
  const Glyph = sectionIcon[field.key];
  if (!Glyph || field.labelIcon != null) return field;
  return { ...field, labelIcon: <Glyph className="shrink-0 text-[15px] text-muted-foreground" /> };
};

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

// SettingsDialog edits one .gavel.yaml file as a schema-driven form, split into
// tabs by top-level section. It loads and saves a single layer (never the merged
// view) so editing the project file does not bake in global values. Overridable
// AI prompts render with a segmented Inline/File editor that shows the built-in
// default. Note: saving rewrites the file from the form, so hand-written comments
// are not preserved.
export function SettingsDialog({ open, onClose, scope, repoOptions, onSaved }: Props) {
  const [schema, setSchema] = useState<JsonSchemaObject | null>(schemaCache);
  const [registry, setRegistry] = useState<Record<string, PromptDescriptor> | null>(promptsCache);
  const [value, setValue] = useState<Record<string, unknown>>({});
  const [path, setPath] = useState('');
  const [tab, setTab] = useState(TABS[0].id);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [saved, setSaved] = useState(false);

  const reg = useProjectRegistration(open, scope.kind === 'project' ? scope.project : null);
  const projectName = scope.kind === 'project' ? scope.project.name : null;
  const isProjectTab = scope.kind === 'project' && tab === PROJECT_TAB;

  const query = useMemo(() => scopeQuery(scope), [scope]);
  const post = useMemo<PostExtension[]>(() => (registry ? [promptPost(registry)] : []), [registry]);
  const pre = useMemo<PreExtension[]>(() => [sectionIconPre], []);
  const sectionTab = TABS.find(t => t.id === tab) ?? TABS[0];
  const tabSchema = useMemo(
    () => (schema ? sectionSchema(schema, sectionTab.sections) : null),
    [schema, sectionTab],
  );

  // Reset to the first relevant tab whenever the dialog opens or the target
  // project changes: Project for a workspace, AI Review for the global config.
  useEffect(() => {
    if (open) setTab(scope.kind === 'project' ? PROJECT_TAB : TABS[0].id);
  }, [open, scope.kind, projectName]);

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

  // Save the .gavel.yaml layer (schema-form tabs).
  async function saveConfig() {
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

  // The footer Save persists whatever the active tab edits: the workspace
  // registration on the Project tab, otherwise the .gavel.yaml config.
  async function onSave() {
    if (isProjectTab) {
      if (await reg.save()) { setSaved(true); onSaved(); }
    } else {
      await saveConfig();
    }
  }

  async function onDelete() {
    if (await reg.remove()) { onSaved(); onClose(); }
  }

  const configReady = schema && registry && tabSchema;
  // clicky-ui ships React 18 types; our icon components are typed against React
  // 19, so the structurally-identical icon prop needs a cast at this boundary.
  const sectionTabItems: TabItem[] = TABS.map(t => ({
    id: t.id,
    label: t.label,
    icon: sectionIcon[t.sections[0]] as unknown as TabItem['icon'],
  }));
  const tabItems: TabItem[] = scope.kind === 'project'
    ? [{ id: PROJECT_TAB, label: 'Project', icon: UiFolder as unknown as TabItem['icon'] }, ...sectionTabItems]
    : sectionTabItems;

  return (
    <Modal
      open
      onClose={onClose}
      title={scopeTitle(scope)}
      size="xl"
      className="max-w-5xl"
      footer={
        <div className="flex items-center justify-between gap-2">
          {isProjectTab ? (
            <Button variant="destructive" size="sm" onClick={onDelete} loading={reg.deleting}>Delete</Button>
          ) : (
            <span className="truncate font-mono text-xs text-muted-foreground" title={path}>{path}</span>
          )}
          <div className="flex items-center gap-2">
            {saved && <span className="text-xs text-green-600 dark:text-green-400">Saved</span>}
            <Button variant="outline" onClick={onClose}>Close</Button>
            <Button
              onClick={onSave}
              loading={isProjectTab ? reg.saving : saving}
              disabled={isProjectTab ? reg.saving : (loading || !configReady)}
            >
              Save
            </Button>
          </div>
        </div>
      }
    >
      <div className="flex flex-col gap-3">
        {error && <div className="text-sm text-destructive">{error}</div>}
        <Tabs
          tabs={tabItems}
          value={tab}
          onChange={setTab}
          className="flex-nowrap whitespace-nowrap overflow-x-auto"
        />
        <div className="max-h-[62vh] space-y-3 overflow-y-auto pr-1">
          {isProjectTab ? (
            <ProjectFields reg={reg} repoOptions={repoOptions} />
          ) : loading || !configReady ? (
            <div className="flex items-center gap-2 py-8 text-sm text-muted-foreground">
              <GavelIcon name="svg-spinners:ring-resize" /> Loading…
            </div>
          ) : (
            <Suspense fallback={<div className="py-8 text-sm text-muted-foreground">Loading editor…</div>}>
              <JsonSchemaForm
                schema={tabSchema}
                value={value}
                onChange={(next) => { setValue(next); setSaved(false); }}
                pre={pre}
                post={post}
                size="sm"
              />
            </Suspense>
          )}
        </div>
      </div>
    </Modal>
  );
}
