import { afterEach, describe, expect, it, vi } from 'vitest';
import { fetchLatestSeq, traceProfileId } from './traces';

describe('traceProfileId', () => {
  it.each([
    ['process_samples', 'profile-traces-process-samples'],
    ['process_tree', 'profile-traces-process-tree'],
    ['http', 'profile-traces-http'],
  ])('names the %s result profile as the query server surfaces it', (kind, id) => {
    expect(traceProfileId(kind)).toBe(id);
  });
});

describe('fetchLatestSeq', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('reads the newest seq of the stream from the kind result profile', async () => {
    const fetchMock = vi.fn(async () => new Response(JSON.stringify([{ seq: 42 }])));
    vi.stubGlobal('fetch', fetchMock);
    await expect(fetchLatestSeq('/q', { kind: 'process_samples', stream: 's1' })).resolves.toBe(42);
    const url = new URL(String((fetchMock.mock.calls[0] as unknown[])[0]), 'http://host');
    expect([url.pathname, url.searchParams.get('stream'), url.searchParams.get('sort')]).toEqual([
      '/q/api/v1/profile/profile-traces-process-samples',
      's1',
      'seq',
    ]);
  });
});
