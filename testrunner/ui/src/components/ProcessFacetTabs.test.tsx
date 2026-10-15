import { afterEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { ProcessDetailPanel } from './ProcessDetailPanel';
import type { ProcessTraceNode } from '../types';

// DataTable is mocked to a plain table of each column's cell text.
vi.mock('@flanksource/clicky-ui/data', () => ({
  DataTable: ({ data, columns, emptyMessage }: {
    data: Record<string, unknown>[];
    columns: Array<{ key: string; accessor?: (row: Record<string, unknown>) => unknown }>;
    emptyMessage?: string;
  }) => data.length === 0 ? <div>{emptyMessage}</div> : (
    <table><tbody>{data.map((row, i) => (
      <tr key={i}>{columns.map(c => <td key={c.key}>{String((c.accessor ? c.accessor(row) : row[c.key]) ?? '')}</td>)}</tr>
    ))}</tbody></table>
  ),
}));

const running: ProcessTraceNode = {
  key: '4242@1', pid: 4242, ppid: 1, name: 'gavel', command: 'gavel test --ui', status: 'sleeping', isRoot: true,
  startedAt: '2026-01-01T00:00:00Z', lastSeenAt: '2026-01-01T00:00:09Z',
  cpuPercent: 1, rssBytes: 1024, peakCpuPercent: 1, peakRssBytes: 1024,
};
const exited: ProcessTraceNode = { ...running, key: '4243@2', pid: 4243, isRoot: false, status: 'exited', exitedAt: '2026-01-01T00:00:07Z' };

function stubFetch(responses: Record<string, Response>) {
  const fetch = vi.fn(async (url: string) => {
    const response = responses[url];
    if (!response) throw new Error(`unexpected ${url}`);
    return response.clone();
  });
  vi.stubGlobal('fetch', fetch);
  return fetch;
}

describe('ProcessDetailPanel facet tabs', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('loads a process environment only when its tab is opened, sorted by name', async () => {
    const fetch = stubFetch({
      '/api/process/env?pid=4242': new Response(JSON.stringify({ ZETA: 'last', ACME_HOME: '/opt/acme' })),
    });
    render(<ProcessDetailPanel node={running} samples={[]} />);
    expect(fetch).not.toHaveBeenCalled();

    fireEvent.click(screen.getByText('Environment'));

    await waitFor(() => expect(screen.getByText('/opt/acme')).toBeTruthy());
    const names = screen.getAllByRole('row').map(row => row.firstElementChild?.textContent);
    expect(names).toEqual(['ACME_HOME', 'ZETA']);
    expect(fetch).toHaveBeenCalledTimes(1);
  });

  it('lists listening ports and refetches on refresh', async () => {
    const fetch = stubFetch({
      '/api/process/ports?pid=4242': new Response(JSON.stringify([
        { fd: '5', type: 'IPv4', name: '127.0.0.1:9092', protocol: 'TCP', state: 'LISTEN', localAddr: '127.0.0.1:9092', port: 9092 },
      ])),
    });
    render(<ProcessDetailPanel node={running} samples={[]} />);

    fireEvent.click(screen.getByText('Ports'));
    await waitFor(() => expect(screen.getByText('9092')).toBeTruthy());
    expect(screen.getByText('127.0.0.1:9092')).toBeTruthy();

    fireEvent.click(screen.getByText('Refresh'));
    await waitFor(() => expect(fetch).toHaveBeenCalledTimes(2));
  });

  it('shows the server refusal for a facet', async () => {
    stubFetch({
      '/api/process/env?pid=4242': new Response('a process environment is only served to a client on this machine\n', { status: 403 }),
    });
    render(<ProcessDetailPanel node={running} samples={[]} />);

    fireEvent.click(screen.getByText('Environment'));
    await waitFor(() => expect(screen.getByText('a process environment is only served to a client on this machine')).toBeTruthy());
  });

  it('does not query an exited process', () => {
    const fetch = stubFetch({});
    render(<ProcessDetailPanel node={exited} samples={[]} />);

    fireEvent.click(screen.getByText('Open files'));
    expect(screen.getByText(/process has exited/)).toBeTruthy();
    expect(fetch).not.toHaveBeenCalled();
  });
});
