// Tests the organization's sandbox size panel: the size its substrate environments get, the
// default it falls back to, and a platform admin setting or resetting it.
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, expect, it, vi } from 'vitest';
import { queryTestWrapper } from '../todos/queryTestWrapper';
import { SandboxSizePanel, type PlatformSandboxSize } from './SandboxSizePanel';

const url = '/api/platform/orgs/o-alpha/sandbox-size';
const defaults = { cpuMillicores: 1000, memoryBytes: 4 * 2 ** 30 };

// stubServer answers the sandbox size route from state, recording each request, or with refusal
// for a PUT when it is set.
function stubServer(state: PlatformSandboxSize, refusal?: { status: number; error: string }) {
  const requests: { method: string; body?: unknown }[] = [];
  vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    if (String(input) !== url) throw new Error(`unexpected fetch ${String(input)}`);
    const method = init?.method ?? 'GET';
    const body = init?.body ? JSON.parse(String(init.body)) : undefined;
    requests.push({ method, body });
    if (method === 'PUT' && refusal) return new Response(JSON.stringify({ error: refusal.error }), { status: refusal.status });
    if (method === 'PUT') state.size = body;
    if (method === 'DELETE') state.size = null;
    return new Response(JSON.stringify(state));
  }));
  return requests;
}

afterEach(() => {
  vi.unstubAllGlobals();
});

it('shows the default an organization without its own size gets', async () => {
  stubServer({ size: null, default: defaults });
  render(<SandboxSizePanel org="o-alpha" />, { wrapper: queryTestWrapper() });

  expect(await screen.findByText('1 CPU · 4 GiB per environment, the default')).toBeTruthy();
  expect(screen.queryByRole('button', { name: 'Use the default' }), 'there is no own size to reset').toBeNull();
});

it('sets the organization\'s own size in cores and GiB', async () => {
  const requests = stubServer({ size: null, default: defaults });
  render(<SandboxSizePanel org="o-alpha" />, { wrapper: queryTestWrapper() });
  await screen.findByText('1 CPU · 4 GiB per environment, the default');

  fireEvent.change(screen.getByLabelText('CPU (cores)'), { target: { value: '2.5' } });
  fireEvent.change(screen.getByLabelText('Memory (GiB)'), { target: { value: '8' } });
  fireEvent.click(screen.getByRole('button', { name: 'Save size' }));

  expect(await screen.findByText('2.5 CPU · 8 GiB per environment, set for this organization')).toBeTruthy();
  expect(requests).toContainEqual({ method: 'PUT', body: { cpuMillicores: 2500, memoryBytes: 8 * 2 ** 30 } });
});

it('returns the organization to the default', async () => {
  const requests = stubServer({ size: { cpuMillicores: 4000, memoryBytes: 16 * 2 ** 30 }, default: defaults });
  render(<SandboxSizePanel org="o-alpha" />, { wrapper: queryTestWrapper() });
  expect(await screen.findByText('4 CPU · 16 GiB per environment, set for this organization')).toBeTruthy();
  expect(screen.getByText('The default is 1 CPU · 4 GiB.')).toBeTruthy();

  fireEvent.click(screen.getByRole('button', { name: 'Use the default' }));
  expect(await screen.findByText('1 CPU · 4 GiB per environment, the default')).toBeTruthy();
  expect(requests.map(r => r.method)).toContain('DELETE');
});

it('says why a size wasn\'t saved', async () => {
  stubServer({ size: null, default: defaults }, { status: 400, error: "an organization's sandbox size needs both CPU and memory" });
  render(<SandboxSizePanel org="o-alpha" />, { wrapper: queryTestWrapper() });
  await screen.findByText('1 CPU · 4 GiB per environment, the default');

  fireEvent.change(screen.getByLabelText('CPU (cores)'), { target: { value: '2' } });
  fireEvent.change(screen.getByLabelText('Memory (GiB)'), { target: { value: '4' } });
  fireEvent.click(screen.getByRole('button', { name: 'Save size' }));

  await waitFor(() => expect(screen.getByRole('alert').textContent).toContain("an organization's sandbox size needs both CPU and memory"));
});

it('says when the default can\'t be read', async () => {
  stubServer({ size: null, defaultError: 'read actor template ate-env/gavel-runner: unavailable' });
  render(<SandboxSizePanel org="o-alpha" />, { wrapper: queryTestWrapper() });

  expect(await screen.findByText(/the default: read actor template ate-env\/gavel-runner: unavailable/)).toBeTruthy();
});

it('says when serve has no substrate to size', async () => {
  vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({ error: 'serve runs without --sandbox substrate' }), { status: 404 })));
  render(<SandboxSizePanel org="o-alpha" />, { wrapper: queryTestWrapper() });

  expect((await screen.findByRole('alert')).textContent).toContain('serve runs without --sandbox substrate');
});
