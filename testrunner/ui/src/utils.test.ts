import { describe, expect, it } from 'vitest';
import { isLintOnlyPhase, processLinksByPid, stripTestTaskGroup, timeoutArgValue, TEST_TASK_GROUP_NAME, LINT_TASK_GROUP_NAME } from './utils';
import type { Test } from './types';

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
  it('is false once the run has finished or lint already reported', () => {
    expect(isLintOnlyPhase([], false, false)).toBe(false);
    expect(isLintOnlyPhase([], true, true)).toBe(false);
  });

  it('is false while the test task group is still running or pending', () => {
    expect(isLintOnlyPhase([testGroup('running')], true, false)).toBe(false);
    expect(isLintOnlyPhase([testGroup('pending')], true, false)).toBe(false);
  });

  it('is true when the test task group is gone or settled but the run is still active', () => {
    expect(isLintOnlyPhase([], true, false)).toBe(true);
    expect(isLintOnlyPhase([testGroup('success')], true, false)).toBe(true);
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
