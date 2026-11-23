// Tests that the connections panel reads its form schema and dashboard from the server's
// connection catalog, gives its YAML and JSON editors Monaco's workers, and stacks its legend.
import { render, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { ConnectionsPanel, LEGEND_SELECTOR } from './ConnectionsPanel';

vi.mock('monaco-editor/esm/vs/editor/editor.worker?worker', () => ({ default: class EditorWorker {} }));
vi.mock('monaco-editor/esm/vs/language/json/json.worker?worker', () => ({ default: class JsonWorker {} }));
vi.mock('monaco-yaml/yaml.worker?worker', () => ({ default: class YamlWorker {} }));

/** The global hook Monaco asks for a language's worker. */
type MonacoEnvironmentScope = { MonacoEnvironment?: { getWorker?: (moduleId: string, label: string) => object } };

function renderPanel() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <ConnectionsPanel />
    </QueryClientProvider>,
  );
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('ConnectionsPanel', () => {
  it('reads the connection catalog the server serves under /api/catalog', async () => {
    const fetchMock = vi.fn().mockImplementation(async () => new Response(JSON.stringify({ connections: [] }), {
      status: 200, headers: { 'Content-Type': 'application/json' },
    }));
    vi.stubGlobal('fetch', fetchMock);

    renderPanel();

    await waitFor(() => {
      const urls = fetchMock.mock.calls.map(([url]) => String(url instanceof Request ? url.url : url));
      expect(urls.some(url => url.includes('/api/catalog/connections/dashboard'))).toBe(true);
      expect(urls.some(url => url.endsWith('/api/catalog/connection') || url.includes('/api/catalog/connection?'))).toBe(true);
    });
  });

  it("gives the connection editor's YAML and JSON tabs Monaco's workers for their languages", async () => {
    vi.stubGlobal('fetch', vi.fn().mockImplementation(async () => new Response(JSON.stringify({ connections: [] }), {
      status: 200, headers: { 'Content-Type': 'application/json' },
    })));
    renderPanel();

    const scope = globalThis as MonacoEnvironmentScope;
    await waitFor(() => expect(scope.MonacoEnvironment?.getWorker).toBeTypeOf('function'));
    const worker = (label: string) => scope.MonacoEnvironment!.getWorker!('', label).constructor.name;
    expect(worker('yaml')).toBe('YamlWorker');
    expect(worker('json')).toBe('JsonWorker');
    expect(worker('editorWorkerService')).toBe('EditorWorker');
  });

  it('stacks the health legend one state per line', async () => {
    const dashboard = {
      connections: [{ id: 'c1', name: 'gavel', namespace: '', type: 'github', secretCount: 0, inlineCredential: true,
        insecureTLS: false, profileCount: 0, updatedAt: '2026-11-23T00:00:00Z' }],
      generatedAt: '2026-11-23T00:00:00Z',
    };
    vi.stubGlobal('fetch', vi.fn().mockImplementation(async (input: RequestInfo | URL) => {
      const url = String(input instanceof Request ? input.url : input);
      return new Response(JSON.stringify(url.includes('/connections/dashboard') ? dashboard : { connections: [] }), {
        status: 200, headers: { 'Content-Type': 'application/json' },
      });
    }));
    const { container } = renderPanel();

    await waitFor(() => expect(container.textContent).toContain('check did not complete'));
    const panel = container.querySelector('[data-connections-panel]');
    const matched = panel?.querySelectorAll(LEGEND_SELECTOR) ?? [];
    expect(matched).toHaveLength(1);
    expect(matched[0]?.textContent).toContain('not checked');
    expect(panel?.className).toContain(`[&_${LEGEND_SELECTOR}]:flex-col`);
  });
});
