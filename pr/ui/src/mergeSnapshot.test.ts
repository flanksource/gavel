import { describe, expect, it } from 'vitest';
import { mergeSnapshot } from './useAppQueries';
import type { Snapshot } from './types';

function snapshot(overrides: Partial<Snapshot> = {}): Snapshot {
  return {
    prs: [{ number: 1, title: 'first', repo: 'flanksource/gavel' }],
    fetchedAt: '2026-06-23T10:00:00Z',
    nextFetchIn: 60,
    incremental: false,
    paused: false,
    config: { repos: ['flanksource/gavel'], org: 'flanksource', all: false },
    ...overrides,
  } as Snapshot;
}

// cached is what the query cache actually holds in steady state: a snapshot
// that has already been through one merge, so it carries the keys the merge
// adds (unread, viewer, rateLimit, …). Comparing against a raw fixture instead
// would test a shape that never reaches mergeSnapshot in production.
function cached(overrides: Partial<Snapshot> = {}): Snapshot {
  return mergeSnapshot(snapshot(overrides), snapshot(overrides));
}

describe('mergeSnapshot', () => {
  // The stream's liveness cadence can resend state we already hold. Because a
  // resend repeats fetchedAt, the newer-than check cannot reject it, so the
  // identity guard is the only thing standing between an inert frame and a
  // full-app re-render.
  it('returns the cached object itself when a resent frame changes nothing', () => {
    const current = cached();
    const incoming = snapshot();

    expect(mergeSnapshot(current, incoming)).toBe(current);
  });

  it('is stable across repeated resends, not just the first', () => {
    const current = cached();

    const once = mergeSnapshot(current, snapshot());
    const twice = mergeSnapshot(once, snapshot());
    const thrice = mergeSnapshot(twice, snapshot());

    expect(thrice).toBe(current);
  });

  it('returns a new object when the PR list actually changes', () => {
    const current = cached();
    const incoming = snapshot({
      prs: [
        { number: 1, title: 'first', repo: 'flanksource/gavel' },
        { number: 2, title: 'second', repo: 'flanksource/gavel' },
      ],
    } as Partial<Snapshot>);

    const merged = mergeSnapshot(current, incoming);
    expect(merged).not.toBe(current);
    expect(merged.prs).toHaveLength(2);
  });

  it('returns a new object when a field outside the PR list changes', () => {
    const current = cached();
    const incoming = snapshot({ paused: true });

    const merged = mergeSnapshot(current, incoming);
    expect(merged).not.toBe(current);
    expect(merged.paused).toBe(true);
  });

  it('keeps rejecting frames older than the cached snapshot', () => {
    const current = cached({ fetchedAt: '2026-06-23T10:05:00Z' });
    const incoming = snapshot({ fetchedAt: '2026-06-23T10:00:00Z', paused: true });

    expect(mergeSnapshot(current, incoming)).toBe(current);
  });
});
