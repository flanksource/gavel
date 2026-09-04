import type React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { TestRunResults } from './TestRunResults';
import type { RunSnapshot } from './types';

vi.mock('@flanksource/clicky-ui/data', () => ({
  TestRunner: () => <div>Test results</div>,
  emptyTestFilters: () => ({ status: [], framework: [] }),
  filterTests: (tests: unknown[]) => tests,
}));

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement>) => <button {...props}>{children}</button>,
}));

vi.mock('./CreateTodoFromRunDialog', () => ({
  CreateTodoFromRunDialog: ({ open, projectName, projectDir, runId, candidates }: {
    open: boolean;
    projectName: string;
    projectDir: string;
    runId: string;
    candidates: unknown[];
  }) => open ? <output data-testid="run-todo-dialog">{projectName}|{projectDir}|{runId}|{candidates.length}</output> : null,
}));

const failedSnapshot: RunSnapshot = {
  status: { running: false },
  tests: [{ name: 'TestSave', failed: true }],
};

describe('TestRunResults failure todo action', () => {
  it('opens the grouped todo dialog for a completed failing run', () => {
    render(
      <TestRunResults
        snapshot={failedSnapshot}
        done
        runKey="run-123"
        projectName="gavel"
        projectDir="/work/gavel"
      />,
    );

    fireEvent.click(screen.getByRole('button', { name: 'Create todo' }));

    expect(screen.getByTestId('run-todo-dialog').textContent).toBe('gavel|/work/gavel|run-123|1');
  });

  it('does not offer a todo until the run is complete or when it has no failures', () => {
    const { rerender } = render(
      <TestRunResults
        snapshot={failedSnapshot}
        done={false}
        runKey="run-123"
        projectName="gavel"
        projectDir="/work/gavel"
      />,
    );
    expect(screen.queryByRole('button', { name: 'Create todo' })).toBeNull();

    rerender(
      <TestRunResults
        snapshot={{ status: { running: false }, tests: [{ name: 'TestSave', passed: true }] }}
        done
        runKey="run-124"
        projectName="gavel"
        projectDir="/work/gavel"
      />,
    );
    expect(screen.queryByRole('button', { name: 'Create todo' })).toBeNull();
  });
});
