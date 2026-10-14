import { describe, expect, it } from 'vitest';
import { hostControlsVisible, tabFor, visibleTabs } from './hostedUI';
import type { Tab } from './routes';

const allTabs: Tab[] = ['prs', 'projects', 'todos', 'sessions', 'tasks', 'prompts', 'activity'];

describe('hosted UI', () => {
  it('keeps every tab and host control locally', () => {
    expect(visibleTabs('none', allTabs)).toEqual(allTabs);
    expect(hostControlsVisible('none')).toBe(true);
    expect(tabFor('none', 'sessions')).toBe('sessions');
  });

  it('shows a hosted org only the PR tab and no host controls', () => {
    expect(visibleTabs('clerk', allTabs)).toEqual(['prs']);
    expect(hostControlsVisible('clerk')).toBe(false);
  });

  it('opens the PR tab for a hosted link to a host-only tab', () => {
    for (const tab of allTabs) {
      expect(tabFor('clerk', tab)).toBe('prs');
    }
  });
});
