import { cleanup, fireEvent, render, screen, within } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { ProjectFileTree } from './ProjectFileTree';
import { projectStatusResponse } from './ProjectStatusView.fixture';
import type { ProjectFileStatus } from './ProjectStatusView';

vi.mock('@flanksource/clicky-ui/components', async importOriginal => {
  const doubles = await import('./projectTreeDoubles');
  return {
    ...await importOriginal<typeof import('@flanksource/clicky-ui/components')>(),
    DropdownMenu: doubles.DropdownMenuDouble,
    ConfirmDialog: doubles.ConfirmDialogDouble,
  };
});

const file = (path: string, state: ProjectFileStatus['state'], extra: Partial<ProjectFileStatus> = {}): ProjectFileStatus => ({
  ...projectStatusResponse.files[0]!,
  path,
  state,
  ...extra,
} as ProjectFileStatus);

function renderTree(files: ProjectFileStatus[], props: Partial<React.ComponentProps<typeof ProjectFileTree>> = {}) {
  const handlers = {
    onStage: vi.fn(),
    onDiscard: vi.fn(),
    onIgnore: vi.fn(),
    onOpenInEditor: vi.fn(),
    onCopyPath: vi.fn(),
    onComment: vi.fn(),
    onRunPrompt: vi.fn(),
  };
  render(
    <ProjectFileTree
      files={files}
      search=""
      onSearchChange={vi.fn()}
      disabled={false}
      showResults={false}
      editors={[]}
      {...handlers}
      {...props}
    />,
  );
  return handlers;
}

// A folder's treeitem wraps its children, so scope to the item's own row.
const row = (path: string) => within(screen.getByRole('treeitem', { name: path }).firstElementChild as HTMLElement);
const agentItems = ['Add comment…', 'Run prompt…'];
const menuLabels = (path: string) => row(path).getAllByRole('menuitem').map(item => item.textContent).slice(agentItems.length);

afterEach(cleanup);

describe('ProjectFileTree line counts', () => {
  it('keeps assets stageable without displaying empty line counts', () => {
    const source = projectStatusResponse.files[0]!;
    renderTree([
      source as ProjectFileStatus,
      file('image.svg', 'unstaged', { adds: 0, dels: 0 }),
      file('report.pdf', 'unstaged', { adds: 0, dels: 0 }),
    ]);
    const sourceRow = row(source.path);
    expect(sourceRow.getByText(`+${source.adds}`)).toBeTruthy();
    expect(sourceRow.getByText(`−${source.dels}`)).toBeTruthy();
    for (const path of ['image.svg', 'report.pdf']) {
      expect(row(path).getByRole('checkbox', { name: `Stage ${path}` })).toBeTruthy();
      expect(row(path).queryByText('+0')).toBeNull();
      expect(row(path).queryByText('−0')).toBeNull();
    }
  });
});

describe('ProjectFileTree compact rows', () => {
  it('shows each file state as a one-letter status token', () => {
    renderTree([file('one.go', 'unstaged'), file('c.go', 'conflict'), file('u.log', 'untracked'), file('s.go', 'staged'), file('b.go', 'both')]);
    const expected = [
      { path: 'one.go', letter: 'M', label: 'Modified, not staged' },
      { path: 'c.go', letter: 'C', label: 'Conflict' },
      { path: 'u.log', letter: 'U', label: 'Untracked' },
      { path: 's.go', letter: 'S', label: 'Staged' },
      { path: 'b.go', letter: 'M', label: 'Partly staged' },
    ];
    for (const { path, letter, label } of expected) {
      expect(row(path).getByTitle(label).textContent).toBe(letter);
    }
  });
});

describe('ProjectFileTree stage checkboxes', () => {
  it.each([
    { state: 'staged' as const, checked: true, indeterminate: false },
    { state: 'both' as const, checked: false, indeterminate: true },
    { state: 'unstaged' as const, checked: false, indeterminate: false },
    { state: 'untracked' as const, checked: false, indeterminate: false },
  ])('renders a $state file as checked=$checked indeterminate=$indeterminate', ({ state, checked, indeterminate }) => {
    renderTree([file('a.go', state)]);
    const box = row('a.go').getByRole('checkbox', { name: 'Stage a.go' }) as HTMLInputElement;
    expect([box.checked, box.indeterminate]).toEqual([checked, indeterminate]);
  });

  it.each([
    { state: 'unstaged' as const, unstage: false },
    { state: 'untracked' as const, unstage: false },
    { state: 'both' as const, unstage: false },
    { state: 'staged' as const, unstage: true },
  ])('sends a file in state $state to stage with unstage=$unstage', ({ state, unstage }) => {
    const { onStage } = renderTree([file('a.go', state)]);
    fireEvent.click(row('a.go').getByRole('checkbox', { name: 'Stage a.go' }));
    expect(onStage).toHaveBeenCalledWith(['a.go'], unstage);
  });

  it('aggregates a folder over its descendants and stages only the unstaged ones', () => {
    const { onStage } = renderTree([file('pkg/a.go', 'staged'), file('pkg/b.go', 'unstaged'), file('pkg/c.go', 'untracked')]);
    const box = row('pkg').getByRole('checkbox', { name: 'Stage directory pkg' }) as HTMLInputElement;
    expect([box.checked, box.indeterminate]).toEqual([false, true]);
    fireEvent.click(box);
    expect(onStage).toHaveBeenCalledWith(['pkg/b.go', 'pkg/c.go'], false);
  });

  it('unstages every staged file when a folder is fully staged', () => {
    const { onStage } = renderTree([file('pkg/a.go', 'staged'), file('pkg/b.go', 'staged')]);
    const box = row('pkg').getByRole('checkbox', { name: 'Stage directory pkg' }) as HTMLInputElement;
    expect(box.checked).toBe(true);
    fireEvent.click(box);
    expect(onStage).toHaveBeenCalledWith(['pkg/a.go', 'pkg/b.go'], true);
  });

  it('disables a conflicted file with the resolve-conflicts reason and skips conflicts when staging a folder', () => {
    const { onStage } = renderTree([file('pkg/c.go', 'conflict'), file('pkg/a.go', 'unstaged')]);
    const conflicted = row('pkg/c.go').getByRole('checkbox', { name: 'Stage pkg/c.go' });
    expect(conflicted.hasAttribute('disabled')).toBe(true);
    expect(conflicted.getAttribute('title')).toBe('Resolve conflicts before staging');
    fireEvent.click(row('pkg').getByRole('checkbox', { name: 'Stage directory pkg' }));
    expect(onStage).toHaveBeenCalledWith(['pkg/a.go'], false);
  });

  it('disables files claimed by a queued commit group', () => {
    renderTree([file('a.go', 'staged')], { locked: new Map([['a.go', 2]]) });
    expect(row('a.go').getByRole('checkbox', { name: 'Stage a.go' }).hasAttribute('disabled')).toBe(true);
    expect(row('a.go').getByText('queued in group 2')).toBeTruthy();
  });

  it('puts a tri-state Stage all checkbox in the filter toolbar that stages every visible stageable file', () => {
    const { onStage } = renderTree([file('a.go', 'staged'), file('b.go', 'unstaged'), file('c.go', 'conflict')]);
    const all = screen.getByRole('checkbox', { name: 'Stage all' }) as HTMLInputElement;
    expect([all.checked, all.indeterminate]).toEqual([false, true]);
    expect(all.closest('[role="presentation"]')?.querySelector('input[type="search"]')).not.toBeNull();
    fireEvent.click(all);
    expect(onStage).toHaveBeenCalledWith(['b.go'], false);
  });

  it('renames the toolbar checkbox to Unstage all once everything is staged', () => {
    const { onStage } = renderTree([file('a.go', 'staged'), file('b.go', 'staged')]);
    const all = screen.getByRole('checkbox', { name: 'Unstage all' }) as HTMLInputElement;
    expect(all.checked).toBe(true);
    fireEvent.click(all);
    expect(onStage).toHaveBeenCalledWith(['a.go', 'b.go'], true);
  });
});

describe('ProjectFileTree row menu', () => {
  const editors = [{ id: 'vscode', label: 'VS Code' }, { id: 'zed', label: 'Zed' }];

  it('offers copy, editors, delete and ignore items for an untracked file', () => {
    renderTree([file('generated/out.log', 'untracked')], { editors });
    expect(menuLabels('generated/out.log')).toEqual([
      'Copy path', 'Open in VS Code', 'Open in Zed', 'Delete',
      'Ignore out.log', 'Ignore all *.log files',
    ]);
    const ignore = row('generated/out.log').getByRole('menuitem', { name: 'Ignore out.log' });
    expect(ignore.hasAttribute('disabled')).toBe(false);
  });

  it('offers discard and disables path ignoring for a tracked file', () => {
    renderTree([file('one.go', 'unstaged')], { editors: [] });
    expect(menuLabels('one.go')).toEqual(['Copy path', 'Discard changes', 'Ignore one.go', 'Ignore all *.go files']);
    const ignore = row('one.go').getByRole('menuitem', { name: 'Ignore one.go' });
    expect(ignore.hasAttribute('disabled')).toBe(true);
    expect(ignore.getAttribute('title')).toBe('.gitignore only hides untracked paths — tracked changes stay');
  });

  it('offers a folder ignore that is enabled only when every descendant is untracked', () => {
    renderTree([file('gen/a.log', 'untracked'), file('gen/b.log', 'untracked'), file('src/x.go', 'unstaged'), file('src/y.log', 'untracked')], { editors });
    expect(menuLabels('gen')).toEqual(['Copy path', 'Open in VS Code', 'Open in Zed', 'Delete', 'Ignore gen/']);
    expect(row('gen').getByRole('menuitem', { name: 'Ignore gen/' }).hasAttribute('disabled')).toBe(false);
    expect(row('src').getByRole('menuitem', { name: 'Ignore src/' }).hasAttribute('disabled')).toBe(true);
    expect(row('src').getByRole('menuitem', { name: 'Discard changes' })).toBeTruthy();
  });

  it('has no ignore-by-extension item for dotfiles and extensionless names', () => {
    renderTree([file('.gitignore', 'unstaged'), file('Makefile', 'unstaged')]);
    expect(menuLabels('.gitignore')).toEqual(['Copy path', 'Discard changes', 'Ignore .gitignore']);
    expect(menuLabels('Makefile')).toEqual(['Copy path', 'Discard changes', 'Ignore Makefile']);
  });

  it('leads every menu with comment and run-prompt requests for the row path', () => {
    const { onComment, onRunPrompt } = renderTree([file('src/x.go', 'unstaged')]);
    expect(row('src').getAllByRole('menuitem').slice(0, agentItems.length).map(item => item.textContent)).toEqual(agentItems);
    fireEvent.click(row('src').getByRole('menuitem', { name: 'Add comment…' }));
    expect(onComment).toHaveBeenCalledWith({ path: 'src', folder: true });
    fireEvent.click(row('src/x.go').getByRole('menuitem', { name: 'Run prompt…' }));
    expect(onRunPrompt).toHaveBeenCalledWith({ path: 'src/x.go', folder: false });
  });

  it('dispatches copy, open-in-editor and ignore requests', () => {
    const { onCopyPath, onOpenInEditor, onIgnore } = renderTree([file('gen/a.log', 'untracked'), file('gen/b.log', 'untracked')], { editors });
    fireEvent.click(row('gen/a.log').getByRole('menuitem', { name: 'Copy path' }));
    expect(onCopyPath).toHaveBeenCalledWith('gen/a.log');
    fireEvent.click(row('gen/a.log').getByRole('menuitem', { name: 'Open in Zed' }));
    expect(onOpenInEditor).toHaveBeenCalledWith('gen/a.log', 'zed');
    fireEvent.click(row('gen/a.log').getByRole('menuitem', { name: 'Ignore all *.log files' }));
    expect(onIgnore).toHaveBeenLastCalledWith({ path: 'gen/a.log', extension: '.log' });
    fireEvent.click(row('gen/a.log').getByRole('menuitem', { name: 'Ignore a.log' }));
    expect(onIgnore).toHaveBeenLastCalledWith({ path: 'gen/a.log' });
    fireEvent.click(row('gen').getByRole('menuitem', { name: 'Ignore gen/' }));
    expect(onIgnore).toHaveBeenLastCalledWith({ path: 'gen', directory: true });
  });

  it.each([
    { path: 'one.go', state: 'unstaged' as const, item: 'Discard changes', title: 'Discard changes to one.go?', confirm: 'Discard changes' },
    { path: 'gen/a.log', state: 'untracked' as const, item: 'Delete', title: 'Delete gen/a.log?', confirm: 'Delete' },
  ])('confirms before discarding $path and sends nothing until confirmed', ({ path, state, item, title, confirm }) => {
    const { onDiscard } = renderTree([file(path, state)]);
    fireEvent.click(row(path).getByRole('menuitem', { name: item }));
    expect(onDiscard).not.toHaveBeenCalled();
    const dialog = screen.getByRole('alertdialog', { name: title });
    expect(within(dialog).getByText('This cannot be undone.')).toBeTruthy();
    fireEvent.click(within(dialog).getByRole('button', { name: 'Cancel' }));
    expect(onDiscard).not.toHaveBeenCalled();
    fireEvent.click(row(path).getByRole('menuitem', { name: item }));
    fireEvent.click(within(screen.getByRole('alertdialog')).getByRole('button', { name: confirm }));
    expect(onDiscard).toHaveBeenCalledWith(path);
  });

  it('asks about the whole folder when discarding a directory', () => {
    renderTree([file('src/x.go', 'unstaged')]);
    fireEvent.click(row('src').getByRole('menuitem', { name: 'Discard changes' }));
    expect(screen.getByRole('alertdialog', { name: 'Discard changes to src/?' })).toBeTruthy();
  });

  // The trigger is hidden only where there is a pointer that hovers and room for
  // it (md and up), so touch and phone widths never depend on a variant that
  // would have to out-rank a base opacity-0. Hover, focus and an open menu
  // reveal it with higher specificity.
  it('hides the menu trigger only for hovering pointers at md and up, and keeps a 32px target', () => {
    renderTree([file('one.go', 'unstaged')]);
    const classes = row('one.go').getByRole('button', { name: 'Actions for one.go' }).className.split(/\s+/);
    expect(classes).toEqual(expect.arrayContaining(['size-8', 'md:[@media(hover:hover)]:opacity-0', 'group-hover:opacity-100', 'focus-visible:opacity-100', 'aria-expanded:opacity-100']));
    expect(classes).not.toContain('opacity-0');
  });
});
