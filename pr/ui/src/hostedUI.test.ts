import { describe, expect, it } from 'vitest';
import { githubScopeVisible, hostControlsVisible, runtimeChoiceVisible, tabFor, todoVisibilityVisible, visibleTabs } from './hostedUI';
import type { Tab } from './routes';

const allTabs: Tab[] = ['prs', 'projects', 'todos', 'sessions', 'tasks', 'prompts', 'activity'];

describe('hosted UI', () => {
  it('keeps every tab and host control locally', () => {
    expect(visibleTabs('none', allTabs)).toEqual(allTabs);
    expect(hostControlsVisible('none')).toBe(true);
    expect(tabFor('none', 'sessions')).toBe('sessions');
  });

  it('shows a hosted org only the PR and todo tabs and no host controls', () => {
    expect(visibleTabs('clerk', allTabs)).toEqual(['prs', 'todos']);
    expect(hostControlsVisible('clerk')).toBe(false);
  });

  it('lets only a local dashboard choose a run\'s runtime; a hosted run uses the repository\'s', () => {
    expect(runtimeChoiceVisible('none')).toBe(true);
    expect(runtimeChoiceVisible('clerk')).toBe(false);
  });

  it('shows todo visibility only to a hosted org, whose members share its todos', () => {
    expect(todoVisibilityVisible('clerk')).toBe(true);
    expect(todoVisibilityVisible('none')).toBe(false);
  });

  it('shows the GitHub scope chooser only locally, where repos come from the user\'s token', () => {
    expect(githubScopeVisible('none')).toBe(true);
    expect(githubScopeVisible('clerk')).toBe(false);
  });

  it('opens the PR tab for a hosted link to a host-only tab', () => {
    for (const tab of allTabs.filter(tab => tab !== 'todos')) {
      expect(tabFor('clerk', tab)).toBe('prs');
    }
    expect(tabFor('clerk', 'todos')).toBe('todos');
  });
});
