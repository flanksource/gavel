import type React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { TracesStatus } from '@flanksource/gavel/testrunner';
import { TestRunResults } from './TestRunResults';
import type { RunSnapshot } from './types';

const processTrace = vi.hoisted(() => ({ options: [] as unknown[] }));

vi.mock('@flanksource/clicky-ui/data', async importOriginal => ({
  ...(await importOriginal<typeof import('@flanksource/clicky-ui/data')>()),
  TestRunner: () => <div>Test results</div>,
}));

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement>) => <button {...props}>{children}</button>,
}));

vi.mock('@flanksource/gavel/testrunner', async importOriginal => ({
  ...(await importOriginal<typeof import('@flanksource/gavel/testrunner')>()),
  useProcessTrace: (options: unknown) => {
    processTrace.options.push(options);
    return {
      nodes: [], samples: [], stacks: [], goprof: [], loading: false, canCollect: false,
      collecting: false, collectStack: async () => {}, refreshing: false, refreshFacets: async () => {},
    };
  },
  RunProcesses: ({ selectedKey }: { selectedKey: string | null }) => (
    <output data-testid="run-processes">{String(selectedKey)}</output>
  ),
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

describe('TestRunResults lint view', () => {
  const lintSnapshot: RunSnapshot = {
    status: { running: false, lint_run: true },
    tests: null,
    lint: [
      {
        linter: 'golangci-lint',
        success: false,
        violations: [{ rule: { method: 'errcheck' }, file: 'pkg/store/save.go', line: 23, message: 'return value is not checked' }],
      },
      { linter: 'eslint', skipped: true, violations: null },
    ],
  };

  it('renders the linters that ran through clicky LintResults, grouped linter then rule, and drops skipped ones', () => {
    render(<TestRunResults snapshot={lintSnapshot} done runKey="run-1" projectName="gavel" projectDir="/work/gavel" />);

    expect(screen.getByText('golangci-lint')).toBeTruthy();
    expect(screen.getByText('errcheck')).toBeTruthy();
    expect(screen.queryByText('eslint')).toBeNull();
  });

  it('says so when the run has a lint phase but no findings', () => {
    render(
      <TestRunResults
        snapshot={{ status: { running: false, lint_run: true }, tests: null, lint: [] }}
        done
        runKey="run-2"
        projectName="gavel"
        projectDir="/work/gavel"
      />,
    );

    expect(screen.getByText('No lint results')).toBeTruthy();
  });
});

describe('TestRunResults processes view', () => {
  const traces = {
    process_run: 'run-1',
    process: {
      tree: { kind: 'process_tree', stream: 't' },
      samples: { kind: 'process_samples', stream: 's' },
      stacks: { kind: 'stacks', stream: 'k' },
    },
    runs: { 'run-1': { mount: '/api/traces/k/query/run-1' } },
  } as TracesStatus;

  it('reads the process capture through the base of the server that reported it, and shows it in the Processes view', () => {
    render(
      <TestRunResults
        snapshot={{ ...failedSnapshot, status: { running: false, traces } }}
        traces={traces}
        base="/api/project-runs/run-9"
        done
        runKey="run-9"
        projectName="gavel"
        projectDir="/work/gavel"
      />,
    );

    expect(processTrace.options.at(-1)).toMatchObject({ traces, base: '/api/project-runs/run-9', done: true });
    fireEvent.click(screen.getByRole('button', { name: /Processes/ }));
    expect(screen.getByTestId('run-processes').textContent).toBe('null');
  });

  it('offers no Processes view for a run without a process capture', () => {
    render(<TestRunResults snapshot={failedSnapshot} done runKey="run-1" projectName="gavel" projectDir="/work/gavel" />);

    expect(screen.queryByRole('button', { name: /Processes/ })).toBeNull();
  });
});
