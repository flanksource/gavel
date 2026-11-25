import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { CommitIdentityPanel, type CommitIdentity } from './CommitIdentityPanel';

const githubDefault = { name: 'ada', email: '42+ada@users.noreply.github.com' };

function renderPanel() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <CommitIdentityPanel />
    </QueryClientProvider>,
  );
}

// stubServer answers /api/commit-identity from identity, or with saveResponse for a save, and
// records requests.
function stubServer(identity: CommitIdentity, saveResponse?: Response) {
  const requests: { method: string; url: string; body?: unknown }[] = [];
  vi.stubGlobal('fetch', vi.fn(async (url: string, init?: RequestInit) => {
    const method = init?.method ?? 'GET';
    const body = init?.body ? JSON.parse(String(init.body)) : undefined;
    requests.push({ method, url, body });
    if (method === 'PUT') {
      if (saveResponse) return saveResponse;
      Object.assign(identity, body);
    }
    return new Response(JSON.stringify(identity));
  }));
  return requests;
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('CommitIdentityPanel', () => {
  it('shows the GitHub identity a field left empty commits as', async () => {
    stubServer({ name: '', email: '', default: githubDefault });
    renderPanel();
    const name = await screen.findByLabelText('Name');
    await waitFor(() => expect(name.getAttribute('placeholder')).toBe('ada'));
    expect(screen.getByLabelText('Email').getAttribute('placeholder')).toBe('42+ada@users.noreply.github.com');
    expect(screen.getByText('ada <42+ada@users.noreply.github.com>')).toBeTruthy();
  });

  it('saves the name and email the user sets', async () => {
    const requests = stubServer({ name: '', email: '', default: githubDefault });
    renderPanel();
    await waitFor(() => expect(screen.getByLabelText('Name').getAttribute('placeholder')).toBe('ada'));
    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Ada Lovelace' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save commit identity' }));
    expect(await screen.findByText('Ada Lovelace <42+ada@users.noreply.github.com>')).toBeTruthy();
    expect(requests).toContainEqual({ method: 'PUT', url: '/api/commit-identity', body: { name: 'Ada Lovelace', email: '' } });
  });

  it('shows why the server refused an identity', async () => {
    stubServer(
      { name: '', email: '', default: githubDefault },
      new Response(JSON.stringify({ error: 'invalid commit identity: the email must be one address, like ada@example.com' }), { status: 400 }),
    );
    renderPanel();
    await waitFor(() => expect(screen.getByLabelText('Name').getAttribute('placeholder')).toBe('ada'));
    fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'ada' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save commit identity' }));
    expect(await screen.findByText(/the email must be one address/)).toBeTruthy();
  });
});
