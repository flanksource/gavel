import { describe, expect, it } from 'vitest';
import { githubScopeVisible, hostControlsVisible, runtimeChoiceVisible, tabFor, todoVisibilityVisible, visibleTabs } from './hostedUI';
import type { Tab } from './routes';

const allTabs: Tab[] = ['prs', 'projects', 'todos', 'comments', 'sessions', 'tasks', 'prompts', 'activity', 'runs', 'platform'];

describe('hosted UI', () => {
  it('keeps every tab but the sandbox runs and the platform dashboard, and every host control, locally', () => {
    expect(visibleTabs('none', allTabs)).toEqual(allTabs.filter(tab => tab !== 'runs' && tab !== 'platform'));
    expect(visibleTabs('none', allTabs, true), 'a local dashboard has no platform').not.toContain('platform');
    expect(hostControlsVisible('none')).toBe(true);
    expect(tabFor('none', 'sessions')).toBe('sessions');
    expect(tabFor('none', 'runs'), 'a local dashboard runs nothing in a sandbox').toBe('prs');
    expect(tabFor('none', 'platform')).toBe('prs');
  });

  it('shows a hosted org the PR, project, todo, comments and sandbox run tabs', () => {
    expect(visibleTabs('clerk', allTabs)).toEqual(['prs', 'projects', 'todos', 'comments', 'runs']);
    expect(tabFor('clerk', 'projects')).toBe('projects');
    expect(tabFor('clerk', 'runs')).toBe('runs');
    expect(hostControlsVisible('clerk')).toBe(false);
  });

  it('shows the platform tab only to a hosted platform admin, and opens it for anyone hosted', () => {
    expect(visibleTabs('clerk', allTabs, true)).toEqual(['prs', 'projects', 'todos', 'comments', 'runs', 'platform']);
    expect(visibleTabs('clerk', allTabs, false)).not.toContain('platform');
    expect(tabFor('clerk', 'platform'), 'a link opens the tab, which says why a non-admin cannot read it').toBe('platform');
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
    for (const tab of allTabs.filter(tab => !['projects', 'todos', 'comments', 'runs', 'platform'].includes(tab))) {
      expect(tabFor('clerk', tab)).toBe('prs');
    }
    expect(tabFor('clerk', 'todos')).toBe('todos');
  });
});
