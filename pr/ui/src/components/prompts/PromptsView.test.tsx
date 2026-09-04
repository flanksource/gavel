import type { ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { PromptCatalogEntry } from '@flanksource/clicky-ui/ai';
import { PromptsView } from './PromptsView';

vi.mock('@flanksource/clicky-ui/ai', () => ({
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
}));

vi.mock('@flanksource/clicky-ui/components', () => ({
  Combobox: ({ value, options }: { value: string; options: Array<{ value: string; label: string }> }) => (
    <div data-testid="scope">{options.find(option => option.value === value)?.label}</div>
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
  const calls: string[] = [];
  vi.stubGlobal('fetch', vi.fn(async (url: string) => {
    calls.push(url);
    if (url.startsWith('/api/settings/prompts/catalog')) {
      return { ok: true, json: async () => [catalogEntry] };
    }
    if (url === '/api/settings/prompts') {
      return { ok: true, json: async () => [{ id: 'commit.message', title: 'Commit message', configPath: 'commit.message', default: defaultRaw }] };
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

    expect(calls).toContain('/api/settings/prompts/catalog?project=acme');
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
});
