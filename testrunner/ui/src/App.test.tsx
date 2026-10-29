import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { App, sourceUrl } from './App';
import type { Snapshot, Test } from './types';


class FakeEventSource {
  static CLOSED = 2;
  readyState = 1;
  onerror: (() => void) | null = null;
  constructor(readonly url: string) {}
  addEventListener() {}
  close() {}
}

const failing: Test = {
  name: 'TestParse', framework: 'go test', failed: true, package_path: './pkg', work_dir: '/repo', message: 'expected 3 but got 5',
  file: 'parse_test.go', line: 12,
};
const passing: Test = { name: 'TestOk', framework: 'go test', passed: true, package_path: './pkg', work_dir: '/repo' };

function snapshot(over: Partial<Snapshot> = {}): Snapshot {
  return {
    status: { running: false, test_edit_supported: false },
    tests: [{ name: 'pkg/', framework: 'go test', children: [failing, passing] }],
    metadata: { sequence: 1, kind: 'initial', started: '2026-01-01T00:00:00Z', ended: '2026-01-01T00:00:03Z' },
    ...over,
  };
}

interface Routes {
  [route: string]: (init?: RequestInit) => Response | Promise<Response>;
}

function serve(snap: Snapshot, routes: Routes = {}) {
  const calls: Array<{ url: string; init?: RequestInit }> = [];
  vi.stubGlobal('fetch', vi.fn(async (url: string, init?: RequestInit) => {
    calls.push({ url, init });
    if (url === '/api/tests') return new Response(JSON.stringify(snap));
    const route = routes[`${init?.method ?? 'GET'} ${url}`];
    if (route) return route(init);
    throw new Error(`unexpected ${init?.method ?? 'GET'} ${url}`);
  }));
  return calls;
}

// detailHeader is the header of the selected node's detail pane, whose actions are the node's own.
async function detailHeader(): Promise<HTMLElement> {
  return await waitFor(() => {
    const header = document.querySelector<HTMLElement>('[data-detail-header]');
    if (!header) throw new Error('no node is selected');
    return header;
  });
}

function open(url: string) {
  window.history.replaceState({}, '', url);
  return render(<App />);
}

describe('sourceUrl', () => {
  const browse = 'https://mc.example.com/entity/transaction/tx-1/xml';
  it.each([
    ['adds the finding\'s line', browse, 3, `${browse}?line=3`],
    ['keeps a line the source already names', `${browse}?line=9`, 3, `${browse}?line=9`],
    ['keeps other query params', `${browse}?doc=Math`, 3, `${browse}?doc=Math&line=3`],
    ['opens the file when the finding has no line', browse, undefined, browse],
    ['offers nothing without a URL', undefined, 3, undefined],
  ])('%s', (_name, url, line, want) => {
    expect(sourceUrl(url, line)).toBe(want);
  });
});

describe('App', () => {
  beforeEach(() => {
    vi.stubGlobal('EventSource', FakeEventSource);
  });
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
    window.history.replaceState({}, '', '/');
  });

  it('shows only the Tests tab for a plain test run and lists the tree under the failing-only default', async () => {
    serve(snapshot());
    open('/');

    expect(await screen.findByText(/> TestParse/)).toBeTruthy();
    expect(screen.queryByRole('tab', { name: /Lint/ })).toBeNull();
    expect(screen.queryByRole('tab', { name: /Bench/ })).toBeNull();
    expect(screen.queryByRole('tab', { name: /Processes/ })).toBeNull();
    expect(screen.queryByText('TestOk')).toBeNull();
  });

  it('adds a Lint tab for a run that linted, and switches to it with the URL', async () => {
    serve(snapshot({
      status: { running: false, lint_run: true },
      lint: [{ linter: 'eslint', work_dir: '/repo', success: false, duration: 1, violations: [{ file: '/repo/a.ts', line: 1, message: 'no-var', rule: { method: 'no-var' } }] }],
    }));
    open('/');

    const tab = await screen.findByRole('tab', { name: /Lint/ });
    fireEvent.click(tab);

    expect(window.location.pathname).toBe('/lint');
    expect(await screen.findByText('Lint Results')).toBeTruthy();
    expect(await screen.findByText('eslint')).toBeTruthy();
  });

  it('opens the node a deep link names, with its error', async () => {
    serve(snapshot());
    open('/tests/pkg-testparse');

    await waitFor(() => expect(screen.getAllByText('expected 3 but got 5').length).toBeGreaterThan(0));
    expect(window.location.pathname).toBe('/tests/pkg-testparse');
  });

  it('reruns the selected test\'s package and says when another rerun is already running', async () => {
    const calls = serve(snapshot(), { 'POST /api/rerun': () => new Response('rerun already in progress', { status: 409 }) });
    open('/tests/pkg-testparse');

    fireEvent.click(within(await detailHeader()).getByRole('button', { name: /Rerun/ }));

    await waitFor(() => expect(screen.getByText(/Another rerun is in progress/)).toBeTruthy());
    const rerun = calls.find(call => call.url === '/api/rerun')!;
    expect(JSON.parse(rerun.init!.body as string)).toEqual({
      package_paths: ['./pkg'], work_dir: '/repo', test_name: 'TestParse', suite: [], framework: 'go test',
    });
  });

  it('copies the AI prompt for a node from the markdown export narrowed to failed and timed-out tests', async () => {
    const written: string[] = [];
    vi.stubGlobal('navigator', { clipboard: { writeText: async (text: string) => { written.push(text); } } });
    const calls = serve(snapshot(), {
      // The server names the node by its uncollapsed ancestry and display label, not the view's "pkg-testparse".
      'GET /tests/pkg/parse.md?status=failed%2Ctimedout':() => new Response('# TestParse\nexpected 3 but got 5'),
    });
    open('/tests/pkg-testparse');

    fireEvent.click(within(await detailHeader()).getByRole('button', { name: /Copy AI Prompt/ }));

    await waitFor(() => expect(written).toHaveLength(1));
    expect(written[0]).toContain('expected 3 but got 5');
    expect(calls.some(call => call.url === '/tests/pkg/parse.md?status=failed%2Ctimedout')).toBe(true);
  });

  it('reads a running attempt\'s live CPU from the run\'s process samples, by the pid of its process', async () => {
    const attempt = { sequence: 1, pid: 4300, process_key: '4300@2', started: '2026-01-01T00:00:01Z' };
    const traces = {
      runs: { 'run-1': { mount: '/query/run-1' } },
      process_run: 'run-1',
      process: {
        tree: { kind: 'process_tree', stream: 't' },
        samples: { kind: 'process_samples', stream: 's' },
        stacks: { kind: 'stacks', stream: 'k' },
      },
    };
    const node = { key: '4300@2', pid: 4300, ppid: 1, name: 'pkg.test', command: 'pkg.test', status: 'active', startedAt: '2026-01-01T00:00:01Z', lastSeenAt: '2026-01-01T00:00:03Z', cpuPercent: 42.5, rssBytes: 3145728, peakCpuPercent: 42.5, peakRssBytes: 3145728 };
    const reading = { sampledAt: '2026-01-01T00:00:03Z', tick: 3, key: '4300@2', pid: 4300, memoryPercent: 1, rssBytes: 3145728, cpuPercent: 42.5 };
    const ndjson = (record: unknown) => () => new Response(`${JSON.stringify(record)}\n`);
    serve(snapshot({
      status: { running: true, test_edit_supported: false, traces },
      tests: [{ name: 'pkg/', framework: 'go test', children: [{ name: 'TestSlow', framework: 'go test', running: true, package_path: './pkg', work_dir: '/repo', attempts: [attempt] }] }],
    }), {
      'GET /query/run-1/api/v1/traces/process_tree/export?stream=t&format=ndjson': ndjson(node),
      'GET /query/run-1/api/v1/profile/profile-traces-process-samples?stream=s&sort=seq&order=desc&limit=1': () => new Response(JSON.stringify([{ seq: 3 }])),
      'GET /query/run-1/api/v1/traces/process_samples/export?stream=s&format=ndjson': ndjson(reading),
    });
    open('/tests/pkg-testslow');

    const metrics = await waitFor(() => {
      const section = document.querySelector<HTMLElement>('[data-attempt-sequence="1"]');
      if (!section || !section.textContent?.includes('42.5%')) throw new Error('attempt metrics not read from samples yet');
      return section;
    }, { timeout: 4000 });
    expect(metrics.textContent).toContain('4300');
  });

  it('stops the active run from the header', async () => {
    const calls = serve(snapshot({ status: { running: true, stop_supported: true, test_edit_supported: false } }), {
      'POST /api/stop': () => new Response('{}'),
    });
    open('/');

    fireEvent.click(await screen.findByRole('button', { name: /Stop/ }));

    await waitFor(() => expect(calls.some(call => call.url === '/api/stop')).toBe(true));
    expect(JSON.parse(calls.find(call => call.url === '/api/stop')!.init!.body as string)).toEqual({ scope: 'global' });
  });

  it('shows the run\'s timeout arguments as pills in the header', async () => {
    serve(snapshot({ metadata: { sequence: 1, kind: 'initial', started: '2026-01-01T00:00:00Z', args: { timeout: '30m', test_timeout: '5m' } } }));
    open('/');

    const pills = await screen.findByText(/global 30m/);
    expect(within(pills.closest('[data-timeout-pills]') as HTMLElement).getByText(/per-test 5m/)).toBeTruthy();
  });

  it('shows the live prompt tasks above the lint tree and counts the linters in the header', async () => {
    serve(snapshot({
      status: { running: true, lint_run: true },
      tests: [{
        name: 'lint prompt', framework: 'task', context: { status: 'running', type: 'group' },
        children: [{ name: 'messaging-review:report.mdx', framework: 'task', context: { status: 'running', type: 'task' }, message: 'tool: read_file' }],
      }],
      lint: [
        { linter: 'messaging-review:report.mdx', work_dir: '/repo', success: false, duration: 0, violations: [], status: 'running', run_id: 'run-1' },
        { linter: 'markdownlint', work_dir: '/repo', success: false, duration: 0, violations: [], status: 'pending' },
      ],
    }));
    open('/lint');

    const strip = await screen.findByRole('region', { name: 'Tasks' });
    expect(within(strip).getByText('tool: read_file')).toBeTruthy();
    expect((await screen.findAllByText('Running linters... (2 running · 0 done · 0 failed)')).length).toBeGreaterThan(0);
  });

  it('ignores a lint finding through the server and says it was saved', async () => {
    const calls = serve(snapshot({
      status: { running: false, lint_run: true },
      lint: [{ linter: 'eslint', work_dir: '/repo', success: false, duration: 1, violations: [{ file: '/repo/a.ts', line: 1, message: 'no-var', rule: { method: 'no-var' } }] }],
    }), { 'POST /api/lint/ignore': () => new Response('{}') });
    open('/lint/eslint');

    fireEvent.click(await screen.findByRole('button', { name: /Disable eslint entirely/ }));

    await waitFor(() => expect(screen.getByText('Ignore rule saved to .gavel.yaml')).toBeTruthy());
    expect(JSON.parse(calls.find(call => call.url === '/api/lint/ignore')!.init!.body as string)).toEqual({ source: 'eslint', work_dir: '/repo' });
  });

  describe('files tab', () => {
    // App loads the tab with React.lazy; transforming it (and clicky-ui's file-findings
    // view) the first time takes longer than testing-library's 1s find timeout.
    beforeAll(() => import('./FilesTab'), 30_000);

    const lintSnapshot = () => snapshot({
      git: { root: '/repo' },
      status: { running: false, lint_run: true },
      lint: [{ linter: 'eslint', work_dir: '/repo', success: false, duration: 1, violations: [{ file: '/repo/src/a.ts', line: 4, message: 'no-var', rule: { method: 'no-var' } }] }],
    });

    it('offers a Files tab only when a finding names a file', async () => {
      serve(snapshot({ tests: [{ name: 'pkg/', framework: 'go test', children: [{ name: 'TestNoFile', framework: 'go test', failed: true, work_dir: '/repo' }] }] }));
      open('/');
      expect(await screen.findByText(/> TestNoFile/)).toBeTruthy();
      expect(screen.queryByRole('tab', { name: /Files/ })).toBeNull();
      cleanup();

      serve(snapshot());
      open('/');
      expect(await screen.findByRole('tab', { name: /Files/ })).toBeTruthy();
    });

    it('requests a deep-linked file by its work dir and relative path and shows the server error in full', async () => {
      const calls = serve(snapshot(), {
        'GET /api/file?work_dir=%2Frepo&path=parse_test.go': () => new Response('file not found: parse_test.go', { status: 404 }),
      });
      open('/files/parse_test.go?dir=%2Frepo&line=12');

      const alert = await screen.findByRole('alert');
      expect(alert.textContent).toContain('file not found: parse_test.go');
      expect(calls.some(call => call.url === '/api/file?work_dir=%2Frepo&path=parse_test.go')).toBe(true);
    });

    it('lists a failed test that has no work dir under the snapshot git root and opens it with that root', async () => {
      const failingNoDir: Test = { name: 'TestParse', framework: 'go test', failed: true, file: '/repo/pkg/parse_test.go', line: 12, message: 'boom' };
      const calls = serve(snapshot({ git: { root: '/repo' }, tests: [{ name: 'pkg/', framework: 'go test', children: [failingNoDir] }] }), {
        'GET /api/file?work_dir=%2Frepo&path=pkg%2Fparse_test.go': () => new Response('missing', { status: 404 }),
      });
      open('/files');

      fireEvent.click(await screen.findByText(/parse_test\.go/));

      await waitFor(() => expect(calls.some(call => call.url === '/api/file?work_dir=%2Frepo&path=pkg%2Fparse_test.go')).toBe(true));
      expect(window.location.pathname).toBe('/files/pkg/parse_test.go');
      expect(window.location.search).toBe('?dir=%2Frepo');
    });

    it('opens a lint violation in the Files tab at its line from the Lint tab', async () => {
      serve(lintSnapshot(), { 'GET /api/file?work_dir=%2Frepo&path=src%2Fa.ts': () => new Response('missing', { status: 404 }) });
      open('/lint?grouping=linter-file');

      fireEvent.click(await screen.findByRole('button', { name: /Expand all/ }));
      fireEvent.click(await screen.findByText(/a\.ts/));
      fireEvent.click(await screen.findByRole('button', { name: /Open in Files/ }));

      await waitFor(() => expect(window.location.pathname).toBe('/files/src/a.ts'));
      expect(window.location.search).toBe('?dir=%2Frepo&line=4');
    });
  });

  describe('run history', () => {
    const lintRun = {
      runId: 'run-2026-01-02T00-00-00Z', path: '/repo/.gavel/run-2026-01-02T00-00-00Z.json', kind: 'lint' as const,
      started: '2026-01-02T00:00:00Z', ended: '2026-01-02T00:00:10Z',
      passed: 0, failed: 0, skipped: 0, warned: 0, total: 0, lintViolations: 4, lintLinters: 1,
      href: '/runs/run-2026-01-02T00-00-00Z/',
    };
    const testRun = {
      ...lintRun, runId: 'run-2026-01-01T00-00-00Z', path: '/repo/.gavel/run-2026-01-01T00-00-00Z.json', kind: 'test' as const,
      started: '2026-01-01T00:00:00Z', ended: '2026-01-01T00:00:03Z',
      passed: 1, failed: 1, total: 2, lintViolations: 0, lintLinters: 0, href: '/runs/run-2026-01-01T00-00-00Z/',
    };
    const page = { latest: [lintRun.runId], latestHref: '/', runs: [lintRun, testRun] };
    const withHistory = { 'GET /api/runs': () => new Response(JSON.stringify(page)) };

    function openWithNavigate(url: string) {
      const navigate = vi.fn();
      window.history.replaceState({}, '', url);
      render(<App navigate={navigate} />);
      return navigate;
    }

    it('offers no History tab or switcher when the host serves no history', async () => {
      serve(snapshot(), { 'GET /api/runs': () => new Response('404 page not found', { status: 404 }) });
      open('/');

      expect(await screen.findByText(/> TestParse/)).toBeTruthy();
      expect(screen.queryByRole('tab', { name: /History/ })).toBeNull();
      expect(document.querySelector('[data-run-switcher]')).toBeNull();
    });

    it('switches to a recorded run on the tab holding its results', async () => {
      serve(snapshot(), withHistory);
      const navigate = openWithNavigate('/');

      const switcher = await waitFor(() => {
        const select = document.querySelector<HTMLSelectElement>('[data-run-switcher] select');
        if (!select || select.options.length < 3) throw new Error('history not loaded');
        return select;
      });
      expect(switcher.value).toBe('');
      fireEvent.change(switcher, { target: { value: lintRun.runId } });

      expect(navigate).toHaveBeenCalledWith('/runs/run-2026-01-02T00-00-00Z/lint');
    });

    it('lists recorded runs on the History tab and opens the one picked', async () => {
      serve(snapshot(), withHistory);
      const navigate = openWithNavigate('/history');

      const link = await screen.findByRole('link', { name: new Date(testRun.started).toLocaleString() });
      expect(link.getAttribute('href')).toBe(testRun.href);
      fireEvent.click(link.closest('tr')!);

      expect(navigate).toHaveBeenCalledWith(testRun.href);
    });
  });
});
