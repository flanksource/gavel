import type { ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, fireEvent } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { PRDetailPanel } from './PRDetail';
import type { GavelResultsSummary, PRItem } from '../types';

// clicky-ui's SplitButton / DropdownMenu pull @floating-ui/react, which resolves
// a duplicate React 18 under vitest and crashes on render (see PRActions.test.tsx).
// Stub the components so the test exercises PRDetailPanel's own close-button
// wiring, not clicky internals.
const openModal = vi.hoisted(() => vi.fn());
vi.mock('@flanksource/clicky-ui/components', () => ({
  useModal: () => ({ openModal }),
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

  // The shard detail must render through clicky-ui's TestRunner /
  // LintResults, so a failing shard reads the same as the results page it links
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

    // The failing test opens as a clicky-ui test runner list row carrying its
    // message, and the violation as a `gavel lint`-style line under its linter and file.
    expect(screen.getAllByText('saves records').length).toBeGreaterThan(0);
    expect(screen.getAllByText(/expected record to persist/).length).toBeGreaterThan(0);
    const file = document.querySelector('[data-lint-list-linter="golangci-lint"] [data-lint-list-file="pkg/store/save.go"]');
    expect(file?.querySelector('[data-lint-line]')?.textContent).toBe(':23 return value is not checked (errcheck)');
    // Counts stay exact even though the payload carries only a sample.
    expect(screen.getByText('showing 1 of 2')).toBeTruthy();
    expect(screen.getByText('showing 1 of 3')).toBeTruthy();
  });

  it('previews the first five failures across all shards in the overview, linking the rest to the merged results', () => {
    const failing = (prefix: string, count: number) =>
      Array.from({ length: count }, (_, i) => ({ name: `${prefix}${i + 1}`, failed: true, message: `${prefix} ${i + 1} broke` }));
    renderPRDetail(
      <PRDetailPanel
        pr={makePR()}
        detail={{
          gavelResults: [
            makeGavelShard({ artifactId: 1, stickyId: 'gavel-pg15', testsTotal: 3, testsFailed: 3, failures: failing('Pg15Case', 3) }),
            makeGavelShard({ artifactId: 2, stickyId: 'gavel-pg16', testsTotal: 4, testsFailed: 4, failures: failing('Pg16Case', 4) }),
          ],
        }}
        loading={false}
      />,
    );

    const rows = Array.from(document.querySelectorAll('[data-test-list-row]')).map(row => row.querySelector('[title]')?.getAttribute('title'));
    const showAll = new URL(screen.getByRole('link', { name: /Show all 7/ }).getAttribute('href')!, 'http://dashboard');
    expect({
      rows,
      path: showAll.pathname,
      filter: showAll.searchParams.get('filter'),
      shards: showAll.searchParams.getAll('shard'),
    }).toEqual({
      rows: ['Pg15Case1', 'Pg15Case2', 'Pg15Case3', 'Pg16Case1', 'Pg16Case2'],
      path: `/results/${makePR().repo}/1,2/tests`,
      filter: 'failed',
      shards: ['1:gavel-pg15', '2:gavel-pg16'],
    });
  });

  it('previews the first five failures as a list and links the rest to the full results view', () => {
    const failures = Array.from({ length: 7 }, (_, i) => ({ name: `TestCase${i + 1}`, failed: true, message: `case ${i + 1} broke` }));
    renderPRDetail(
      <PRDetailPanel
        pr={makePR()}
        detail={{ gavelResults: [makeGavelShard({ artifactId: 4242, testsTotal: 9, testsPassed: 2, testsFailed: 7, failures })] }}
        loading={false}
      />,
    );

    const rows = Array.from(document.querySelectorAll('[data-test-list-row]')).map(row => row.querySelector('[title]')?.getAttribute('title'));
    const showAll = screen.getByRole('link', { name: /Show all 7/ });
    expect({
      rows,
      showing: screen.getByText('showing 5 of 7') !== null,
      href: new URL(showAll.getAttribute('href')!, 'http://dashboard').pathname + new URL(showAll.getAttribute('href')!, 'http://dashboard').searchParams.get('filter'),
      viewToggle: screen.getByRole('radiogroup', { name: 'View' }) !== null,
    }).toEqual({
      rows: ['TestCase1', 'TestCase2', 'TestCase3', 'TestCase4', 'TestCase5'],
      showing: true,
      href: `/results/${makePR().repo}/4242/testsfailed`,
      viewToggle: true,
    });
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

  // The commands are the whole point of the section for someone triaging a red
  // PR: they turn "these two tests failed in CI" into something runnable.
  it('shows the gavel commands that re-run the PR failures', () => {
    renderPRDetail(
      <PRDetailPanel
        pr={makePR()}
        detail={{
          gavelResults: [makeGavelShard({
            testsTotal: 12,
            testsFailed: 2,
            commands: ['gavel test --pr 57', 'gavel lint --pr 57'],
          })],
        }}
        loading={false}
      />,
    );

    expect(screen.getByText('Reproduce locally')).toBeTruthy();
    expect(screen.getByText('gavel test --pr 57')).toBeTruthy();
    expect(screen.getByText('gavel lint --pr 57')).toBeTruthy();
  });

  // Timed-out tests are already inside the failed count, so they annotate the
  // failed figure rather than adding a separate total.
  it('annotates the failed count with timed-out tests only when there are some', () => {
    const { unmount } = renderPRDetail(
      <PRDetailPanel
        pr={makePR()}
        detail={{ gavelResults: [makeGavelShard({ testsTotal: 12, testsPassed: 9, testsFailed: 3, testsTimedOut: 1 })] }}
        loading={false}
      />,
    );

    expect(screen.getByText('need triage (1 timed out)')).toBeTruthy();
    unmount();

    renderPRDetail(
      <PRDetailPanel
        pr={makePR()}
        detail={{ gavelResults: [makeGavelShard({ testsTotal: 12, testsPassed: 9, testsFailed: 3 })] }}
        loading={false}
      />,
    );

    expect(screen.getByText('need triage')).toBeTruthy();
    expect(screen.queryByText(/timed out/)).toBeNull();
  });

  it('sums timed-out tests across shards and annotates each shard badge', () => {
    renderPRDetail(
      <PRDetailPanel
        pr={makePR()}
        detail={{
          gavelResults: [
            makeGavelShard({ artifactId: 1, stickyId: 'gavel-test-pg15', testsTotal: 6, testsFailed: 3, testsTimedOut: 1 }),
            makeGavelShard({ artifactId: 2, stickyId: 'gavel-test-pg16', testsTotal: 6, testsFailed: 2 }),
          ],
        }}
        loading={false}
      />,
    );

    expect(screen.getByText('need triage (1 timed out)')).toBeTruthy();

    fireEvent.click(screen.getByText('Per-shard breakdown'));

    expect(screen.getByTitle('3 failed (1 timed out)')).toBeTruthy();
    expect(screen.getByTitle('2 failed')).toBeTruthy();
  });

  it('links the multi-shard lint card to every mergeable shard merged into one filterable results page', () => {
    renderPRDetail(
      <PRDetailPanel
        pr={makePR()}
        detail={{
          gavelResults: [
            makeGavelShard({ artifactId: 11, stickyId: 'gavel-lint', lintLinters: 7, lintViolations: 1300 }),
            makeGavelShard({ artifactId: 12, stickyId: 'gavel-test', lintLinters: 1, lintViolations: 12 }),
            makeGavelShard({ artifactId: 13, stickyId: 'gavel-crashed', error: 'exit status 2' }),
          ],
        }}
        loading={false}
      />,
    );

    const href = screen.getByText('Lint').closest('a')!.getAttribute('href')!;
    const url = new URL(href, 'http://dashboard');
    expect({ path: url.pathname, shards: url.searchParams.getAll('shard') }).toEqual({
      path: `/results/${makePR().repo}/11,12/lint`,
      shards: ['11:gavel-lint', '12:gavel-test'],
    });
  });

  it('omits the reproduce block for a shard with nothing to re-run', () => {
    renderPRDetail(
      <PRDetailPanel
        pr={makePR()}
        detail={{ gavelResults: [makeGavelShard({ testsTotal: 4, testsPassed: 4 })] }}
        loading={false}
      />,
    );

    expect(screen.queryByText('Reproduce locally')).toBeNull();
  });
});

describe('PRDetailPanel linked todos', () => {
  afterEach(() => { vi.unstubAllGlobals(); });

  it('lists the todos linked to the PR on the overview tab, looked up by the PR url', async () => {
    const fetchMock = vi.fn(async (_input: RequestInfo | URL) => ({
      ok: true,
      json: async () => [{ id: 'full-1', shortId: 'ab12', title: 'Fix ledger rounding', status: 'pending', priority: 'medium', relation: 'implements', source: 'manual' }],
    }) as unknown as Response);
    vi.stubGlobal('fetch', fetchMock);

    renderPRDetail(<PRDetailPanel pr={makePR()} detail={null} loading={false} />);

    expect(await screen.findByRole('heading', { name: 'Linked TODOs' })).toBeTruthy();
    expect(screen.getByRole('link', { name: 'Fix ledger rounding' }).getAttribute('href')).toBe('/todos/ab12');
    expect(String(fetchMock.mock.calls[0]![0])).toBe(`/api/prs/todos?url=${encodeURIComponent(makePR().url)}`);
  });
});
