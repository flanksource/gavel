import type React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { FixtureEditorProps } from '@flanksource/clicky-ui/data';
import type { TodoItem, TodoSessionDetailResponse } from '../../types';
import { TodoVerification } from './TodoVerification';
import { queryTestWrapper } from './queryTestWrapper';
import { todoQueryKeys } from './todoQueries';
import { workspaceTodoBatchKeys } from './workspaceTodoQueries';

const fixtureEditorCalls = vi.hoisted(() => ({
  props: [] as FixtureEditorProps[],
}));

vi.mock('@flanksource/clicky-ui/data', () => ({
  FixtureEditor: (props: FixtureEditorProps) => {
    fixtureEditorCalls.props.push(props);
    return <div data-testid="fixture-editor" />;
  },
}));

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({
    children,
    loading: _loading,
    ...props
  }: React.ButtonHTMLAttributes<HTMLButtonElement> & { loading?: boolean }) => (
    // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for the Clicky Button itself.
    <button type="button" {...props}>
      {children}
    </button>
  ),
  DropdownMenu: ({
    trigger,
    children,
  }: {
    trigger: React.ReactNode;
    children: (close: () => void) => React.ReactNode;
  }) => (
    <div>
      {trigger}
      {children(() => {})}
    </div>
  ),
  Field: ({ children }: { children?: React.ReactNode }) => <div>{children}</div>,
  Modal: ({ children, open }: { children?: React.ReactNode; open?: boolean }) => (open === false ? null : <div>{children}</div>),
  SegmentedControl: () => null,
}));

vi.mock('@flanksource/clicky-ui/chat', () => ({
  ModelSelector: () => null,
  providerIcon: () => (props: React.SVGProps<SVGSVGElement>) => <svg {...props} />,
}));

vi.mock('@flanksource/clicky-ui/ai', () => ({
  effortOptionsForModel: (_model: unknown, fallback: string[]) => fallback,
  promptRuntimeValueToPayload: (value: unknown) => ({ spec: value }),
  reconcileModelCapabilities: (value: unknown) => value,
  RuntimeBar: ({ ariaLabel }: { ariaLabel?: string }) => <button type="button" aria-label={ariaLabel}>Runtime</button>,
  SpecRuntimeEditor: () => <div>Verification runtime editor</div>,
}));

vi.mock('@flanksource/clicky-ui/icons', async (importOriginal) => ({
  ...(await importOriginal<object>()),
  UiBeaker: (props: React.SVGProps<SVGSVGElement>) => <svg {...props} />,
  UiCopy: (props: React.SVGProps<SVGSVGElement>) => <svg {...props} />,
}));

vi.mock('./AcceptanceCriteria', () => ({
  AcceptanceCriteria: () => <div data-testid="acceptance-criteria" />,
}));

vi.mock('./TodoVerificationAttempts', () => ({
  TodoVerificationAttempts: ({ detail }: { detail: TodoSessionDetailResponse | null }) => (
    <div data-testid="verification-attempts">{detail ? `${detail.attempts.length} attempts` : 'loading'}</div>
  ),
}));

const todo: TodoItem = {
  ref: 'todo-1',
  title: 'Fix fixture editor',
  status: 'pending',
  priority: 'medium',
  verificationMarkdown: '',
};

const testSchema = { type: 'object' as const, properties: { paths: { type: 'array' } } };
const lintSchema = { type: 'object' as const, properties: { files: { type: 'array' } } };
const verificationRunContext = {
  defaultBackend: 'codex-agent',
  efforts: ['low', 'medium', 'high'],
  tools: [],
  backends: [{
    id: 'codex-agent',
    label: 'Codex Agent',
    provider: 'openai',
    agent: 'codex',
    defaultModel: 'gpt-5.6-sol',
    driver: 'codex-headless',
    mechanisms: [{ value: 'agent', label: 'Agent', driver: 'codex-headless' }],
    models: [{ id: 'gpt-5.6-sol', provider: 'openai', label: 'GPT-5.6 Sol', reasoning: true, configured: true }],
    configured: true,
  }],
};

function mockSchemaFetch() {
  const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
    if (String(input) === '/api/todos/run/context') {
      return {
        ok: true,
        json: async () => verificationRunContext,
        text: async (): Promise<string> => '',
      };
    }
    return {
      ok: true,
      json: async () => ({
        fences: {
          test: { schema: testSchema, aliases: ['yaml test'] },
          lint: { schema: lintSchema, aliases: ['yaml lint'] },
        },
      }),
      text: async (): Promise<string> => '',
    };
  });
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
}

describe('TodoVerification', () => {
  beforeEach(() => {
    fixtureEditorCalls.props.length = 0;
    const store: Record<string, string> = {};
    vi.stubGlobal('localStorage', {
      getItem: vi.fn((key: string) => store[key] ?? null),
      setItem: vi.fn((key: string, value: string) => {
        store[key] = String(value);
      }),
      removeItem: vi.fn((key: string) => {
        delete store[key];
      }),
      clear: vi.fn(() => {
        for (const key of Object.keys(store)) delete store[key];
      }),
    });
    mockSchemaFetch();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('passes the Gavel fixture fence types to the editor menu', async () => {
    render(
      <TodoVerification
        dir="/workspace"
        todo={todo}
        onChanged={() => {}}
        attempts={null}
      />,
      { wrapper: queryTestWrapper() },
    );

    await waitFor(() => {
      expect(fixtureEditorCalls.props.at(-1)?.allowedFences).toEqual([
        { info: 'yaml test', label: 'test', description: 'Gavel test options' },
        { info: 'yaml lint', label: 'lint', description: 'Gavel lint options' },
        { info: 'ai', label: 'ai', description: 'Reviewer instructions' },
        { info: 'exec', label: 'exec', description: 'Shell command or script' },
        { info: 'bash', label: 'bash', description: 'Bash command block' },
      ]);
    });
  });

  it('passes fixture schemas from the schema endpoint to the editor', async () => {
    render(
      <TodoVerification
        dir="/workspace"
        todo={todo}
        onChanged={() => {}}
        attempts={null}
      />,
      { wrapper: queryTestWrapper() },
    );

    await waitFor(() => {
      expect(fixtureEditorCalls.props.at(-1)?.schemas).toEqual({
        test: testSchema,
        'yaml test': testSchema,
        lint: lintSchema,
        'yaml lint': lintSchema,
      });
    });
  });

  it('reuses the cached verification schema after remount', async () => {
    const fetchMock = mockSchemaFetch();
    const wrapper = queryTestWrapper();
    const first = render(
      <TodoVerification dir="/workspace" todo={todo} onChanged={() => {}} attempts={null} />,
      { wrapper },
    );
    await waitFor(() => expect(fixtureEditorCalls.props.at(-1)?.schemas?.test).toEqual(testSchema));
    first.unmount();

    render(
      <TodoVerification dir="/workspace" todo={todo} onChanged={() => {}} attempts={null} />,
      { wrapper },
    );
    await waitFor(() => expect(fixtureEditorCalls.props.at(-1)?.schemas?.test).toEqual(testSchema));

    expect(fetchMock.mock.calls.filter(([input]) => String(input) === '/api/todos/verification/schema')).toHaveLength(1);
  });

  it('saves a dirty fixture before running the shared verification endpoint', async () => {
    const updatedTodo = { ...todo, verificationMarkdown: '### command: smoke' };
    const onChanged = vi.fn();
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const invalidateQueries = vi.spyOn(client, 'invalidateQueries');
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      if (url === '/api/todos/verification/schema') {
        return { ok: true, json: async () => ({ fences: {} }), text: async (): Promise<string> => '' };
      }
      if (url === '/api/todos/run/context') {
        return { ok: true, json: async () => verificationRunContext, text: async (): Promise<string> => '' };
      }
      if (url.startsWith('/api/todos/verification/fixture')) {
        expect(init?.method).toBe('POST');
        return { ok: true, json: async () => updatedTodo, text: async (): Promise<string> => '' };
      }
      if (url.startsWith('/api/todos/verification/run')) {
        expect(init?.method).toBe('POST');
        return {
          ok: true,
          json: async () => ({
            verification: {
              allPassed: true,
              duration: 10,
              output: {
                results: [{ name: 'smoke', status: 'PASS', command: 'echo ok', stdout: 'ok\n' }],
              },
            },
            todo: { ...updatedTodo, status: 'verified' },
          }),
          text: async (): Promise<string> => '',
        };
      }
      throw new Error(`unexpected fetch ${url}`);
    });
    vi.stubGlobal('fetch', fetchMock);
    localStorage.setItem(
      'gavel.pr-ui.promptRunChoices.v2',
      JSON.stringify({
        verification: {
          last: { spec: { backend: 'codex-agent', model: 'gpt-5.6-sol', effort: 'high' } },
          recent: [],
        },
      }),
    );

    const attempts: TodoSessionDetailResponse = {
      attempts: [
        {
          promptRunId: 'run-1',
          ordinal: 1,
          step: 'verify',
          requested: {},
          resolved: {},
          status: 'succeeded',
          processActive: false,
          state: 'succeeded',
          phase: 'done',
          queuedAt: '2026-07-30T09:00:00Z',
          admissionSessionId: 'admission-1',
          createdAt: '2026-07-30T09:00:00Z',
          updatedAt: '2026-07-30T09:01:00Z',
          resultJson: { definitionOfDone: { ran: true, passed: true } },
        },
      ],
      diagnostics: [],
    };
    render(<TodoVerification dir="/workspace" todo={todo} onChanged={onChanged} attempts={attempts} />, {
      wrapper: ({ children }) => <QueryClientProvider client={client}>{children}</QueryClientProvider>,
    });
    await waitFor(() => expect(fixtureEditorCalls.props.at(-1)).toBeDefined());
    act(() => fixtureEditorCalls.props.at(-1)?.onChange('### command: smoke'));
    fireEvent.click(screen.getByRole('button', { name: /^run verification$/i }));

    await waitFor(() => expect(onChanged).toHaveBeenCalledTimes(2));
    // Results live in the persisted attempt list, not in this view's state, so a
    // run neither renders its own copy nor clears what is already listed.
    expect(screen.queryByText('Verification passed')).toBeNull();
    expect(screen.getByTestId('verification-attempts').textContent).toBe('1 attempts');
    const mutationURLs = fetchMock.mock.calls
      .map(call => String(call[0]))
      .filter(url => url.includes('/api/todos/verification/') && !url.endsWith('/schema'));
    expect(mutationURLs[0]).toContain('/api/todos/verification/fixture');
    expect(mutationURLs[1]).toContain('/api/todos/verification/run');
    const runCall = fetchMock.mock.calls.find(call => String(call[0]).includes('/api/todos/verification/run'));
    expect(JSON.parse(String(runCall?.[1]?.body))).toMatchObject({
      ref: 'todo-1',
      spec: { model: 'gpt-5.6-sol', effort: 'high' },
    });
    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Advanced' }));
    });
    expect(screen.getByText('Verification runtime editor')).toBeTruthy();
    expect(onChanged).toHaveBeenCalledTimes(2);
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: workspaceTodoBatchKeys.all });
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: todoQueryKeys.sessionDetail('/workspace', 'todo-1', undefined, true),
    });
    expect(client.getQueryState(todoQueryKeys.verificationSchema())?.isInvalidated).toBe(false);
  });

  // A pre-execution failure creates no prompt run, so it can never reach the
  // attempt list — the banner is the only place it can be seen.
  it('surfaces a pre-execution verification error in the banner', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url === '/api/todos/verification/schema') {
        return { ok: true, json: async () => ({ fences: {} }), text: async (): Promise<string> => '' };
      }
      if (url === '/api/todos/run/context') {
        return { ok: true, json: async () => verificationRunContext, text: async (): Promise<string> => '' };
      }
      if (url.startsWith('/api/todos/verification/run')) {
        return {
          ok: true,
          json: async () => ({
            verification: { allPassed: false, duration: 0, error: 'no verification fixture, acceptance criteria, or configured checks' },
            todo,
          }),
          text: async (): Promise<string> => '',
        };
      }
      throw new Error(`unexpected fetch ${url}`);
    });
    vi.stubGlobal('fetch', fetchMock);

    render(<TodoVerification dir="/workspace" todo={todo} onChanged={() => {}} attempts={{ attempts: [], diagnostics: [] }} />, {
      wrapper: queryTestWrapper(),
    });
    await waitFor(() => expect(fixtureEditorCalls.props.at(-1)).toBeDefined());
    fireEvent.click(screen.getByRole('button', { name: /^run verification$/i }));

    await waitFor(() =>
      expect(screen.getByText('no verification fixture, acceptance criteria, or configured checks')).toBeTruthy()
    );
  });

  it('surfaces verification request failures with action context', async () => {
    vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url === '/api/todos/verification/schema') {
        return { ok: true, json: async () => ({ fences: {} }), text: async (): Promise<string> => '' };
      }
      if (url === '/api/todos/run/context') {
        return { ok: true, json: async () => verificationRunContext, text: async (): Promise<string> => '' };
      }
      if (url.startsWith('/api/todos/verification/run')) {
        return {
          ok: false,
          status: 409,
          json: async () => ({ error: 'verification is already running' }),
          text: async () => JSON.stringify({ error: 'verification is already running' }),
        };
      }
      throw new Error(`unexpected fetch ${url}`);
    }));
    render(<TodoVerification dir="/workspace" todo={todo} onChanged={() => {}} attempts={{ attempts: [], diagnostics: [] }} />, {
      wrapper: queryTestWrapper(),
    });
    await waitFor(() => expect(fixtureEditorCalls.props.at(-1)).toBeDefined());

    fireEvent.click(screen.getByRole('button', { name: /^run verification$/i }));

    expect(await screen.findByText(/verification run failed.*verification is already running/i)).toBeTruthy();
  });
});
