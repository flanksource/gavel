import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { TodoItem } from '../../types';
import { TriageCreationRecovery } from './TriageAfterCreation';

afterEach(() => vi.unstubAllGlobals());

describe('Triage creation recovery', () => {
  it('retries triage on the saved TODO instead of creating another', async () => {
    const fetch = vi.fn().mockResolvedValue({ ok: true, status: 200, json: async () => ({ status: 'started' }) });
    vi.stubGlobal('fetch', fetch);
    const onDone = vi.fn();
    render(<TriageCreationRecovery dir="/work/acme" result={{ todo: { ref: 'saved-todo' } as TodoItem, triage: { status: 'failed', error: 'Provider unavailable' } }} onDone={onDone} />);
    expect(screen.getByRole('link', { name: 'Open todo' }).getAttribute('href')).toBe('/todos/saved-todo');
    fireEvent.click(screen.getByRole('button', { name: 'Retry triage' }));
    await waitFor(() => expect(onDone).toHaveBeenCalledOnce());
    expect(fetch).toHaveBeenCalledOnce();
    const [url, request] = fetch.mock.calls[0];
    expect(url).toBe('/api/todos/run');
    expect(JSON.parse(request.body)).toEqual({ dir: '/work/acme', ref: 'saved-todo', step: 'triage.new' });
  });
});
