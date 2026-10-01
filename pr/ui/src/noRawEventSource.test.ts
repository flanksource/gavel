import { existsSync, readdirSync, readFileSync, statSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';

// A raw `new EventSource(...)` bypasses eventHub's connection multiplexing
// and re-introduces the six-connections-per-host starvation this hub exists
// to fix (see eventHub.ts). Every stream must open through
// openEventStream()/useEventSourceFactory() instead, so this guards the whole
// source tree — eventHub.ts itself is the one legitimate caller, and test
// files stub the global constructor directly, which is not this pattern.
const srcDir = dirname(fileURLToPath(import.meta.url));
const rawEventSourcePattern = /\bnew\s+EventSource\s*\(/;

function collectSourceFiles(dir: string): string[] {
  const files: string[] = [];
  for (const entry of readdirSync(dir)) {
    const path = join(dir, entry);
    const stats = statSync(path);
    if (stats.isDirectory()) {
      files.push(...collectSourceFiles(path));
      continue;
    }
    if (!/\.(ts|tsx)$/.test(entry)) continue;
    if (/\.test\.(ts|tsx)$/.test(entry)) continue;
    if (path === join(srcDir, 'eventHub.ts')) continue;
    files.push(path);
  }
  return files;
}

describe('no raw EventSource construction outside eventHub', () => {
  it('finds every stream opened through openEventStream()/useEventSourceFactory(), never `new EventSource(...)` directly', () => {
    const offenders = collectSourceFiles(srcDir)
      .map(path => ({ path, content: readFileSync(path, 'utf8') }))
      .filter(({ content }) => rawEventSourcePattern.test(content))
      .map(({ path }) => path);

    expect(offenders).toEqual([]);
  });
});

// The source-tree scan above can't see what actually shipped: a build could
// inline, re-export, or re-wrap a raw EventSource in a way that reads clean
// per-file but still bypasses the hub in the bundle a browser loads. This
// guards the BUILT bundle (`pnpm run build` output) instead, so it needs
// `dist/` to exist — skipped with a named reason when it doesn't (e.g. a
// vitest run that hasn't built the UI yet; CI always builds both UIs before
// Go tests per project memory, and this build's own vitest run above already
// requires source-level compliance).
const distDir = join(srcDir, '..', 'dist');
const distBuilt = existsSync(distDir);

// Minification erases names and whitespace, so instead of matching source
// syntax we classify every `new EventSource(...)` call site in the bundle by
// its ARGUMENT shape. Exactly two shapes are legitimate:
//
//  1. A pass-through factory: a function whose entire job is `new
//     EventSource(<its own single parameter>)` — this is clicky-ui's default
//     `EventSourceFactoryContext` value ((e) => new EventSource(e)) and
//     testrunner/ui's `defaultCreateEventSource` (same shape, as a named
//     function after minification). Both stay legitimate only because every
//     *caller* that matters (pr/ui's index.tsx, ProjectActionRunDialog.tsx)
//     overrides them with `openEventStream` instead of ever invoking the
//     default — this test doesn't re-verify that call-site wiring, only that
//     the two shapes in the bundle are these known-safe forms and nothing
//     else new EventSource(...)-shaped snuck in.
//  2. eventHub's own singleton connection to `/api/events` (ensureRealOpen in
//     eventHub.ts): the call argument is a bare identifier, and that same
//     identifier is assigned the string literal "/api/events" elsewhere in
//     the file (the minified REAL_URL constant).
//
// Any other call — a literal URL, a member expression, a function whose body
// does more than pass its argument through — is an offender: something is
// constructing a raw EventSource outside the hub in what actually ships.
const identifier = String.raw`[A-Za-z_$][\w$]*`;
const passThroughArrow = new RegExp(String.raw`\(\s*(${identifier})\s*\)\s*=>\s*new EventSource\(\s*\1\s*\)`, 'g');
const passThroughFunction = new RegExp(
  String.raw`function\s+${identifier}\s*\(\s*(${identifier})\s*\)\s*\{\s*return new EventSource\(\s*\1\s*\)\s*;?\s*\}`,
  'g',
);
const eventSourceCall = /new EventSource\(\s*([^()]*?)\s*\)/g;
const stringLiteral = /^(["'`])(.*)\1$/;

function findAllowedSpans(content: string): Array<[number, number]> {
  const spans: Array<[number, number]> = [];
  for (const pattern of [passThroughArrow, passThroughFunction]) {
    pattern.lastIndex = 0;
    for (let m = pattern.exec(content); m; m = pattern.exec(content)) {
      spans.push([m.index, m.index + m[0].length]);
    }
  }
  return spans;
}

function isWithinAllowedSpan(index: number, spans: Array<[number, number]>): boolean {
  return spans.some(([start, end]) => index >= start && index < end);
}

function resolvesToHubUrl(content: string, arg: string): boolean {
  const literal = arg.match(stringLiteral);
  if (literal && literal[2] === '/api/events') return true; // inlined literal, still the hub's own URL
  if (!/^[A-Za-z_$][\w$]*$/.test(arg)) return false; // not a bare identifier — can't resolve
  const assignment = new RegExp(String.raw`\b${arg}\s*=\s*(["'\`])/api/events\1`);
  return assignment.test(content);
}

function collectDistFiles(): string[] {
  const files = [join(distDir, 'prui.js')].filter(existsSync);
  const chunksDir = join(distDir, 'chunks');
  if (existsSync(chunksDir)) {
    for (const entry of readdirSync(chunksDir)) {
      if (entry.endsWith('.js')) files.push(join(chunksDir, entry));
    }
  }
  return files;
}

describe('no raw EventSource construction in the built bundle', () => {
  it.skipIf(!distBuilt)(
    'classifies every `new EventSource(...)` call site as the pass-through default factory or eventHub\'s own /api/events connection — run `pnpm run build` first if this is skipped',
    () => {
      const offenders: string[] = [];
      const sites: string[] = [];

      for (const file of collectDistFiles()) {
        const content = readFileSync(file, 'utf8');
        const allowedSpans = findAllowedSpans(content);
        eventSourceCall.lastIndex = 0;
        for (let m = eventSourceCall.exec(content); m; m = eventSourceCall.exec(content)) {
          const arg = m[1];
          sites.push(`${file.slice(distDir.length + 1)} @${m.index}: new EventSource(${arg})`);
          if (isWithinAllowedSpan(m.index, allowedSpans)) continue; // pass-through factory
          if (resolvesToHubUrl(content, arg)) continue; // eventHub's own connection
          offenders.push(`${file}:${m.index}: new EventSource(${arg}) — not a recognized pass-through factory or hub connection`);
        }
      }

      // Every recognized call site, for the report: proves the matcher found
      // the known sites (clicky-ui's default factory, testrunner/ui's default
      // factory, and eventHub's own /api/events connection) rather than
      // silently matching nothing because the bundle shape moved.
      expect(sites.length).toBeGreaterThan(0);
      expect(offenders).toEqual([]);
    },
  );
});
