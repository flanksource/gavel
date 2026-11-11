// Tests the platform dashboard's substrate inventory panel: workers' capacity against what their
// actors use, actors by owner and state, and why one of the lists is missing.
import { render, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { queryTestWrapper } from '../todos/queryTestWrapper';
import { SubstrateInventoryPanel, type SubstrateInventory } from './SubstrateInventoryPanel';

let response: { status: number; body: unknown };

beforeEach(() => {
  vi.stubGlobal('fetch', vi.fn((input: RequestInfo | URL) => {
    const url = String(input);
    if (url !== '/api/platform/substrate/inventory') throw new Error(`unexpected fetch ${url}`);
    return Promise.resolve(new Response(JSON.stringify(response.body), { status: response.status }));
  }));
});

afterEach(() => {
  vi.unstubAllGlobals();
});

const gib = 1 << 30;

function cells(name: string) {
  const table = screen.getByRole('table', { name });
  return within(table).getAllByRole('row').slice(1).map(row => within(row).getAllByRole('cell').map(cell => cell.textContent));
}

it('shows the workers\' capacity and use, and every actor by owner and state', async () => {
  const inventory: SubstrateInventory = {
    workers: [
      { name: 'pool-a-0', pool: 'pool-a', node: 'node-1', state: 'active',
        capacity: { cpuMillicores: 4000, memoryBytes: 8 * gib, actors: 10 }, allocated: { cpuMillicores: 1500, memoryBytes: 2 * gib, actors: 3 } },
      { name: 'pool-a-1', pool: 'pool-a', node: 'node-2', state: 'draining',
        capacity: { cpuMillicores: 4000, memoryBytes: 8 * gib, actors: 10 }, allocated: { cpuMillicores: 0, memoryBytes: 0, actors: 0 } },
    ],
    actors: [
      { atespace: 'ate-env', name: 'gavel-run-1', state: 'running', run: '1', org: { org: 'a', orgName: 'Alpha', kind: 'agent' } },
      { atespace: 'ate-env', name: 'gavel-run-2', state: 'crashed', run: '2', org: { org: 'a', orgName: 'Alpha', kind: 'test' } },
      { atespace: 'ate-env', name: 'gavel-run-3', state: 'suspended', run: '3' },
      { atespace: 'ate-env', name: 'default-template-0', state: 'running' },
    ],
  };
  response = { status: 200, body: inventory };
  render(<SubstrateInventoryPanel />, { wrapper: queryTestWrapper() });

  expect(await screen.findByText('1 of 2 workers active')).toBeTruthy();
  expect(screen.getByText('3 of 20 actor slots used')).toBeTruthy();
  expect(screen.getByText('1.5 of 8 cores allocated')).toBeTruthy();
  expect(screen.getByText('2 of 16 GiB allocated')).toBeTruthy();
  expect(cells('Substrate workers')).toEqual([
    ['pool-a-0', 'pool-a', 'node-1', 'active', '1.5 / 4', '2 / 8', '3 / 10'],
    ['pool-a-1', 'pool-a', 'node-2', 'draining', '0 / 4', '0 / 8', '0 / 10'],
  ]);
  expect(cells('Substrate actors by owner')).toEqual([
    ['Alpha', '1', '0', '1', '0', '2'],
    ['A run gavel doesn\'t know', '0', '1', '0', '0', '1'],
    ['Not a gavel run, in ate-env', '1', '0', '0', '0', '1'],
  ]);
});

it('says why one list is missing and shows the other', async () => {
  response = { status: 200, body: { workers: [], workersError: 'list substrate\'s workers: PermissionDenied', actors: [] } };
  render(<SubstrateInventoryPanel />, { wrapper: queryTestWrapper() });

  expect((await screen.findByRole('alert')).textContent).toContain('PermissionDenied');
  expect(screen.getByText('No actors in any atespace.')).toBeTruthy();
});

it('says how to list substrate when serve runs without it', async () => {
  response = { status: 404, body: { error: 'substrate isn\'t configured: start gavel serve with --sandbox substrate and its --substrate-* flags' } };
  render(<SubstrateInventoryPanel />, { wrapper: queryTestWrapper() });

  expect((await screen.findByRole('alert')).textContent).toContain('--sandbox substrate');
});
