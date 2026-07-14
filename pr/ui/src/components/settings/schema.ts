// Shared schema machinery for the settings surface: the tab→section map, human
// labels, and the decorate/scope helpers that adapt the raw .gavel.yaml JSON
// Schema (from /api/settings/schema) for rendering. Extracted so both the page
// and its tests can import them without pulling in React.

import type { JsonSchemaObject, JsonSchemaProperty } from '@flanksource/clicky-ui/components';

// A settings tab groups one or more top-level .gavel.yaml sections under a human
// label; the form for a tab is the schema scoped to just those sections.
export interface TabDef {
  id: string;
  label: string;
  sections: string[];
}

// The 7 config tabs from the redesign. `status` and `test` are folded into
// AI Review and Tests so every config section is reachable (they have no tab of
// their own in the mockup but must not be silently uneditable).
export const TABS: TabDef[] = [
  { id: 'review', label: 'AI Review', sections: ['verify', 'status'] },
  { id: 'commit', label: 'Commit', sections: ['commit'] },
  { id: 'lint', label: 'Linting', sections: ['lint', 'secrets'] },
  { id: 'tests', label: 'Tests', sections: ['fixtures', 'test'] },
  { id: 'todos', label: 'Todos', sections: ['todos', 'checks'] },
  { id: 'processes', label: 'Processes', sections: ['procfile'] },
  { id: 'hooks', label: 'Hooks', sections: ['pre', 'post', 'ssh'] },
];

// The workspace-registration tab (project scope only) — edits the project
// record (dir/repos), not a .gavel.yaml section, so it is handled specially.
export const WORKSPACE_TAB = 'workspace';

// Human labels for the top-level sections (used as in-form section headings) and
// for a handful of nested keys the automatic humanizer would render awkwardly.
export const SECTION_TITLES: Record<string, string> = {
  verify: 'AI code review',
  status: 'Status summary',
  commit: 'Commit',
  lint: 'Lint rules',
  secrets: 'Secret scanning',
  fixtures: 'Fixture tests',
  test: 'Test outline',
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

const ACRONYMS = new Set(['ssh', 'ai', 'pr', 'cpu', 'id', 'url', 'ci', 'yaml', 'cel']);

// humanizeKey turns a config key (camelCase / snake / kebab) into a sentence-case
// label: "maxIterations" → "Max iterations", "prContentPrompt" → "PR content prompt".
export function humanizeKey(key: string): string {
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
export interface PromptDescriptor {
  id: string;
  title: string;
  description: string;
  configPath: string;
  default: string;
}

// promptIdOf returns a schema node's x-prompt-id extension, marking it an
// overridable prompt linked to a registry descriptor.
export function promptIdOf(node: JsonSchemaProperty | undefined): string | undefined {
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

// The dotted config path (e.g. "commit.gitignore") stamped onto each property
// node so extensions can recover a field's full path. The form preserves x-*
// extension keys on FieldControl.schema (the same way it carries x-prompt-id),
// so reading this off field.schema is more robust than node identity.
const SETTINGS_PATH_KEY = 'x-settings-path';

function stampPaths(node: JsonSchemaProperty | JsonSchemaObject, prefix: string): void {
  if (!node.properties) return;
  for (const [key, child] of Object.entries(node.properties)) {
    const path = prefix ? `${prefix}.${key}` : key;
    (child as Record<string, unknown>)[SETTINGS_PATH_KEY] = path;
    stampPaths(child, path);
  }
}

// pathOf returns the dotted config path stamped on a schema node by
// decorateSchema, or undefined for the (unstamped) root.
export function pathOf(node: JsonSchemaProperty | undefined): string | undefined {
  const path = node?.[SETTINGS_PATH_KEY as keyof JsonSchemaProperty];
  return typeof path === 'string' ? path : undefined;
}

export function decorateSchema(schema: JsonSchemaObject): JsonSchemaObject {
  const clone: JsonSchemaObject = structuredClone(schema);
  decorateNode(clone, true);
  stampPaths(clone, '');
  return clone;
}

// sectionSchema scopes the full schema to a tab's sections so each tab renders
// only its slice. The full value is still passed to the form (it edits keys in
// place), so switching tabs never drops another tab's settings.
export function sectionSchema(schema: JsonSchemaObject, sections: string[]): JsonSchemaObject {
  const properties: Record<string, JsonSchemaProperty> = {};
  const order: string[] = [];
  for (const key of sections) {
    const prop = schema.properties?.[key];
    if (prop) {
      properties[key] = prop;
      order.push(key);
    }
  }
  return { type: 'object', properties, 'x-order': order };
}
