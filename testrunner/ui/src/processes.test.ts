import { describe, expect, it } from 'vitest';
import { buildProcessForest, countLiveProcesses, isProcessExited, processLinksByPid, processStateColor, processStateIcon, pruneExitedProcesses } from './utils';
import type { ProcessTraceNode, Test } from './types';

function node(key: string, overrides: Partial<ProcessTraceNode> = {}): ProcessTraceNode {
  const [pid] = key.split('@');
  return {
    key,
    pid: Number(pid),
    ppid: 1,
    name: `proc-${pid}`,
    command: `proc-${pid} --flag`,
    status: 'sleeping',
    startedAt: '2026-01-01T00:00:00Z',
    lastSeenAt: '2026-01-01T00:00:05Z',
    cpuPercent: 0,
    rssBytes: 0,
    peakCpuPercent: 0,
    peakRssBytes: 0,
    ...overrides,
  };
}

describe('buildProcessForest', () => {
  it('returns an empty forest for no nodes', () => {
    expect(buildProcessForest([])).toEqual([]);
  });

  it('nests children under their parentKey and sorts siblings by startedAt', () => {
    const root = node('100@1', { isRoot: true });
    const late = node('102@3', { parentKey: '100@1', startedAt: '2026-01-01T00:00:03Z' });
    const early = node('101@2', { parentKey: '100@1', startedAt: '2026-01-01T00:00:02Z' });
    const grandchild = node('103@4', { parentKey: '101@2' });

    const forest = buildProcessForest([late, grandchild, root, early]);

    expect(forest.map(f => f.node.key)).toEqual(['100@1']);
    expect(forest[0].children.map(f => f.node.key)).toEqual(['101@2', '102@3']);
    expect(forest[0].children[0].children.map(f => f.node.key)).toEqual(['103@4']);
  });

  it('treats a node whose parent is not in the set as a root', () => {
    const root = node('100@1', { isRoot: true, startedAt: '2026-01-01T00:00:00Z' });
    const orphan = node('200@9', { parentKey: '999@0', startedAt: '2026-01-01T00:00:09Z' });

    const forest = buildProcessForest([orphan, root]);

    expect(forest.map(f => f.node.key)).toEqual(['100@1', '200@9']);
  });

  it('keeps two nodes that share a pid but differ by start second as separate entries', () => {
    const root = node('100@1', { isRoot: true });
    const first = node('555@2', { parentKey: '100@1', startedAt: '2026-01-01T00:00:02Z', status: 'exited' });
    const reused = node('555@7', { parentKey: '100@1', startedAt: '2026-01-01T00:00:07Z' });

    const forest = buildProcessForest([root, reused, first]);

    expect(forest[0].children.map(f => f.node.key)).toEqual(['555@2', '555@7']);
  });

  it('does not loop on a node that names itself as parent', () => {
    const forest = buildProcessForest([node('100@1', { parentKey: '100@1' })]);
    expect(forest.map(f => f.node.key)).toEqual(['100@1']);
  });
});

describe('pruneExitedProcesses', () => {
  it('drops exited leaves but keeps an exited parent with a live child', () => {
    const forest = buildProcessForest([
      node('1@1', { isRoot: true }),
      node('2@2', { parentKey: '1@1', status: 'exited' }),
      node('3@3', { parentKey: '1@1', status: 'exited' }),
      node('4@4', { parentKey: '3@3' }),
    ]);

    const pruned = pruneExitedProcesses(forest);

    expect(pruned[0].children.map(c => c.node.key)).toEqual(['3@3']);
    expect(pruned[0].children[0].children.map(c => c.node.key)).toEqual(['4@4']);
  });

  it('returns an empty forest when every process has exited', () => {
    const forest = buildProcessForest([node('1@1', { status: 'exited' })]);
    expect(pruneExitedProcesses(forest)).toEqual([]);
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

    const links = processLinksByPid(tests);

    expect(Object.fromEntries(links)).toEqual({
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

describe('process state helpers', () => {
  it('counts only processes that have not exited', () => {
    const nodes = [
      node('1@1'),
      node('2@2', { status: 'exited' }),
      node('3@3', { status: 'active', exitedAt: '2026-01-01T00:00:09Z' }),
    ];
    expect(nodes.map(isProcessExited)).toEqual([false, true, true]);
    expect(countLiveProcesses(nodes)).toBe(1);
  });

  it('renders exited as muted gray with its own icon', () => {
    expect(processStateColor('exited')).toBe('text-gray-400');
    expect(processStateIcon('exited')).toBe('codicon:circle-slash');
    expect(processStateIcon('exited')).not.toBe(processStateIcon('sleeping'));
  });
});
