import { describe, expect, it } from 'vitest';
import { ProcessNotFoundError } from '@flanksource/clicky-ui/data';
import { fetchProcessMetricsFromSamples } from './processMetrics';
import type { ProcessSample, ProcessTraceNode } from './types';

const node = (pid: number, over: Partial<ProcessTraceNode> = {}): ProcessTraceNode => ({
  key: `${pid}@1`, pid, ppid: 1, name: 'pkg.test', command: 'pkg.test -test.v', status: 'active',
  startedAt: '2026-01-01T00:00:00Z', lastSeenAt: '2026-01-01T00:00:09Z',
  cpuPercent: null, rssBytes: 0, peakCpuPercent: null, peakRssBytes: 0, ...over,
});
const sample = (pid: number, tick: number, over: Partial<ProcessSample> = {}): ProcessSample => ({
  sampledAt: `2026-01-01T00:00:0${tick}Z`, tick, key: `${pid}@1`, pid, memoryPercent: 1, rssBytes: 1000 * tick, cpuPercent: 10 * tick, ...over,
});

describe('fetchProcessMetricsFromSamples', () => {
  it('reads the latest sample of the pid, not of another process', async () => {
    const fetchMetrics = fetchProcessMetricsFromSamples({
      samples: [sample(7, 1), sample(7, 3, { openFiles: 12 }), sample(7, 2), sample(8, 9)],
      nodes: [node(7), node(8)],
    });

    await expect(fetchMetrics(7)).resolves.toEqual({ pid: 7, rss: 3000, cpu_percent: 30, open_files: 12 });
  });

  it('leaves out a CPU% the sample does not know yet, and a facet count the session did not arm', async () => {
    const fetchMetrics = fetchProcessMetricsFromSamples({ samples: [sample(7, 1, { cpuPercent: null })], nodes: [node(7)] });

    const metrics = await fetchMetrics(7);

    expect(metrics).toEqual({ pid: 7, rss: 1000 });
    expect('cpu_percent' in metrics).toBe(false);
  });

  it('rejects with ProcessNotFoundError, which stops the poll, once the process has exited', async () => {
    const fetchMetrics = fetchProcessMetricsFromSamples({
      samples: [sample(7, 1)],
      nodes: [node(7, { status: 'exited', exitedAt: '2026-01-01T00:00:05Z' })],
    });

    await expect(fetchMetrics(7)).rejects.toBeInstanceOf(ProcessNotFoundError);
  });

  it('rejects with a plain error, which keeps the poll going, while a live pid has no sample yet', async () => {
    const fetchMetrics = fetchProcessMetricsFromSamples({ samples: [sample(8, 1)], nodes: [node(7)] });

    const failure = await fetchMetrics(7).catch((e: unknown) => e);

    expect(failure).toBeInstanceOf(Error);
    expect(failure).not.toBeInstanceOf(ProcessNotFoundError);
    expect((failure as Error).message).toBe('no process sample for pid 7 yet');
  });
});
