// Tests the platform dashboard's infrastructure panel: gavel's catalog by type, its health checks
// when it has any, and changes, their health at a glance, and why it, or one list, is missing.
import { render, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { queryTestWrapper } from '../todos/queryTestWrapper';
import { InfrastructurePanel, healthTone, infrastructureSubsections, severityTone, type Infrastructure } from './InfrastructurePanel';
import { formatDateTime } from './dates';

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
      { id: 'c2', agent: '', name: 'gavel-postgres', namespace: 'gavel', type: 'Kubernetes::StatefulSet', health: 'unhealthy', status: 'CrashLoopBackOff' },
      { id: 'c1', agent: '', name: 'gavel', namespace: 'gavel', type: 'Kubernetes::Deployment', health: 'healthy', status: 'Running', updated_at: '2026-11-10T10:00:00Z' },
      { id: 'c3', agent: '', name: 'gavel-data', namespace: 'gavel', type: 'Kubernetes::PersistentVolumeClaim', health: 'healthy', status: 'Bound' },
    ],
    checks: [{ id: 'k1', agent: '', name: 'gavel-http', namespace: 'gavel', type: 'http', health: 'unhealthy', status: 'failing' }],
    changes: [{ id: 'ch1', config_id: 'c1', name: 'gavel', type: 'Kubernetes::Deployment', change_type: 'diff', summary: 'status.replicas', severity: 'info', created_at: '2026-11-10T09:00:00Z' }],
    changesDays: 7,
  };
  response = { status: 200, body: infrastructure };
  render(<InfrastructurePanel />, { wrapper: queryTestWrapper() });

  expect(await screen.findByText('2 healthy, 1 unhealthy')).toBeTruthy();
  expect(screen.getByText('0 healthy, 1 unhealthy')).toBeTruthy();
  const catalog = screen.getByRole('table', { name: 'Catalog' });
  expect(within(catalog).getAllByRole('row').slice(1).map(row => within(row).getAllByRole('cell').map(cell => cell.textContent).slice(0, 4)), 'the catalog shows names first, ordered by type').toEqual([
    ['gavel', 'Kubernetes::Deployment', 'healthy', 'Running'],
    ['gavel-data', 'Kubernetes::PersistentVolumeClaim', 'healthy', 'Bound'],
    ['gavel-postgres', 'Kubernetes::StatefulSet', 'unhealthy', 'CrashLoopBackOff'],
  ]);
  expect(within(catalog).getByRole('link', { name: 'gavel-postgres' }).getAttribute('href'), 'each name opens the config item in Mission Control')
    .toBe('https://mc.example.com/catalog/c2');
  expect(within(within(catalog).getAllByRole('row')[1]).getAllByRole('cell')[5].textContent, 'with when it was last updated')
    .toBe(formatDateTime('2026-11-10T10:00:00Z'));
  expect(within(screen.getByRole('table', { name: 'Health checks' })).getByText('gavel-http')).toBeTruthy();
  const changes = screen.getByRole('table', { name: 'Changes' });
  expect(within(changes).getAllByRole('row').slice(1).map(row => within(row).getAllByRole('cell').map(cell => cell.textContent)), 'each change names its config item and its type, and says when it happened').toEqual([
    ['gavel', 'Kubernetes::Deployment', 'diff', 'status.replicas', 'info', formatDateTime('2026-11-10T09:00:00Z')],
  ]);
  expect(within(changes).getByRole('link', { name: 'gavel' }).getAttribute('href')).toBe('https://mc.example.com/catalog/c1');
  expect(within(changes).getAllByRole('columnheader').map(cell => cell.textContent)).toEqual(['Config', 'Type', 'Change', 'Summary', 'Severity', 'When']);
  expect(within(changes).getByText('Kubernetes::Deployment').closest('td')?.firstElementChild?.tagName, 'the config type is a badge, like the change type')
    .toBe(within(changes).getByText('diff').closest('td')?.firstElementChild?.tagName);
  expect(screen.getByText('1 in 7 days'), 'the changes count says how far back it reaches').toBeTruthy();
  expect(screen.getAllByRole('heading', { level: 3 }).map(heading => [heading.textContent, heading.closest('[id]')?.id]), 'each part is a subsection the sidebar scrolls to').toEqual([
    ['Catalog', 'platform-infrastructure-catalog'],
    ['Health checks', 'platform-infrastructure-checks'],
    ['Changes', 'platform-infrastructure-changes'],
  ]);
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
    changesDays: 7,
  };
  response = { status: 200, body: infrastructure };
  render(<InfrastructurePanel />, { wrapper: queryTestWrapper() });

  expect((await screen.findByRole('alert')).textContent).toContain('canaries are down');
  expect(within(screen.getByRole('table', { name: 'Catalog' })).getByText('Kubernetes::Deployment')).toBeTruthy();
  expect(screen.queryByRole('table', { name: 'Health checks' })).toBeNull();
  expect(screen.queryByText('No health checks match the search.'), 'a list that failed isn\'t called empty').toBeNull();
});

it('leaves health checks out when gavel has none', async () => {
  response = { status: 200, body: { missionControl: 'https://mc.example.com/api', configs: [], checks: [], changes: [], changesDays: 7 } satisfies Infrastructure };
  render(<InfrastructurePanel />, { wrapper: queryTestWrapper() });

  expect(await screen.findByText('No config items match the search.')).toBeTruthy();
  expect(screen.getByText('No changes to these config items in the last 7 days.')).toBeTruthy();
  expect(screen.getAllByRole('heading', { level: 3 }).map(heading => heading.textContent)).toEqual(['Catalog', 'Changes']);
  expect(screen.queryByText('Health checks')).toBeNull();
});

it('names the subsections the sidebar lists: health checks only when gavel has some, or they couldn\'t be read', () => {
  const base: Infrastructure = { missionControl: '', configs: [], checks: [], changes: [], changesDays: 7 };
  const check = { id: 'k1', agent: '', name: 'gavel-http', namespace: 'gavel', type: 'http' };
  expect(infrastructureSubsections(undefined)).toEqual([]);
  expect(infrastructureSubsections(base).map(s => s.label)).toEqual(['Catalog', 'Changes']);
  expect(infrastructureSubsections({ ...base, checks: [check] }).map(s => s.label)).toEqual(['Catalog', 'Health checks', 'Changes']);
  expect(infrastructureSubsections({ ...base, checksError: 'down' }).map(s => s.id)).toEqual(['infrastructure-catalog', 'infrastructure-checks', 'infrastructure-changes']);
});

it('colors health and change severity badges by what they say', () => {
  expect(['healthy', 'unhealthy', 'warning', 'unknown', ''].map(healthTone)).toEqual(['success', 'danger', 'warning', 'neutral', 'neutral']);
  expect(['critical', 'high', 'medium', 'low', 'info', ''].map(severityTone)).toEqual(['danger', 'danger', 'warning', 'info', 'neutral', 'neutral']);
});
