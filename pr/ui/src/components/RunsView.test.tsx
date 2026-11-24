import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { queryTestWrapper } from './todos/queryTestWrapper';
import { RunsView, type SandboxRun } from './RunsView';

const pushCheck: SandboxRun = {
  id: 'run-push', kind: 'push-check', status: 'failed', exitCode: 3, ref: 'refs/heads/main', sha: 'abc1234def',
  repo: 'app', user: 'Ada Lovelace', backend: 'substrate', hasLog: true,
  createdAt: '2026-10-22T10:00:00Z', startedAt: '2026-10-22T10:00:05Z', finishedAt: '2026-10-22T10:01:05Z',
};
const agentRun: SandboxRun = {
  id: 'run-agent', kind: 'agent', status: 'running', ref: 'refs/heads/main', sha: 'fedcba9876', repo: 'app',
  user: 'Grace Hopper', todoId: 'todo-1', todoTitle: 'Fix the build', backend: 'substrate', hasLog: false,
  createdAt: '2026-10-22T10:05:00Z', startedAt: '2026-10-22T10:05:02Z',
};

let runs: SandboxRun[];
let logs: Record<string, { status: number; body: string }>;
let fetchMock: ReturnType<typeof vi.fn>;

function respond(status: number, body: string) {
  return Promise.resolve(new Response(body, { status }));
}

beforeEach(() => {
  runs = [agentRun, pushCheck];
  logs = { 'run-push': { status: 200, body: 'checking\nexit 3\n' } };
  fetchMock = vi.fn((input: RequestInfo | URL) => {
    const url = String(input);
    if (url === '/api/sandbox-runs') return respond(200, JSON.stringify(runs));
    const log = url.match(/^\/api\/sandbox-runs\/([^/]+)\/log$/);
    if (log) {
      const found = logs[decodeURIComponent(log[1])];
      return found ? respond(found.status, found.body) : respond(404, JSON.stringify({ error: `sandbox run ${log[1]} has no stored log` }));
    }
    throw new Error(`unexpected fetch ${url}`);
  });
  vi.stubGlobal('fetch', fetchMock);
});

afterEach(() => {
  vi.unstubAllGlobals();
});

function renderRuns(props: Partial<React.ComponentProps<typeof RunsView>> = {}) {
  return render(<RunsView selectedId="" onSelect={vi.fn()} {...props} />, { wrapper: queryTestWrapper() });
}

describe('RunsView', () => {
  it('lists the organization\'s runs with what ran, where, how it ended and how long it took', async () => {
    renderRuns();

    const [agent, push] = await screen.findAllByRole('option');
    const pushRow = within(push);
    expect(pushRow.getByText('Push check')).toBeTruthy();
    expect(pushRow.getByText('app')).toBeTruthy();
    expect(pushRow.getByText('main@abc1234')).toBeTruthy();
    expect(push.textContent).toContain('failed (exit 3)');
    expect(pushRow.getByText('1m 0s')).toBeTruthy();
    expect(within(agent).getByText('Agent run')).toBeTruthy();
    expect(agent.textContent).toContain('running');
  });

  it('says why a run waits or why it never ran', async () => {
    runs = [
      { ...agentRun, id: 'run-queued', status: 'queued', statusMessage: 'queued: this organization\'s runs hold 8 of its 8 concurrent cores, and no run is ahead of this one' },
      { ...pushCheck, id: 'run-refused', status: 'failed', exitCode: undefined, statusMessage: 'budget used up: this organization has used $50.00 of its $50.00 model cost budget' },
    ];
    renderRuns({ selectedId: 'run-queued' });

    expect(await screen.findByText('queued: this organization\'s runs hold 8 of its 8 concurrent cores, and no run is ahead of this one')).toBeTruthy();
    fireEvent.click(screen.getAllByRole('option')[1]);
    cleanup();
    renderRuns({ selectedId: 'run-refused' });
    expect(await screen.findByText('budget used up: this organization has used $50.00 of its $50.00 model cost budget')).toBeTruthy();
  });

  it('selects a run from its row', async () => {
    const onSelect = vi.fn();
    renderRuns({ onSelect });

    fireEvent.click(await screen.findByText('Push check'));

    expect(onSelect).toHaveBeenCalledExactlyOnceWith('run-push');
  });

  it('shows the selected run\'s log', async () => {
    renderRuns({ selectedId: 'run-push' });

    const log = await screen.findByRole('log');
    expect(log.textContent).toContain('checking\nexit 3');
  });

  it('says why a selected run has no log', async () => {
    renderRuns({ selectedId: 'run-agent' });

    expect(await screen.findByText(/sandbox run run-agent has no stored log/)).toBeTruthy();
  });

  it('says when the organization has no runs yet', async () => {
    runs = [];
    renderRuns();

    expect(await screen.findByText('No sandbox runs yet. Pushes and hosted agent runs show up here.')).toBeTruthy();
  });

  it('reports a failure to list the runs', async () => {
    fetchMock.mockImplementationOnce(() => respond(500, JSON.stringify({ error: 'list sandbox runs: database is down' })));
    renderRuns();

    await waitFor(() => expect(screen.getByRole('alert').textContent).toContain('list sandbox runs: database is down'));
  });
});
