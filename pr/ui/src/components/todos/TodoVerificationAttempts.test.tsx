import type React from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { TodoSessionAttempt, TodoSessionDetailResponse } from '../../types';
import type { RunSnapshot } from '../tests/types';
import { TodoVerificationAttempts } from './TodoVerificationAttempts';

const runResultsCalls = vi.hoisted(() => ({ props: [] as { snapshot: RunSnapshot; runKey: string }[] }));
const snapshotFetch = vi.hoisted(() => ({ fn: vi.fn() }));

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({
    children,
    ...props
  }: React.ButtonHTMLAttributes<HTMLButtonElement>) => (
    // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for the Clicky Button itself.
    <button type="button" {...props}>
      {children}
    </button>
  ),
}));

vi.mock('@flanksource/clicky-ui/icons', async (importOriginal) => ({
  ...(await importOriginal<object>()),
  UiBeaker: (props: React.SVGProps<SVGSVGElement>) => <svg {...props} />,
  UiError: (props: React.SVGProps<SVGSVGElement>) => <svg {...props} />,
  UiPass: (props: React.SVGProps<SVGSVGElement>) => <svg {...props} />,
  UiStop: (props: React.SVGProps<SVGSVGElement>) => <svg {...props} />,
  UiWarningTriangle: (props: React.SVGProps<SVGSVGElement>) => <svg {...props} />,
}));

vi.mock('../tests/TestRunResults', () => ({
  TestRunResults: (props: { snapshot: RunSnapshot; runKey: string }) => {
    runResultsCalls.props.push(props);
    return <div data-testid="test-run-results">{props.runKey}</div>;
  },
}));

vi.mock('../tests/types', async (importOriginal) => ({
  ...(await importOriginal<object>()),
  fetchRunSnapshot: (opts: { dir?: string; runId: string }) => snapshotFetch.fn(opts),
}));

vi.mock('./VerificationAttemptSession', () => ({
  VerificationAttemptSession: ({ sessionId }: { sessionId: string }) => (
    <div data-testid="attempt-session">{sessionId}</div>
  ),
}));

function attempt(ordinal: number, overrides: Partial<TodoSessionAttempt> = {}): TodoSessionAttempt {
  return {
    promptRunId: `run-${ordinal}`,
    ordinal,
    step: 'verify',
    requested: {},
    resolved: {},
    status: 'succeeded',
    processActive: false,
    state: 'succeeded',
    phase: 'done',
    queuedAt: `2026-07-30T09:0${ordinal}:00Z`,
    startedAt: `2026-07-30T09:0${ordinal}:00Z`,
    admissionSessionId: `admission-${ordinal}`,
    createdAt: `2026-07-30T09:0${ordinal}:00Z`,
    updatedAt: `2026-07-30T09:0${ordinal}:30Z`,
    ...overrides,
  };
}

const testRun = {
  run_id: 'run-2026-07-30T09-00-00Z',
  kind: 'test',
  total: 12,
  passed: 10,
  failed: 2,
  warned: 0,
  skipped: 0,
  failures: [
    { name: 'renders the badge', suite: 'TodoDetail', status: 'FAIL', message: 'expected 2, got 0' },
    { name: 'lists attempts', suite: 'TodoDetail', status: 'FAIL' },
  ],
  truncated: 3,
};

function detailOf(...attempts: TodoSessionAttempt[]): TodoSessionDetailResponse {
  return { attempts, diagnostics: [] };
}

const snapshot: RunSnapshot = {
  metadata: { kind: 'test' },
  status: { running: false },
  tests: [],
} as unknown as RunSnapshot;

describe('TodoVerificationAttempts', () => {
  beforeEach(() => {
    runResultsCalls.props.length = 0;
    snapshotFetch.fn = vi.fn(async () => snapshot);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('lists attempts newest-first regardless of the order the endpoint returns', () => {
    const detail = detailOf(
      attempt(1, { resultJson: { definitionOfDone: { ran: true, passed: true } } }),
      attempt(2, { resultJson: { definitionOfDone: { ran: true, passed: false } } }),
    );
    render(<TodoVerificationAttempts dir="/workspace" todoRef="todo-1" detail={detail} />);

    const ordinals = screen.getAllByText(/^#\d+$/).map((node) => node.textContent);
    expect(ordinals).toEqual(['#2', '#1']);
  });

  it('distinguishes loading from an empty attempt list', () => {
    const { rerender } = render(<TodoVerificationAttempts dir="/workspace" todoRef="todo-1" detail={null} />);
    expect(screen.getByText('Loading attempts…')).toBeTruthy();

    rerender(<TodoVerificationAttempts dir="/workspace" todoRef="todo-1" detail={detailOf()} />);
    expect(screen.queryByText('Loading attempts…')).toBeNull();
    expect(screen.getByText(/No verification has run yet/)).toBeTruthy();
  });

  it('reports unreadable payloads instead of dropping the attempt silently', () => {
    const detail = detailOf(attempt(4, { resultJson: { definitionOfDone: 'nope' } }));
    render(<TodoVerificationAttempts dir="/workspace" todoRef="todo-1" detail={detail} />);

    expect(screen.getByText('1 attempt(s) could not be read:')).toBeTruthy();
    expect(screen.getByText(/#4 — definitionOfDone is string/)).toBeTruthy();
  });

  it('renders the recorded summary before the full snapshot arrives, then hands it to TestRunResults', async () => {
    let resolveSnapshot: (value: RunSnapshot) => void = () => {};
    snapshotFetch.fn = vi.fn(() => new Promise<RunSnapshot>((resolve) => { resolveSnapshot = resolve; }));
    const detail = detailOf(
      attempt(1, {
        resultJson: {
          definitionOfDone: {
            ran: true,
            passed: false,
            output: { results: [{ name: 'go test', type: 'test', status: 'FAIL', run: testRun }] },
          },
        },
      }),
    );
    render(<TodoVerificationAttempts dir="/workspace" todoRef="todo-1" detail={detail} />);

    // The counts and failure heads come from the attempt payload, so they must be
    // legible while the .gavel artifact is still in flight.
    expect(screen.getByText('10/12 passed · 2 failed')).toBeTruthy();
    expect(screen.getByText('TodoDetail > renders the badge')).toBeTruthy();
    expect(screen.getByText('…and 3 more')).toBeTruthy();
    expect(snapshotFetch.fn).toHaveBeenCalledWith({ dir: '/workspace', runId: testRun.run_id });

    resolveSnapshot(snapshot);
    await waitFor(() => expect(screen.getByTestId('test-run-results')).toBeTruthy());
    expect(runResultsCalls.props.at(-1)?.snapshot).toBe(snapshot);
    expect(runResultsCalls.props.at(-1)?.runKey).toBe(testRun.run_id);
  });

  it('keeps the recorded summary when the artifact file is gone', async () => {
    snapshotFetch.fn = vi.fn(async () => {
      throw new Error('run snapshot not found');
    });
    const detail = detailOf(
      attempt(1, {
        resultJson: {
          definitionOfDone: {
            ran: true,
            passed: false,
            output: { results: [{ name: 'go test', type: 'test', status: 'FAIL', run: testRun }] },
          },
        },
      }),
    );
    render(<TodoVerificationAttempts dir="/workspace" todoRef="todo-1" detail={detail} />);

    await waitFor(() => expect(screen.getByText('Full results unavailable: run snapshot not found')).toBeTruthy());
    expect(screen.getByText('10/12 passed · 2 failed')).toBeTruthy();
  });

  it('disables sub-tabs the attempt has no evidence for, with the reason', async () => {
    const detail = detailOf(
      attempt(1, {
        resultJson: {
          definitionOfDone: {
            ran: true,
            passed: false,
            output: { results: [{ name: 'go test', type: 'test', status: 'FAIL', run: testRun }] },
          },
        },
      }),
    );
    render(<TodoVerificationAttempts dir="/workspace" todoRef="todo-1" detail={detail} />);

    const lint = screen.getByRole('button', { name: /^Lint$/ });
    expect(lint.hasAttribute('disabled')).toBe(true);
    expect(lint.getAttribute('title')).toBe('This attempt ran no lint step');
    const session = screen.getByRole('button', { name: /^Session$/ });
    expect(session.hasAttribute('disabled')).toBe(true);
    expect(session.getAttribute('title')).toBe('This attempt recorded no agent session');
    await waitFor(() => expect(screen.getByTestId('test-run-results')).toBeTruthy());
  });

  it('shows the selected attempt session when one was recorded', () => {
    const detail = detailOf(
      attempt(1, {
        executionSessionId: 'session-abc',
        resultJson: {
          definitionOfDone: {
            ran: true,
            passed: false,
            output: { results: [{ name: 'smoke', type: 'exec', status: 'FAIL', command: 'make test' }] },
          },
        },
      }),
    );
    render(<TodoVerificationAttempts dir="/workspace" todoRef="todo-1" detail={detail} />);

    fireEvent.click(screen.getByRole('button', { name: /^Session$/ }));
    expect(screen.getByTestId('attempt-session').textContent).toBe('session-abc');
  });

  it('switches evidence when another attempt is selected', async () => {
    const detail = detailOf(
      attempt(1, {
        resultJson: {
          definitionOfDone: {
            ran: true,
            passed: false,
            output: { results: [{ name: 'first', type: 'exec', status: 'FAIL', command: 'make one' }] },
          },
        },
      }),
      attempt(2, {
        resultJson: {
          definitionOfDone: {
            ran: true,
            passed: true,
            output: { results: [{ name: 'second', type: 'exec', status: 'PASS', command: 'make two' }] },
          },
        },
      }),
    );
    render(<TodoVerificationAttempts dir="/workspace" todoRef="todo-1" detail={detail} />);

    expect(screen.getByText('make two')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: /#1/ }));
    await waitFor(() => expect(screen.getByText('make one')).toBeTruthy());
  });

  it('surfaces a fetch error for the attempt list itself', () => {
    render(<TodoVerificationAttempts dir="/workspace" todoRef="todo-1" detail={null} error="session detail unavailable" />);
    expect(screen.getByRole('alert').textContent).toBe('session detail unavailable');
  });
});
