// Tests the hosted Projects tab: the organization's repositories, test and lint started on a
// picked branch, and the repository's recent test and lint runs.
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { Project } from '../types';
import { queryTestWrapper } from './todos/queryTestWrapper';
import { HostedProjectPane, HostedProjectsSidebar } from './HostedProjectsView';
import type { SandboxRun } from './RunsView';

const app: Project = { name: 'acme/app', dir: 'repo:1', repos: ['acme/app'] };
const site: Project = { name: 'acme/site', dir: 'repo:2', repos: ['acme/site'] };

const run = (overrides: Partial<SandboxRun>): SandboxRun => ({
  id: 'run', kind: 'test', status: 'succeeded', ref: 'refs/heads/main', sha: 'abc1234def', repo: 'acme/app',
  user: 'Ada Lovelace', backend: 'substrate', hasLog: true, createdAt: '2026-10-25T10:00:00Z', ...overrides,
});

let runs: SandboxRun[];
let started: { action: string; branch: string }[];
let fetchMock: ReturnType<typeof vi.fn>;

function respond(status: number, body: unknown) {
  return Promise.resolve(new Response(JSON.stringify(body), { status }));
}

beforeEach(() => {
  started = [];
  runs = [
    run({ id: 'run-lint', kind: 'lint', status: 'failed', exitCode: 1, ref: 'refs/heads/feat/x' }),
    run({ id: 'run-push', kind: 'push-check' }),
    run({ id: 'run-site', repo: 'acme/site' }),
    run({ id: 'run-test' }),
  ];
  fetchMock = vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input);
    if (url === '/api/sandbox-runs') return respond(200, runs);
    if (url === '/api/projects/acme%2Fapp/branches') return respond(200, { default: 'main', names: ['feat/x', 'main'] });
    if (url === '/api/projects/acme%2Fapp/actions' && init?.method === 'POST') {
      const body = JSON.parse(String(init.body));
      if (body.branch === 'feat/x' && body.action === 'lint') return respond(400, { error: 'unknown branch "feat/x": acme/app has no such branch' });
      started.push(body);
      return respond(202, { action: body.action, runId: 'run-new', href: '/runs/run-new', running: true });
    }
    throw new Error(`unexpected fetch ${init?.method ?? 'GET'} ${url}`);
  });
  vi.stubGlobal('fetch', fetchMock);
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('HostedProjectsSidebar', () => {
  it('lists the organization\'s repositories and selects one', () => {
    const onSelect = vi.fn();
    render(<HostedProjectsSidebar projects={[app, site]} selectedName="acme/app" onSelect={onSelect} />);

    expect(screen.getByRole('button', { name: 'acme/app' }).getAttribute('aria-current')).toBe('true');
    fireEvent.click(screen.getByRole('button', { name: 'acme/site' }));
    expect(onSelect).toHaveBeenCalledWith('acme/site');
  });
});

describe('HostedProjectPane', () => {
  it('starts test on the default branch and opens its run', async () => {
    const onSelectRun = vi.fn();
    render(<HostedProjectPane project={app} onSelectRun={onSelectRun} />, { wrapper: queryTestWrapper() });

    const branch = await screen.findByRole('combobox', { name: 'Branch' });
    await waitFor(() => expect((branch as HTMLSelectElement).value).toBe('main'));
    fireEvent.click(screen.getByRole('button', { name: 'Test' }));

    await waitFor(() => expect(onSelectRun).toHaveBeenCalledWith('run-new'));
    expect(started).toEqual([{ action: 'test', branch: 'main' }]);
  });

  it('runs on the branch the member picks and shows why a run could not start', async () => {
    render(<HostedProjectPane project={app} onSelectRun={vi.fn()} />, { wrapper: queryTestWrapper() });

    const branch = await screen.findByRole('combobox', { name: 'Branch' });
    await within(branch).findByRole('option', { name: 'feat/x' });
    fireEvent.change(branch, { target: { value: 'feat/x' } });
    fireEvent.click(screen.getByRole('button', { name: 'Lint' }));

    expect((await screen.findByRole('alert')).textContent).toContain('unknown branch "feat/x": acme/app has no such branch');
    expect(started).toEqual([]);
  });

  it('lists the repository\'s test and lint runs, newest first, each opening its run', async () => {
    const onSelectRun = vi.fn();
    render(<HostedProjectPane project={app} onSelectRun={onSelectRun} />, { wrapper: queryTestWrapper() });

    const list = await screen.findByRole('list', { name: 'Recent runs' });
    const items = await within(list).findAllByRole('button');
    expect(items.map(item => item.textContent)).toEqual([
      expect.stringContaining('Lint · feat/x · failed (exit 1)'),
      expect.stringContaining('Test · main · succeeded'),
    ]);
    fireEvent.click(items[0]);
    expect(onSelectRun).toHaveBeenCalledWith('run-lint');
  });
});
