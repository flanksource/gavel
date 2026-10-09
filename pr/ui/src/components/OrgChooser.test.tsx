import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { OrgChooser } from './OrgChooser';
import type { Project, SearchConfig } from '../types';

const projects: Project[] = [
  { name: 'gavel', dir: '/work/gavel', repos: ['acme/gavel'] },
  { name: 'clicky', dir: '/work/clicky', repos: ['acme/clicky', 'acme/gavel'] },
  { name: 'local-only', dir: '/work/local-only', repos: [] },
];

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('OrgChooser', () => {
  it('loads the full organization list once and reuses it when reopened', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => [{ login: 'acme', avatarUrl: '' }],
    });
    vi.stubGlobal('fetch', fetchMock);
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const config = { all: false, org: '', repos: [], ignoredOrgs: [] } as SearchConfig;
    render(
      <QueryClientProvider client={queryClient}>
        <OrgChooser config={config} projects={projects} onChange={vi.fn()} />
      </QueryClientProvider>,
    );

    const trigger = screen.getByTitle('Switch project / GitHub scope');
    fireEvent.click(trigger);
    expect(await screen.findByText('acme')).toBeTruthy();
    fireEvent.click(trigger);
    fireEvent.click(trigger);
    await waitFor(() => expect(screen.getByText('acme')).toBeTruthy());

    expect(fetchMock).toHaveBeenCalledOnce();
    expect(fetchMock).toHaveBeenCalledWith('/api/orgs?include-ignored=1', {
      signal: expect.any(AbortSignal),
    });
  });

  it('selects a project as the active repo and workspace scope', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: async () => [] }));
    const onChange = vi.fn();
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={queryClient}>
        <OrgChooser
          config={{ all: true, org: 'acme', repos: [] }}
          projects={projects}
          onChange={onChange}
        />
      </QueryClientProvider>,
    );

    fireEvent.click(screen.getByTitle('Switch project / GitHub scope'));
    fireEvent.click(screen.getByRole('button', { name: 'Filter by gavel project' }));

    expect(onChange).toHaveBeenCalledWith({
      project: 'gavel',
      repos: ['acme/gavel'],
      org: '',
      all: false,
    });
  });

  it('selects every configured project and its repositories', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: async () => [] }));
    const onChange = vi.fn();
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={queryClient}>
        <OrgChooser
          config={{ project: 'gavel', repos: ['acme/gavel'] }}
          projects={projects}
          onChange={onChange}
        />
      </QueryClientProvider>,
    );

    fireEvent.click(screen.getByTitle('Switch project / GitHub scope'));
    fireEvent.click(screen.getByRole('button', { name: 'Filter by all projects' }));

    expect(onChange).toHaveBeenCalledWith({
      project: '',
      repos: ['acme/gavel', 'acme/clicky'],
      org: '',
      all: false,
    });
  });

  it('shows the combined project scope as active', () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: async () => [] }));
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={queryClient}>
        <OrgChooser
          config={{ project: '', repos: ['acme/clicky', 'acme/gavel'], all: false }}
          projects={projects}
          onChange={vi.fn()}
        />
      </QueryClientProvider>,
    );

    expect(screen.getByTitle('Switch project / GitHub scope').textContent).toContain('All projects');
  });

  it('keeps personal scope active until the project catalog loads', () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: async () => [] }));
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={queryClient}>
        <OrgChooser
          config={{ project: '', repos: [], all: false }}
          projects={[]}
          onChange={vi.fn()}
        />
      </QueryClientProvider>,
    );

    expect(screen.getByTitle('Switch project / GitHub scope').textContent).toContain('@me');
  });

  it('keeps local-only projects selectable for todo scoping', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: async () => [] }));
    const onChange = vi.fn();
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={queryClient}>
        <OrgChooser
          config={{ repos: [], project: 'gavel' }}
          projects={projects}
          onChange={onChange}
        />
      </QueryClientProvider>,
    );

    expect(screen.getByTitle('Switch project / GitHub scope').textContent).toContain('gavel');
    fireEvent.click(screen.getByTitle('Switch project / GitHub scope'));
    fireEvent.click(screen.getByRole('button', { name: 'Filter by local-only project' }));

    expect(onChange).toHaveBeenCalledWith({
      project: 'local-only',
      repos: [],
      org: '',
      all: false,
    });
  });

  it('clears the project scope when a GitHub organization is selected', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: async () => [{ login: 'acme', avatarUrl: '' }],
    }));
    const onChange = vi.fn();
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={queryClient}>
        <OrgChooser
          config={{ repos: ['acme/gavel'], project: 'gavel' }}
          projects={projects}
          onChange={onChange}
        />
      </QueryClientProvider>,
    );

    fireEvent.click(screen.getByTitle('Switch project / GitHub scope'));
    fireEvent.click(await screen.findByRole('button', { name: 'Select acme organization' }));

    expect(onChange).toHaveBeenCalledWith({ project: '', org: 'acme', all: true, repos: [] });
  });
});
