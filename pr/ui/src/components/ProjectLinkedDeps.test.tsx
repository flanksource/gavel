import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { ProjectStatusView, type ProjectFileStatus } from './ProjectStatusView';
import { ProjectChangesView } from './ProjectChangesView';
import { projectStatusResponse } from './ProjectStatusView.fixture';

vi.mock('../useGitFocus', () => ({ useGitFocus: () => ({ error: '' }) }));
vi.mock('./ProjectCommitTasks', () => ({ ProjectCommitTasks: () => null }));
vi.mock('../useIsMobile', () => ({ useIsMobile: () => true }));

const linkedFile: ProjectFileStatus = {
  ...projectStatusResponse.files[0], state: 'unstaged', path: 'generated/package.json',
  linkedDependencies: [{ name: '@acme/ui', target: 'link:../../ui' }],
};
const response = { ...projectStatusResponse, files: [linkedFile, { ...linkedFile, path: 'go.mod' }, ...projectStatusResponse.files] };

afterEach(() => vi.unstubAllGlobals());

function renderStatus(status: unknown) {
  const fetchMock = vi.fn(async (_url: string, init?: RequestInit) => Response.json(init?.method === 'POST' ? { ok: true } : status));
  vi.stubGlobal('fetch', fetchMock);
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(<QueryClientProvider client={client}><ProjectStatusView project={response.project} /></QueryClientProvider>);
  const stageBodies = () => fetchMock.mock.calls
    .filter(([url, init]) => init?.method === 'POST' && url.includes('/stage'))
    .map(([, init]) => JSON.parse(String(init?.body)));
  return { stageBodies };
}

describe('linked dependency staging', () => {
  it('disables linked manifests and excludes them from folder and stage-all requests', async () => {
    const { stageBodies } = renderStatus(response);
    const checkbox = await screen.findByRole('checkbox', { name: `Stage ${linkedFile.path}` });
    expect(checkbox.hasAttribute('disabled')).toBe(true);
    expect(checkbox.getAttribute('title')).toBe('Cannot stage dependencies linked outside this repository');
    const row = screen.getByRole('treeitem', { name: linkedFile.path });
    expect(within(row).getByText('Linked dependencies').title).toContain('@acme/ui → link:../../ui');

    fireEvent.click(screen.getByRole('checkbox', { name: 'Stage directory generated' }));
    await waitFor(() => expect(stageBodies()).toHaveLength(1));
    expect(stageBodies()[0]).toEqual({ paths: ['generated/output.log', 'generated/cache/data.json'], unstage: false });

    fireEvent.click(screen.getByRole('checkbox', { name: 'Stage all' }));
    await waitFor(() => expect(stageBodies()).toHaveLength(2));
    expect(stageBodies()[1]).toEqual({ paths: ['one.go', 'generated/output.log', 'generated/cache/data.json'], unstage: false });
  });

  it('still lets a manifest that gained a link after it was staged be unstaged', async () => {
    const staged = { ...response, files: response.files.map(file => file.path === 'go.mod' ? { ...file, state: 'staged' } : file) };
    const { stageBodies } = renderStatus(staged);

    const box = await screen.findByRole('checkbox', { name: 'Stage go.mod' }) as HTMLInputElement;
    expect(box.disabled).toBe(false);
    expect(box.checked).toBe(true);
    fireEvent.click(box);

    await waitFor(() => expect(stageBodies()).toHaveLength(1));
    expect(stageBodies()[0]).toEqual({ paths: ['go.mod'], unstage: true });
  });

  it('keeps mobile diffs reviewable while disabling staging', () => {
    render(<ProjectChangesView files={[linkedFile]} locked={new Map()} disabled={false} diffPath={linkedFile.path} onDiffPathChange={vi.fn()} onToggleStaged={vi.fn()} fileList={null} diff={<div>Manifest diff</div>} />);
    expect(screen.getByText('Manifest diff')).toBeTruthy();
    expect(screen.getByRole('checkbox', { name: `Stage ${linkedFile.path}` }).hasAttribute('disabled')).toBe(true);
    expect(screen.getByText('Linked dependencies')).toBeTruthy();
  });
});
