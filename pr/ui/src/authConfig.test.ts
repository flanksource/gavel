import { afterEach, describe, expect, it, vi } from 'vitest';
import { loadAuthConfig } from './authConfig';

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('loadAuthConfig', () => {
  it('reads the sign-in settings from the server', async () => {
    const config = { mode: 'clerk', publishableKey: 'pk_test_abc', issuer: 'https://clerk.example.com', clientId: 'cli' };
    const fetchMock = vi.fn(async () => new Response(JSON.stringify(config)));
    vi.stubGlobal('fetch', fetchMock);
    await expect(loadAuthConfig()).resolves.toEqual(config);
    expect(fetchMock).toHaveBeenCalledWith('/api/auth/config');
  });

  it('says why when the settings cannot be read', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response('down', { status: 503 })));
    await expect(loadAuthConfig()).rejects.toThrow('load the sign-in settings: HTTP 503');
  });

  it('rejects an unknown mode instead of guessing', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({ mode: 'saml' }))));
    await expect(loadAuthConfig()).rejects.toThrow('unknown sign-in mode "saml"');
  });
});
