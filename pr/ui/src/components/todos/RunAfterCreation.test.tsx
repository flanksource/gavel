import { useState } from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { TodoItem } from '../../types';
import { RunAfterCreation, RunCreationRecovery, useRunAfterCreation } from './RunAfterCreation';
import { readCreateStepChoice } from './runChoiceStorage';
import { Button } from '@flanksource/clicky-ui/components';
import { rememberTodoRunOptions } from './run';
import { CREATE_MODEL as MODEL, CREATE_RUNTIME_CONTEXT as context } from './createRuntimeTestHelpers';
const runtime = vi.hoisted(() => ({ loading: false, error: '', dirs: [] as string[] }));

vi.mock('./run', async importOriginal => ({
  ...await importOriginal<typeof import('./run')>(),
  useTodoRunContext: ({ dir }: { dir: string }) => {
    runtime.dirs.push(dir);
    return { context: runtime.loading ? null : runtime.error ? { ...context, modes: [] } : context, loading: runtime.loading, error: runtime.error };
  },
}));

vi.mock('./TodoRunActionButton', async () => ({ TodoRunRuntimeBar: (await import('./createRuntimeTestHelpers')).CreateRuntimeBarStub }));

const DIR = '/work/acme';
const STEPS_BY_STATUS: Record<string, { name: string; label: string; prompt: string; auxiliary: boolean }[]> = {
  draft: [
    { name: 'plan', label: 'Plan', prompt: 'plan', auxiliary: false },
    { name: 'triage.new', label: 'Triage new', prompt: 'triage-new', auxiliary: true },
  ],
  pending: [
    { name: 'run', label: 'Run', prompt: 'run', auxiliary: false },
    { name: 'plan', label: 'Plan', prompt: 'plan', auxiliary: false },
  ],
};

function stubSteps() {
  const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
    const url = new URL(String(input), 'http://gavel.test');
    expect(url.pathname).toBe('/api/todos/new/steps');
    expect(url.searchParams.get('dir')).toMatch(/^\/work\//);
    return { ok: true, status: 200, json: async () => ({ steps: STEPS_BY_STATUS[url.searchParams.get('status') ?? ''] ?? [] }) };
  });
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
}

function Harness({ initialStatus }: { initialStatus: string }) {
  const [status, setStatus] = useState(initialStatus);
  const [dir, setDir] = useState(DIR);
  const runAfter = useRunAfterCreation({ dir, status });
  return (
    <>
      <Button type="button" onClick={() => setStatus('pending')}>to pending</Button>
      <Button onClick={() => setDir('/work/other')}>Other workspace</Button>
      <RunAfterCreation state={runAfter} disabled={false} />
      <output aria-label="payload">{JSON.stringify({ title: 't', ...runAfter.fields })}</output>
      <Button disabled={runAfter.blocked}>Save</Button>
      <Button onClick={runAfter.capture}>Capture run</Button>
      <output aria-label="submitted">{JSON.stringify(runAfter.submitted)}</output>
    </>
  );
}

function renderHarness(initialStatus = 'draft') {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(<QueryClientProvider client={client}><Harness initialStatus={initialStatus} /></QueryClientProvider>);
}

const options = () => Array.from(screen.getByLabelText<HTMLSelectElement>('Run after creation').options).map(option => option.textContent);
const payload = () => JSON.parse(screen.getByLabelText('payload').textContent ?? '');

beforeEach(() => {
  localStorage.clear();
  runtime.loading = false;
  runtime.error = '';
  runtime.dirs = [];
});
afterEach(() => vi.unstubAllGlobals());

describe('RunAfterCreation', () => {
  it('lists None followed by the steps the endpoint returns for the status', async () => {
    const fetchMock = stubSteps();
    renderHarness('draft');
    await waitFor(() => expect(options()).toEqual(['None', 'Plan', 'Triage new']));
    expect(fetchMock).toHaveBeenCalledOnce();
    expect(String(fetchMock.mock.calls[0]![0])).toContain('status=draft');
  });

  it('defaults to None and sends no step until one is chosen', async () => {
    stubSteps();
    renderHarness('draft');
    await waitFor(() => expect(options()).toContain('Plan'));
    expect(screen.getByLabelText<HTMLSelectElement>('Run after creation').value).toBe('');
    expect(payload()).toEqual({ title: 't' });
    expect('triage' in payload()).toBe(false);
  });

  it('sends the chosen step, and no triage field, in the create payload', async () => {
    stubSteps();
    renderHarness('draft');
    await waitFor(() => expect(options()).toContain('Triage new'));
    fireEvent.change(screen.getByLabelText('Run after creation'), { target: { value: 'triage.new' } });
    expect(payload()).toEqual({ title: 't', step: 'triage.new', spec: {} });
  });

  it('refetches the steps when the status changes', async () => {
    const fetchMock = stubSteps();
    renderHarness('draft');
    await waitFor(() => expect(options()).toEqual(['None', 'Plan', 'Triage new']));
    fireEvent.click(screen.getByText('to pending'));
    await waitFor(() => expect(options()).toEqual(['None', 'Run', 'Plan']));
    expect(fetchMock.mock.calls.map(call => new URL(String(call[0]), 'http://gavel.test').searchParams.get('status'))).toEqual(['draft', 'pending']);
  });

  it('reads a remembered step the new status does not offer as None', async () => {
    stubSteps();
    renderHarness('draft');
    await waitFor(() => expect(options()).toContain('Triage new'));
    fireEvent.change(screen.getByLabelText('Run after creation'), { target: { value: 'triage.new' } });
    fireEvent.click(screen.getByText('to pending'));
    await waitFor(() => expect(options()).toEqual(['None', 'Run', 'Plan']));
    expect(payload()).toEqual({ title: 't' });
  });

  it('remembers the last choice for the next form and falls back to None when nothing was chosen', async () => {
    stubSteps();
    expect(readCreateStepChoice()).toBe('');
    const first = renderHarness('draft');
    await waitFor(() => expect(options()).toContain('Plan'));
    fireEvent.change(screen.getByLabelText('Run after creation'), { target: { value: 'plan' } });
    expect(readCreateStepChoice()).toBe('plan');
    first.unmount();

    renderHarness('draft');
    await waitFor(() => expect(screen.getByLabelText<HTMLSelectElement>('Run after creation').value).toBe('plan'));
  });

  it('shows a failed steps request instead of silently offering only None', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => ({ ok: false, status: 500, json: async () => ({ error: 'registry unavailable' }), clone() { return this; }, text: async () => '' })));
    renderHarness('draft');
    expect((await screen.findByRole('alert')).textContent).toContain('registry unavailable');
  });

  it('edits and remembers runtime settings per step and omits them for None', async () => {
    stubSteps();
    renderHarness();
    await waitFor(() => expect(options()).toContain('Plan'));
    expect(screen.queryByRole('button', { name: 'plan runtime' })).toBeNull();
    fireEvent.change(screen.getByLabelText('Run after creation'), { target: { value: 'plan' } });
    fireEvent.click(screen.getByRole('button', { name: 'plan runtime' }));
    const selected = { model: MODEL, mode: 'api', effort: 'high' };
    expect(payload()).toEqual({ title: 't', step: 'plan', spec: selected });
    fireEvent.change(screen.getByLabelText('Run after creation'), { target: { value: 'triage.new' } });
    expect(payload()).toEqual({ title: 't', step: 'triage.new', spec: {} });
    fireEvent.change(screen.getByLabelText('Run after creation'), { target: { value: 'plan' } });
    expect(payload().spec).toEqual(selected);
    fireEvent.change(screen.getByLabelText('Run after creation'), { target: { value: '' } });
    expect(payload()).toEqual({ title: 't' });
  });

  it('preserves remembered presets and advanced spec fields through runtime edits', async () => {
    rememberTodoRunOptions('plan', { step: 'plan', presets: ['reviewer'], spec: { budget: { timeout: '3m' } } });
    stubSteps();
    renderHarness();
    await waitFor(() => expect(options()).toContain('Plan'));
    fireEvent.change(screen.getByLabelText('Run after creation'), { target: { value: 'plan' } });
    fireEvent.click(screen.getByRole('button', { name: 'plan runtime' }));
    expect(payload()).toEqual({ title: 't', step: 'plan', presets: ['reviewer'], spec: { model: MODEL, mode: 'api', effort: 'high', budget: { timeout: '3m' } } });
  });

  it.each(['loading', 'error'] as const)('blocks a selected run when runtime context is %s but allows None', async state => {
    rememberTodoRunOptions('plan', { step: 'plan', spec: { model: MODEL, mode: 'api' } });
    runtime.loading = state === 'loading';
    runtime.error = state === 'error' ? 'Catalog unavailable' : '';
    stubSteps();
    renderHarness();
    await waitFor(() => expect(options()).toContain('Plan'));
    fireEvent.change(screen.getByLabelText('Run after creation'), { target: { value: 'plan' } });
    expect(screen.getByRole<HTMLButtonElement>('button', { name: 'Save' }).disabled).toBe(true);
    if (runtime.error) expect(screen.getByRole('alert').textContent).toContain(runtime.error);
    fireEvent.change(screen.getByLabelText('Run after creation'), { target: { value: '' } });
    expect(screen.getByRole<HTMLButtonElement>('button', { name: 'Save' }).disabled).toBe(false);
  });

  it('reloads the runtime catalog for the selected workspace', async () => {
    stubSteps();
    renderHarness();
    await waitFor(() => expect(options()).toContain('Plan'));
    fireEvent.change(screen.getByLabelText('Run after creation'), { target: { value: 'plan' } });
    fireEvent.click(screen.getByRole('button', { name: 'Other workspace' }));
    await waitFor(() => expect(runtime.dirs).toContain('/work/other'));
  });

  it('keeps the submitted runtime and workspace when the form selection changes', async () => {
    stubSteps();
    renderHarness();
    await waitFor(() => expect(options()).toContain('Plan'));
    fireEvent.change(screen.getByLabelText('Run after creation'), { target: { value: 'plan' } });
    fireEvent.click(screen.getByRole('button', { name: 'plan runtime' }));
    fireEvent.click(screen.getByRole('button', { name: 'Capture run' }));
    fireEvent.click(screen.getByRole('button', { name: 'Other workspace' }));
    fireEvent.change(screen.getByLabelText('Run after creation'), { target: { value: '' } });
    expect(JSON.parse(screen.getByLabelText('submitted').textContent ?? '')).toEqual({
      dir: DIR, step: 'plan', spec: { model: MODEL, mode: 'api', effort: 'high' },
    });
  });
});

describe('RunCreationRecovery', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('retries the chosen step on the saved todo instead of creating another', async () => {
    const fetchMock = vi.fn().mockResolvedValue({ ok: true, status: 200, json: async () => ({ status: 'started' }) });
    vi.stubGlobal('fetch', fetchMock);
    const onDone = vi.fn();
    const request = { dir: DIR, step: 'triage.new', spec: { model: MODEL, mode: 'api', effort: 'high' }, presets: ['reviewer'] };
    render(<RunCreationRecovery request={request} result={{ todo: { ref: 'saved-todo' } as TodoItem, run: { status: 'failed', error: 'Provider unavailable' } }} onDone={onDone} />);
    expect(screen.getByText(/Provider unavailable/)).toBeTruthy();
    expect(screen.getByRole('link', { name: 'Open todo' }).getAttribute('href')).toBe('/todos/saved-todo');
    fireEvent.click(screen.getByRole('button', { name: 'Retry triage.new' }));
    await waitFor(() => expect(onDone).toHaveBeenCalledOnce());
    expect(fetchMock).toHaveBeenCalledOnce();
    const [url, init] = fetchMock.mock.calls[0]!;
    expect(url).toBe('/api/todos/run');
    expect(JSON.parse(init.body)).toEqual({ ...request, ref: 'saved-todo' });
  });
});
