import type React from 'react';
import type { ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { DropdownMenuItem } from '@flanksource/clicky-ui/components';
import { ProjectStatusView } from './ProjectStatusView';
import { RebaseActionContext } from './rebaseAction';

// The focus lease has its own spec (useGitFocus.test.tsx); here it would only
// add a POST to the requests these tests assert on.
vi.mock('../useGitFocus', () => ({ useGitFocus: () => ({ error: '' }) }));

/* oxlint-disable clicky-ui/prefer-clicky-components --
   These raw <button>s ARE the test doubles for clicky-ui's Button/SplitButton, not a
   rebuild of them: this factory replaces the '@flanksource/clicky-ui/components'
   module, so rendering clicky-ui's own Button here would be circular. The mock exists
   because the real SplitButton mounts floating-ui, which crashes under vitest. */
vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement>) => <button {...props}>{children}</button>,
  SplitButton: ({ label, onClick, items, disabled, title }: {
    label: ReactNode;
    onClick: () => void;
    items: { label: ReactNode; onSelect?: () => void }[];
    disabled?: boolean;
    title?: string;
  }) => (
    <div role="group" aria-label={title}>
      <button type="button" onClick={onClick} disabled={disabled}>{label}</button>
      {/* react-doctor-disable-next-line react-doctor/no-array-index-as-key -- test stub renders a fixed menu list */}
      {items.map((item, index) => <button key={index} type="button" onClick={item.onSelect} disabled={disabled}>{item.label}</button>)}
    </div>
  ),
  SplitPane: ({ left, right }: { left: ReactNode; right: ReactNode }) => <div>{left}{right}</div>,
  FilterBar: () => null,
}));
/* oxlint-enable clicky-ui/prefer-clicky-components */

vi.mock('./ProjectActionDialog', () => ({ ProjectActionDialog: () => null }));
vi.mock('./ProjectActionRunDialog', () => ({ ProjectActionRunDialog: () => null }));
vi.mock('./ProjectCommitTasks', () => ({ ProjectCommitTasks: () => null }));
vi.mock('./ProjectDiffView', () => ({ ProjectDiffView: () => null }));
vi.mock('./ProjectFileTree', () => ({ ProjectFileTree: () => null }));

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

const project = { name: 'gavel', dir: '/work/gavel', repos: ['acme/gavel'] };
const changedFile = 'one.go';

function renderProject(state = 'unstaged', rebaseAction: DropdownMenuItem | null = null) {
  const status = {
    project,
    workDir: project.dir,
    branch: 'feature/project-actions',
    files: [{
      path: changedFile,
      state,
      adds: 1,
      dels: 0,
      testStatus: { passed: 1, failed: 0, skipped: 0 },
      lintStatus: { errors: 0, warnings: 0, infos: 0 },
      resultsStale: false,
    }],
    resultsStale: false,
    action: { running: false },
  };
  const fetchMock = vi.fn(async (_input: RequestInfo | URL, init?: RequestInit) => (
    new Response(JSON.stringify(
      init?.method === 'POST' ? { runId: 'pr-run-1' } : status,
    ), {
      headers: { 'content-type': 'application/json' },
      status: 200,
    })
  ));
  vi.stubGlobal('fetch', fetchMock);
  render(
    <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
      <RebaseActionContext.Provider value={rebaseAction}>
        <ProjectStatusView project={project} />
      </RebaseActionContext.Provider>
    </QueryClientProvider>,
  );
  return fetchMock;
}

async function postedQueueRequest(fetchMock: ReturnType<typeof renderProject>) {
  await waitFor(() => expect(fetchMock.mock.calls.some(([, init]) => init?.method === 'POST')).toBe(true));
  const request = fetchMock.mock.calls.find(([, init]) => init?.method === 'POST');
  return { url: request?.[0], body: JSON.parse(String(request?.[1]?.body)) };
}

describe('Project Open PR action', () => {
  it('queues a push-only Open PR from the standalone button when nothing is staged', async () => {
    const fetchMock = renderProject();
    const openPR = await screen.findByRole('button', { name: 'Open PR' });
    expect(openPR.hasAttribute('disabled')).toBe(false);

    fireEvent.click(openPR);

    expect(await postedQueueRequest(fetchMock)).toEqual({
      url: '/api/projects/gavel/commit-queue',
      body: { action: 'open-pr', files: [] },
    });
  });

  it('commits the staged files and opens a PR from the standalone button', async () => {
    const fetchMock = renderProject('staged');
    fireEvent.click(await screen.findByRole('button', { name: 'Commit & open PR (1)' }));

    expect(await postedQueueRequest(fetchMock)).toEqual({
      url: '/api/projects/gavel/commit-queue',
      body: { action: 'open-pr', files: [changedFile] },
    });
  });

  it('offers the ref pane\'s Rebase… on the Open PR split button', async () => {
    const onSelect = vi.fn();
    renderProject('unstaged', { label: 'Rebase…', onSelect });
    const prGroup = await screen.findByRole('group', { name: 'Pull request options' });
    fireEvent.click(within(prGroup).getByRole('button', { name: 'Rebase…' }));
    expect(onSelect).toHaveBeenCalledTimes(1);
  });

  it('no longer offers Open PR inside the Commit split button', async () => {
    renderProject();
    const commitGroup = await screen.findByRole('group', { name: 'Commit options' });
    expect(within(commitGroup).queryByRole('button', { name: /open PR/i })).toBeNull();
  });
});
