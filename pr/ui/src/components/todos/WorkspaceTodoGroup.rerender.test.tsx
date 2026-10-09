import type React from 'react';
import { render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Project, TodoItem, TodoListResponse } from '../../types';
import { WorkspaceTodoGroup } from './WorkspaceTodoGroup';
import { emptyCounts } from './format';

// The rows are stubbed so the test can observe exactly what the group hands
// them. The subject here is the group's prop stability, which is what decides
// whether the real TodoRow's memo holds; the row's own rendering is covered by
// format.test.tsx.
const rowProps: Array<Record<string, unknown>> = [];
vi.mock('./format', async (importOriginal) => {
  const actual = await importOriginal<Record<string, unknown>>();
  return {
    ...actual,
    TodoRow: (props: Record<string, unknown>) => {
      rowProps.push(props);
      return <div data-testid={`row-${String(props.todo && (props.todo as TodoItem).ref)}`} />;
    },
  };
});

vi.mock('../RepoIcon', () => ({ RepoIcon: () => <span /> }));

const workspace: Project = { name: 'gavel', dir: '/work/gavel', repos: [] };

function todo(ref: string, title: string): TodoItem {
  return { ref, title, status: 'pending', priority: 'medium' };
}

function data(items: TodoItem[]): TodoListResponse {
  return { dir: workspace.dir, counts: emptyCounts, items };
}

function renderGroup(props: Partial<React.ComponentProps<typeof WorkspaceTodoGroup>> = {}) {
  rowProps.length = 0;
  const onSelect = vi.fn();
  const utils = render(
    <WorkspaceTodoGroup
      workspace={workspace}
      data={data([todo('a', 'first'), todo('b', 'second')])}
      selectedRef=""
      onSelect={onSelect}
      {...props}
    />,
  );
  return { ...utils, onSelect };
}

describe('WorkspaceTodoGroup re-render cost', () => {
  // Every ambient update re-renders this group. If it hands its rows a new
  // onSelect each time, every row re-renders and commits DOM no matter how well
  // the row itself is memoised — which is what profiling measured as the
  // dominant main-thread cost on an idle todos page.
  it('hands rows the same onSelect identity across parent re-renders', () => {
    const { rerender, onSelect } = renderGroup();
    const first = rowProps.map(p => p.onSelect);

    rerender(
      <WorkspaceTodoGroup
        workspace={workspace}
        data={data([todo('a', 'first'), todo('b', 'second')])}
        selectedRef=""
        onSelect={onSelect}
      />,
    );

    const second = rowProps.slice(first.length).map(p => p.onSelect);
    expect(second).toHaveLength(first.length);
    second.forEach((handler, i) => expect(handler).toBe(first[i]));
  });

  it('passes the caller its own handler so selection needs no per-row closure', () => {
    const { onSelect } = renderGroup();

    expect(rowProps[0].onSelect).toBe(onSelect);

    (rowProps[0].onSelect as (t: { dir: string; ref: string }) => void)({ dir: workspace.dir, ref: 'a' });
    expect(onSelect).toHaveBeenCalledWith({ dir: workspace.dir, ref: 'a' });
  });

  // Rows used to be one branch of a ternary whose other branch was the empty
  // note, so the first arriving todo replaced one subtree with another and
  // React remounted the whole group. That remount is what shifted the layout as
  // late data landed.
  it('keeps the section element mounted when the first todos arrive', () => {
    const { container, rerender, onSelect } = renderGroup({ data: data([]) });
    const sectionBefore = container.firstElementChild;

    rerender(
      <WorkspaceTodoGroup
        workspace={workspace}
        data={data([todo('a', 'first')])}
        selectedRef=""
        onSelect={onSelect}
      />,
    );

    expect(container.firstElementChild).toBe(sectionBefore);
  });
});

describe('WorkspaceTodoGroup child todos', () => {
  const parentId = '0c4a6e8f-1b3d-4a5c-9e7f-2d4b6a8c0e1f';

  it('lists top-level todos only and does not count a child as filtered away', () => {
    const child = { ...todo('child', 'sub-task'), parentId };
    const { queryByTestId, getByText } = renderGroup({
      data: data([todo('parent', 'parent'), child]),
      filters: { statuses: { pending: 'exclude' }, priorities: {}, external: {}, tags: {}, workspaces: {} },
    });

    expect(queryByTestId('row-child')).toBeNull();
    expect(getByText('1 todo hidden by filter')).toBeTruthy();
  });

  it("hands a parent row its children's done/total and a childless row neither", () => {
    const kids = [
      { ...todo('kid-1', 'one'), parentId, status: 'completed' as const },
      { ...todo('kid-2', 'two'), parentId },
      { ...todo('kid-3', 'three'), parentId, status: 'verified' as const },
    ];
    renderGroup({ data: data([{ ...todo('parent', 'parent'), id: parentId }, todo('plain', 'plain'), ...kids]) });

    const byRef = Object.fromEntries(rowProps.map(p => [(p.todo as TodoItem).ref, p]));
    expect([byRef.parent.childDone, byRef.parent.childTotal]).toEqual([2, 3]);
    expect([byRef.plain.childDone, byRef.plain.childTotal]).toEqual([undefined, undefined]);
  });

  it('shows the empty note when a workspace holds nothing but children', () => {
    const { queryByTestId, getByText } = renderGroup({ data: data([{ ...todo('child', 'sub-task'), parentId }]) });

    expect(queryByTestId('row-child')).toBeNull();
    expect(getByText('No todos')).toBeTruthy();
  });
});
