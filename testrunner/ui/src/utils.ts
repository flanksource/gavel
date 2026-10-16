import type { Test } from './types';

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

// isLintOnlyPhase reports whether the run is in the window where the test
// phase has settled (no test task group is still running) but linters have
// not finished yet. The server keeps `running: true` until lint completes,
// and `lint_run` flips true only when SetLintResults is called.
export function isLintOnlyPhase(tests: Test[], running: boolean, lintRun: boolean): boolean {
  if (!running || lintRun) return false;
  const testGroup = tests.find(t => isVirtualTaskGroup(t, TEST_TASK_GROUP_NAME));
  if (!testGroup) return true;
  // Group's running/pending state is encoded in its context.status. The
  // server only emits the group while running/pending (handler.go:381), so
  // its presence alone implies tests are still in flight.
  const status = (testGroup.context as Record<string, unknown> | undefined)?.status;
  return status !== 'running' && status !== 'pending';
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
