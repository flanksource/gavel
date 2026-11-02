import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { ModelKeysPanel, type ModelKeySummary } from './ModelKeysPanel';

const kinds = [
  { name: 'anthropic-api-key', provider: 'anthropic', label: 'Anthropic API key' },
  { name: 'claude-code-oauth-token', provider: 'anthropic', label: 'Claude Code OAuth token' },
];
const oauthKey: ModelKeySummary = {
  provider: 'anthropic', kind: 'claude-code-oauth-token', hint: 'WXYZ', updatedAt: '2026-10-29T10:00:00Z', updatedBy: 'u1',
};

function renderPanel() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <ModelKeysPanel />
    </QueryClientProvider>,
  );
}

// stubServer answers the /api/model-keys routes from an in-memory list and records requests.
function stubServer(keys: ModelKeySummary[], canEdit: boolean, setResponse?: Response) {
  const requests: { method: string; url: string; body?: unknown }[] = [];
  vi.stubGlobal('fetch', vi.fn(async (url: string, init?: RequestInit) => {
    const method = init?.method ?? 'GET';
    const body = init?.body ? JSON.parse(String(init.body)) : undefined;
    requests.push({ method, url, body });
    if (method === 'GET') return new Response(JSON.stringify({ kinds, keys, canEdit }));
    if (method === 'PUT') {
      if (setResponse) return setResponse;
      const saved = { ...oauthKey, kind: body.kind, hint: body.key.slice(-4) };
      keys.splice(0, keys.length, saved);
      return new Response(JSON.stringify(saved));
    }
    keys.splice(0, keys.length);
    return new Response(null, { status: 204 });
  }));
  return requests;
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('ModelKeysPanel', () => {
  it('shows each provider\'s key without its value', async () => {
    stubServer([oauthKey], false);
    renderPanel();
    expect(await screen.findByText('Anthropic')).toBeTruthy();
    expect(screen.getByText('Claude Code OAuth token ending in WXYZ')).toBeTruthy();
  });

  it('tells members only an admin can change keys', async () => {
    stubServer([], false);
    renderPanel();
    expect(await screen.findByText('Not set')).toBeTruthy();
    expect(screen.getByText(/Only an organization admin can change model keys/)).toBeTruthy();
    expect(screen.queryByLabelText('Anthropic key')).toBeNull();
    expect(screen.queryByRole('button', { name: 'Save Anthropic key' })).toBeNull();
  });

  it('lets an admin set a key of the chosen kind', async () => {
    const requests = stubServer([], true);
    renderPanel();
    fireEvent.change(await screen.findByLabelText('Anthropic key kind'), { target: { value: 'anthropic-api-key' } });
    fireEvent.change(screen.getByLabelText('Anthropic key'), { target: { value: 'sk-ant-api03-new-key-ABCD' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save Anthropic key' }));
    expect(await screen.findByText('Anthropic API key ending in ABCD')).toBeTruthy();
    expect(requests).toContainEqual({
      method: 'PUT', url: '/api/model-keys/anthropic', body: { kind: 'anthropic-api-key', key: 'sk-ant-api03-new-key-ABCD' },
    });
    expect((screen.getByLabelText('Anthropic key') as HTMLInputElement).value).toBe('');
    expect((screen.getByLabelText('Anthropic key') as HTMLInputElement).type).toBe('password');
  });

  it('shows why the server refused a key', async () => {
    stubServer([], true, new Response(JSON.stringify({ error: 'invalid model key: a Claude Code OAuth token starts with "sk-ant-oat"' }), { status: 400 }));
    renderPanel();
    fireEvent.change(await screen.findByLabelText('Anthropic key'), { target: { value: 'sk-ant-api03-x' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save Anthropic key' }));
    expect(await screen.findByText(/starts with "sk-ant-oat"/)).toBeTruthy();
  });

  it('lets an admin remove a key', async () => {
    const requests = stubServer([oauthKey], true);
    renderPanel();
    fireEvent.click(await screen.findByRole('button', { name: 'Remove Anthropic key' }));
    await waitFor(() => expect(screen.getByText('Not set')).toBeTruthy());
    expect(requests).toContainEqual({ method: 'DELETE', url: '/api/model-keys/anthropic', body: undefined });
  });
});
