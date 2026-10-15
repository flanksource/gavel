import { afterEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { TracesSection } from './TracesSection';
import type { TraceResult, TracesStatus } from '../types';

vi.mock('@flanksource/clicky-ui/data', () => ({
  TraceView: ({ kind, rows }: { kind: string; rows: unknown[] }) => <div>{`${kind} viewer of ${rows.length} records`}</div>,
}));

const traced: TraceResult = {
  run_id: 'run-1', kind: 'http_proxy', stream: 's1', afterSeq: 3, toSeq: 7, count: 4,
  summary: { count: 4, errors: 1, hosts: ['api.example.com'] },
};
const kinds = [{ name: 'http_proxy', title: 'HTTP', exportFormats: [
  { name: 'ndjson', ext: 'ndjson', mime: 'application/x-ndjson' }, { name: 'har', ext: 'har', mime: 'application/json' },
] }];

describe('TracesSection', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('links each window into the query app and its export formats, and views its records inline', async () => {
    vi.stubGlobal('fetch', vi.fn(async (url: string) => {
      if (url === '/query/run-1/api/v1/traces/kinds') return new Response(JSON.stringify(kinds));
      if (url === '/query/run-1/api/v1/traces/http_proxy/export?stream=s1&afterSeq=3&toSeq=7&format=ndjson') {
        return new Response('{"url":"/a"}\n{"url":"/b"}\n');
      }
      throw new Error(`unexpected ${url}`);
    }));

    render(<TracesSection windows={[traced]} traces={{ runs: { 'run-1': { mount: '/query/run-1' } } }} />);

    expect(screen.getByText('Open in query').getAttribute('href')).toBe('/query/run-1/traces/http_proxy?stream=s1&afterSeq=3&toSeq=7');
    expect(screen.getByText('4 records')).toBeTruthy();
    expect(screen.getByText('1')).toBeTruthy();
    await waitFor(() => expect(screen.getByText('HAR').getAttribute('href'))
      .toBe('/query/run-1/api/v1/traces/http_proxy/export?stream=s1&afterSeq=3&toSeq=7&format=har'));

    fireEvent.click(screen.getByText('View'));
    await waitFor(() => expect(screen.getByText('http_proxy viewer of 2 records')).toBeTruthy());
  });

  it('opens each window in the store of the run it names', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify(kinds))));
    const status: TracesStatus = { runs: { 'run-1': { mount: '/query/run-1' }, 'run-2': { mount: '/query/run-2' } } };
    const { container } = render(<TracesSection windows={[traced, { ...traced, run_id: 'run-2', stream: 's2' }]} traces={status} />);

    const links = within(container).getAllByText('Open in query').map(link => link.getAttribute('href'));
    expect(links).toEqual([
      '/query/run-1/traces/http_proxy?stream=s1&afterSeq=3&toSeq=7',
      '/query/run-2/traces/http_proxy?stream=s2&afterSeq=3&toSeq=7',
    ]);
    await waitFor(() => expect(within(container).getAllByText('HAR')).toHaveLength(2));
  });

  it('says why a run whose store is gone cannot open its windows', () => {
    vi.stubGlobal('fetch', vi.fn());
    render(<TracesSection windows={[{ ...traced, error: 'summarize: stream gone' }]}
      traces={{ runs: { 'run-1': { unavailable: 'trace store .gavel/traces/run-1 is gone' } } }} />);

    expect(screen.getByText('trace store .gavel/traces/run-1 is gone')).toBeTruthy();
    expect(screen.getByText('summarize: stream gone')).toBeTruthy();
    expect(screen.queryByText('Open in query')).toBeNull();
    expect(fetch).not.toHaveBeenCalled();
  });

  it('says a window of a run this view does not serve cannot be opened', () => {
    vi.stubGlobal('fetch', vi.fn());
    render(<TracesSection windows={[{ ...traced, run_id: 'run-other' }]} traces={{ runs: { 'run-1': { mount: '/query/run-1' } } }} />);
    expect(screen.getByText('The trace store of run run-other is not served by this view.')).toBeTruthy();
    expect(screen.queryByText('Open in query')).toBeNull();
  });
});
