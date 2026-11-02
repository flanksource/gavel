import type { LinterResult, Test } from './types';

// Names of the clicky task groups injected as virtual pseudo-tests by the
// Go server (testrunner/ui/task_groups.go). They appear at the top of the
// snapshot tree while their underlying tasks are running/pending.
export const TEST_TASK_GROUP_NAME = 'Running tests across packages';
export const LINT_TASK_GROUP_NAME = 'Running linters';

function isVirtualTaskGroup(t: Test, groupName: string): boolean {
  return t.framework === 'task' && t.name === groupName;
}

// stripTestTaskGroup removes the "Running tests across packages" virtual task
// group from the rendered tree. Used while linters are still running but tests
// have settled, so the only "in-flight" indicator left is the lint task group.
// Real parsed test results are untouched.
export function stripTestTaskGroup(tests: Test[]): Test[] {
  return tests.filter(t => !isVirtualTaskGroup(t, TEST_TASK_GROUP_NAME));
}

function taskStatus(t: Test): unknown {
  return (t.context as Record<string, unknown> | undefined)?.status;
}

function isLive(t: Test): boolean {
  const status = taskStatus(t);
  return status === 'running' || status === 'pending';
}

function hasLiveTestGroup(tests: Test[]): boolean {
  const testGroup = tests.find(t => isVirtualTaskGroup(t, TEST_TASK_GROUP_NAME));
  // The server only emits the group while running/pending (handler.go:381), so
  // its presence alone implies tests are still in flight; the status check
  // covers a group that has just settled.
  return !!testGroup && isLive(testGroup);
}

// isLintOnlyPhase reports whether the run is in the window where the test
// phase has settled (no test task group is still running) but the run has not
// finished yet. The server keeps `running: true` until linters complete. A lint
// result can appear while the run is in flight (lint_run flips true on the first
// pending linter), so the lint state does not take part.
export function isLintOnlyPhase(tests: Test[], running: boolean): boolean {
  return running && !hasLiveTestGroup(tests);
}

// isLintOnlyRun reports whether the run is a lint run with no tests: lint ran
// and nothing but virtual task rows (which are not test results) is in the tree.
export function isLintOnlyRun(tests: Test[], lintRun: boolean): boolean {
  return lintRun && !hasLiveTestGroup(tests) && tests.every(t => t.framework === 'task');
}

export interface LintStatusCounts {
  running: number;
  done: number;
  failed: number;
}

// lintStatusCounts buckets linter results: in-flight (pending or running) is
// running; a linter that failed to run (an error, a timeout, or a failure with
// no finding to explain it) is failed; the rest, including skipped linters and
// ones that reported findings, are done.
export function lintStatusCounts(lint: LinterResult[] | undefined): LintStatusCounts {
  const counts: LintStatusCounts = { running: 0, done: 0, failed: 0 };
  for (const r of lint ?? []) {
    if (r.status) counts.running++;
    else if (!r.skipped && (r.error || r.timed_out || (!r.success && !r.violations?.length))) counts.failed++;
    else counts.done++;
  }
  return counts;
}

// liveTaskGroups returns the virtual task groups that are still running or
// pending, with only their live children, excluding the test and linter groups
// (which the Tests tab and the lint tree already present). These are the tasks
// behind a prompt lint run, such as the `lint prompt` group and its per-review
// tasks.
export function liveTaskGroups(tests: Test[]): Test[] {
  return tests
    .filter(t => t.framework === 'task' && t.name !== TEST_TASK_GROUP_NAME && t.name !== LINT_TASK_GROUP_NAME && isLive(t))
    .map(t => ({ ...t, children: (t.children ?? []).filter(isLive) }));
}

// timeoutArgValue reads one of the run's timeout knobs (SnapshotMetadata.Args:
// timeout, test_timeout, lint_timeout), null when it is unset.
export function timeoutArgValue(args: Record<string, unknown> | undefined, key: string): string | null {
  if (!args) return null;
  const raw = args[key];
  if (typeof raw !== 'string') return null;
  const trimmed = raw.trim();
  return trimmed === '' ? null : trimmed;
}

export interface ProcessStepLink {
  pid: number;
  routePath: string;
  name: string;
}

// processLinksByPid maps each pid recorded in a fixture step's fixture.pids to
// the test node that started it. The first step to claim a pid wins.
export function processLinksByPid(tests: Test[]): Map<number, ProcessStepLink> {
  const links = new Map<number, ProcessStepLink>();
  const walk = (list: Test[]) => {
    for (const test of list) {
      for (const pid of test.fixture?.pids || []) {
        if (!links.has(pid) && test.route_path) {
          links.set(pid, { pid, routePath: test.route_path, name: test.name });
        }
      }
      if (test.children) walk(test.children);
    }
  };
  walk(tests);
  return links;
}
