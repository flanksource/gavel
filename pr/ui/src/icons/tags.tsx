// Tag icon set — Phosphor (`ph:*-thin`) glyphs bundled offline via
// @iconify/react, keyed by the clicky icon-registry name the server stores.
//
// Two-stage resolution, deliberately: a curated key renders from a bundled
// glyph and never touches the network, while any other icon falls back to the
// Iconify name the server resolved (see icons/iconifyFallback.tsx). That keeps
// the built-in tags fully offline — which matters because this dashboard is
// routinely used without a network — while still letting someone name an
// arbitrary Iconify glyph in the tag editor.
import type { ComponentType, ReactNode } from 'react';
import type { IconMenuOption } from '@flanksource/clicky-ui/components';
import { UiProhibit } from '@flanksource/clicky-ui/icons';
import { Icon, type IconProps as IconifyProps } from '@iconify/react/offline';

import bugThin from '@iconify-icons/ph/bug-thin';
import lockKeyThin from '@iconify-icons/ph/lock-key-thin';
import booksThin from '@iconify-icons/ph/books-thin';
import lightningThin from '@iconify-icons/ph/lightning-thin';
import flaskThin from '@iconify-icons/ph/flask-thin';
import hammerThin from '@iconify-icons/ph/hammer-thin';
import paletteThin from '@iconify-icons/ph/palette-thin';
import globeThin from '@iconify-icons/ph/globe-thin';
import gitBranchThin from '@iconify-icons/ph/git-branch-thin';
import warningThin from '@iconify-icons/ph/warning-thin';
import tagSimpleThin from '@iconify-icons/ph/tag-simple-thin';
import rocketThin from '@iconify-icons/ph/rocket-thin';
import databaseThin from '@iconify-icons/ph/database-thin';
import shieldCheckThin from '@iconify-icons/ph/shield-check-thin';
import gearThin from '@iconify-icons/ph/gear-thin';
import flagThin from '@iconify-icons/ph/flag-thin';
import questionThin from '@iconify-icons/ph/question-thin';
import broomThin from '@iconify-icons/ph/broom-thin';
import sparkleThin from '@iconify-icons/ph/sparkle-thin';

type Glyph = IconifyProps['icon'];
type GlyphComponent = ComponentType<Omit<IconifyProps, 'icon'>>;

// Keys are clicky icon-registry names — the same values `gavel todos labels set
// --icon` accepts and the server stores — so one name drives both the terminal
// glyph and this one.
const TAG_GLYPHS = {
  debug: bugThin,
  lock: lockKeyThin,
  docs: booksThin,
  performance: lightningThin,
  test: flaskThin,
  refactor: hammerThin,
  style: paletteThin,
  http: globeThin,
  ci: gitBranchThin,
  warning: warningThin,
  rocket: rocketThin,
  database: databaseThin,
  security: shieldCheckThin,
  config: gearThin,
  flag: flagThin,
  question: questionThin,
  cleanup: broomThin,
  feature: sparkleThin,
  tag: tagSimpleThin,
} satisfies Record<string, Glyph>;

export type TagGlyphKey = keyof typeof TAG_GLYPHS;

/** The registry keys with a bundled glyph — the editor's offline-safe picks. */
export const TAG_GLYPH_KEYS = Object.keys(TAG_GLYPHS) as TagGlyphKey[];

function offlineGlyph(icon: Glyph): GlyphComponent {
  return function OfflineTagGlyph(props) {
    return <Icon icon={icon} {...props} />;
  };
}

export const TAG_GLYPH_OPTIONS: IconMenuOption<string>[] = [
  { value: '', label: 'No icon', icon: UiProhibit },
  ...TAG_GLYPH_KEYS.map(key => ({ value: key, label: key, icon: offlineGlyph(TAG_GLYPHS[key]) })),
];

export function hasTagGlyph(key: string | undefined): key is TagGlyphKey {
  return key != null && key in TAG_GLYPHS;
}

export function boundTagGlyph(iconKey: string, iconify?: string): GlyphComponent {
  return function BoundTagGlyph({ className }) {
    return <TagIcon iconKey={iconKey} iconify={iconify} className={className} />;
  };
}

/**
 * TagIcon renders a tag's glyph. `iconKey` is the stored clicky registry name
 * and wins when it is one we bundle; `iconify` is the server-resolved Iconify
 * name used otherwise, which the registered fallback provider resolves at
 * runtime. With neither, nothing renders — a chip is never padded with a
 * placeholder box.
 *
 * The glyph takes its colour from the chip via currentColor, so it can never
 * disagree with the tint.
 */
export function TagIcon({ iconKey, iconify, className }: {
  iconKey?: string;
  iconify?: string;
  className?: string;
}) {
  if (hasTagGlyph(iconKey)) {
    return <Icon icon={TAG_GLYPHS[iconKey]} className={className} aria-hidden="true" />;
  }
  if (iconify) {
    return <RuntimeTagIcon name={iconify} className={className} />;
  }
  return null;
}

// Split out so the offline path above never pulls in the runtime resolver.
function RuntimeTagIcon({ name, className }: { name: string; className?: string }) {
  const Fallback = getFallbackIcon();
  if (!Fallback) return null;
  return <Fallback name={name} className={className} />;
}

// Resolved lazily so a test that never registers a provider simply renders
// nothing rather than importing the network-capable Iconify entry.
let fallbackIcon: ((props: { name: string; className?: string }) => ReactNode) | null = null;

export function setTagFallbackIcon(component: typeof fallbackIcon): void {
  fallbackIcon = component;
}

function getFallbackIcon() {
  return fallbackIcon;
}

export const TagGlyphDefault = tagSimpleThin;
