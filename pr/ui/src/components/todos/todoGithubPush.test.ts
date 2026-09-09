import { afterEach, describe, expect, it, vi } from 'vitest';
import { TodoMutationError, todoMutationJSON } from './todoMutations';
import type { GithubPushRequest, GithubPushResponse } from './todoMutations';

afterEach(() => {
  vi.unstubAllGlobals();
});

function pushRequest(dir: string, { ref, force, update, issue }: GithubPushRequest) {
  return todoMutationJSON<GithubPushResponse>(
    '/api/todos/github',
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ref, dir, force, update, issue }),
    },
    update || issue ? `Failed to update todo ${ref} on GitHub` : `Failed to push todo ${ref} to GitHub`,
  );
}

function jsonResponse(body: unknown) {
  return { ok: true, status: 200, json: async () => body };
}

describe('github push request', () => {
  it('posts the ref and workspace and returns the created issue', async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({
      todo: { ref: '3f2a1b' },
      repo: 'acme/api',
      number: 11,
      url: 'https://github.com/acme/api/issues/11',
      alias: 'acme/api#11',
      updated: false,
    }));
    vi.stubGlobal('fetch', fetchMock);

    const response = await pushRequest('/workspace/api', { ref: '3f2a1b' });

    expect(fetchMock).toHaveBeenCalledOnce();
    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe('/api/todos/github');
    expect(init.method).toBe('POST');
    expect(JSON.parse(String(init.body))).toEqual({ ref: '3f2a1b', dir: '/workspace/api' });
    expect(response.url).toBe('https://github.com/acme/api/issues/11');
    expect(response.alias).toBe('acme/api#11');
    expect(response.updated).toBe(false);
  });

  it('re-pushes onto the linked issue when asked to update', async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({
      todo: { ref: '3f2a1b' },
      repo: 'acme/api',
      number: 3,
      url: 'https://github.com/acme/api/issues/3',
      alias: 'acme/api#3',
      updated: true,
    }));
    vi.stubGlobal('fetch', fetchMock);

    const response = await pushRequest('/workspace/api', { ref: '3f2a1b', update: true });

    const [, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(JSON.parse(String(init.body))).toEqual({ ref: '3f2a1b', dir: '/workspace/api', update: true });
    expect(response.number).toBe(3);
    expect(response.updated).toBe(true);
  });

  it('sends an explicitly targeted issue reference', async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({
      todo: { ref: '3f2a1b' },
      repo: 'acme/api',
      number: 412,
      url: 'https://github.com/acme/api/issues/412',
      alias: 'acme/api#412',
      updated: true,
    }));
    vi.stubGlobal('fetch', fetchMock);

    await pushRequest('/workspace/api', { ref: '3f2a1b', issue: 'acme/api#412' });

    const [, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(JSON.parse(String(init.body))).toEqual({ ref: '3f2a1b', dir: '/workspace/api', issue: 'acme/api#412' });
  });

  it('surfaces the already-linked conflict with its status so the caller can offer an update', async () => {
    const conflict = {
      ok: false,
      status: 409,
      clone: () => conflict,
      json: async () => ({
        error: 'todo is already linked to a GitHub issue: acme/api#3 '
          + '(re-run with --update to rewrite it, or --force to open a second issue)',
      }),
      text: async () => '',
    };
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(conflict));

    const failure = await pushRequest('/workspace/api', { ref: '3f2a1b' }).catch((err: unknown) => err);

    expect(failure).toBeInstanceOf(TodoMutationError);
    expect((failure as TodoMutationError).status).toBe(409);
    expect((failure as TodoMutationError).message).toMatch(/already linked to a GitHub issue: acme\/api#3/);
  });
});
