import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { TodoDetailTabs, type TodoDetailTabKey } from './TodoDetailTabs';
import { verificationAttempts, verificationBadge } from './verificationReport';
import type { TodoSessionAttempt } from '../../types';

describe('verification failure tab', () => {
  it.each<TodoDetailTabKey>(['session', 'verification'])('highlights the entire tab while %s is selected', selected => {
    const onSelect = vi.fn();
    render(<TodoDetailTabs tab={selected} onSelect={onSelect} sessionAttempts={1} verification={{ count: 1, failing: true, title: 'Latest verification failed' }} />);
    const tab = screen.getByRole('button', { name: /Verification/ });
    expect(tab.className).toContain('bg-red-500/10');
    expect(tab.className).toContain('text-red-700');
    expect(tab.className).toContain('[[data-theme=dark]_&]:text-red-400');
    expect(tab.getAttribute('aria-pressed')).toBe(String(selected === 'verification'));
    if (selected === 'verification') expect(tab.className).toContain('border-red-500');
    expect(onSelect).not.toHaveBeenCalled();
    fireEvent.click(tab);
    expect(onSelect).toHaveBeenCalledWith('verification');
  });

  it('clears the tab highlight after a newer verification passes', () => {
    const failed: TodoSessionAttempt = {
      promptRunId: 'failed-run', admissionSessionId: 'admission', ordinal: 1, step: 'run',
      requested: {}, resolved: {}, status: 'failed', processActive: false, state: 'failed', phase: 'verify',
      queuedAt: '2026-10-11T10:00:00Z', createdAt: '2026-10-11T10:00:00Z', updatedAt: '2026-10-11T10:01:00Z',
      verification: { kind: 'todo', ran: true, passed: false, state: 'failed', summary: { total: 1, passed: 0, failed: 1, warned: 0, skipped: 0, pending: 0, running: 0, timedout: 0 } },
    };
    const passed: TodoSessionAttempt = {
      ...failed, promptRunId: 'passed-run', ordinal: 2, queuedAt: '2026-10-12T10:00:00Z',
      verification: { ...failed.verification!, passed: true, state: 'passed', summary: { ...failed.verification!.summary, passed: 1, failed: 0 } },
    };
    const props = { tab: 'session' as const, onSelect: vi.fn(), sessionAttempts: 2 };
    const { rerender } = render(<TodoDetailTabs {...props} verification={verificationBadge(verificationAttempts({ attempts: [failed] }))} />);
    expect(screen.getByRole('button', { name: /Verification/ }).className).toContain('bg-red-500/10');
    rerender(<TodoDetailTabs {...props} verification={verificationBadge(verificationAttempts({ attempts: [passed, failed] }))} />);
    expect(screen.getByRole('button', { name: /Verification/ }).className).not.toContain('bg-red-500/10');
  });
});
