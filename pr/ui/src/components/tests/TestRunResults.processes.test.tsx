import { afterEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import type { TracesStatus } from '@flanksource/gavel/testrunner';
import { TestRunResults } from './TestRunResults';
import type { RunSnapshot } from './types';

vi.mock('@flanksource/clicky-ui/data', async importOriginal => ({
  ...(await importOriginal<typeof import('@flanksource/clicky-ui/data')>()),
  TestRunner: () => <div>Test results</div>,
}));

vi.mock('./CreateTodoFromRunDialog', () => ({ CreateTodoFromRunDialog: () => null }));

const base = '/api/project-runs/run-9';
const mount = '/api/traces/k/query/run-1';
const traces = {
  process_run: 'run-1',
  process: {
    tree: { kind: 'process_tree', stream: 't' },
    samples: { kind: 'process_samples', stream: 's' },
    stacks: { kind: 'stacks', stream: 'k' },
  },
  // A stored snapshot can carry the session ids of the run it recorded; the run it came from is over.
  sessions: { tree: 'st', samples: 'ss', stacks: 'sk' },
  runs: { 'run-1': { mount } },
} as TracesStatus;

const root = {
  key: '100@1', pid: 100, ppid: 1, name: 'gavel', command: 'gavel test ./...', status: 'sleeping', isRoot: true,
  startedAt: '2026-01-01T00:00:00Z', lastSeenAt: '2026-01-01T00:00:09Z',
  cpuPercent: 3, rssBytes: 4096, peakCpuPercent: 9, peakRssBytes: 8192,
  stackTransports: [{ transport: 'native', destructive: false }],
};
const dump = {
  checkpointId: 'c1', takenAt: '2026-01-01T00:00:05Z', reason: 'package-timeout:pkg', transport: 'sigquit', key: root.key, pid: 100,
  format: 'native', raw: 'Thread 1:\n  recorded_wait()\n',
};

const snapshot: RunSnapshot = { status: { running: false, traces }, tests: [{ name: 'TestSave', failed: true }] };

function stubTraceStore() {
  vi.stubGlobal('fetch', vi.fn(async (url: string) => {
    const ndjson = (record: unknown) => new Response(`${JSON.stringify(record)}\n`);
    if (url === `${base}${mount}/api/v1/traces/process_tree/export?stream=t&format=ndjson`) return ndjson(root);
    if (url === `${base}${mount}/api/v1/traces/stacks/export?stream=k&format=ndjson`) return ndjson(dump);
    if (url.startsWith(`${base}${mount}/api/v1/profile/`)) return new Response('[]');
    throw new Error(`unexpected ${url}`);
  }));
}

function openProcesses(done: boolean) {
  render(<TestRunResults snapshot={snapshot} traces={traces} base={base} done={done} runKey="run-9" projectName="gavel" projectDir="/work/gavel" />);
  fireEvent.click(screen.getByRole('button', { name: /Processes/ }));
}

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe('TestRunResults Processes view', () => {
  it('shows the stack a stored run recorded and offers no way to collect another', async () => {
    stubTraceStore();
    openProcesses(true);

    expect(await screen.findByText(/recorded_wait\(\)/)).toBeTruthy();
    expect(screen.queryByRole('button', { name: /native sample/i })).toBeNull();
    expect(screen.queryByRole('button', { name: /stack trace/i })).toBeNull();
  });

  it('offers to collect a stack from the live run server while the run is going', async () => {
    stubTraceStore();
    openProcesses(false);

    expect(await screen.findByRole('button', { name: /native sample/i })).toBeTruthy();
  });
});
