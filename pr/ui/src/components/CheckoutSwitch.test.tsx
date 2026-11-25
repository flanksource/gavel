import { describe, expect, it } from 'vitest';
import type { CheckoutSwitchPlan } from '../types/checkoutSwitch';
import { checkoutSyncChoices, defaultCheckoutSync } from './CheckoutSwitch';

const plan = (overrides: Partial<CheckoutSwitchPlan>): CheckoutSwitchPlan => ({
  repo: '/src/acme', current: 'feature', head: 'aaaa', branch: 'main', branchHead: 'bbbb',
  upstream: 'origin/main', upstreamHead: 'cccc', ahead: 0, behind: 3, unpreserved: [], dirtyCount: 0,
  ...overrides,
});

const reasons = (input: CheckoutSwitchPlan) =>
  Object.fromEntries(checkoutSyncChoices(input).map(choice => [choice.id, choice.reason]));

describe('checkoutSyncChoices', () => {
  it('allows every move for a branch that is only behind its upstream', () => {
    expect(reasons(plan({}))).toEqual({ 'fast-forward': '', reset: '', none: '' });
  });

  it('refuses a fast-forward but allows a reset when the diverged commits are held elsewhere', () => {
    const choices = checkoutSyncChoices(plan({ ahead: 6, behind: 28 }));
    expect(Object.fromEntries(choices.map(choice => [choice.id, choice.reason]))).toEqual({
      'fast-forward': 'main has 6 commits origin/main lacks, so it cannot fast-forward.',
      reset: '',
      none: '',
    });
    expect(defaultCheckoutSync(choices)).toBe('reset');
  });

  it('refuses a reset that would orphan commits only the branch holds', () => {
    const unpreserved = [{ sha: '0f8161680000', subject: 'feat: local only' }];
    expect(reasons(plan({ ahead: 1, unpreserved })).reset).toBe('Resetting would orphan 1 commit no other branch holds.');
  });

  it('refuses every move while tracked files are dirty', () => {
    expect(reasons(plan({ dirtyCount: 2 }))).toEqual({
      'fast-forward': '2 tracked files in the checkout have uncommitted changes: commit or discard them first.',
      reset: '2 tracked files in the checkout have uncommitted changes: commit or discard them first.',
      none: '2 tracked files in the checkout have uncommitted changes: commit or discard them first.',
    });
  });

  it('refuses syncing a branch without an upstream but still allows a plain switch', () => {
    expect(reasons(plan({ upstream: undefined, upstreamHead: undefined }))).toEqual({
      'fast-forward': 'main tracks no upstream.', reset: 'main tracks no upstream.', none: '',
    });
  });

  it('refuses a plain switch onto the branch already checked out', () => {
    expect(reasons(plan({ current: 'main' })).none).toBe('The checkout is already on main.');
  });

  it('refuses every move when another worktree has the branch checked out', () => {
    expect(reasons(plan({ checkedOutIn: '/src/acme-main' })).none).toBe('main is checked out in /src/acme-main.');
  });
});
