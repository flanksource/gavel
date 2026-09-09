import { describe, expect, it } from 'vitest';
import type { TodoItem, TodoTagDef } from '../../types';
import { buildTagIndex, tagCandidates, todoReservedLabels, todoTagTokens, todoVisibleLabels } from './tagResolve';
import { TAG_PALETTE, tagHash, tagKey, normalizeTag } from './tagPalette';

function todo(labels: string[]): TodoItem {
  return { ref: 'r', title: 't', status: 'pending', priority: 'medium', labels };
}

const defs: TodoTagDef[] = [
  { name: 'bug', color: 'red', icon: 'debug', iconify: 'ion:bug', scope: 'builtin', description: 'Broken' },
  { name: 'area', color: 'teal', scope: 'global' },
  { name: 'area/ui', color: 'rose', scope: 'workspace' },
];

describe('buildTagIndex', () => {
  const index = buildTagIndex(defs);

  it('resolves an exact name to its definition', () => {
    const tag = index.resolve('bug');
    expect(tag.color).toBe('red');
    expect(tag.icon).toBe('debug');
    expect(tag.iconify).toBe('ion:bug');
    expect(tag.description).toBe('Broken');
    expect(tag.defined).toBe(true);
  });

  it('falls back to the namespace key for an undefined value', () => {
    const tag = index.resolve('area/api');
    expect(tag.color).toBe('teal');
    expect(tag.key).toBe('area');
    expect(tag.value).toBe('api');
    expect(tag.defined).toBe(true);
  });

  it('prefers an exact match over the key match', () => {
    expect(index.resolve('area/ui').color).toBe('rose');
  });

  it('derives a stable palette colour for an undefined tag', () => {
    const tag = index.resolve('nobody-defined-this');
    expect(tag.defined).toBe(false);
    expect(tag.icon).toBeUndefined();
    expect(TAG_PALETTE).toContain(tag.color);
    expect(index.resolve('nobody-defined-this').color).toBe(tag.color);
  });

  it('normalizes case and surrounding space', () => {
    expect(index.resolve('  BUG  ').token).toBe('bug');
    expect(index.resolve('  BUG  ').color).toBe('red');
  });

  it('splits key and value for a colon-namespaced label', () => {
    const tag = index.resolve('source:todo');
    expect(tag.key).toBe('source');
    expect(tag.value).toBe('todo');
  });

  it('leaves a flat label with no key', () => {
    const tag = index.resolve('flaky');
    expect(tag.key).toBeUndefined();
    expect(tag.value).toBe('flaky');
  });
});

describe('label partitioning', () => {
  const item = todo(['bug', 'status:open', 'priority:high', 'session:abc', 'area/ui']);

  it('hides machine-managed lifecycle prefixes from the tag surfaces', () => {
    expect(todoVisibleLabels(item)).toEqual(['bug', 'area/ui']);
  });

  it('keeps the lifecycle labels available for write paths', () => {
    expect(todoReservedLabels(item)).toEqual(['status:open', 'priority:high', 'session:abc']);
  });

  // A write path sends visible + reserved back. If these two ever overlapped or
  // dropped a label, saving a tag edit would silently delete lifecycle state.
  it('partitions every label exactly once', () => {
    const all = [...todoVisibleLabels(item), ...todoReservedLabels(item)].sort();
    expect(all).toEqual([...(item.labels ?? [])].sort());
  });

  it('tolerates a todo with no labels', () => {
    expect(todoVisibleLabels(todo([]))).toEqual([]);
    expect(todoReservedLabels({ ref: 'r', title: 't', status: 'pending', priority: 'medium' })).toEqual([]);
  });
});

describe('todoTagTokens', () => {
  it('emits the full token and the bare key for a namespaced label', () => {
    expect(todoTagTokens(todo(['area/ui'])).sort()).toEqual(['area', 'area/ui']);
  });

  it('emits only the token for a flat label', () => {
    expect(todoTagTokens(todo(['flaky']))).toEqual(['flaky']);
  });

  it('excludes reserved lifecycle labels', () => {
    expect(todoTagTokens(todo(['status:open', 'bug']))).toEqual(['bug']);
  });

  it('de-duplicates a key shared by two values', () => {
    expect(todoTagTokens(todo(['area/ui', 'area/api'])).sort()).toEqual(['area', 'area/api', 'area/ui']);
  });
});

describe('tagPalette', () => {
  it('splits a key on the first : or /', () => {
    expect(tagKey('source:todo')).toBe('source');
    expect(tagKey('area/ui')).toBe('area');
    expect(tagKey('a:b:c')).toBe('a');
    expect(tagKey('flaky')).toBe('');
    expect(tagKey(':leading')).toBe('');
  });

  it('normalizes to lowercase and trims', () => {
    expect(normalizeTag('  BUG ')).toBe('bug');
  });

  // The Go resolver hashes the same way over the same palette order, so a tag
  // with no definition must land on the same hue in both. A drift here shows up
  // as the terminal and the dashboard disagreeing.
  it('hashes a namespace as a family', () => {
    expect(tagHash('area/ui')).toBe(tagHash('area/api'));
    expect(tagHash('area/ui')).toBe(tagHash('area'));
  });

  it('is stable and stays inside the palette', () => {
    for (const label of ['flaky', 'infra', 'ünïcode', 'source:todo']) {
      expect(tagHash(label)).toBe(tagHash(label));
      expect(TAG_PALETTE).toContain(tagHash(label));
    }
  });
});

describe('tagCandidates', () => {
  const index = buildTagIndex([
    { name: 'bug', color: 'red', scope: 'builtin' },
    { name: 'flaky', color: 'amber', scope: 'workspace' },
  ]);

  // A label typed free-hand is stored on the todo and defined nowhere. If the
  // picker only offered definitions it would be invisible, and the next person
  // would type a near-miss instead of reusing it.
  it('offers labels that are in use but undefined', () => {
    const names = tagCandidates(index, { 'legacy-thing': 4 }).map(def => def.name);
    expect(names).toContain('legacy-thing');
  });

  it('gives an undefined label the colour it already renders with', () => {
    const candidate = tagCandidates(index, { 'legacy-thing': 4 })
      .find(def => def.name === 'legacy-thing');
    expect(candidate?.color).toBe(tagHash('legacy-thing'));
    expect(candidate?.scope).toBe('derived');
  });

  it('does not duplicate a label that is both defined and in use', () => {
    const names = tagCandidates(index, { flaky: 9, bug: 2 }).map(def => def.name);
    expect(names.filter(name => name === 'flaky')).toHaveLength(1);
    expect(names.filter(name => name === 'bug')).toHaveLength(1);
  });

  it('never offers machine-managed lifecycle labels', () => {
    const names = tagCandidates(index, { 'status:open': 3, 'priority:high': 1, 'session:abc': 1 })
      .map(def => def.name);
    expect(names).not.toContain('status:open');
    expect(names).not.toContain('priority:high');
    expect(names).not.toContain('session:abc');
  });

  it('is just the definitions when nothing is in use', () => {
    expect(tagCandidates(index).map(def => def.name)).toEqual(['bug', 'flaky']);
  });
});
