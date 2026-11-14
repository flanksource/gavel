import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { RebaseContext, RebaseOperation } from '../types/branchRebase';
import { ProjectRebase } from './ProjectRebase';

const prompt = vi.hoisted(() => ({ start: vi.fn(), success: undefined as undefined | (() => void) }));
vi.mock('./promptRun', () => ({
  usePromptRun: ({ prompt: name, onSuccess }: { prompt: string; onSuccess?: () => void }) => {
    expect(name).toBe('branch-rebase');
    prompt.success = onSuccess;
    return { state: { status: 'idle' }, active: false, start: prompt.start, stop: vi.fn() };
  },
}));
vi.mock('@flanksource/clicky-ui/ai', () => ({ SessionTail: () => null }));

const WORKTREE = '/src/acme/.worktrees/topic';
const operation: RebaseOperation = { id: 'op-123', worktree: WORKTREE, branch: 'topic', target: 'refs/heads/main', targetSha: 'target', originalHead: 'before', stashSha: 'saved', phase: 'rebase', error: 'Conflicting rebase', conflicts: ['shared file.txt'], aborted: false };
const context: RebaseContext = { executable: 'gavel', branch: 'topic', head: 'before', dirtyCount: 126, targets: [{ ref: 'refs/remotes/origin/main', label: 'origin/main' }, { ref: 'refs/heads/main', label: 'main' }], operation: null };
const json = (value: unknown, status = 200) => new Response(JSON.stringify(value), { status, headers: { 'Content-Type': 'application/json' } });

function setup(value = context, fail = false) {
  let live = value;
  const fetch = vi.fn(async (_url: string, init?: RequestInit) => {
    if (init?.method === 'POST') {
      if (fail) return json({ error: operation.error, operation }, 409);
      return json({ ...operation, phase: 'done' });
    }
    return json(live);
  });
  vi.stubGlobal('fetch', fetch);
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  render(<QueryClientProvider client={client}><ProjectRebase projectName="acme" worktree={WORKTREE} branch={value.branch} /></QueryClientProvider>);
  return { fetch, client, setContext: (next: RebaseContext) => { live = next; } };
}

afterEach(() => { cleanup(); vi.unstubAllGlobals(); vi.clearAllMocks(); });

describe('Project rebase', () => {
  it('previews the source and local edits and defaults to origin/main', async () => {
    const { fetch } = setup();
    await waitFor(() => expect(fetch).toHaveBeenCalled());
    fireEvent.click(screen.getByRole('button', { name: 'Rebase…' }));
    expect(await screen.findByRole('combobox', { name: 'Rebase from' })).toBeTruthy();
    expect(screen.getByText(/126 uncommitted files/)).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Rebase' }));
    await waitFor(() => expect(fetch.mock.calls.some(([, init]) => init?.method === 'POST')).toBe(true));
    const [, init] = fetch.mock.calls.find(([, init]) => init?.method === 'POST')!;
    expect(JSON.parse(String(init?.body))).toEqual({ worktree: WORKTREE, branch: 'topic', target: 'refs/remotes/origin/main', expectedHead: 'before' });
  });

  it('requires an explicit click before handing a failure to the new rebase prompt', async () => {
    setup({ ...context, branch: '', operation });
    const send = await screen.findByRole('button', { name: 'Send to rebase prompt' });
    expect(prompt.start).not.toHaveBeenCalled();
    fireEvent.click(send);
    expect(prompt.start).toHaveBeenCalledWith(expect.objectContaining({ operation: operation.id, worktree: WORKTREE, targetSha: 'target', stashSha: 'saved' }));
    expect(screen.getByText('shared file.txt')).toBeTruthy();
  });

  it('clears the failed request after successful AI recovery instead of keeping a stale operation', async () => {
    const { client, setContext } = setup(context, true);
    await waitFor(() => expect(client.getQueryData(['projects', 'acme', 'rebase', WORKTREE, undefined])).toBeTruthy());
    fireEvent.click(screen.getByRole('button', { name: 'Rebase…' }));
    fireEvent.click(await screen.findByRole('button', { name: 'Rebase' }));
    await screen.findByRole('button', { name: 'Send to rebase prompt' });
    setContext(context);
    prompt.success?.();
    await client.invalidateQueries();
    await waitFor(() => expect(screen.queryByRole('button', { name: 'Send to rebase prompt' })).toBeNull());
    expect((await screen.findByRole('status')).textContent).toBe('Rebase complete; saved edits restored.');
  });
});
