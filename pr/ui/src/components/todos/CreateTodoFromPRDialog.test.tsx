import type React from 'react';
import { createContext, useContext } from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { PRDetail, PRItem, Project } from '../../types';
import { CreateTodoFromPRDialog } from './CreateTodoFromPRDialog';

interface Selection {
  isSelected: (key: string) => boolean;
  toggle: (key: string) => void;
}

const SelectionContext = createContext<Selection | null>(null);

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, loading: _loading, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement> & { loading?: boolean }) => <button {...props}>{children}</button>,
  Field: ({ label, children }: { label: string; children: React.ReactNode }) => <label>{label}{children}</label>,
  ListMenu: ({ children, selection }: { children: React.ReactNode; selection: Selection }) => (
    <SelectionContext.Provider value={selection}>{children}</SelectionContext.Provider>
  ),
  ListMenuHeader: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  ListMenuItem: ({ children, itemKey, onClick }: { children: React.ReactNode; itemKey: string; onClick?: () => void }) => {
    const selection = useContext(SelectionContext);
    if (!selection) throw new Error('ListMenuItem requires selection context');
    return (
      <label>
        <input
          type="checkbox"
          aria-label={`Include ${itemKey}`}
          checked={selection.isSelected(itemKey)}
          onChange={() => selection.toggle(itemKey)}
        />
        <button type="button" onClick={onClick}>{children}</button>
      </label>
    );
  },
  Modal: ({ open, title, children, footer }: { open: boolean; title: string; children: React.ReactNode; footer?: React.ReactNode }) => open ? <section aria-label={title}>{children}{footer}</section> : null,
  Select: (props: React.SelectHTMLAttributes<HTMLSelectElement>) => <select {...props} />,
  SplitPane: ({ left, right }: { left: React.ReactNode; right: React.ReactNode }) => <div>{left}{right}</div>,
  useListMenuSelection: ({
    selectedKeys,
    onSelectionChange,
  }: {
    keys: string[];
    selectedKeys: string[];
    onSelectionChange: (keys: string[]) => void;
  }) => ({
    isSelected: (key: string) => selectedKeys.includes(key),
    toggle: (key: string) => onSelectionChange(
      selectedKeys.includes(key)
        ? selectedKeys.filter(selected => selected !== key)
        : [...selectedKeys, key],
    ),
  }),
}));

vi.mock('@flanksource/clicky-ui/icons', async importOriginal => ({
  ...await importOriginal<typeof import('@flanksource/clicky-ui/icons')>(),
  UiBeaker: () => <span />,
  UiComment: () => <span />,
  UiError: () => <span />,
  UiLinkExternal: () => <span />,
  UiPass: () => <span />,
  UiWarningTriangle: () => <span />,
}));

vi.mock('../Markdown', () => ({
  Markdown: ({ text }: { text: string }) => <span>{text}</span>,
}));

vi.mock('../Avatar', () => ({
  Avatar: () => <span />,
}));

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

const pr: PRItem = {
  number: 17,
  title: 'Keep PR failure context',
  author: 'octocat',
  repo: 'acme/widget',
  source: 'failure-details',
  target: 'main',
  state: 'OPEN',
  isDraft: false,
  url: 'https://github.com/acme/widget/pull/17',
  updatedAt: '2026-07-26T12:00:00Z',
  checkStatus: {
    passed: 0,
    failed: 1,
    running: 0,
    pending: 0,
    failures: [{
      name: 'unit-tests',
      failedSteps: ['Run tests'],
      logTail: 'check log must stay out of the body',
    }],
  },
};

const detail: PRDetail = {
  gavelResults: [{
    artifactId: 101,
    artifactUrl: 'https://example.test/artifacts/101',
    testsPassed: 10,
    testsFailed: 1,
    testsSkipped: 0,
    testsTotal: 11,
    lintViolations: 1,
    lintLinters: 1,
    hasBench: false,
    topFailures: [{
      suite: 'storage',
      name: 'saves records',
      file: 'pkg/store/save_test.go',
      line: 41,
      message: 'expected record to persist',
      details: 'stderr: persistence failed',
    }],
    topLintViolations: [{
      linter: 'golangci-lint',
      rule: 'errcheck',
      file: 'pkg/store/save.go',
      line: 23,
      message: 'return value is not checked',
    }],
  }],
  comments: [{
    id: 901,
    body: 'review comment body must stay out of the failure details',
    author: 'reviewer',
    url: 'https://github.com/acme/widget/pull/17#discussion_r901',
    createdAt: '2026-07-26T12:30:00Z',
  }],
};

const workspaces: Project[] = [{
  name: 'gavel',
  dir: '/work/gavel',
  repos: ['acme/widget'],
}];

function createdResponse() {
  return {
    ok: true,
    json: async () => ({
      todo: {
        ref: 'todo/123',
        title: 'Address feedback on acme/widget#17',
        status: 'pending',
        priority: 'medium',
        criteria: [],
      },
    }),
  } as Response;
}

describe('CreateTodoFromPRDialog', () => {
  it('submits selected test and lint details with criteria and PR verification', async () => {
    const onCreated = vi.fn();
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) => createdResponse());
    vi.stubGlobal('fetch', fetchMock);
    render(
      <CreateTodoFromPRDialog
        open
        onClose={vi.fn()}
        pr={pr}
        detail={detail}
        workspaces={workspaces}
        onCreated={onCreated}
      />,
    );

    fireEvent.change(screen.getByLabelText('Notes'), { target: { value: 'Keep the public API stable.' } });
    fireEvent.click(screen.getByRole('button', { name: 'Add todo' }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
    expect(fetchMock.mock.calls[0]?.[0]).toBe('/api/todos/new?dir=%2Fwork%2Fgavel');
    const payload = JSON.parse(String(fetchMock.mock.calls[0]?.[1]?.body));
    expect(payload.body).toContain('Keep the public API stable.');
    expect(payload.body).toContain('## Failure details');
    expect(payload.body).toContain('expected record to persist');
    expect(payload.body).toContain('stderr: persistence failed');
    expect(payload.body).toContain('golangci-lint (errcheck)');
    expect(payload.body).toContain('pkg/store/save.go:23');
    expect(payload.body).toContain('return value is not checked');
    expect(payload.body).not.toContain('check log must stay out of the body');
    expect(payload.body).not.toContain('review comment body must stay out of the failure details');
    expect(payload.criteria).toEqual([
      { text: 'Test `storage › saves records` passes' },
      { text: 'Resolve golangci-lint (errcheck) violation at pkg/store/save.go:23' },
    ]);
    expect(payload.prVerification).toEqual({
      prNumber: 17,
      repo: 'acme/widget',
      actions: ['*'],
    });
    expect(onCreated).toHaveBeenCalledTimes(1);
    expect((await screen.findByRole('link', { name: 'Open todo' })).getAttribute('href')).toBe('/todos/todo/123');
  });

  it('uses the same deselected candidates for the body and criteria', async () => {
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) => createdResponse());
    vi.stubGlobal('fetch', fetchMock);
    render(
      <CreateTodoFromPRDialog
        open
        onClose={vi.fn()}
        pr={pr}
        detail={detail}
        workspaces={workspaces}
      />,
    );

    fireEvent.click(screen.getByLabelText('Include lint:0:0:pkg/store/save.go:23:errcheck'));
    fireEvent.click(screen.getByRole('button', { name: 'Add todo' }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
    const payload = JSON.parse(String(fetchMock.mock.calls[0]?.[1]?.body));
    expect(payload.body).toContain('expected record to persist');
    expect(payload.body).not.toContain('golangci-lint');
    expect(payload.body).not.toContain('return value is not checked');
    expect(payload.criteria).toEqual([
      { text: 'Test `storage › saves records` passes' },
    ]);
    expect(payload.prVerification).toEqual({
      prNumber: 17,
      repo: 'acme/widget',
      actions: ['*'],
    });
  });
});
