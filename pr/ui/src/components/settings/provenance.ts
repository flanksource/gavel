// Provenance derives per-field layer badges for the settings form from the
// backend config trace (/api/settings/gavel/trace). The trace lists each
// contributing .gavel.yaml's parsed config in merge order (user-home, then
// git-root, then target-directory); a later source overrides an earlier one, so
// the effective owner of a field is the LAST source that sets a value at its path.

// The two layers the settings UI exposes. Backend origins collapse onto these:
// user-home → 'user'; git-root / target-directory → 'project'.
export type SettingsLayer = 'project' | 'user';

export interface TraceSource {
  origin: string;
  path: string;
  config: Record<string, unknown>;
}

export interface GavelTrace {
  sources?: TraceSource[];
  merged: Record<string, unknown>;
}

export function layerOfOrigin(origin: string): SettingsLayer {
  return origin === 'user-home' ? 'user' : 'project';
}

// isSet decides whether a layer actually contributes a value at a path. Empty
// strings / arrays / objects are treated as unset (an omitempty-dropped or
// zero-valued field), while a meaningful false / 0 counts as set.
function isSet(value: unknown): boolean {
  if (value === undefined || value === null) return false;
  if (typeof value === 'string') return value !== '';
  if (Array.isArray(value)) return value.length > 0;
  if (typeof value === 'object') return Object.keys(value as object).length > 0;
  return true;
}

function getByPath(root: unknown, path: string[]): unknown {
  let cur: unknown = root;
  for (const key of path) {
    if (cur == null || typeof cur !== 'object') return undefined;
    cur = (cur as Record<string, unknown>)[key];
  }
  return cur;
}

// provenanceForPath returns which layer owns the effective value at a dotted
// path, or undefined when no layer sets it (the built-in default is in force).
export function provenanceForPath(trace: GavelTrace | null, path: string): SettingsLayer | undefined {
  if (!trace?.sources?.length) return undefined;
  const segments = path.split('.').filter(Boolean);
  if (segments.length === 0) return undefined;
  let owner: SettingsLayer | undefined;
  for (const source of trace.sources) {
    if (isSet(getByPath(source.config, segments))) owner = layerOfOrigin(source.origin);
  }
  return owner;
}
