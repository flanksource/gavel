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

// Routes the chooser's two reads: the unfiltered org list and the
// hidden-github-orgs filter whose values decide which of them are hidden.
function stubOrgFetch({ orgs = [], hidden = [] }: { orgs?: { login: string; avatarUrl: string }[]; hidden?: string[] } = {}) {
  const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
    const url = String(input);
    const body = url === '/api/orgs' ? orgs
      : url === '/api/filters/hidden-github-orgs' ? { name: 'hidden-github-orgs', kind: 'github-org', values: hidden }
        : undefined;
    if (body === undefined) throw new Error(`unexpected request: ${url}`);
    return new Response(JSON.stringify(body), { status: 200, headers: { 'Content-Type': 'application/json' } });
  });
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
}

describe('OrgChooser', () => {
  it('loads the full organization list and the hidden-org filter once and reuses them when reopened', async () => {
    const fetchMock = stubOrgFetch({ orgs: [{ login: 'acme', avatarUrl: '' }] });
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const config = { all: false, org: '', repos: [] } as SearchConfig;
    render(
      <QueryClientProvider client={queryClient}>
        <OrgChooser config={config} projects={projects} onChange={vi.fn()} onSetProjectHidden={vi.fn()} onSetOrgHidden={vi.fn()} />
      </QueryClientProvider>,
    );

    const trigger = screen.getByTitle('Switch project / GitHub scope');
    fireEvent.click(trigger);
    expect(await screen.findByText('acme')).toBeTruthy();
    fireEvent.click(trigger);
    fireEvent.click(trigger);
    await waitFor(() => expect(screen.getByText('acme')).toBeTruthy());

    expect(fetchMock.mock.calls.map(([input]) => String(input)).sort()).toEqual([
      '/api/filters/hidden-github-orgs',
      '/api/orgs',
    ]);
    expect(fetchMock).toHaveBeenCalledWith('/api/orgs', {
      signal: expect.any(AbortSignal),
    });
  });

  it('lists an org in the hidden-github-orgs filter only under Manage hidden and unhides it through the callback', async () => {
    stubOrgFetch({ orgs: [{ login: 'acme', avatarUrl: '' }, { login: 'noisy', avatarUrl: '' }], hidden: ['noisy'] });
    const onSetOrgHidden = vi.fn();
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={queryClient}>
        <OrgChooser config={{ repos: [] }} projects={projects} onChange={vi.fn()} onSetProjectHidden={vi.fn()} onSetOrgHidden={onSetOrgHidden} />
      </QueryClientProvider>,
    );

    fireEvent.click(screen.getByTitle('Switch project / GitHub scope'));
    expect(await screen.findByRole('button', { name: 'Select acme organization' })).toBeTruthy();
    expect(screen.queryByRole('button', { name: 'Select noisy organization' })).toBeNull();

    fireEvent.click(screen.getByRole('button', { name: /Manage hidden \(1\)/ }));
    fireEvent.click(screen.getByTitle('Unhide noisy'));
    expect(onSetOrgHidden).toHaveBeenCalledExactlyOnceWith('noisy', false);
  });

  it('hides an org through the callback and drops back to @me when it was the active scope', async () => {
    stubOrgFetch({ orgs: [{ login: 'acme', avatarUrl: '' }] });
    const onSetOrgHidden = vi.fn();
    const onChange = vi.fn();
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={queryClient}>
        <OrgChooser config={{ all: true, org: 'acme', repos: [] }} projects={projects} onChange={onChange} onSetProjectHidden={vi.fn()} onSetOrgHidden={onSetOrgHidden} />
      </QueryClientProvider>,
    );

    fireEvent.click(screen.getByTitle('Switch project / GitHub scope'));
    await screen.findByRole('button', { name: 'Select acme organization' });
    fireEvent.click(screen.getByTitle('Hide acme from this list'));

    expect(onSetOrgHidden).toHaveBeenCalledExactlyOnceWith('acme', true);
    expect(onChange).toHaveBeenCalledExactlyOnceWith({ org: '', all: false, repos: [] });
  });

  it('selects a project as the active repo and workspace scope', async () => {
    stubOrgFetch();
    const onChange = vi.fn();
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={queryClient}>
        <OrgChooser
          config={{ all: true, org: 'acme', repos: [] }}
          projects={projects}
          onChange={onChange}
          onSetProjectHidden={vi.fn()}
          onSetOrgHidden={vi.fn()}
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
    stubOrgFetch();
    const onChange = vi.fn();
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={queryClient}>
        <OrgChooser
          config={{ project: 'gavel', repos: ['acme/gavel'] }}
          projects={projects}
          onChange={onChange}
          onSetProjectHidden={vi.fn()}
          onSetOrgHidden={vi.fn()}
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
    stubOrgFetch();
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={queryClient}>
        <OrgChooser
          config={{ project: '', repos: ['acme/clicky', 'acme/gavel'], all: false }}
          projects={projects}
          onChange={vi.fn()}
          onSetProjectHidden={vi.fn()}
          onSetOrgHidden={vi.fn()}
        />
      </QueryClientProvider>,
    );

    expect(screen.getByTitle('Switch project / GitHub scope').textContent).toContain('All projects');
  });

  it('keeps personal scope active until the project catalog loads', () => {
    stubOrgFetch();
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={queryClient}>
        <OrgChooser
          config={{ project: '', repos: [], all: false }}
          projects={[]}
          onChange={vi.fn()}
          onSetProjectHidden={vi.fn()}
          onSetOrgHidden={vi.fn()}
        />
      </QueryClientProvider>,
    );

    expect(screen.getByTitle('Switch project / GitHub scope').textContent).toContain('@me');
  });

  it('keeps local-only projects selectable for todo scoping', async () => {
    stubOrgFetch();
    const onChange = vi.fn();
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={queryClient}>
        <OrgChooser
          config={{ repos: [], project: 'gavel' }}
          projects={projects}
          onChange={onChange}
          onSetProjectHidden={vi.fn()}
          onSetOrgHidden={vi.fn()}
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
    stubOrgFetch({ orgs: [{ login: 'acme', avatarUrl: '' }] });
    const onChange = vi.fn();
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={queryClient}>
        <OrgChooser
          config={{ repos: ['acme/gavel'], project: 'gavel' }}
          projects={projects}
          onChange={onChange}
          onSetProjectHidden={vi.fn()}
          onSetOrgHidden={vi.fn()}
        />
      </QueryClientProvider>,
    );

    fireEvent.click(screen.getByTitle('Switch project / GitHub scope'));
    fireEvent.click(await screen.findByRole('button', { name: 'Select acme organization' }));

    expect(onChange).toHaveBeenCalledWith({ project: '', org: 'acme', all: true, repos: [] });
  });
});

describe('OrgChooser hidden projects', () => {
  const withHidden: Project[] = [
    ...projects,
    { name: 'archive', dir: '/work/archive', repos: ['acme/archive'], hidden: true },
  ];

  function renderChooser({ config, onChange = vi.fn(), onSetProjectHidden = vi.fn(), onSetOrgHidden = vi.fn() }: {
    config: SearchConfig;
    onChange?: (partial: Partial<SearchConfig>) => void;
    onSetProjectHidden?: (name: string, hidden: boolean) => void;
    onSetOrgHidden?: (login: string, hidden: boolean) => void;
  }) {
    stubOrgFetch();
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={queryClient}>
        <OrgChooser config={config} projects={withHidden} onChange={onChange} onSetProjectHidden={onSetProjectHidden} onSetOrgHidden={onSetOrgHidden} />
      </QueryClientProvider>,
    );
    fireEvent.click(screen.getByTitle('Switch project / GitHub scope'));
  }

  it('lists a hidden project only after expanding Manage hidden', () => {
    renderChooser({ config: { repos: [] } });

    expect(screen.queryByRole('button', { name: 'Filter by archive project' })).toBeNull();
    expect(screen.queryByText('archive')).toBeNull();

    fireEvent.click(screen.getByRole('button', { name: /Manage hidden \(1\)/ }));

    expect(screen.getByText('archive')).toBeTruthy();
    expect(screen.queryByRole('button', { name: 'Filter by archive project' })).toBeNull();
  });

  it('hides a visible project through the callback without selecting it', () => {
    const onChange = vi.fn();
    const onSetProjectHidden = vi.fn();
    renderChooser({ config: { repos: [] }, onChange, onSetProjectHidden });

    fireEvent.click(screen.getByRole('button', { name: 'Hide clicky project' }));

    expect(onSetProjectHidden).toHaveBeenCalledExactlyOnceWith('clicky', true);
    expect(onChange).not.toHaveBeenCalled();
  });

  it('unhides a hidden project through the callback', () => {
    const onSetProjectHidden = vi.fn();
    renderChooser({ config: { repos: [] }, onSetProjectHidden });

    fireEvent.click(screen.getByRole('button', { name: /Manage hidden \(1\)/ }));
    fireEvent.click(screen.getByRole('button', { name: 'Unhide archive project' }));

    expect(onSetProjectHidden).toHaveBeenCalledExactlyOnceWith('archive', false);
  });

  it('computes all-projects scope from visible projects only', () => {
    const onChange = vi.fn();
    renderChooser({ config: { project: '', repos: ['acme/gavel', 'acme/clicky'], all: false }, onChange });

    expect(screen.getByTitle('Switch project / GitHub scope').textContent).toContain('All projects');
    fireEvent.click(screen.getByRole('button', { name: 'Filter by all projects' }));

    expect(onChange).toHaveBeenCalledWith({
      project: '',
      repos: ['acme/gavel', 'acme/clicky'],
      org: '',
      all: false,
    });
  });
});
