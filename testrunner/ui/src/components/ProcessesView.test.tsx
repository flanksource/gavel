import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import { ProcessesView } from './ProcessesView';
import { ProcessDetailPanel } from './ProcessDetailPanel';
import type { ProcessTraceNode } from '../types';

const root: ProcessTraceNode = {
  key: '4242@1', pid: 4242, ppid: 1, name: 'gavel', command: 'gavel test --ui ./testrunner/ui', status: 'sleeping', isRoot: true,
  startedAt: '2026-01-01T00:00:00Z', lastSeenAt: '2026-01-01T00:00:09Z',
  cpuPercent: 3.5, rssBytes: 64 * 1024 * 1024, peakCpuPercent: 9, peakRssBytes: 80 * 1024 * 1024,
};
const exitedChild: ProcessTraceNode = {
  key: '4243@2', parentKey: '4242@1', pid: 4243, ppid: 4242, name: 'go', command: 'go test ./pkg', status: 'exited',
  startedAt: '2026-01-01T00:00:02Z', exitedAt: '2026-01-01T00:00:07Z', lastSeenAt: '2026-01-01T00:00:07Z',
  cpuPercent: 0, rssBytes: 0, peakCpuPercent: 40, peakRssBytes: 10 * 1024 * 1024,
};

describe('ProcessesView', () => {
  const baseProps = { nodes: [root, exitedChild], selectedKey: root.key, expandAll: null, hideExited: false, onHideExitedChange: () => {}, onSelect: () => {} };

  it('renders every process with name, pid, command and the exited lifetime', () => {
    render(<ProcessesView {...baseProps} />);

    expect(screen.getByText('pid 4242')).toBeTruthy();
    expect(screen.getByText('pid 4243')).toBeTruthy();
    expect(screen.getByText('gavel test --ui ./testrunner/ui')).toBeTruthy();
    expect(screen.getByText('go test ./pkg')).toBeTruthy();
    expect(screen.getByText('5s')).toBeTruthy();
  });

  it('hides exited processes when the toggle is on and reports toggle changes', () => {
    const onHideExitedChange = vi.fn();
    const { rerender } = render(<ProcessesView {...baseProps} hideExited onHideExitedChange={onHideExitedChange} />);

    expect(screen.queryByText('go test ./pkg')).toBeNull();
    expect(screen.getByText('gavel test --ui ./testrunner/ui')).toBeTruthy();

    fireEvent.click(screen.getByLabelText('Hide exited'));
    expect(onHideExitedChange).toHaveBeenCalledWith(false);
    rerender(<ProcessesView {...baseProps} onHideExitedChange={onHideExitedChange} />);
    expect(screen.getByText('go test ./pkg')).toBeTruthy();
  });

  it('selects a process by its key', () => {
    const onSelect = vi.fn();
    render(<ProcessesView {...baseProps} onSelect={onSelect} />);

    fireEvent.click(screen.getByText('pid 4243'));
    expect(onSelect).toHaveBeenCalledWith('4243@2');
  });
});

describe('ProcessDetailPanel', () => {
  it('offers Collect stack trace on the root and renders the returned goroutine frames', () => {
    const onCollectStack = vi.fn();
    const { rerender } = render(<ProcessDetailPanel node={root} samples={[]} onCollectStack={onCollectStack} />);

    fireEvent.click(screen.getByText('Collect stack trace'));
    expect(onCollectStack).toHaveBeenCalledTimes(1);

    rerender(<ProcessDetailPanel node={root} samples={[]} onCollectStack={onCollectStack}
      stack={{ text: 'goroutine 1 [running]:\nmain.first()\n\t/src/main.go:10 +0x1\n', collectedAt: '2026-01-01T00:00:09Z' }} />);
    expect(screen.getByText('main.first')).toBeTruthy();
    expect(screen.getByText('Refresh stack trace')).toBeTruthy();
  });

  it('shows the checkpoint rejection message', () => {
    render(<ProcessDetailPanel node={root} samples={[]} onCollectStack={() => {}} collectMessage="run has ended" />);
    expect(screen.getByText('run has ended')).toBeTruthy();
  });

  it('has no stack section for a non-root process and links to the step that started it', () => {
    const onOpenStep = vi.fn();
    const link = { pid: 4243, routePath: 'suite/go-test', name: 'go test step' };
    render(<ProcessDetailPanel node={exitedChild} samples={[]} stepLink={link} onOpenStep={onOpenStep} onCollectStack={() => {}} />);

    expect(screen.queryByText('Collect stack trace')).toBeNull();
    fireEvent.click(screen.getByText('go test step'));
    expect(onOpenStep).toHaveBeenCalledWith(link);
    expect(screen.getByText('Exited')).toBeTruthy();
  });

  it('draws CPU and RSS sparklines from samples', () => {
    const samples = [0, 1, 2].map(i => ({
      sampledAt: `2026-01-01T00:00:0${i}Z`, tick: i, key: root.key, pid: root.pid, cpuPercent: i * 10, memoryPercent: 1, rssBytes: (i + 1) * 1024,
    }));
    const { container } = render(<ProcessDetailPanel node={root} samples={samples} />);

    expect(container.querySelectorAll('polyline')).toHaveLength(2);
  });
});
