import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { SSHKeysPanel } from './SSHKeysPanel';

const laptop = { id: 'k1', name: 'laptop', fingerprint: 'SHA256:abc', createdAt: '2026-10-14T10:00:00Z', lastUsedAt: null };

function renderPanel() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <SSHKeysPanel />
    </QueryClientProvider>,
  );
}

// stubServer answers the /api/ssh-keys routes from an in-memory list and records requests.
function stubServer(keys: typeof laptop[], addResponse?: Response) {
  const requests: { method: string; url: string; body?: unknown }[] = [];
  vi.stubGlobal('fetch', vi.fn(async (url: string, init?: RequestInit) => {
    const method = init?.method ?? 'GET';
    requests.push({ method, url, body: init?.body ? JSON.parse(String(init.body)) : undefined });
    if (method === 'GET') return new Response(JSON.stringify(keys));
    if (method === 'POST') {
      if (addResponse) return addResponse;
      keys.push({ ...laptop, id: 'k2', name: 'desktop', fingerprint: 'SHA256:def' });
      return new Response(JSON.stringify(keys[keys.length - 1]), { status: 201 });
    }
    keys.splice(keys.findIndex(key => url.endsWith(key.id)), 1);
    return new Response(null, { status: 204 });
  }));
  return requests;
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('SSHKeysPanel', () => {
  it('lists the user\'s keys', async () => {
    stubServer([laptop]);
    renderPanel();
    expect(await screen.findByText('laptop')).toBeTruthy();
    expect(screen.getByText('SHA256:abc')).toBeTruthy();
    expect(screen.getByText('Never used')).toBeTruthy();
  });

  it('adds a key and shows it', async () => {
    const requests = stubServer([laptop]);
    renderPanel();
    await screen.findByText('laptop');
    fireEvent.change(screen.getByLabelText('Key name'), { target: { value: 'desktop' } });
    fireEvent.change(screen.getByLabelText('Public key'), { target: { value: 'ssh-ed25519 AAAA desktop' } });
    fireEvent.click(screen.getByRole('button', { name: 'Add SSH key' }));
    expect(await screen.findByText('desktop')).toBeTruthy();
    expect(requests).toContainEqual({ method: 'POST', url: '/api/ssh-keys', body: { name: 'desktop', publicKey: 'ssh-ed25519 AAAA desktop' } });
  });

  it('shows why the server refused a key', async () => {
    stubServer([], new Response(JSON.stringify({ error: 'unsupported SSH key: ssh-rsa keys must be at least 3072 bits, this one is 2048' }), { status: 400 }));
    renderPanel();
    fireEvent.change(await screen.findByLabelText('Public key'), { target: { value: 'ssh-rsa AAAA' } });
    fireEvent.click(screen.getByRole('button', { name: 'Add SSH key' }));
    expect(await screen.findByText(/at least 3072 bits/)).toBeTruthy();
  });

  it('deletes a key', async () => {
    const requests = stubServer([laptop]);
    renderPanel();
    fireEvent.click(await screen.findByRole('button', { name: 'Delete SSH key laptop' }));
    await waitFor(() => expect(screen.queryByText('laptop')).toBeNull());
    expect(requests).toContainEqual({ method: 'DELETE', url: '/api/ssh-keys/k1', body: undefined });
  });
});
