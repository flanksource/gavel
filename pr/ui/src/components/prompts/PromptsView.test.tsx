import type { ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { PromptCatalogEntry } from '@flanksource/clicky-ui/ai';
import { PromptsView } from './PromptsView';

vi.mock('@flanksource/clicky-ui/ai', () => ({
  familiesFromRuntimeCatalog: () => [],
  PromptCatalogTable: ({ entries, onSelect }: { entries: PromptCatalogEntry[]; onSelect: (entry: PromptCatalogEntry) => void }) => (
    <ul>
      {entries.map(entry => (
        <li key={entry.id}>
          <button onClick={() => onSelect(entry)}>{entry.title}</button>
        </li>
      ))}
    </ul>
  ),
  PromptPage: ({ entry }: { entry: PromptCatalogEntry }) => (
    <div data-testid="prompt-page">
      page:{entry.id} default:{entry.defaultRaw ?? 'none'}
    </div>
  ),
  RuntimeProfilesWorkspace: ({ presets, store, persistence }: {
    presets: Array<{ id: string; name: string; description?: string }>;
    store: { updatePreset: (preset: unknown) => void; deletePreset: (id: string) => void };
    persistence: { onSave: () => void };
  }) => (
    <div data-testid="preset-workspace">
      {presets.map(preset => (
        <div key={preset.id}>
          <span>{preset.name}</span>
          <button onClick={() => store.updatePreset({ ...preset, description: 'Edited description' })}>Edit {preset.name}</button>
          <button onClick={() => store.deletePreset(preset.id)}>Delete {preset.name}</button>
        </div>
      ))}
      <button onClick={persistence.onSave}>Save presets</button>
    </div>
  ),
}));

vi.mock('@flanksource/clicky-ui/components', () => ({
  Combobox: ({ value, options }: { value: string; options: Array<{ value: string; label: string }> }) => (
    <div data-testid="scope">{options.find(option => option.value === value)?.label}</div>
  ),
  Tabs: ({ tabs, onChange }: { tabs: Array<{ id: string; label: string }>; onChange: (id: string) => void }) => (
    <div>
      {tabs.map(tab => <button key={tab.id} onClick={() => onChange(tab.id)}>{tab.label}</button>)}
    </div>
  ),
}));

const catalogEntry: PromptCatalogEntry = {
  id: 'commit.message',
  title: 'Commit message',
  owner: 'gavel',
  source: 'inline',
  effective: { model: 'claude-sonnet-4-6', modelSource: 'operation' },
  layers: [],
};

const defaultRaw = '---\nmodel: x\n---\nbody';

function stubApi() {
  const calls: Array<{ url: string; init?: RequestInit }> = [];
  vi.stubGlobal('fetch', vi.fn(async (url: string, init?: RequestInit) => {
    calls.push({ url, init });
    if (url.startsWith('/api/settings/prompts/catalog')) {
      return { ok: true, json: async () => [catalogEntry] };
    }
    if (url === '/api/settings/prompts') {
      return { ok: true, json: async () => [{ id: 'commit.message', title: 'Commit message', configPath: 'commit.message', default: defaultRaw }] };
    }
    if (url === '/api/todos/run/context?dir=%2Fwork%2Facme') {
      return { ok: true, json: async () => ({ runtimes: [], models: [], tools: [], modes: [], efforts: [], lifecycle: { steps: [] } }) };
    }
    if (url === '/api/settings/runtime-presets?project=acme' && !init?.method) {
      return { ok: true, json: async () => ({
        presets: [{ id: 'preset-1', key: 'review', name: 'Review preset', scope: 'context', spec: {}, presets: [], source: { kind: 'file', id: 'project', label: 'Project presets', writable: true, records: ['preset'] }, updatedAt: '2026-09-15T10:00:00Z' }],
        profiles: [],
        sources: [{ kind: 'file', id: 'project', label: 'Project presets', writable: true, records: ['preset'] }],
      }) };
    }
    if (url === '/api/settings/runtime-presets/preset-1?project=acme' && init?.method === 'PUT') {
      return { ok: true, json: async () => ({ id: 'preset-1', key: 'review', name: 'Review preset', description: 'Edited description', scope: 'context', spec: {}, presets: [], source: { kind: 'file', id: 'project', label: 'Project presets', writable: true, records: ['preset'] }, updatedAt: '2026-09-15T10:01:00Z' }) };
    }
    if (url === '/api/settings/runtime-presets/preset-1?project=acme' && init?.method === 'DELETE') {
      return { ok: true, status: 204, text: async () => '' };
    }
    if (url.startsWith('/api/settings/runtime-presets/resolve?')) {
      return { ok: true, json: async () => ({ resolved: { spec: {}, trace: [] }, tools: [], permissions: {}, permissionSupport: {}, effectivePolicy: [] }) };
    }
    return { ok: false, status: 404, text: async () => 'not found' };
  }));
  return calls;
}

type Props = Parameters<typeof PromptsView>[0];

function renderView(props: Partial<Props> = {}) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const onNavigate = vi.fn();
  const view: ReactNode = (
    <QueryClientProvider client={client}>
      <PromptsView
        projects={[{ name: 'acme', dir: '/work/acme', repos: [] }]}
        scopeProject=""
        selectedId=""
        onNavigate={onNavigate}
        {...props}
      />
    </QueryClientProvider>
  );
  render(view);
  return { onNavigate };
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('PromptsView', () => {
  it('lists the catalog for the scope and navigates to a selected prompt within it', async () => {
    const calls = stubApi();
    const { onNavigate } = renderView({ scopeProject: 'acme' });

    fireEvent.click(await screen.findByRole('button', { name: 'Commit message' }));

    expect(calls.map(call => call.url)).toContain('/api/settings/prompts/catalog?project=acme');
    expect(screen.getByTestId('scope').textContent).toBe('acme');
    expect(onNavigate).toHaveBeenCalledWith('commit.message', 'acme');
  });

  it('opens the page for the selected prompt with its built-in default attached', async () => {
    stubApi();
    renderView({ selectedId: 'commit.message' });

    const page = await screen.findByTestId('prompt-page');
    expect(page.textContent).toBe(`page:commit.message default:${defaultRaw}`);
  });

  it('explains a selected id the scope does not resolve instead of a blank page', async () => {
    stubApi();
    renderView({ selectedId: 'todos.prompts.missing' });

    const alert = await screen.findByRole('alert');
    expect(alert.textContent).toContain('No prompt todos.prompts.missing in this scope.');
  });

  it('keeps the preset workspace in the prompt route', () => {
    stubApi();
    const { onNavigate } = renderView({ scopeProject: 'acme' });

    fireEvent.click(screen.getByRole('button', { name: 'Presets' }));

    expect(onNavigate).toHaveBeenCalledWith('presets', 'acme');
  });

  it('lists, edits, saves, and deletes presets in the selected scope', async () => {
    const calls = stubApi();
    renderView({ scopeProject: 'acme', selectedId: 'presets' });

    expect(await screen.findByText('Review preset')).toBeTruthy();
    expect(calls.map(call => call.url)).toContain('/api/settings/runtime-presets?project=acme');

    fireEvent.click(screen.getByRole('button', { name: 'Edit Review preset' }));
    fireEvent.click(screen.getByRole('button', { name: 'Save presets' }));
    await waitFor(() => expect(calls.some(call => call.url === '/api/settings/runtime-presets/preset-1?project=acme' && call.init?.method === 'PUT')).toBe(true));
    const update = calls.find(call => call.init?.method === 'PUT');
    expect(JSON.parse(String(update?.init?.body))).toEqual({ name: 'Review preset', description: 'Edited description', scope: 'context', spec: {}, presets: [] });

    fireEvent.click(screen.getByRole('button', { name: 'Delete Review preset' }));
    await waitFor(() => expect(calls.some(call => call.url === '/api/settings/runtime-presets/preset-1?project=acme' && call.init?.method === 'DELETE')).toBe(true));
  });
});
