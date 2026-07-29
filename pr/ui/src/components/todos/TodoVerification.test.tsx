import type React from 'react';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { FixtureEditorProps } from '@flanksource/clicky-ui/data';
import type { TodoItem } from '../../types';
import { TodoVerification } from './TodoVerification';

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
  vi.stubGlobal(
    'fetch',
    vi.fn(async (input: RequestInfo | URL) => {
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
    }),
  );
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
      />,
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
      />,
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

  it('saves a dirty fixture before running the shared verification endpoint', async () => {
    const updatedTodo = { ...todo, verificationMarkdown: '### command: smoke' };
    const onChanged = vi.fn();
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

    render(<TodoVerification dir="/workspace" todo={todo} onChanged={onChanged} />);
    await waitFor(() => expect(fixtureEditorCalls.props.at(-1)).toBeDefined());
    act(() => fixtureEditorCalls.props.at(-1)?.onChange('### command: smoke'));
    fireEvent.click(screen.getByRole('button', { name: /^run verification \(/i }));

    await waitFor(() => expect(screen.getByText('Verification passed')).toBeTruthy());
    expect(screen.getByText('smoke')).toBeTruthy();
    expect(screen.getByText('echo ok')).toBeTruthy();
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
  });
});
