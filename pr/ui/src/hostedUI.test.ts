import { describe, expect, it } from 'vitest';
import { githubScopeVisible, hostControlsVisible, runtimeChoiceVisible, tabFor, todoVisibilityVisible, visibleTabs } from './hostedUI';
import type { Tab } from './routes';

const allTabs: Tab[] = ['prs', 'projects', 'todos', 'sessions', 'tasks', 'prompts', 'activity', 'runs'];

describe('hosted UI', () => {
  it('keeps every tab but the sandbox runs and every host control locally', () => {
    expect(visibleTabs('none', allTabs)).toEqual(allTabs.filter(tab => tab !== 'runs'));
    expect(hostControlsVisible('none')).toBe(true);
    expect(tabFor('none', 'sessions')).toBe('sessions');
    expect(tabFor('none', 'runs'), 'a local dashboard runs nothing in a sandbox').toBe('prs');
  });

  it('shows a hosted org only the PR, todo and sandbox run tabs and no host controls', () => {
    expect(visibleTabs('clerk', allTabs)).toEqual(['prs', 'todos', 'runs']);
    expect(tabFor('clerk', 'runs')).toBe('runs');
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
    for (const tab of allTabs.filter(tab => tab !== 'todos' && tab !== 'runs')) {
      expect(tabFor('clerk', tab)).toBe('prs');
    }
    expect(tabFor('clerk', 'todos')).toBe('todos');
  });
});
