import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { ProjectStatusView, type ProjectFileStatus } from './ProjectStatusView';
import { ProjectChangesView } from './ProjectChangesView';
import { projectStatusResponse } from './ProjectStatusView.fixture';
import { queryKeys } from '../query';

vi.mock('../useGitFocus', () => ({ useGitFocus: () => ({ error: '' }) }));
vi.mock('./ProjectCommitTasks', () => ({ ProjectCommitTasks: () => null }));
vi.mock('../useIsMobile', () => ({ useIsMobile: () => true }));

const linkedFile: ProjectFileStatus = {
  ...projectStatusResponse.files[0], state: 'unstaged', path: 'generated/package.json',
  linkedDependencies: [{ name: '@acme/ui', target: 'link:../../ui' }],
};
const response = { ...projectStatusResponse, files: [linkedFile, { ...linkedFile, path: 'go.mod' }, ...projectStatusResponse.files] };

afterEach(() => vi.unstubAllGlobals());

describe('linked dependency selection', () => {
  it('disables linked manifests and excludes them from folder and select-all selection', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify(response))));
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(<QueryClientProvider client={client}><ProjectStatusView project={response.project} /></QueryClientProvider>);
    const checkbox = await screen.findByRole('checkbox', { name: `Select ${linkedFile.path}` });
    expect(checkbox.hasAttribute('disabled')).toBe(true);
    const row = screen.getByRole('treeitem', { name: linkedFile.path });
    expect(within(row).getByText('Linked dependencies').title).toContain('@acme/ui → link:../../ui');
    fireEvent.click(screen.getByRole('checkbox', { name: 'Select directory generated' }));
    expect(screen.getByRole('button', { name: 'Commit selected (2)' })).toBeTruthy();
    expect((checkbox as HTMLInputElement).checked).toBe(false);
    fireEvent.click(screen.getByRole('checkbox', { name: 'Select all files' }));
    expect(screen.getByRole('button', { name: 'Commit selected (3)' })).toBeTruthy();
    expect((screen.getByRole('checkbox', { name: 'Select go.mod' }) as HTMLInputElement).checked).toBe(false);
  });

  it('removes an existing selection when a refreshed manifest gains a link', async () => {
    const initial = { ...response, files: response.files.map(file => ({ ...file, linkedDependencies: [] })) };
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify(initial))));
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(<QueryClientProvider client={client}><ProjectStatusView project={response.project} /></QueryClientProvider>);
    fireEvent.click(await screen.findByRole('checkbox', { name: `Select ${linkedFile.path}` }));
    expect(screen.getByRole('button', { name: 'Commit selected (1)' })).toBeTruthy();
    await act(async () => { client.setQueryData(queryKeys.projectStatus(response.project.name, false), response); });
    expect(await screen.findByRole('button', { name: 'Commit selected (0)' })).toBeTruthy();
  });

  it('keeps mobile diffs reviewable while disabling inclusion in a commit', () => {
    const toggle = vi.fn();
    render(<ProjectChangesView files={[linkedFile]} selected={new Set()} locked={new Map()} disabled={false} diffPath={linkedFile.path} onDiffPathChange={vi.fn()} onToggleFile={toggle} fileList={null} diff={<div>Manifest diff</div>} />);
    expect(screen.getByText('Manifest diff')).toBeTruthy();
    expect(screen.getByRole('checkbox', { name: `Include ${linkedFile.path} in commit` }).hasAttribute('disabled')).toBe(true);
    expect(screen.getByText('Linked dependencies')).toBeTruthy();
  });
});
