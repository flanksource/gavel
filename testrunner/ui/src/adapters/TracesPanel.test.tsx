import { afterEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { TracesPanel } from './TracesPanel';
import type { TraceResult, TracesStatus } from '../types';

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

const traced: TraceResult = {
  run_id: 'run-1', kind: 'sql_audit', stream: 's1', afterSeq: 3, toSeq: 7, count: 4,
  summary: { count: 4, errors: 1, hosts: ['api.example.com'] },
};
const kinds = [{ name: 'sql_audit', title: 'Audit', exportFormats: [
  { name: 'ndjson', ext: 'ndjson', mime: 'application/x-ndjson' }, { name: 'csv', ext: 'csv', mime: 'text/csv' },
] }];

describe('TracesPanel', () => {
  it('links each window into the query app and its export formats, and views its records inline', async () => {
    vi.stubGlobal('fetch', vi.fn(async (url: string) => {
      if (url === '/query/run-1/api/v1/traces/kinds') return new Response(JSON.stringify(kinds));
      if (url === '/query/run-1/api/v1/traces/sql_audit/export?stream=s1&afterSeq=3&toSeq=7&format=ndjson') {
        return new Response('{"statement":"select 1"}\n{"statement":"select 2"}\n');
      }
      throw new Error(`unexpected ${url}`);
    }));

    render(<TracesPanel windows={[traced]} traces={{ runs: { 'run-1': { mount: '/query/run-1' } } }} />);

    expect(screen.getByText('Open in query').getAttribute('href')).toBe('/query/run-1/traces/sql_audit?stream=s1&afterSeq=3&toSeq=7');
    expect(screen.getByText('4 records')).toBeTruthy();
    await waitFor(() => expect(screen.getByText('CSV').getAttribute('href'))
      .toBe('/query/run-1/api/v1/traces/sql_audit/export?stream=s1&afterSeq=3&toSeq=7&format=csv'));

    fireEvent.click(screen.getByText('View'));
    expect(await screen.findByText('select 2')).toBeTruthy();
  });

  it('opens each window in the store of the run it names, under the base it was rebased onto', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify(kinds))));
    const status: TracesStatus = { runs: { 'run-1': { mount: '/base/query/run-1' }, 'run-2': { mount: '/base/query/run-2' } } };
    const { container } = render(<TracesPanel windows={[traced, { ...traced, run_id: 'run-2', stream: 's2' }]} traces={status} />);

    const links = within(container).getAllByText('Open in query').map(link => link.getAttribute('href'));
    expect(links).toEqual([
      '/base/query/run-1/traces/sql_audit?stream=s1&afterSeq=3&toSeq=7',
      '/base/query/run-2/traces/sql_audit?stream=s2&afterSeq=3&toSeq=7',
    ]);
    await waitFor(() => expect(within(container).getAllByText('CSV')).toHaveLength(2));
  });

  describe('goprof windows', () => {
    const goprofWindow: TraceResult = {
      run_id: 'run-1', kind: 'goprof', stream: 'g1', count: 1,
      summary: { count: 1, profiles: 'cpu' },
    };
    const goprofKinds = [
      ...kinds,
      { name: 'goprof', title: 'Go profiles', exportFormats: [
        { name: 'ndjson', ext: 'ndjson', mime: 'application/x-ndjson' }, { name: 'pprof', ext: 'pb.gz', mime: 'application/gzip' },
      ] },
    ];
    const profileRecord = {
      checkpointId: 'p1', takenAt: '2026-01-01T00:00:06Z', reason: 'goProfiles.cpu', source: 'file', key: '4243@2', pid: 4243,
      profile: 'cpu', sampleType: 'cpu', unit: 'nanoseconds', value: 5_000_000, stackHash: 'h',
      frames: [{ func: 'main.hotLoop', file: 'main.go', line: 12 }],
    };
    const goprofExport = '/query/run-1/api/v1/traces/goprof/export?stream=g1&format=';

    it('renders the window\'s records with the Go profile view and links its pprof download', async () => {
      vi.stubGlobal('fetch', vi.fn(async (url: string) => {
        if (url === '/query/run-1/api/v1/traces/kinds') return new Response(JSON.stringify(goprofKinds));
        if (url === `${goprofExport}ndjson`) return new Response(`${JSON.stringify(profileRecord)}\n`);
        throw new Error(`unexpected ${url}`);
      }));

      const { container } = render(<TracesPanel windows={[goprofWindow]} traces={{ runs: { 'run-1': { mount: '/query/run-1' } } }} />);

      await waitFor(() => expect(screen.getByText('PPROF').getAttribute('href')).toBe(`${goprofExport}pprof`));
      fireEvent.click(screen.getByText('View'));
      await waitFor(() => expect(container.querySelector('[data-go-profile="cpu"]')).not.toBeNull());
      expect(screen.getByText('main.hotLoop')).toBeTruthy();
    });

    it('shows one profile view per process when the window holds the records of several', async () => {
      vi.stubGlobal('fetch', vi.fn(async (url: string) => {
        if (url === '/query/run-1/api/v1/traces/kinds') return new Response(JSON.stringify(goprofKinds));
        return new Response([profileRecord, { ...profileRecord, key: '9@9', pid: 9 }].map(record => JSON.stringify(record)).join('\n'));
      }));

      const { container } = render(<TracesPanel windows={[goprofWindow]} traces={{ runs: { 'run-1': { mount: '/query/run-1' } } }} />);
      fireEvent.click(screen.getByText('View'));

      await waitFor(() => expect(container.querySelectorAll('[data-go-profile="cpu"]')).toHaveLength(2));
      expect(screen.getByText('4243@2')).toBeTruthy();
      expect(screen.getByText('9@9')).toBeTruthy();
    });

    it('says why a goprof window of a run whose store is gone cannot be opened', () => {
      vi.stubGlobal('fetch', vi.fn());
      render(<TracesPanel windows={[goprofWindow]} traces={{ runs: { 'run-1': { unavailable: 'trace store .gavel/traces/run-1 is gone' } } }} />);

      expect(screen.getByText('trace store .gavel/traces/run-1 is gone')).toBeTruthy();
      expect(screen.queryByText('View')).toBeNull();
      expect(fetch).not.toHaveBeenCalled();
    });

    it('keeps other kinds on the generic window viewer next to a goprof window', async () => {
      vi.stubGlobal('fetch', vi.fn(async (url: string) => {
        if (url === '/query/run-1/api/v1/traces/kinds') return new Response(JSON.stringify(goprofKinds));
        if (url.endsWith('format=ndjson') && url.includes('/sql_audit/')) return new Response('{"statement":"select 2"}\n');
        throw new Error(`unexpected ${url}`);
      }));

      const { container } = render(<TracesPanel windows={[traced, goprofWindow]} traces={{ runs: { 'run-1': { mount: '/query/run-1' } } }} />);
      await waitFor(() => expect(within(container).getAllByText('Open in query')).toHaveLength(2));

      fireEvent.click(within(container.querySelector('[data-trace-kind="sql_audit"]') as HTMLElement).getByText('View'));
      expect(await screen.findByText('select 2')).toBeTruthy();
    });
  });

  describe('stacks windows', () => {
    const stacksWindow: TraceResult = { run_id: 'run-1', kind: 'stacks', stream: 'k1', count: 2, summary: { count: 2, checkpoints: 1 } };
    const stackBase = { checkpointId: 'c1', takenAt: '2026-01-01T00:00:06Z', reason: 'timeout', transport: 'inprocess', key: '4243@2', pid: 4243, format: 'go' };
    const stackRecords = [
      { ...stackBase, goroutineId: 1, state: 'running', frames: [{ func: 'main.main', file: '/app/main.go', line: 12 }], raw: 'goroutine 1 [running]:' },
      { ...stackBase, goroutineId: 7, state: 'select', frames: [{ func: 'main.serve', file: '/app/serve.go', line: 40 }], raw: 'goroutine 7 [select]:' },
    ];

    it('renders the window\'s records as a stack dump per checkpoint through the generic window viewer', async () => {
      vi.stubGlobal('fetch', vi.fn(async (url: string) => {
        if (url === '/query/run-1/api/v1/traces/kinds') return new Response(JSON.stringify([{ name: 'stacks', title: 'Stacks', exportFormats: [] }]));
        if (url === '/query/run-1/api/v1/traces/stacks/export?stream=k1&format=ndjson') return new Response(stackRecords.map(record => JSON.stringify(record)).join('\n'));
        throw new Error(`unexpected ${url}`);
      }));

      const { container } = render(<TracesPanel windows={[stacksWindow]} traces={{ runs: { 'run-1': { mount: '/query/run-1' } } }} />);
      fireEvent.click(screen.getByText('View'));

      await waitFor(() => expect(container.querySelector('[data-stack-checkpoint="c1"]')).not.toBeNull());
      expect(screen.getByText('2 / 2 goroutines')).toBeTruthy();
      expect(screen.getByText('g7')).toBeTruthy();
    });
  });

  it('says why a run whose store is gone cannot open its windows', () => {
    vi.stubGlobal('fetch', vi.fn());
    render(<TracesPanel windows={[{ ...traced, error: 'summarize: stream gone' }]}
      traces={{ runs: { 'run-1': { unavailable: 'trace store .gavel/traces/run-1 is gone' } } }} />);

    expect(screen.getByText('trace store .gavel/traces/run-1 is gone')).toBeTruthy();
    expect(screen.getByText('summarize: stream gone')).toBeTruthy();
    expect(screen.queryByText('Open in query')).toBeNull();
    expect(fetch).not.toHaveBeenCalled();
  });

  it('says a window of a run this view does not serve cannot be opened', () => {
    vi.stubGlobal('fetch', vi.fn());
    render(<TracesPanel windows={[{ ...traced, run_id: 'run-other' }]} traces={{ runs: { 'run-1': { mount: '/query/run-1' } } }} />);
    expect(screen.getByText('The trace store of run run-other is not served by this view.')).toBeTruthy();
    expect(screen.queryByText('Open in query')).toBeNull();
  });
});
