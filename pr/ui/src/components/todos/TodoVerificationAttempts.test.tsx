import type React from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { VerificationResultsProps, VerifyReport } from '@flanksource/clicky-ui/data';
import type { TodoSessionAttempt, TodoSessionDetailResponse } from '../../types';
import { TodoVerificationAttempts } from './TodoVerificationAttempts';
import { queryTestWrapper } from './queryTestWrapper';

const verificationResultsCalls = vi.hoisted(() => ({ props: [] as VerificationResultsProps[] }));

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

// VerificationResults itself (the shared TestRunner over a VerifyReport) is
// clicky-ui's own concern — already exercised by its own test suite — so it is
// mocked here to a probe that records what this component hands it. That keeps
// these tests focused on the adapter/wiring this component owns: which report
// reaches which attempt, and whether the run is treated as live.
vi.mock('@flanksource/clicky-ui/data', () => ({
  VerificationResults: (props: VerificationResultsProps) => {
    verificationResultsCalls.props.push(props);
    if (!props.report) return <div data-testid="verification-results">{props.emptyText}</div>;
    const failing = props.report.tests?.find(node => node.failed);
    return (
      <div data-testid="verification-results">
        <span data-testid="vr-state">{props.report.state}</span>
        {props.report.reason && <span data-testid="vr-reason">{props.report.reason}</span>}
        {failing?.context && <span data-testid="vr-cel">{JSON.stringify(failing.context)}</span>}
      </div>
    );
  },
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
    verification: null,
    ...overrides,
  };
}

function detailOf(...attempts: TodoSessionAttempt[]): TodoSessionDetailResponse {
  return { attempts, diagnostics: [] };
}

function report(overrides: Partial<VerifyReport> = {}): VerifyReport {
  return {
    kind: 'todo',
    ran: true,
    passed: true,
    state: 'passed',
    summary: { total: 1, passed: 1, failed: 0, warned: 0, skipped: 0, pending: 0, running: 0, timedout: 0 },
    ...overrides,
  };
}

function renderAttempts(detail: TodoSessionDetailResponse | null, error?: string) {
  return render(
    <TodoVerificationAttempts dir="/workspace" todoRef="todo-1" detail={detail} error={error} />,
    { wrapper: queryTestWrapper() },
  );
}

describe('TodoVerificationAttempts', () => {
  beforeEach(() => {
    verificationResultsCalls.props.length = 0;
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('lists attempts newest-first regardless of the order the endpoint returns', () => {
    const detail = detailOf(
      attempt(1, { verification: report() }),
      attempt(2, { verification: report({ passed: false, state: 'failed' }) }),
    );
    renderAttempts(detail);

    const ordinals = screen.getAllByText(/^#\d+$/).map((node) => node.textContent);
    expect(ordinals).toEqual(['#2', '#1']);
  });

  it('distinguishes loading from an empty attempt list', () => {
    const { rerender } = renderAttempts(null);
    expect(screen.getByText('Loading attempts…')).toBeTruthy();

    rerender(<TodoVerificationAttempts dir="/workspace" todoRef="todo-1" detail={detailOf()} />);
    expect(screen.queryByText('Loading attempts…')).toBeNull();
    expect(screen.getByText(/No verification has run yet/)).toBeTruthy();
  });

  it('renders a passing report through VerificationResults', () => {
    const detail = detailOf(attempt(1, { verification: report() }));
    renderAttempts(detail);

    expect(screen.getByTestId('vr-state').textContent).toBe('passed');
    expect(verificationResultsCalls.props.at(-1)?.report).toEqual(report());
  });

  it('surfaces a failing CEL leaf so its context reaches the renderer', () => {
    const failing = report({
      passed: false,
      state: 'failed',
      summary: { total: 1, passed: 0, failed: 1, warned: 0, skipped: 0, pending: 0, running: 0, timedout: 0 },
      tests: [{
        name: 'file count matches',
        framework: 'fixture',
        failed: true,
        context: { cel_expression: 'files.size() == 3', cel_vars: { 'files.size()': 2 }, expected: 3, actual: 2 },
      }],
    });
    const detail = detailOf(attempt(1, { verification: failing }));
    renderAttempts(detail);

    const cel = screen.getByTestId('vr-cel');
    expect(cel.textContent).toContain('files.size() == 3');
    expect(cel.textContent).toContain('"actual":2');
  });

  it('opens a running attempt on the live snapshot', () => {
    const running = report({ state: 'running', passed: false });
    const detail = detailOf(attempt(1, { state: 'running', status: 'running', verification: running }));
    renderAttempts(detail);

    expect(screen.getByTestId('vr-state').textContent).toBe('running');
  });

  it('shows an errored report reason', () => {
    const errored = report({
      passed: false,
      ran: false,
      state: 'errored',
      reason: 'no verification fixture, acceptance criteria, or configured checks',
    });
    const detail = detailOf(attempt(1, { verification: errored }));
    renderAttempts(detail);

    expect(screen.getByTestId('vr-reason').textContent).toBe('no verification fixture, acceptance criteria, or configured checks');
  });

  it('shows the selected attempt session when one was recorded', () => {
    const detail = detailOf(attempt(1, { executionSessionId: 'session-abc', verification: report() }));
    renderAttempts(detail);

    fireEvent.click(screen.getByRole('button', { name: /^Session$/ }));
    expect(screen.getByTestId('attempt-session').textContent).toBe('session-abc');
  });

  it('has no session tab when the attempt recorded no session', () => {
    const detail = detailOf(attempt(1, { verification: report() }));
    renderAttempts(detail);

    expect(screen.queryByRole('button', { name: /^Session$/ })).toBeNull();
  });

  it('switches evidence when another attempt is selected', async () => {
    const detail = detailOf(
      attempt(1, { verification: report({ passed: false, state: 'failed' }) }),
      attempt(2, { verification: report() }),
    );
    renderAttempts(detail);

    expect(screen.getByTestId('vr-state').textContent).toBe('passed');
    fireEvent.click(screen.getByRole('button', { name: /#1/ }));
    await waitFor(() => expect(screen.getByTestId('vr-state').textContent).toBe('failed'));
  });

  it('surfaces a fetch error for the attempt list itself', () => {
    renderAttempts(null, 'session detail unavailable');
    expect(screen.getByRole('alert').textContent).toBe('session detail unavailable');
  });
});
