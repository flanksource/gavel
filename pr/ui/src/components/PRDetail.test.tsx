import type { ReactNode } from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { PRDetailPanel } from './PRDetail';
import type { GavelResultsSummary, PRItem } from '../types';

// clicky-ui's SplitButton / DropdownMenu pull @floating-ui/react, which resolves
// a duplicate React 18 under vitest and crashes on render (see PRActions.test.tsx).
// Stub the components so the test exercises PRDetailPanel's own close-button
// wiring, not clicky internals. useListMenuSelection is called unconditionally
// by CreateTodoFromPRDialog (which PRDetailPanel always mounts, closed) before
// its own `if (!open) return null` — its return value is unused while closed,
// so the stub only needs to avoid throwing.
vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, onClick, title, disabled, ...rest }: any) => (
    <button onClick={onClick} title={title} disabled={disabled} aria-label={rest['aria-label']}>{children}</button>
  ),
  SplitButton: ({ label, onClick, title, disabled }: { label: ReactNode; onClick?: () => void; title?: string; disabled?: boolean }) => (
    <button onClick={onClick} title={title} disabled={disabled}>{label}</button>
  ),
  DropdownMenu: ({ trigger }: { trigger: ReactNode }) => <div>{trigger}</div>,
  Modal: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  useListMenuSelection: () => ({ selectedKeys: [], toggle: () => {}, selectAll: () => {}, clear: () => {} }),
}));

function makePR(overrides: Partial<PRItem> = {}): PRItem {
  return {
    number: 7,
    title: 'Test PR',
    author: 'octocat',
    repo: 'acme/widget',
    source: 'feature',
    target: 'main',
    state: 'OPEN',
    isDraft: false,
    url: 'https://github.com/acme/widget/pull/7',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

describe('PRDetailPanel', () => {
  it.each([
    ['absent', undefined],
    ['empty', []],
  ])('renders New todo when projects are %s', (_label, projects) => {
    render(<PRDetailPanel pr={makePR()} detail={null} loading={false} projects={projects} />);

    expect(screen.getByRole('button', { name: 'New todo' })).toBeTruthy();
  });

  it('renders a close button and fires onClose when clicked', () => {
    const onClose = vi.fn();
    render(<PRDetailPanel pr={makePR()} detail={null} loading={false} onClose={onClose} />);

    const closeButton = screen.getByLabelText('Close pull request details');
    fireEvent.click(closeButton);

    expect(onClose).toHaveBeenCalledOnce();
  });

  it('omits the close button when onClose is not provided', () => {
    render(<PRDetailPanel pr={makePR()} detail={null} loading={false} />);

    expect(screen.queryByLabelText('Close pull request details')).toBeNull();
  });
});

function makeGavelShard(overrides: Partial<GavelResultsSummary> = {}): GavelResultsSummary {
  return {
    artifactId: 4242,
    artifactUrl: 'https://github.com/acme/widget/actions/runs/1/artifacts/4242',
    testsPassed: 0,
    testsFailed: 0,
    testsSkipped: 0,
    testsTotal: 0,
    lintViolations: 0,
    lintLinters: 0,
    hasBench: false,
    ...overrides,
  };
}

// A crashed run produces an artifact with no tests, no lint and no bench. The
// reason it produced nothing is the only useful thing to show; falling back to
// "No test, lint, or bench data" reads as if the artifact was never found.
describe('Gavel Results section', () => {
  const crashMessage = 'pre-build: compiling Go test binaries failed (exit 1)';

  it('shows the crash reason, exit code and log tail instead of the no-data text', () => {
    render(
      <PRDetailPanel
        pr={makePR()}
        detail={{
          gavelResults: [makeGavelShard({
            error: crashMessage,
            exitCode: 1,
            logTail: 'go: updates to go.mod needed; to update it:\n\tgo mod tidy\n',
          })],
        }}
        loading={false}
      />,
    );

    expect(screen.getByText(crashMessage)).toBeTruthy();
    expect(screen.getByText('1')).toBeTruthy();
    expect(screen.getByText(/go mod tidy/)).toBeTruthy();
    expect(screen.queryByText(/No test, lint, or bench data/)).toBeNull();
  });

  it('still reports no data when an empty artifact carries no error', () => {
    render(
      <PRDetailPanel
        pr={makePR()}
        detail={{ gavelResults: [makeGavelShard()] }}
        loading={false}
      />,
    );

    expect(screen.getByText(/No test, lint, or bench data/)).toBeTruthy();
  });
});
