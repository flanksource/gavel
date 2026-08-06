import type React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { VerificationFixtureResult } from './verificationAttempts';
import { VerificationStepOutput } from './VerificationStepOutput';

const logViewerCalls = vi.hoisted(() => ({ props: [] as { logs: string; collapsedLines?: number }[] }));

vi.mock('@flanksource/clicky-ui/icons', async (importOriginal) => ({
  ...(await importOriginal<object>()),
  UiError: (props: React.SVGProps<SVGSVGElement>) => <svg data-testid="icon-fail" {...props} />,
  UiPass: (props: React.SVGProps<SVGSVGElement>) => <svg data-testid="icon-pass" {...props} />,
}));

vi.mock('../LogViewer', () => ({
  LogViewer: (props: { logs: string; collapsedLines?: number }) => {
    logViewerCalls.props.push(props);
    return <div data-testid="log-viewer">{props.logs}</div>;
  },
}));

const failingExec: VerificationFixtureResult = {
  name: 'smoke test',
  type: 'exec',
  status: 'FAIL',
  duration: 2_500_000_000,
  command: 'go test ./todos/...',
  cwd: '/workspace',
  exit_code: 2,
  stdout: '[31mFAIL[0m todos 2.5s\n',
  stderr: 'exit status 1\n',
  cel_expression: 'results.failed == 0',
  cel_vars: { failed: 3, changed_files: ['todos/runtime/lifecycle.go'] },
  expected: 0,
  actual: 3,
};

describe('VerificationStepOutput', () => {
  it('renders the command, working dir, duration and exit code of a step', () => {
    render(<VerificationStepOutput steps={[failingExec]} checklist={[]} />);

    expect(screen.getByText('go test ./todos/...')).toBeTruthy();
    expect(screen.getByText('in /workspace')).toBeTruthy();
    // Go marshals durations as nanoseconds; 2.5e9 ns is 2.5s, not 2500000000ms.
    expect(screen.getByText('2.5s')).toBeTruthy();
    expect(screen.getByText('exit 2')).toBeTruthy();
  });

  it('routes stdout and stderr through LogViewer rather than a raw pre block', () => {
    logViewerCalls.props.length = 0;
    render(<VerificationStepOutput steps={[failingExec]} checklist={[]} />);

    expect(logViewerCalls.props.map((props) => props.logs)).toEqual([failingExec.stdout, failingExec.stderr]);
    expect(screen.getByText('stdout')).toBeTruthy();
    expect(screen.getByText('stderr')).toBeTruthy();
  });

  it('explains a failing verdict with the CEL expression and the values it saw', () => {
    render(<VerificationStepOutput steps={[failingExec]} checklist={[]} />);

    expect(screen.getByText('results.failed == 0')).toBeTruthy();
    const varRow = screen.getByText('failed').closest('tr');
    expect(varRow?.textContent).toBe('failed3');
    expect(screen.getByText('["todos/runtime/lifecycle.go"]')).toBeTruthy();
    // The verdict pair: expected 0, actual 3.
    expect(screen.getByText('expected').parentElement?.textContent).toBe('expected 0');
    expect(screen.getByText('actual').parentElement?.textContent).toBe('actual 3');
  });

  it('prefers the native CEL trace over the plain expression', () => {
    const trace = 'cel: results.failed == 0\n     │              │\n     │              └─ int(0)\n     └─ int(3)';
    render(<VerificationStepOutput steps={[{ ...failingExec, cel_trace: trace }]} checklist={[]} />);

    expect(screen.getByText((_, element) => element?.textContent === trace)).toBeTruthy();
    expect(screen.queryByText('results.failed == 0')).toBeNull();
  });

  // A command step with no CEL assertion records `actual` as the whole command
  // result object — the card already renders every field of it above.
  it('does not re-dump the raw command record a command step stores as actual', () => {
    const commandStep: VerificationFixtureResult = {
      name: 'failing smoke',
      type: 'exec',
      status: 'FAIL',
      command: 'bash -c echo hi\nexit 3',
      exit_code: 3,
      stdout: 'hi\n',
      actual: { command: 'bash', exit_code: 3, pid: 14346, status: 'failed', stdout: 'hi\n' },
    };
    render(<VerificationStepOutput steps={[commandStep]} checklist={[]} />);

    expect(screen.queryByText('actual')).toBeNull();
    expect(screen.queryByText(/"pid":14346/)).toBeNull();
  });

  it('keeps a structured actual when a CEL expression compares it', () => {
    const celStep: VerificationFixtureResult = {
      ...failingExec,
      expected: undefined,
      actual: ['todos/runtime/lifecycle.go'],
    };
    render(<VerificationStepOutput steps={[celStep]} checklist={[]} />);

    expect(screen.getByText('actual').parentElement?.textContent).toBe('actual ["todos/runtime/lifecycle.go"]');
  });

  it('omits CEL detail for a passing step', () => {
    const passing: VerificationFixtureResult = { ...failingExec, status: 'PASS', exit_code: 0 };
    render(<VerificationStepOutput steps={[passing]} checklist={[]} />);

    expect(screen.queryByText('results.failed == 0')).toBeNull();
    expect(screen.getByTestId('icon-pass')).toBeTruthy();
  });

  it('lists the acceptance criteria the definition of done evaluated', () => {
    render(
      <VerificationStepOutput
        steps={[]}
        checklist={[
          { item: 'Attempts survive a reload', passed: true },
          { item: 'Failures show on the badge', passed: false, message: 'badge counted criteria' },
        ]}
      />,
    );

    expect(screen.getByText('Acceptance criteria')).toBeTruthy();
    expect(screen.getByText('Failures show on the badge')).toBeTruthy();
    expect(screen.getByText('badge counted criteria')).toBeTruthy();
  });

  it('says so when the attempt produced neither output nor a checklist', () => {
    render(<VerificationStepOutput steps={[]} checklist={[]} />);
    expect(screen.getByText('This attempt produced no command output or checklist.')).toBeTruthy();
  });
});
