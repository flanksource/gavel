import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { PRItem } from '../../types';
import { MenubarView } from './MenubarView';

// The layout test only needs the shell's own containers; every panel it hosts
// is replaced by a marker so the roots can be found and measured by class.
vi.mock('../PRList', () => ({ PRList: () => <div>pr-list</div> }));
vi.mock('../PRDetail', () => ({ PRDetailPanel: () => <div>pr-detail</div> }));
vi.mock('../MenubarTodos', () => ({ MenubarTodos: () => <div>todos</div> }));
vi.mock('../OrgChooser', () => ({ OrgChooser: () => <div>org</div> }));
vi.mock('../ProjectsPlaceholder', () => ({ ProjectsPlaceholder: () => <div>placeholder</div> }));
vi.mock('../ProcessTable', () => ({ WorkspaceGroup: () => <div>workspace</div> }));
vi.mock('../sessions/SessionsView', () => ({ SessionsSidebar: () => <div>sessions</div> }));
vi.mock('../sessions/MobileSessionScreen', () => ({ MobileSessionScreen: () => <div data-testid="session-screen" /> }));
vi.mock('../../procStatusQuery', () => ({ useProcStatus: () => ({}) }));
vi.mock('@flanksource/clicky-ui/ai', () => ({ ChatButton: () => <div>chat-button</div> }));

const pr = { repo: 'acme/widgets', number: 7, title: 'Fix parser', url: 'https://example.test/pr/7' } as PRItem;

const classes = (el: Element | null) => [...(el?.classList ?? [])];

function renderShell(overrides: Partial<Parameters<typeof MenubarView>[0]> = {}) {
  return render(
    <MenubarView
      prs={[]}
      selected={null}
      detail={null}
      detailLoading={false}
      unread={{}}
      config={{} as never}
      projects={[]}
      availableProjects={[]}
      projectsLoaded
      projectsByRepo={{}}
      syncStatus={{}}
      gavelResults={{}}
      onSelect={vi.fn()}
      onBack={vi.fn()}
      onScopeChange={vi.fn()}
      onProcChanged={vi.fn()}
      fetchedAt=""
      tab="prs"
      onTabChange={vi.fn()}
      sessions={{ data: undefined, error: null } as never}
      sessionId=""
      onSessionChange={vi.fn()}
      chatEnabled
      {...overrides}
    />,
  );
}

describe('MenubarView full-screen roots', () => {
  it('sizes the tab shell to the dynamic viewport and clips it so the document cannot scroll', () => {
    const { container } = renderShell();

    expect(classes(container.firstElementChild)).toEqual(expect.arrayContaining(['h-dvh', 'overflow-hidden', 'flex-col']));
    expect(classes(container.firstElementChild)).not.toContain('h-screen');
  });

  it('sizes the pull request screen and its body to the dynamic viewport', () => {
    const { container } = renderShell({ selected: pr });

    const root = container.firstElementChild as HTMLElement;
    expect(classes(root)).toEqual(expect.arrayContaining(['h-dvh', 'overflow-hidden']));
    expect(root.innerHTML).toContain('h-[calc(100dvh-2.75rem)]');
    expect(root.innerHTML).not.toContain('100vh');
  });

  it('hands a session id to the session screen, which owns its own dynamic-viewport root', () => {
    const { container } = renderShell({ tab: 'sessions', sessionId: 'sess-1' });

    expect(screen.getByTestId('session-screen')).toBeTruthy();
    expect(container.firstElementChild).toBe(screen.getByTestId('session-screen'));
  });
});
