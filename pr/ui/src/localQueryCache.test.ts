import { beforeEach, describe, expect, it } from 'vitest';
import { readLocalCache, writeLocalCache } from './localQueryCache';

const key = 'gavel.pr-ui.cache.test.v1';
const parseNames = (value: unknown): string[] => {
  if (!Array.isArray(value) || !value.every(v => typeof v === 'string')) throw new Error('invalid names');
  return value;
};

beforeEach(() => localStorage.clear());

describe('local query cache', () => {
  it('round-trips a value through the parser', () => {
    writeLocalCache(key, ['alpha', 'beta']);
    expect(readLocalCache(key, parseNames)).toEqual(['alpha', 'beta']);
  });

  it('has nothing to offer before the first write', () => {
    expect(readLocalCache(key, parseNames)).toBeUndefined();
  });

  // A cached value only seeds an optimistic render until the network answers,
  // so a stale shape from an older build is dropped rather than rendered.
  it.each([
    ['unparseable JSON', '{not json'],
    ['a shape the parser rejects', JSON.stringify([{ name: 'alpha' }])],
  ])('drops %s', (_, raw) => {
    localStorage.setItem(key, raw);
    expect(readLocalCache(key, parseNames)).toBeUndefined();
  });
});
