import { afterEach, describe, expect, it } from 'vitest';
import { cleanup, render, screen, within } from '@testing-library/react';
import { LintTasks } from './LintTasks';
import type { Test } from '../types';

afterEach(cleanup);

const task = (name: string, status: string, message?: string): Test => ({
  name,
  framework: 'task',
  context: { status, type: 'task' },
  ...(message ? { message } : {}),
});

const promptGroup = (children: Test[]): Test => ({
  name: 'lint prompt',
  framework: 'task',
  context: { status: 'running', type: 'group' },
  children,
});

describe('LintTasks', () => {
  it('lists each live prompt task under its group with its status text', () => {
    render(<LintTasks tests={[promptGroup([
      task('messaging-review:report.mdx', 'running', 'tool: read_file'),
      task('tone-review:report.mdx', 'pending'),
    ])]} />);

    const strip = screen.getByRole('region', { name: 'Tasks' });
    const rows = within(strip).getAllByRole('listitem');
    expect(rows.map(row => row.textContent)).toEqual([
      'messaging-review:report.mdxtool: read_file',
      'tone-review:report.mdxpending',
    ]);
    expect(within(strip).getByText('lint prompt')).toBeTruthy();
  });

  it('hides the finished tasks of a live group', () => {
    render(<LintTasks tests={[promptGroup([task('a', 'success'), task('b', 'running', 'thinking')])]} />);

    expect(screen.getAllByRole('listitem').map(row => row.textContent)).toEqual(['bthinking']);
  });

  it('renders nothing when no task is live', () => {
    const { container } = render(<LintTasks tests={[
      { name: 'TestFoo', framework: 'go test', passed: true },
      { name: 'lint prompt', framework: 'task', context: { status: 'success' } },
    ]} />);

    expect(container.innerHTML).toBe('');
  });
});
