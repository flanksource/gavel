import { afterEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import type { GoProfileRecord, StackRecord } from '@flanksource/clicky-ui/data';
import { RunProcesses, type RunProcessesProps } from './RunProcesses';
import type { ProcessTraceState } from '../hooks/use-process-trace';
import type { ProcessSample, ProcessTraceNode } from '../types';

afterEach(() => {
  cleanup();
});

const root: ProcessTraceNode = {
  key: '100@1', pid: 100, ppid: 1, name: 'gavel', command: 'gavel test ./...', status: 'sleeping', isRoot: true,
  startedAt: '2026-01-01T00:00:00Z', lastSeenAt: '2026-01-01T00:00:09Z',
  cpuPercent: 3, rssBytes: 4096, peakCpuPercent: 9, peakRssBytes: 8192,
  stackTransports: [{ transport: 'inprocess', destructive: false }],
};
const exited: ProcessTraceNode = {
  key: '101@2', parentKey: root.key, pid: 101, ppid: 100, name: 'sleep', command: 'sleep 1', status: 'exited',
  startedAt: '2026-01-01T00:00:01Z', exitedAt: '2026-01-01T00:00:02Z', lastSeenAt: '2026-01-01T00:00:02Z',
  cpuPercent: null, rssBytes: 0, peakCpuPercent: null, peakRssBytes: 1024,
};

const armedSample: ProcessSample = {
  sampledAt: '2026-01-01T00:00:03Z', tick: 1, key: root.key, pid: 100, memoryPercent: 1, rssBytes: 4096, cpuPercent: 3,
  openFiles: 1, openFileList: [{ fd: '3', type: 'REG', name: '/var/log/gavel.log' }],
};
const dump: StackRecord = {
  checkpointId: 'c1', takenAt: '2026-01-01T00:00:05Z', reason: 'manual', transport: 'native', key: root.key, pid: 100,
  format: 'native', raw: 'Thread 1:\n  poll_wait()\n',
};
const profileRecord: GoProfileRecord = {
  checkpointId: 'p1', takenAt: '2026-01-01T00:00:06Z', reason: 'manual', source: 'pprof', key: root.key, pid: 100,
  profile: 'cpu', sampleType: 'cpu', unit: 'nanoseconds', value: 5_000_000, stackHash: 'h',
  frames: [{ func: 'main.hotLoop', file: 'main.go', line: 12 }],
};

function traceState(over: Partial<ProcessTraceState> = {}): ProcessTraceState {
  return {
    nodes: [root, exited], samples: [], stacks: [], goprof: [], loading: false,
    canCollect: true, collecting: false, collectStack: vi.fn(async () => {}),
    refreshing: false, refreshFacets: vi.fn(async () => {}),
    ...over,
  };
}

function renderProcesses(over: Partial<RunProcessesProps> = {}, state: Partial<ProcessTraceState> = {}) {
  const props: RunProcessesProps = {
    trace: traceState(state),
    selectedKey: root.key,
    onSelect: vi.fn(),
    hideExited: false,
    onHideExitedChange: vi.fn(),
    expandAll: null,
    onExpandAllChange: vi.fn(),
    ...over,
  };
  render(<RunProcesses {...props} />);
  return props;
}

describe('RunProcesses', () => {
  it('shows the tree with its process count and reports the hide-exited toggle and a selection', () => {
    const props = renderProcesses();

    expect(screen.getByText('2 processes')).toBeTruthy();
    fireEvent.click(screen.getByLabelText('Hide exited'));
    expect(props.onHideExitedChange).toHaveBeenCalledWith(true);

    fireEvent.click(screen.getByText('sleep 1', { selector: '[data-process-command]' }));
    expect(props.onSelect).toHaveBeenCalledWith('101@2');
  });

  it('drops exited processes while hide-exited is on', () => {
    renderProcesses({ hideExited: true });
    expect(screen.queryByText('sleep 1', { selector: '[data-process-command]' })).toBeNull();
    expect(screen.getByText('gavel test ./...', { selector: '[data-process-command]' })).toBeTruthy();
  });

  it('shows a waiting state until the first trace arrives and an error banner when it cannot be read', () => {
    renderProcesses({}, { nodes: [], loading: true });
    expect(screen.getByText('Waiting for process trace…')).toBeTruthy();
    cleanup();
    renderProcesses({}, { nodes: [], error: 'read process_tree window failed (500): boom' });
    expect(screen.getByRole('alert').textContent).toContain('boom');
  });

  it('collects a stack over the transport the selected process offers', () => {
    const props = renderProcesses();

    fireEvent.click(screen.getByRole('button', { name: /Collect stack trace/ }));

    expect(props.trace.collectStack).toHaveBeenCalledWith('100@1', 'inprocess');
  });

  it('shows the collected stack of the selected process', () => {
    renderProcesses({}, { stacks: [dump] });
    expect(screen.getByText(/poll_wait\(\)/)).toBeTruthy();
  });

  it('offers no collect or refresh button for a run whose capture is not live, but still shows its stack and facets', () => {
    renderProcesses({}, { canCollect: false, stacks: [dump], samples: [armedSample] });

    expect(screen.queryByRole('button', { name: /stack trace/i })).toBeNull();
    expect(screen.queryByRole('button', { name: 'Refresh facets' })).toBeNull();
    expect(screen.getByText(/poll_wait\(\)/)).toBeTruthy();
    expect(screen.getByRole('tab', { name: 'Open files' })).toBeTruthy();
  });

  it('shows a facet tab only for a facet the samples carry, with the latest stored list', () => {
    renderProcesses({}, { samples: [armedSample] });

    expect(screen.queryByRole('tab', { name: 'Sockets' })).toBeNull();
    fireEvent.click(screen.getByRole('tab', { name: 'Open files' }));
    expect(screen.getByText('/var/log/gavel.log')).toBeTruthy();
  });

  it('shows the environment of a node that recorded it', () => {
    renderProcesses({}, { nodes: [{ ...root, environment: { HOME: '/home/gavel' } }, exited] });

    fireEvent.click(screen.getByRole('tab', { name: 'Environment' }));
    expect(screen.getByText('/home/gavel')).toBeTruthy();
  });

  it('forces a fresh facet reading of the selected process, and reports why one failed', () => {
    const props = renderProcesses({}, { samples: [armedSample], refreshError: 'checkpoint failed (409): busy' });

    fireEvent.click(screen.getByRole('button', { name: 'Refresh facets' }));

    expect(props.trace.refreshFacets).toHaveBeenCalledWith('100@1');
    expect(screen.getByRole('alert').textContent).toContain('busy');
  });

  it('offers no facet refresh when the session armed no facet', () => {
    renderProcesses();
    expect(screen.queryByRole('button', { name: 'Refresh facets' })).toBeNull();
  });

  it('shows the Go profiles of the selected process with the pprof download of the window', () => {
    renderProcesses({}, { goprof: [profileRecord], goprofExportUrl: '/run-1/query/api/v1/traces/goprof/export?stream=g&format=pprof' });

    expect(screen.getByText('main.hotLoop')).toBeTruthy();
    expect(screen.getByRole('link', { name: /pprof/i }).getAttribute('href')).toBe('/run-1/query/api/v1/traces/goprof/export?stream=g&format=pprof');
  });

  it('shows no profiles section for a process without goprof records', () => {
    renderProcesses({}, { goprof: [{ ...profileRecord, key: exited.key }] });
    expect(screen.queryByText('main.hotLoop')).toBeNull();
    expect(screen.queryByText('Go profiles')).toBeNull();
  });

  it('links the process to the fixture step that started it', () => {
    const onOpenStep = vi.fn();
    renderProcesses({ stepLinks: new Map([[100, { pid: 100, routePath: 'suite/start', name: 'start gavel' }]]), onOpenStep });

    fireEvent.click(screen.getByRole('button', { name: /start gavel/ }));

    expect(onOpenStep).toHaveBeenCalledWith({ pid: 100, routePath: 'suite/start', name: 'start gavel' });
  });
});
