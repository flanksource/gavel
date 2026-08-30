import { fnv1a32 } from '../../utils';

// TAG_PALETTE mirrors todos/labels/palette.go Palette() — same hues, same order.
// The order is load-bearing: tagHash indexes into it, and the Go side hashes the
// same way, so a tag with no stored definition renders the same colour in the
// dashboard and in the terminal. Changing the order here without changing it
// there silently desynchronises the two.
export const TAG_PALETTE = [
  'slate', 'red', 'orange', 'amber', 'yellow', 'lime',
  'green', 'emerald', 'teal', 'cyan', 'sky', 'blue',
  'indigo', 'violet', 'purple', 'fuchsia', 'pink', 'rose',
] as const;

export type TagColor = typeof TAG_PALETTE[number];

// Every class is spelled out as a literal rather than interpolated: Tailwind v4
// scans source for class names, so `bg-${hue}-100` would never be emitted and a
// DB-authored colour would render unstyled.
//
// lime/amber/yellow use the 800 text stop — their 700 is unreadable on their own
// 100 tint. This matches Color.TextClass() in Go and AVATAR_PALETTE in utils.ts.
const TAG_CLASSES: Record<TagColor, string> = {
  slate: 'bg-slate-100 text-slate-700 dark:bg-slate-500/15 dark:text-slate-300',
  red: 'bg-red-100 text-red-700 dark:bg-red-500/15 dark:text-red-300',
  orange: 'bg-orange-100 text-orange-700 dark:bg-orange-500/15 dark:text-orange-300',
  amber: 'bg-amber-100 text-amber-800 dark:bg-amber-500/15 dark:text-amber-300',
  yellow: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-500/15 dark:text-yellow-300',
  lime: 'bg-lime-100 text-lime-800 dark:bg-lime-500/15 dark:text-lime-300',
  green: 'bg-green-100 text-green-700 dark:bg-green-500/15 dark:text-green-300',
  emerald: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300',
  teal: 'bg-teal-100 text-teal-700 dark:bg-teal-500/15 dark:text-teal-300',
  cyan: 'bg-cyan-100 text-cyan-700 dark:bg-cyan-500/15 dark:text-cyan-300',
  sky: 'bg-sky-100 text-sky-700 dark:bg-sky-500/15 dark:text-sky-300',
  blue: 'bg-blue-100 text-blue-700 dark:bg-blue-500/15 dark:text-blue-300',
  indigo: 'bg-indigo-100 text-indigo-700 dark:bg-indigo-500/15 dark:text-indigo-300',
  violet: 'bg-violet-100 text-violet-700 dark:bg-violet-500/15 dark:text-violet-300',
  purple: 'bg-purple-100 text-purple-700 dark:bg-purple-500/15 dark:text-purple-300',
  fuchsia: 'bg-fuchsia-100 text-fuchsia-700 dark:bg-fuchsia-500/15 dark:text-fuchsia-300',
  pink: 'bg-pink-100 text-pink-700 dark:bg-pink-500/15 dark:text-pink-300',
  rose: 'bg-rose-100 text-rose-700 dark:bg-rose-500/15 dark:text-rose-300',
};

// Text-only classes, for surfaces that show the glyph without the chip tint.
const TAG_TEXT_CLASSES: Record<TagColor, string> = {
  slate: 'text-slate-600 dark:text-slate-400',
  red: 'text-red-600 dark:text-red-400',
  orange: 'text-orange-600 dark:text-orange-400',
  amber: 'text-amber-700 dark:text-amber-400',
  yellow: 'text-yellow-700 dark:text-yellow-400',
  lime: 'text-lime-700 dark:text-lime-400',
  green: 'text-green-600 dark:text-green-400',
  emerald: 'text-emerald-600 dark:text-emerald-400',
  teal: 'text-teal-600 dark:text-teal-400',
  cyan: 'text-cyan-600 dark:text-cyan-400',
  sky: 'text-sky-600 dark:text-sky-400',
  blue: 'text-blue-600 dark:text-blue-400',
  indigo: 'text-indigo-600 dark:text-indigo-400',
  violet: 'text-violet-600 dark:text-violet-400',
  purple: 'text-purple-600 dark:text-purple-400',
  fuchsia: 'text-fuchsia-600 dark:text-fuchsia-400',
  pink: 'text-pink-600 dark:text-pink-400',
  rose: 'text-rose-600 dark:text-rose-400',
};

export function isTagColor(value: string): value is TagColor {
  return (TAG_PALETTE as readonly string[]).includes(value);
}

/** The chip's tint + text classes. An unknown hue falls back to slate. */
export function tagClasses(color: string): string {
  return TAG_CLASSES[isTagColor(color) ? color : 'slate'];
}

/** Text-only classes, for glyph-only surfaces. */
export function tagTextClasses(color: string): string {
  return TAG_TEXT_CLASSES[isTagColor(color) ? color : 'slate'];
}

/** Normalize mirrors labels.Normalize in Go: lowercase and trim. */
export function normalizeTag(value: string): string {
  return value.trim().toLowerCase();
}

/**
 * tagKey splits a namespaced label into its key — "source:todo" -> "source",
 * "area/ui" -> "area", "bug" -> "". Mirrors labels.Key in Go so both ends group
 * a namespace the same way.
 */
export function tagKey(label: string): string {
  const normalized = normalizeTag(label);
  const index = normalized.search(/[:/]/);
  return index > 0 ? normalized.slice(0, index) : '';
}

/**
 * tagHash is the deterministic colour for a label with no stored definition, so
 * an unconfigured backlog is still visually separable. It hashes the label's key
 * when it has one, so every member of a namespace shares a hue — mirroring
 * labels.Hash in Go over the same palette order and the same FNV-1a.
 */
export function tagHash(label: string): TagColor {
  const seed = tagKey(label) || normalizeTag(label);
  return TAG_PALETTE[fnv1a32(seed) % TAG_PALETTE.length];
}
