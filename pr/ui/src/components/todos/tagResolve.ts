import type { TodoItem, TodoTagDef } from '../../types';
import { normalizeTag, tagHash, tagKey, type TagColor } from './tagPalette';

// Labels prefixed with these are machine-managed lifecycle state, not user
// tags: they are hidden from every tag surface. This is the single definition —
// the detail pane, the list rows and the filter facet must agree on which
// labels are user-facing, or a tag can be filtered on but never seen.
const RESERVED_PREFIXES = ['status:', 'priority:', 'session:'] as const;

function isReserved(label: string): boolean {
  return RESERVED_PREFIXES.some(prefix => label.startsWith(prefix));
}

/** A label resolved to its presentation. */
export interface ResolvedTag {
  /** The raw label as stored on the todo ("area/ui"). */
  token: string;
  /** The key half of a namespaced label, or undefined when flat. */
  key?: string;
  /** The value half, or the whole token when flat. */
  value: string;
  color: TagColor | string;
  /** The clicky icon-registry key ("debug"); resolves to a bundled glyph. */
  icon?: string;
  /** The server-resolved Iconify name ("ion:bug"); the runtime fallback path. */
  iconify?: string;
  description?: string;
  /** False when nothing defined this label and its colour came from the hash. */
  defined: boolean;
}

export interface TagIndex {
  resolve(label: string): ResolvedTag;
  /** Every stored and built-in definition, for the facet and the editor. */
  defs: TodoTagDef[];
  loading: boolean;
}

/**
 * buildTagIndex mirrors the Go resolver's chain: an exact name match, then the
 * label's namespace key, then the deterministic hash. The server already merged
 * workspace over global over built-in, so `defs` is a flat effective set.
 */
export function buildTagIndex(defs: TodoTagDef[], loading = false): TagIndex {
  const byName = new Map<string, TodoTagDef>();
  for (const def of defs) {
    byName.set(normalizeTag(def.name), def);
  }

  const cache = new Map<string, ResolvedTag>();

  const resolve = (label: string): ResolvedTag => {
    const token = normalizeTag(label);
    const cached = cache.get(token);
    if (cached) return cached;

    const key = tagKey(token);
    const value = key ? token.slice(key.length + 1) : token;

    const exact = byName.get(token);
    const viaKey = exact ? undefined : (key ? byName.get(key) : undefined);
    const def = exact ?? viaKey;

    const resolved: ResolvedTag = def
      ? {
        token,
        key: key || undefined,
        value,
        color: def.color || tagHash(token),
        icon: def.icon || undefined,
        iconify: def.iconify || undefined,
        description: def.description,
        defined: true,
      }
      : { token, key: key || undefined, value, color: tagHash(token), defined: false };

    cache.set(token, resolved);
    return resolved;
  };

  return { resolve, defs, loading };
}

/** An index with no definitions — the pre-load state and the test default. */
export function emptyTagIndex(loading = false): TagIndex {
  return buildTagIndex([], loading);
}

/**
 * todoVisibleLabels are the user-facing tags on a todo: everything except the
 * machine-managed lifecycle prefixes.
 */
export function todoVisibleLabels(todo: TodoItem): string[] {
  return (todo.labels ?? [])
    .map(label => label.trim())
    .filter(label => label && !isReserved(label));
}

/**
 * todoReservedLabels is the complement of todoVisibleLabels.
 *
 * Any write path MUST send these back alongside the edited visible labels — the
 * API replaces the whole label set, so omitting them silently deletes the
 * todo's lifecycle state.
 */
export function todoReservedLabels(todo: TodoItem): string[] {
  return (todo.labels ?? [])
    .map(label => label.trim())
    .filter(label => label && isReserved(label));
}

/**
 * tagCandidates is everything a tag picker can offer: the defined taxonomy plus
 * the labels this workspace is already using that nothing has defined yet.
 *
 * The second half is what stops the vocabulary from splintering. A label typed
 * free-hand is stored on the todo but defined nowhere, so offering only
 * definitions would hide it from every picker — and the next person, not seeing
 * it, types their own near-miss. `flaky`, `flakey` and `flake` become three
 * colours nobody chose. An undefined label is offered with the colour it
 * already renders with, so picking it is reuse rather than a new tag.
 */
export function tagCandidates(index: TagIndex, counts: Record<string, number> = {}): TodoTagDef[] {
  const defined = new Set(index.defs.map(def => normalizeTag(def.name)));
  const used = Object.keys(counts)
    .map(normalizeTag)
    .filter(name => name && !defined.has(name) && !isReserved(name))
    .map((name): TodoTagDef => {
      const resolved = index.resolve(name);
      return { name, color: String(resolved.color), scope: 'derived' };
    });
  return [...index.defs, ...used];
}

/**
 * todoTagTokens are the values a todo matches on in the tag facet: each full
 * label plus, for namespaced labels, the bare key. That makes both "only
 * area:ui" and "exclude every area:*" expressible from one facet, which matters
 * as soon as a workspace has more than a handful of namespaced values.
 */
export function todoTagTokens(todo: TodoItem): string[] {
  const tokens = new Set<string>();
  for (const label of todoVisibleLabels(todo)) {
    const token = normalizeTag(label);
    tokens.add(token);
    const key = tagKey(token);
    if (key) tokens.add(key);
  }
  return [...tokens];
}
