import { describe, expect, it } from 'vitest';
import { isLintOnlyPhase, isLintOnlyRun, liveTaskGroups, lintStatusCounts, processLinksByPid, stripTestTaskGroup, timeoutArgValue, TEST_TASK_GROUP_NAME, LINT_TASK_GROUP_NAME } from './utils';
import type { LinterResult, Test } from './types';

const testGroup = (status: string): Test => ({ name: TEST_TASK_GROUP_NAME, framework: 'task', context: { status } });
const lintGroup: Test = { name: LINT_TASK_GROUP_NAME, framework: 'task', context: { status: 'running' } };

describe('stripTestTaskGroup', () => {
  it('removes the test task group while keeping real tests and the lint group', () => {
    const real: Test = { name: 'TestFoo', framework: 'go test', passed: true };

    expect(stripTestTaskGroup([testGroup('running'), real, lintGroup])).toEqual([real, lintGroup]);
  });

  it('only matches the task framework, not real tests sharing the name', () => {
    const lookalike: Test = { name: TEST_TASK_GROUP_NAME, framework: 'go test', passed: true };

    expect(stripTestTaskGroup([lookalike])).toEqual([lookalike]);
  });
});

describe('isLintOnlyPhase', () => {
  it('is false once the run has finished', () => {
    expect(isLintOnlyPhase([], false)).toBe(false);
  });

  it('is false while the test task group is still running or pending', () => {
    expect(isLintOnlyPhase([testGroup('running')], true)).toBe(false);
    expect(isLintOnlyPhase([testGroup('pending')], true)).toBe(false);
  });

  it('is true when the test task group is gone or settled but the run is still active, even after lint has reported', () => {
    expect(isLintOnlyPhase([], true)).toBe(true);
    expect(isLintOnlyPhase([testGroup('success')], true)).toBe(true);
  });
});

describe('isLintOnlyRun', () => {
  const real: Test = { name: 'TestFoo', framework: 'go test', passed: true };

  it('is true for a lint run whose only tests are virtual task rows', () => {
    expect(isLintOnlyRun([], true)).toBe(true);
    expect(isLintOnlyRun([lintGroup], true)).toBe(true);
  });

  it('is false without a lint run, with a live test group, or with real tests', () => {
    expect(isLintOnlyRun([], false)).toBe(false);
    expect(isLintOnlyRun([testGroup('running')], true)).toBe(false);
    expect(isLintOnlyRun([real], true)).toBe(false);
  });
});

describe('lintStatusCounts', () => {
  const linter = (over: Partial<LinterResult>): LinterResult => ({ linter: 'l', success: true, duration: 1, violations: [], ...over });

  it('counts in-flight linters as running, clean or finding linters as done, and failed-to-run linters as failed', () => {
    expect(lintStatusCounts([
      linter({ status: 'running', success: false }),
      linter({ status: 'pending', success: false }),
      linter({}),
      linter({ success: false, violations: [{ file: 'a', message: 'm' }] }),
      linter({ skipped: true, success: false }),
      linter({ success: false, error: 'uir executable not found' }),
      linter({ success: false }),
      linter({ timed_out: true }),
    ])).toEqual({ running: 2, done: 3, failed: 3 });
  });

  it('is all zero without results', () => {
    expect(lintStatusCounts(undefined)).toEqual({ running: 0, done: 0, failed: 0 });
  });
});

describe('liveTaskGroups', () => {
  const group = (name: string, status: string, children: Test[] = []): Test => ({ name, framework: 'task', context: { status }, children });

  it('keeps live groups other than the test and linter groups, with only their live children', () => {
    const live: Test = { name: 'messaging-review:a.mdx', framework: 'task', context: { status: 'running' }, message: 'tool: read_file' };
    const done: Test = { name: 'old', framework: 'task', context: { status: 'success' } };
    const prompt = group('lint prompt', 'running', [live, done]);

    expect(liveTaskGroups([testGroup('running'), lintGroup, prompt, group('settled', 'success')])).toEqual([{ ...prompt, children: [live] }]);
  });

  it('is empty when no task group is live', () => {
    expect(liveTaskGroups([{ name: 'TestFoo', framework: 'go test', passed: true }, group('lint prompt', 'failed')])).toEqual([]);
  });
});

describe('timeoutArgValue', () => {
  it('reads a set knob, trimmed, and reports an unset, blank or non-string one as absent', () => {
    const args = { timeout: ' 30m ', test_timeout: '', lint_timeout: 5 };

    expect(timeoutArgValue(args, 'timeout')).toBe('30m');
    expect(timeoutArgValue(args, 'test_timeout')).toBeNull();
    expect(timeoutArgValue(args, 'lint_timeout')).toBeNull();
    expect(timeoutArgValue(undefined, 'timeout')).toBeNull();
  });
});

describe('processLinksByPid', () => {
  it('returns an empty map for no tests', () => {
    expect(processLinksByPid([]).size).toBe(0);
  });

  it('maps pids from nested fixture steps to the step that started them', () => {
    const tests = [
      {
        name: 'suite',
        route_path: 'suite',
        children: [
          { name: 'start db', route_path: 'suite/start-db', fixture: { key: 'a', kind: 'service', state: 'running', pids: [4243, 4244] } },
          { name: 'no processes', route_path: 'suite/none', fixture: { key: 'b', kind: 'cel', state: 'passed' } },
          { name: 'go test', route_path: 'suite/go-test', fixture: { key: 'c', kind: 'exec', state: 'passed', pids: [4250] } },
        ],
      },
    ] as Test[];

    expect(Object.fromEntries(processLinksByPid(tests))).toEqual({
      4243: { pid: 4243, routePath: 'suite/start-db', name: 'start db' },
      4244: { pid: 4244, routePath: 'suite/start-db', name: 'start db' },
      4250: { pid: 4250, routePath: 'suite/go-test', name: 'go test' },
    });
  });

  it('keeps the first step when two steps report the same pid', () => {
    const tests = [
      { name: 'one', route_path: 'one', fixture: { key: 'a', kind: 'exec', state: 'passed', pids: [7] } },
      { name: 'two', route_path: 'two', fixture: { key: 'b', kind: 'exec', state: 'passed', pids: [7] } },
    ] as Test[];

    expect(processLinksByPid(tests).get(7)?.routePath).toBe('one');
  });
});
