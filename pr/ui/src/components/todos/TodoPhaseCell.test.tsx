import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { TodoItem, TodoPhaseRun } from '../../types';
import {
  TodoPhaseCell,
  phaseElapsedMs,
  phaseFailed,
  phaseProgressLabel,
  phaseRunning,
} from './TodoPhaseCell';

function todoWith(phases: TodoItem['phases']): TodoItem {
  return { ref: 'todo-1', title: 'A todo', status: 'pending', priority: 'medium', phases };
}

const settled = (over: Partial<TodoPhaseRun> = {}): TodoPhaseRun => ({
  phase: 'run',
  state: 'succeeded',
  duration_ms: 134_000,
  ...over,
});

describe('TodoPhaseCell', () => {
  // A phase that has never run must stay visibly different from one that ran
  // and produced nothing — that is the whole reason the server omits absent
  // phases instead of sending zero values.
  it('renders an em-dash for a phase that never ran', () => {
    render(<TodoPhaseCell todo={todoWith({})} phase="triage" />);
    expect(screen.getByText('—')).toBeTruthy();
  });

  it('renders a settled phase with its recorded duration', () => {
    render(<TodoPhaseCell todo={todoWith({ run: settled() })} phase="run" />);
    expect(screen.getByText('2m 14s')).toBeTruthy();
  });

  it('labels the cell with the phase and its state', () => {
    render(<TodoPhaseCell todo={todoWith({ verify: settled({ phase: 'verify', state: 'failed' }) })} phase="verify" />);
    expect(screen.getByTitle(/verif/i)).toBeTruthy();
  });

  it('shows verification progress as its fixture counts it', () => {
    const verify = settled({ phase: 'verify', progress: { done: 3, failed: 1, total: 4 } });
    render(<TodoPhaseCell todo={todoWith({ verify })} phase="verify" />);
    expect(screen.getByText('3/4')).toBeTruthy();
  });

  // A queued phase has not started, so it has no elapsed time. Showing "0s"
  // would read as "finished instantly" rather than "not started yet".
  it('omits the timer for a phase that has not started', () => {
    const queued: TodoPhaseRun = { phase: 'run', state: 'pending' };
    const { container } = render(<TodoPhaseCell todo={todoWith({ run: queued })} phase="run" />);
    expect(container.textContent).toBe('');
  });
});

describe('phaseElapsedMs', () => {
  // duration_ms is a snapshot from when the row was read. A running phase has
  // to be measured against the clock or its timer freezes between list polls.
  it('measures a running phase against now, not the recorded duration', () => {
    const now = Date.parse('2026-08-26T12:00:00Z');
    const run = settled({ state: 'running', started_at: '2026-08-26T11:58:30Z', duration_ms: 1_000 });
    expect(phaseElapsedMs(run, now)).toBe(90_000);
  });

  it('uses the recorded duration once the phase has settled', () => {
    const now = Date.parse('2026-08-26T12:00:00Z');
    const run = settled({ started_at: '2026-08-26T11:00:00Z' });
    expect(phaseElapsedMs(run, now)).toBe(134_000);
  });

  it('falls back to the recorded duration when the start time is unparseable', () => {
    const run = settled({ state: 'running', started_at: 'not-a-date', duration_ms: 500 });
    expect(phaseElapsedMs(run, Date.now())).toBe(500);
  });
});

describe('phase outcome helpers', () => {
  // A verification run that executed cleanly but produced failing checks is a
  // failure, even though the run's own state is `succeeded`.
  it('treats failing checks as a failed phase', () => {
    expect(phaseFailed(settled({ progress: { done: 3, failed: 1, total: 4 } }))).toBe(true);
    expect(phaseFailed(settled({ progress: { done: 4, total: 4 } }))).toBe(false);
    expect(phaseFailed(settled({ state: 'cancelled' }))).toBe(true);
  });

  it('counts a queued phase as running so its cell shows a spinner', () => {
    expect(phaseRunning(settled({ state: 'pending' }))).toBe(true);
    expect(phaseRunning(settled({ state: 'waiting' }))).toBe(false);
  });
});

describe('phaseProgressLabel', () => {
  // "1/1" on every row that succeeded first try is noise, but any failure is
  // worth the space.
  it.each([
    { progress: { done: 1, total: 1 }, want: '' },
    { progress: { done: 3, total: 4 }, want: '3/4' },
    { progress: { done: 0, failed: 2, total: 1 }, want: '2 failed' },
    { progress: undefined, want: '' },
  ])('renders $progress as "$want"', ({ progress, want }) => {
    expect(phaseProgressLabel(settled({ progress }))).toBe(want);
  });
});
