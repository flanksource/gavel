import type { ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
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

function renderPRDetail(children: ReactNode) {
  return render(
    <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
      {children}
    </QueryClientProvider>,
  );
}

describe('PRDetailPanel', () => {
  it.each([
    ['absent', undefined],
    ['empty', []],
  ])('renders New todo when projects are %s', (_label, projects) => {
    renderPRDetail(<PRDetailPanel pr={makePR()} detail={null} loading={false} projects={projects} />);

    expect(screen.getByRole('button', { name: 'New todo' })).toBeTruthy();
  });

  it('renders a close button and fires onClose when clicked', () => {
    const onClose = vi.fn();
    renderPRDetail(<PRDetailPanel pr={makePR()} detail={null} loading={false} onClose={onClose} />);

    const closeButton = screen.getByLabelText('Close pull request details');
    fireEvent.click(closeButton);

    expect(onClose).toHaveBeenCalledOnce();
  });

  it('omits the close button when onClose is not provided', () => {
    renderPRDetail(<PRDetailPanel pr={makePR()} detail={null} loading={false} />);

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
    renderPRDetail(
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
    renderPRDetail(
      <PRDetailPanel
        pr={makePR()}
        detail={{ gavelResults: [makeGavelShard()] }}
        loading={false}
      />,
    );

    expect(screen.getByText(/No test, lint, or bench data/)).toBeTruthy();
  });

  // The shard detail must render through the testrunner UI's own TestNode /
  // LintView, so a failing shard reads the same as the results page it links
  // to. Asserting on the rendered findings is what catches a regression back to
  // a parallel set of row components.
  it('renders failing tests and lint findings from the artifact shapes', () => {
    renderPRDetail(
      <PRDetailPanel
        pr={makePR()}
        detail={{
          gavelResults: [makeGavelShard({
            testsTotal: 12,
            testsPassed: 10,
            testsFailed: 2,
            lintLinters: 1,
            lintViolations: 3,
            failures: [{
              name: 'saves records',
              suite: ['storage'],
              file: 'pkg/store/save_test.go',
              line: 41,
              failed: true,
              message: 'expected record to persist',
            }],
            lint: [{
              linter: 'golangci-lint',
              success: false,
              duration: 0,
              violations: [{
                rule: { method: 'errcheck' },
                file: 'pkg/store/save.go',
                line: 23,
                message: 'return value is not checked',
              }],
            }],
          })],
        }}
        loading={false}
      />,
    );

    // The failing leaf appears in the tree row and again in the detail panel
    // that carries its message — both come from the testrunner UI.
    expect(screen.getAllByText('saves records').length).toBeGreaterThan(0);
    expect(screen.getByText(/expected record to persist/)).toBeTruthy();
    // The lint tree groups linter → rule → file exactly as the results page does.
    expect(screen.getAllByText('golangci-lint').length).toBeGreaterThan(0);
    expect(screen.getAllByText('errcheck').length).toBeGreaterThan(0);
    // Counts stay exact even though the payload carries only a sample.
    expect(screen.getByText('showing 1 of 2')).toBeTruthy();
    expect(screen.getByText('showing 1 of 3')).toBeTruthy();
  });

  // A linter that could not run reports zero violations, so it never reaches
  // the lint tree. Dropping it would let a broken lint shard read as clean.
  it('surfaces a linter that failed to run rather than reporting a clean shard', () => {
    renderPRDetail(
      <PRDetailPanel
        pr={makePR()}
        detail={{
          gavelResults: [makeGavelShard({
            lintLinters: 1,
            lint: [{
              linter: 'golangci-lint',
              success: false,
              duration: 0,
              violations: [],
              error: 'golangci-lint execution failed: exit status 3',
            }],
          })],
        }}
        loading={false}
      />,
    );

    expect(screen.getByText('Linters that failed to run')).toBeTruthy();
    expect(screen.getByText('golangci-lint')).toBeTruthy();
    expect(screen.getByText(/exit status 3/)).toBeTruthy();
  });
});
