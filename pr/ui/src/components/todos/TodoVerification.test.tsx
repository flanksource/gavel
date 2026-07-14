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

function mockSchemaFetch() {
  vi.stubGlobal(
    'fetch',
    vi.fn(async () => ({
      ok: true,
      json: async () => ({
        fences: {
          test: { schema: testSchema, aliases: ['yaml test'] },
          lint: { schema: lintSchema, aliases: ['yaml lint'] },
        },
      }),
      text: async () => '',
    })),
  );
}

describe('TodoVerification', () => {
  beforeEach(() => {
    fixtureEditorCalls.props.length = 0;
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

    render(<TodoVerification dir="/workspace" todo={todo} onChanged={onChanged} />);
    await waitFor(() => expect(fixtureEditorCalls.props.at(-1)).toBeDefined());
    act(() => fixtureEditorCalls.props.at(-1)?.onChange('### command: smoke'));
    fireEvent.click(screen.getByRole('button', { name: /run verification/i }));

    await waitFor(() => expect(screen.getByText('Verification passed')).toBeTruthy());
    expect(screen.getByText('smoke')).toBeTruthy();
    expect(screen.getByText('echo ok')).toBeTruthy();
    const mutationURLs = fetchMock.mock.calls
      .map(call => String(call[0]))
      .filter(url => url.includes('/api/todos/verification/') && !url.endsWith('/schema'));
    expect(mutationURLs[0]).toContain('/api/todos/verification/fixture');
    expect(mutationURLs[1]).toContain('/api/todos/verification/run');
    expect(onChanged).toHaveBeenCalledTimes(2);
  });
});
