// Icon set for the settings dialog — one glyph per top-level .gavel.yaml section,
// shared by the tab strip and the in-form section headers. The AI-review glyph
// comes from clicky-ui's icon set (`icons/ai`); the rest are Phosphor thin glyphs
// bundled offline via @iconify/react so nothing is fetched at runtime.
import { Icon } from '@iconify/react/offline';
import type { IconifyIcon } from '@iconify/react/offline';
import type { ComponentType } from 'react';
import { UiSparkles } from '@flanksource/clicky-ui/icons';
import gitCommitThin from '@iconify-icons/ph/git-commit-thin';
import warningThin from '@iconify-icons/ph/warning-thin';
import shieldCheckThin from '@iconify-icons/ph/shield-check-thin';
import flaskThin from '@iconify-icons/ph/flask-thin';
import listChecksThin from '@iconify-icons/ph/list-checks-thin';
import checkSquareThin from '@iconify-icons/ph/check-square-thin';
import stackThin from '@iconify-icons/ph/stack-thin';
import playThin from '@iconify-icons/ph/play-thin';
import lightningThin from '@iconify-icons/ph/lightning-thin';
import terminalWindowThin from '@iconify-icons/ph/terminal-window-thin';
import chartLineThin from '@iconify-icons/ph/chart-line-thin';
import treeStructureThin from '@iconify-icons/ph/tree-structure-thin';

// SettingsIcon is a plain component so it can be handed straight to clicky-ui's
// Tabs (`icon`) and rendered as a section-header label glyph (`labelIcon`).
export type SettingsIcon = ComponentType<{ className?: string }>;

function ph(glyph: IconifyIcon): SettingsIcon {
  return function PhGlyph({ className }: { className?: string }) {
    return <Icon icon={glyph} className={className} />;
  };
}

// sectionIcon is keyed by the .gavel.yaml top-level section name.
export const sectionIcon: Record<string, SettingsIcon> = {
  verify: UiSparkles,
  status: ph(chartLineThin),
  test: ph(treeStructureThin),
  commit: ph(gitCommitThin),
  lint: ph(warningThin),
  secrets: ph(shieldCheckThin),
  fixtures: ph(flaskThin),
  todos: ph(listChecksThin),
  checks: ph(checkSquareThin),
  procfile: ph(stackThin),
  pre: ph(playThin),
  post: ph(lightningThin),
  ssh: ph(terminalWindowThin),
};
