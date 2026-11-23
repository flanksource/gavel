import type { ReactElement, ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, expect, it, vi } from 'vitest';
import type { FieldControl } from '@flanksource/clicky-ui/components';
import { RuntimeBar, type PromptCatalogEntry } from '@flanksource/clicky-ui/ai';
import { LifecyclesView, stepPromptPicker, stepRuntimeBar } from './LifecyclesView';

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, loading: _loading, ...props }: { children: ReactNode; loading?: boolean }) => <button {...props}>{children}</button>,
  Tabs: () => null,
  JsonSchemaForm: () => <div>Lifecycle fields</div>,
}));

afterEach(() => vi.unstubAllGlobals());

it('requires a fresh preview after reordering and saves only the scoped override', async () => {
  const steps = ['plan', 'verify', 'run'].map(name => ({ name, prompt: `todos.${name}` }));
  const view = { path: '/work/acme/.gavel.yaml', revision: 'original', source: {}, inherited: { steps }, effective: { steps } };
  const calls: RequestInit[] = [];
  vi.stubGlobal('fetch', vi.fn(async (url: string, init?: RequestInit) => {
    if (url === '/api/settings/schema') return Response.json({ properties: { todos: { properties: { lifecycle: { type: 'object' } } } } });
    if (url.startsWith('/api/settings/prompts/catalog')) return Response.json([]);
    if (init?.method !== 'GET') calls.push(init ?? {});
    return Response.json(view);
  }));
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(<QueryClientProvider client={client}><LifecyclesView scopeQuery="project=acme" kind="todos" onSelect={vi.fn()} models={[]} families={[]} /></QueryClientProvider>);
  const save = await screen.findByRole('button', { name: 'Save lifecycle' });
  expect(save.hasAttribute('disabled')).toBe(true);
  expect(screen.getByRole('button', { name: 'Disable verify' }).hasAttribute('disabled')).toBe(true);
  fireEvent.click(screen.getByRole('button', { name: 'Edit plan' }));
  expect(screen.getByRole('button', { name: 'Move plan up' }).hasAttribute('disabled')).toBe(true);
  fireEvent.click(screen.getByRole('button', { name: 'Reset plan' }));
  fireEvent.click(screen.getByRole('button', { name: 'Move run up' }));
  fireEvent.click(screen.getByRole('button', { name: 'Preview lifecycle' }));
  await waitFor(() => expect(save.hasAttribute('disabled')).toBe(false));
  fireEvent.click(save);
  await waitFor(() => expect(calls.some(call => call.method === 'PUT')).toBe(true));
  const payload = JSON.parse(String(calls.find(call => call.method === 'PUT')?.body));
  expect(payload).toEqual({ revision: 'original', source: { steps: [], order: ['plan', 'run', 'verify'] } });
  fireEvent.click(screen.getByRole('button', { name: 'Disable plan' }));
  expect(save.hasAttribute('disabled')).toBe(true);
});

it("picks a step's prompt id from the catalog while still accepting a file: reference, leaving the inline prompt object alone", () => {
  const entries = [
    { id: 'todos.run', title: 'Run a todo' },
    { id: 'todos.plan', title: 'todos.plan' },
  ] as PromptCatalogEntry[];
  const field = (key: string, type: string): FieldControl => ({ key, kind: 'string', label: key, required: true, schema: { type }, value: 'todos.run', onChange: vi.fn() } as FieldControl);
  const picker = stepPromptPicker(entries);

  const stepPrompt = picker(field('prompt', 'string'), { key: 'prompt', prop: { type: 'string' }, value: 'todos.run' }) as FieldControl & { options?: unknown; allowCustomValue?: boolean };
  const inlinePrompt = picker(field('prompt', 'object'), { key: 'prompt', prop: { type: 'object' }, value: {} });
  const name = picker(field('name', 'string'), { key: 'name', prop: { type: 'string' }, value: 'run' });

  expect([stepPrompt.kind, stepPrompt.options, stepPrompt.allowCustomValue, inlinePrompt?.kind, name?.kind]).toEqual([
    'enum',
    [{ value: 'todos.run', label: 'todos.run — Run a todo' }, { value: 'todos.plan', label: 'todos.plan' }],
    true,
    'string',
    'string',
  ]);
});

it("renders a step's Captain spec as one RuntimeBar carrying the full options, and leaves other objects to the form", () => {
  const onChange = vi.fn();
  const specSchema = { type: 'object', properties: { mode: { type: 'string' }, model: { type: 'string' } } };
  const spec = { key: 'spec', kind: 'object', label: 'Spec', required: false, schema: specSchema, value: { model: 'gpt-5.5', tools: ['bash'] }, onChange } as unknown as FieldControl;
  const subject = { ...spec, key: 'subject', schema: { type: 'object', properties: { kind: { type: 'string' } } } } as unknown as FieldControl;
  const nodes = { label: 'label', value: 'form' };
  const extension = stepRuntimeBar({ models: [], families: [] });

  const bar = extension(spec, nodes).value as ReactElement<{ value: unknown; showTimeout?: boolean; showCost?: boolean; actions?: { fields?: { id: string }[] }; onChange: (next: unknown) => void }>;
  bar.props.onChange({ model: 'gpt-5.5', tools: ['bash'], effort: 'high' });

  expect({
    type: bar.type === RuntimeBar,
    value: bar.props.value,
    showTimeout: bar.props.showTimeout,
    showCost: bar.props.showCost,
    fields: bar.props.actions?.fields?.map(field => field.id),
    committed: onChange.mock.calls[0]?.[0],
    subject: extension(subject, nodes).value,
  }).toEqual({
    type: true,
    value: { model: 'gpt-5.5', tools: ['bash'] },
    showTimeout: true,
    showCost: true,
    fields: expect.arrayContaining(['setup.checkout.worktree.mode', 'workflow.commits']),
    committed: { model: 'gpt-5.5', tools: ['bash'], effort: 'high' },
    subject: 'form',
  });
});
