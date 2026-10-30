import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import commitSchema from '../../../../schemas/commit.schema.json';
import lintSchema from '../../../../schemas/lint.schema.json';
import testSchema from '../../../../schemas/test.schema.json';
import { ProjectActionDialog, type ProjectAction } from './ProjectActionDialog';

const storageKey = 'gavel.project-action-options.v2';
const projectName = 'acme/widget';
const documents = { commit: commitSchema, lint: lintSchema, test: testSchema };

afterEach(() => { cleanup(); vi.unstubAllGlobals(); });

function renderAction(action: ProjectAction, modelFailures = 0) {
  const fetchMock = vi.fn(async (url: string) => {
    if (url === '/api/chat/models' && modelFailures-- > 0) {
      return { ok: false, status: 503, text: async () => 'model catalog unavailable' };
    }
    return {
      ok: true,
      json: async () => url === '/api/chat/models' ? [{
        id: 'example-api', provider: 'example', label: 'Example model', reasoning: true,
        capabilitiesKnown: true, supportedEfforts: ['low', 'high'], configured: true,
        runtime: { model: 'example-model', mode: 'api' },
      }] : url === '/api/chat/runtimes' ? [{
        family: 'example', provider: 'example', catalogPrefix: 'example',
        modes: ['api', 'agent'].map(mode => ({ mode, schema: { type: 'object', properties: { model: { type: 'string' }, effort: { type: 'string' } } } })),
      }] : { schemaVersion: 2, action, schema: documents[action], defaults: { 'ai-fix-max-iterations': 0, recursive: true } },
    };
  });
  vi.stubGlobal('fetch', fetchMock);
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const props = { projectName, selectedFiles: ['one.go'], onClose: vi.fn(), onRun: vi.fn().mockResolvedValue(undefined) };
  const view = render(<QueryClientProvider client={client}><ProjectActionDialog {...props} action={action} /></QueryClientProvider>);
  return { ...props, fetchMock, refreshSelection: (selectedFiles: string[]) => {
    view.rerender(<QueryClientProvider client={client}><ProjectActionDialog {...props} selectedFiles={selectedFiles} action={action} /></QueryClientProvider>);
  }, reopen: () => {
    view.rerender(<QueryClientProvider client={client}><ProjectActionDialog {...props} action={null} /></QueryClientProvider>);
    view.rerender(<QueryClientProvider client={client}><ProjectActionDialog {...props} action={action} /></QueryClientProvider>);
  } };
}

describe('ProjectActionDialog sparse v2 options', () => {
  it.each([
    { action: 'commit' as const, button: 'Commit selected', field: 'Commit message', expected: { files: ['one.go'] } },
    { action: 'lint' as const, button: 'Run lint', field: 'Files and directories', expected: { files: ['one.go'] } },
    { action: 'test' as const, button: 'Run tests', field: 'Package paths', expected: {} },
  ])('submits no untouched defaults for $action', async ({ action, button, field, expected }) => {
    const view = renderAction(action);
    await screen.findByText(field, { exact: true });
    fireEvent.click(screen.getByRole('button', { name: button }));
    await waitFor(() => expect(view.onRun).toHaveBeenCalledWith(action, expected));
    expect(JSON.parse(localStorage.getItem(storageKey) ?? '{}')[projectName][action]).toEqual({ schemaVersion: 2, value: {} });
  });

  it('ignores v1 preferences and adds only the first edited field', async () => {
    localStorage.setItem('gavel.project-action-options.v1', JSON.stringify({
      [projectName]: { commit: { schemaVersion: 1, value: { 'ai-fix-max-iterations': 0, 'commit-all': true } } },
    }));
    const view = renderAction('commit');
    fireEvent.change(await screen.findByRole('textbox', { name: 'Commit message' }), { target: { value: 'fix: selected files' } });
    fireEvent.click(screen.getByRole('button', { name: 'Commit selected' }));
    await waitFor(() => expect(view.onRun).toHaveBeenCalledWith('commit', { files: ['one.go'], message: 'fix: selected files' }));
    expect(JSON.parse(localStorage.getItem(storageKey) ?? '{}')[projectName].commit.value).toEqual({ message: 'fix: selected files' });
  });

  it('retains authored options while project status refreshes the selected files', async () => {
    const view = renderAction('commit');
    fireEvent.change(await screen.findByRole('textbox', { name: 'Commit message' }), { target: { value: 'fix: selected files' } });
    view.refreshSelection(['two.go']);
    fireEvent.click(screen.getByRole('button', { name: 'Commit selected' }));
    await waitFor(() => expect(view.onRun).toHaveBeenCalledWith('commit', { files: ['two.go'], message: 'fix: selected files' }));
  });

  it('displays inherited true without storing it and preserves explicit false', async () => {
    const view = renderAction('commit');
    const metadata = await screen.findByRole('checkbox', { name: 'Include session metadata' });
    expect((metadata as HTMLInputElement).checked).toBe(true);
    fireEvent.click(metadata);
    fireEvent.click(screen.getByRole('button', { name: 'Commit selected' }));
    await waitFor(() => expect(view.onRun).toHaveBeenCalledWith('commit', { files: ['one.go'], 'add-metadata': false }));
    view.reopen();
    expect((await screen.findByRole('checkbox', { name: 'Include session metadata' }) as HTMLInputElement).checked).toBe(false);
    expect(view.fetchMock.mock.calls.filter(([url]) => url.includes('/actions/schema'))).toHaveLength(1);
    expect(view.fetchMock.mock.calls.filter(([url]) => url === '/api/chat/models')).toHaveLength(1);
    expect(view.fetchMock.mock.calls.filter(([url]) => url === '/api/chat/runtimes')).toHaveLength(1);
  });

  it('reveals grouping overrides and removes them when grouping is disabled', async () => {
    const view = renderAction('commit');
    fireEvent.change(await screen.findByRole('textbox', { name: 'Commit message' }), { target: { value: 'fix: selected files' } });
    expect(screen.queryByText('Maximum logical commits', { exact: true })).toBeNull();
    const grouping = screen.getByRole('checkbox', { name: 'Split into logical commits' });
    fireEvent.click(grouping);
    fireEvent.change(await screen.findByRole('spinbutton', { name: 'Maximum logical commits' }), { target: { value: '3' } });
    expect(screen.queryByRole('textbox', { name: 'Commit message' })).toBeNull();
    fireEvent.click(grouping);
    expect(screen.queryByText('Maximum logical commits', { exact: true })).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'Commit selected' }));
    await waitFor(() => expect(view.onRun).toHaveBeenCalledWith('commit', { files: ['one.go'], 'commit-all': false }));
  });

  it('preserves explicitly authored automatic zero', async () => {
    const view = renderAction('test');
    fireEvent.change(await screen.findByRole('spinbutton', { name: 'Ginkgo parallel nodes' }), { target: { value: '0' } });
    fireEvent.click(screen.getByRole('button', { name: 'Run tests' }));
    await waitFor(() => expect(view.onRun).toHaveBeenCalledWith('test', { nodes: 0 }));
  });

  it('reveals an empty repair limit and removes an authored limit when repair is disabled', async () => {
    const view = renderAction('commit');
    const repair = within(await screen.findByRole('radiogroup', { name: 'Repair lint findings' }));
    expect(screen.queryByRole('spinbutton', { name: 'Maximum repair iterations' })).toBeNull();
    fireEvent.click(repair.getByRole('radio', { name: 'Enabled' }));
    const limit = await screen.findByRole('spinbutton', { name: 'Maximum repair iterations' });
    expect((limit as HTMLInputElement).value).toBe('');
    fireEvent.change(limit, { target: { value: '4' } });
    fireEvent.click(repair.getByRole('radio', { name: 'Disabled' }));
    expect(screen.queryByRole('spinbutton', { name: 'Maximum repair iterations' })).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'Commit selected' }));
    await waitFor(() => expect(view.onRun).toHaveBeenCalledWith('commit', { files: ['one.go'], 'ai-fix': 'false' }));
  });

  it.each(['Lint checks', 'Secret checks', 'Repair lint findings', 'Tidy Go modules'])('renders %s as inherited segmented toggles', async label => {
    const view = renderAction('commit');
    const toggle = within(await screen.findByRole('radiogroup', { name: label }));
    expect(toggle.getByRole('radio', { name: 'Inherit' }).getAttribute('aria-checked')).toBe('true');
    fireEvent.click(toggle.getByRole('radio', { name: 'Enabled' }));
    fireEvent.click(toggle.getByRole('radio', { name: 'Disabled' }));
    fireEvent.click(toggle.getByRole('radio', { name: 'Inherit' }));
    fireEvent.click(screen.getByRole('button', { name: 'Commit selected' }));
    await waitFor(() => expect(view.onRun).toHaveBeenCalledWith('commit', { files: ['one.go'] }));
  });

  it('uses only native runtime controls and preserves unrelated options', async () => {
    const view = renderAction('commit');
    const bar = within(await screen.findByRole('group', { name: 'Commit AI runtime' }));
    expect(screen.queryByRole('textbox', { name: 'Message model' })).toBeNull();
    expect(screen.queryByText('Fallback models', { exact: true })).toBeNull();
    expect(screen.queryByText('Temperature', { exact: true })).toBeNull();
    expect(screen.queryByText('Bypass AI cache', { exact: true })).toBeNull();
    fireEvent.change(screen.getByRole('textbox', { name: 'Commit message' }), { target: { value: 'fix: selected files' } });
    fireEvent.click(bar.getByTitle('Model — unspecified'));
    fireEvent.mouseDown(await screen.findByRole('option', { name: 'Example model' }));
    fireEvent.click(bar.getByTitle('Runtime options'));
    expect(screen.getAllByRole('menuitem').map(item => item.textContent)).toEqual(['Effort']);
    fireEvent.click(await screen.findByRole('menuitem', { name: 'Effort' }));
    fireEvent.click(await screen.findByRole('menuitem', { name: 'High' }));
    fireEvent.click(screen.getByRole('button', { name: 'Commit selected' }));
    await waitFor(() => expect(view.onRun).toHaveBeenCalledWith('commit', {
      files: ['one.go'], message: 'fix: selected files', model: 'example-model', mode: 'api', effort: 'high',
    }));
  });

  it('authors only the selected mode through the native picker', async () => {
    const view = renderAction('commit');
    const bar = within(await screen.findByRole('group', { name: 'Commit AI runtime' }));
    fireEvent.click(bar.getByTitle('Runtime mode — API'));
    fireEvent.click(await screen.findByRole('menuitem', { name: 'Agent' }));
    fireEvent.click(screen.getByRole('button', { name: 'Commit selected' }));
    await waitFor(() => expect(view.onRun).toHaveBeenCalledWith('commit', { files: ['one.go'], mode: 'agent' }));
  });

  it('excludes removed AI settings from remembered v2 options', async () => {
    localStorage.setItem(storageKey, JSON.stringify({ [projectName]: { commit: { schemaVersion: 2, value: {
      'group-model': 'old-group', fallback: ['old-fallback'], temperature: '0', 'no-cache': false, 'add-metadata': false,
    } } } }));
    const view = renderAction('commit');
    await screen.findByRole('group', { name: 'Commit AI runtime' });
    fireEvent.click(screen.getByRole('button', { name: 'Commit selected' }));
    await waitFor(() => expect(view.onRun).toHaveBeenCalledWith('commit', { files: ['one.go'], 'add-metadata': false }));
    expect(JSON.parse(localStorage.getItem(storageKey) ?? '{}')[projectName].commit.value).toEqual({ 'add-metadata': false });
  });

  it('preserves compact model selectors until a runtime control is edited', async () => {
    const saved = { model: 'api:example-model:high', effort: 'low', 'add-metadata': false };
    localStorage.setItem(storageKey, JSON.stringify({ [projectName]: { commit: { schemaVersion: 2, value: saved } } }));
    const view = renderAction('commit');
    const bar = within(await screen.findByRole('group', { name: 'Commit AI runtime' }));
    expect(bar.getByTitle('Reasoning effort').textContent).toContain('High');
    fireEvent.click(screen.getByRole('button', { name: 'Commit selected' }));
    await waitFor(() => expect(view.onRun).toHaveBeenLastCalledWith('commit', { ...saved, files: ['one.go'] }));
    view.reopen();
    fireEvent.click(within(await screen.findByRole('group', { name: 'Commit AI runtime' })).getByTitle('Reasoning effort'));
    fireEvent.click(await screen.findByRole('menuitem', { name: 'Low' }));
    fireEvent.click(screen.getByRole('button', { name: 'Commit selected' }));
    await waitFor(() => expect(view.onRun).toHaveBeenLastCalledWith('commit', {
      files: ['one.go'], model: 'example-model', mode: 'api', effort: 'low', 'add-metadata': false,
    }));
  });

  it('clears native model and effort overrides without changing other authored fields', async () => {
    localStorage.setItem(storageKey, JSON.stringify({ [projectName]: { commit: { schemaVersion: 2, value: {
      model: 'example-model', effort: 'high', 'add-metadata': false,
    } } } }));
    const view = renderAction('commit');
    const bar = within(await screen.findByRole('group', { name: 'Commit AI runtime' }));
    fireEvent.click(bar.getByTitle('Model — example-model'));
    fireEvent.mouseDown(await screen.findByRole('option', { name: /^-/ }));
    fireEvent.click(bar.getByTitle('Reasoning effort'));
    fireEvent.click(await screen.findByRole('menuitem', { name: /^-/ }));
    fireEvent.click(screen.getByRole('button', { name: 'Commit selected' }));
    await waitFor(() => expect(view.onRun).toHaveBeenCalledWith('commit', { files: ['one.go'], 'add-metadata': false }));
  });

  it.each(['lint', 'test'] as const)('does not load AI catalogs for %s', async action => {
    const view = renderAction(action);
    await screen.findByText(action === 'lint' ? 'Files and directories' : 'Package paths', { exact: true });
    expect(view.fetchMock.mock.calls.map(([url]) => url)).toEqual([
      `/api/projects/acme%2Fwidget/actions/schema?action=${action}`,
    ]);
    expect(screen.queryByRole('group', { name: 'Commit AI runtime' })).toBeNull();
  });

  it('retries catalog failures while preserving authored options', async () => {
    const view = renderAction('commit', 1);
    expect((await screen.findByRole('alert')).textContent).toContain('model catalog unavailable');
    fireEvent.change(screen.getByRole('textbox', { name: 'Commit message' }), { target: { value: 'fix: retry catalog' } });
    fireEvent.click(screen.getByRole('button', { name: 'Retry AI catalogs' }));
    await screen.findByRole('group', { name: 'Commit AI runtime' });
    fireEvent.click(screen.getByRole('button', { name: 'Commit selected' }));
    await waitFor(() => expect(view.onRun).toHaveBeenCalledWith('commit', { files: ['one.go'], message: 'fix: retry catalog' }));
  });

  it.each([
    { action: 'lint' as const, toggle: 'Summarize repeated findings', dependent: 'Examples per rule' },
    { action: 'test' as const, toggle: 'Discover fixture tests', dependent: 'Fixture globs' },
    { action: 'test' as const, toggle: 'Run lint alongside tests', dependent: 'Per-linter timeout' },
  ])('reveals $dependent through native listeners', async ({ action, toggle, dependent }) => {
    renderAction(action);
    const input = await screen.findByRole('checkbox', { name: toggle });
    expect(screen.queryByText(dependent, { exact: true })).toBeNull();
    fireEvent.click(input);
    expect(await screen.findByText(dependent, { exact: true })).toBeTruthy();
  });

  it('does not save rejected requests', async () => {
    const view = renderAction('commit');
    view.onRun.mockRejectedValueOnce(new Error('Selected files changed'));
    await screen.findByRole('textbox', { name: 'Commit message' });
    fireEvent.click(screen.getByRole('button', { name: 'Commit selected' }));
    expect(await screen.findByRole('alert')).toHaveProperty('textContent', 'Selected files changed');
    expect(localStorage.getItem(storageKey)).toBeNull();
    expect(view.onClose).not.toHaveBeenCalled();
  });
});
