import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { act, cleanup, render, screen } from '@testing-library/react';
import { RerunDialog } from './RerunDialog';

class FakeEventSource {
  static CLOSED = 2;
  static instances: FakeEventSource[] = [];
  readyState = 0;
  closed = false;
  onerror: (() => void) | null = null;
  private listeners: Record<string, ((ev: Event) => void)[]> = {};

  constructor(readonly url: string) {
    FakeEventSource.instances.push(this);
  }
  addEventListener(type: string, fn: (ev: Event) => void) {
    (this.listeners[type] ??= []).push(fn);
  }
  emit(type: string, data?: unknown) {
    const ev = { data: data === undefined ? undefined : JSON.stringify(data) } as unknown as Event;
    for (const fn of this.listeners[type] ?? []) fn(ev);
  }
  fail(readyState: number) {
    this.readyState = readyState;
    this.onerror?.();
  }
  close() {
    this.closed = true;
    this.readyState = FakeEventSource.CLOSED;
  }
  static last() {
    return FakeEventSource.instances.at(-1)!;
  }
}

const write = (text: string, stream: 'stdout' | 'stderr' = 'stdout') => ({ text, stream });
const frame = (status: string, lines: ReturnType<typeof write>[] = [], command = 'test ./pkg') => ({ status, lines, command });

describe('RerunDialog', () => {
  beforeEach(() => {
    FakeEventSource.instances = [];
    vi.stubGlobal('EventSource', FakeEventSource);
  });
  afterEach(() => {
    cleanup();
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  it('streams the rerun output from the stream URL under the run\'s base, with its command and status', () => {
    render(<RerunDialog open onClose={() => {}} streamUrl="/run-1/api/rerun/stream" />);
    expect(FakeEventSource.last().url).toBe('/run-1/api/rerun/stream');

    act(() => {
      FakeEventSource.last().emit('open');
      FakeEventSource.last().emit('message', frame('running', [write('=== RUN TestA\n')]));
    });
    expect(screen.getByText('test ./pkg')).toBeTruthy();
    expect(screen.getByText(/=== RUN TestA/)).toBeTruthy();

    act(() => {
      FakeEventSource.last().emit('message', frame('failed', [write('--- FAIL: TestA\n', 'stderr')]));
    });
    expect(screen.getByText('failed')).toBeTruthy();
    expect(screen.getByText(/--- FAIL: TestA/)).toBeTruthy();
  });

  it('says so when a rerun finished without writing any output', () => {
    render(<RerunDialog open onClose={() => {}} streamUrl="/api/rerun/stream" />);
    expect(screen.getByText('Waiting for output...')).toBeTruthy();

    act(() => FakeEventSource.last().emit('message', frame('failed')));

    expect(screen.getByText('The rerun wrote no output.')).toBeTruthy();
    expect(screen.queryByText('Waiting for output...')).toBeNull();
  });

  it('closes the stream when it errors after the rerun finished', () => {
    render(<RerunDialog open onClose={() => {}} streamUrl="/api/rerun/stream" />);
    act(() => FakeEventSource.last().emit('message', frame('success')));

    act(() => FakeEventSource.last().fail(0));

    expect(FakeEventSource.last().closed).toBe(true);
  });

  it('keeps a stream that drops mid-run open so it reconnects, and starts the output over rather than repeating it', () => {
    render(<RerunDialog open onClose={() => {}} streamUrl="/api/rerun/stream" />);
    const source = FakeEventSource.last();
    act(() => {
      source.emit('open');
      source.emit('message', frame('running', [write('line one\n')]));
    });

    act(() => source.fail(0));
    expect(source.closed).toBe(false);

    // The server replays the output from the start to a client that reconnects.
    act(() => {
      source.emit('open');
      source.emit('message', frame('running', [write('line one\n'), write('line two\n')]));
    });
    expect(screen.getAllByText(/line one/)).toHaveLength(1);
    expect(screen.getByText(/line two/)).toBeTruthy();
  });

  it('asks again when the server has no rerun output yet, then gives up with a message', () => {
    vi.useFakeTimers();
    render(<RerunDialog open onClose={() => {}} streamUrl="/api/rerun/stream" />);

    for (let i = 0; i < 25; i++) {
      act(() => FakeEventSource.last().fail(FakeEventSource.CLOSED));
      act(() => { vi.advanceTimersByTime(300); });
    }

    expect(FakeEventSource.instances.length).toBeGreaterThan(5);
    expect(FakeEventSource.instances.length).toBeLessThanOrEqual(22);
    expect(screen.getByText('The rerun output is not available')).toBeTruthy();
  });
});
