// Tests the platform dashboard's infrastructure panel: what Mission Control knows about
// gavel's infrastructure, its health at a glance, and why it, or one of its lists, is missing.
import { render, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { queryTestWrapper } from '../todos/queryTestWrapper';
import { InfrastructurePanel, type Infrastructure } from './InfrastructurePanel';

let response: { status: number; body: unknown };

beforeEach(() => {
  vi.stubGlobal('fetch', vi.fn((input: RequestInfo | URL) => {
    const url = String(input);
    if (url !== '/api/platform/infrastructure') throw new Error(`unexpected fetch ${url}`);
    return Promise.resolve(new Response(JSON.stringify(response.body), { status: response.status }));
  }));
});

afterEach(() => {
  vi.unstubAllGlobals();
});

it('shows gavel\'s configs, checks and changes from Mission Control, with their health', async () => {
  const infrastructure: Infrastructure = {
    missionControl: 'https://mc.example.com/api',
    configs: [
      { id: 'c1', agent: '', name: 'gavel', namespace: 'gavel', type: 'Kubernetes::Deployment', health: 'healthy', status: 'Running', updated_at: '2026-11-10T10:00:00Z' },
      { id: 'c2', agent: '', name: 'gavel-postgres', namespace: 'gavel', type: 'Kubernetes::StatefulSet', health: 'unhealthy', status: 'CrashLoopBackOff' },
      { id: 'c3', agent: '', name: 'gavel-data', namespace: 'gavel', type: 'Kubernetes::PersistentVolumeClaim', health: 'healthy', status: 'Bound' },
    ],
    checks: [{ id: 'k1', agent: '', name: 'gavel-http', namespace: 'gavel', type: 'http', health: 'unhealthy', status: 'failing' }],
    changes: [{ id: 'ch1', agent: '', name: 'gavel', namespace: 'gavel', type: 'diff', severity: 'info', created_at: '2026-11-10T09:00:00Z' }],
  };
  response = { status: 200, body: infrastructure };
  render(<InfrastructurePanel />, { wrapper: queryTestWrapper() });

  expect(await screen.findByText('2 healthy, 1 unhealthy')).toBeTruthy();
  expect(screen.getByText('0 healthy, 1 unhealthy')).toBeTruthy();
  const configs = screen.getByRole('table', { name: 'Config items' });
  expect(within(within(configs).getByRole('row', { name: /gavel-postgres/ })).getAllByRole('cell').map(cell => cell.textContent).slice(0, 4))
    .toEqual(['gavel-postgres', 'Kubernetes::StatefulSet', 'unhealthy', 'CrashLoopBackOff']);
  expect(within(screen.getByRole('table', { name: 'Health checks' })).getByText('gavel-http')).toBeTruthy();
  expect(within(screen.getByRole('table', { name: 'Recent changes' })).getByText('diff')).toBeTruthy();
  expect(screen.getByRole('link', { name: 'https://mc.example.com/api' })).toBeTruthy();
});

it('says how to connect Mission Control when serve runs without one', async () => {
  response = { status: 404, body: { error: 'no Mission Control is configured: start gavel serve with --mission-control-url' } };
  render(<InfrastructurePanel />, { wrapper: queryTestWrapper() });

  expect((await screen.findByRole('alert')).textContent).toContain('--mission-control-url');
});

it('says why Mission Control couldn\'t be read', async () => {
  response = { status: 502, body: { error: 'read gavel\'s infrastructure from Mission Control at https://mc.example.com/api: server 401: invalid access token (check GAVEL_MISSION_CONTROL_TOKEN, a token with Mission Control\'s viewer role)' } };
  render(<InfrastructurePanel />, { wrapper: queryTestWrapper() });

  expect((await screen.findByRole('alert')).textContent).toContain('check GAVEL_MISSION_CONTROL_TOKEN');
});

it('says why one list is missing and shows the others', async () => {
  const infrastructure: Infrastructure = {
    missionControl: 'https://mc.example.com/api',
    configs: [{ id: 'c1', agent: '', name: 'gavel', namespace: 'gavel', type: 'Kubernetes::Deployment', health: 'healthy' }],
    checks: [],
    checksError: 'read gavel\'s health checks from Mission Control at https://mc.example.com/api: server 500: canaries are down',
    changes: [],
  };
  response = { status: 200, body: infrastructure };
  render(<InfrastructurePanel />, { wrapper: queryTestWrapper() });

  expect((await screen.findByRole('alert')).textContent).toContain('canaries are down');
  expect(within(screen.getByRole('table', { name: 'Config items' })).getByText('Kubernetes::Deployment')).toBeTruthy();
  expect(screen.queryByRole('table', { name: 'Health checks' })).toBeNull();
  expect(screen.queryByText('No health checks match the search.'), 'a list that failed isn\'t called empty').toBeNull();
});
