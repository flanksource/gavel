import { existsSync, readdirSync, readFileSync, statSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';

// A raw `new EventSource(...)` bypasses the event hub's connection
// multiplexing and re-introduces the six-connections-per-host starvation the
// hub exists to fix (see eventHub.ts). Every stream must open through
// openEventStream()/useEventSourceFactory() instead, so this guards the whole
// source tree — the hub's one real connection lives in clicky-ui's
// createEventHub, and test files stub the global constructor directly, which
// is not this pattern.
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
    files.push(path);
  }
  return files;
}

describe('no raw EventSource construction in gavel source', () => {
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
const hubUrl = '/api/events';

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
//  2. The hub's own connection to `/api/events` (clicky-ui's createEventHub,
//     `const { url = DEFAULT_URL } = options; new EventSource(url)`): the call
//     argument is a bare identifier assigned either the "/api/events" literal
//     or an identifier holding it — minified, `const b="/api/events"` and the
//     destructuring default `url:n=b` feeding `new EventSource(n)`.
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
const bareIdentifier = new RegExp(String.raw`^${identifier}$`);

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

// escapeRegExp makes text match itself literally inside a RegExp source.
// Minified identifiers routinely carry `$`, a regex metacharacter.
function escapeRegExp(text: string): string {
  return text.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

// assigns reports whether `name = <value>` appears in content, where value is a
// regex source; the lookbehind keeps `xname=` and `obj.name=` from matching.
function assigns(content: string, name: string, value: string): boolean {
  return new RegExp(String.raw`(?<![\w$.])${escapeRegExp(name)}\s*=\s*${value}`).test(content);
}

function resolvesToHubUrl(content: string, arg: string): boolean {
  const literal = arg.match(stringLiteral);
  if (literal) return literal[2] === hubUrl; // inlined literal, still the hub's own URL
  if (!bareIdentifier.test(arg)) return false; // not a bare identifier — can't resolve
  const hubLiteral = ['"', "'", '`'].map(quote => `${quote}${hubUrl}${quote}`).join('|').replace(/^(.*)$/, '(?:$1)');
  if (assigns(content, arg, hubLiteral)) return true;
  const hubConstants = [...content.matchAll(new RegExp(String.raw`(?<![\w$.])(${identifier})\s*=\s*${hubLiteral}`, 'g'))].map(m => m[1]);
  return hubConstants.some(constant => assigns(content, arg, String.raw`${escapeRegExp(constant)}(?![\w$(.])`));
}

// classifyBundle lists every `new EventSource(...)` call site in content and
// the ones that are neither a pass-through factory nor the hub's connection.
function classifyBundle(content: string): { sites: string[]; offenders: string[] } {
  const sites: string[] = [];
  const offenders: string[] = [];
  const allowedSpans = findAllowedSpans(content);
  eventSourceCall.lastIndex = 0;
  for (let m = eventSourceCall.exec(content); m; m = eventSourceCall.exec(content)) {
    const site = `@${m.index}: new EventSource(${m[1]})`;
    sites.push(site);
    if (isWithinAllowedSpan(m.index, allowedSpans)) continue; // pass-through factory
    if (resolvesToHubUrl(content, m[1])) continue; // the hub's own connection
    offenders.push(`${site} — not a recognized pass-through factory or hub connection`);
  }
  return { sites, offenders };
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

describe('bundle EventSource classifier', () => {
  it.each([
    ['the minified clicky-ui hub connection', 'const b="/api/events",F=1e3;function H(o={}){const{url:n=b}=o;const e=new EventSource(n);return e}'],
    ['a hub connection with the URL inlined', 'function H(){const t="/api/events";return new EventSource(t)}'],
    ['the pass-through default factory', 'const f=(e)=>new EventSource(e);'],
    ['a hub connection through `$`-bearing minified names', 'const $b="/api/events";function H(o={}){const{url:n$=$b}=o;return new EventSource(n$)}'],
  ])('accepts %s', (_name, content) => {
    expect(classifyBundle(content)).toEqual({ sites: [expect.any(String)], offenders: [] });
  });

  it.each([
    ['a raw stream URL literal', 'const b="/api/events";function S(){return new EventSource("/api/prs/stream")}'],
    ['an identifier holding another URL', 'const b="/api/events",u="/api/prs/stream";function S(){return new EventSource(u)}'],
    ['a member expression', 'const b="/api/events";function S(o){return new EventSource(o.url)}'],  ])('flags %s', (_name, content) => {
    expect(classifyBundle(content).offenders).toHaveLength(1);
  });
});

describe('no raw EventSource construction in the built bundle', () => {
  it.skipIf(!distBuilt)(
    'classifies every `new EventSource(...)` call site as the pass-through default factory or the hub\'s own /api/events connection — run `pnpm run build` first if this is skipped',
    () => {
      const sites: string[] = [];
      const offenders: string[] = [];
      for (const file of collectDistFiles()) {
        const name = file.slice(distDir.length + 1);
        const result = classifyBundle(readFileSync(file, 'utf8'));
        sites.push(...result.sites.map(site => `${name} ${site}`));
        offenders.push(...result.offenders.map(offender => `${name} ${offender}`));
      }

      // Every recognized call site, for the report: proves the matcher found
      // the known sites (clicky-ui's default factory, testrunner/ui's default
      // factory, and the hub's own /api/events connection) rather than
      // silently matching nothing because the bundle shape moved.
      expect(sites.length).toBeGreaterThan(0);
      expect(offenders).toEqual([]);
    },
  );
});
