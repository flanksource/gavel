import type React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { TodoItem } from '../../types';
import { TodoRow, todoQuery } from './format';

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement>) => (
    // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for the Clicky Button itself.
    <button type="button" {...props}>
      {children}
    </button>
  ),
  ListMenuItem: ({ children, active: _active, selected: _selected, ...props }: { children?: React.ReactNode; active?: boolean; selected?: boolean }) => (
    <div {...props}>{children}</div>
  ),
}));

vi.mock('@flanksource/clicky-ui/icons', async (importOriginal) => ({
  ...(await importOriginal<object>()),
  UiListDashes: (props: React.SVGProps<SVGSVGElement>) => <svg {...props} />,
  UiBeaker: (props: React.SVGProps<SVGSVGElement>) => <svg {...props} />,
}));

const baseTodo: TodoItem = {
  ref: 'todo-1',
  title: 'Ship the feature',
  status: 'pending',
  priority: 'medium',
};

describe('todoQuery', () => {
  it('carries only the native workspace directory', () => {
    expect(todoQuery('/work/repo')).toBe('dir=%2Fwork%2Frepo');
    expect(todoQuery('')).toBe('');
  });
});

describe('TodoRow', () => {
  it('shows a plan indicator when hasPlan is true', () => {
    render(<TodoRow todo={{ ...baseTodo, hasPlan: true, hasVerification: false }} active={false} onClick={() => {}} />);

    expect(screen.getByTitle('Plan available')).toBeTruthy();
    expect(screen.queryByTitle('Verification fixture defined')).toBeNull();
  });

  it('shows a verification indicator when hasVerification is true', () => {
    render(<TodoRow todo={{ ...baseTodo, hasPlan: false, hasVerification: true }} active={false} onClick={() => {}} />);

    expect(screen.queryByTitle('Plan available')).toBeNull();
    expect(screen.getByTitle('Verification fixture defined')).toBeTruthy();
  });

  it('shows neither indicator when the todo has no plan or verification fixture', () => {
    render(<TodoRow todo={{ ...baseTodo, hasPlan: false, hasVerification: false }} active={false} onClick={() => {}} />);

    expect(screen.queryByTitle('Plan available')).toBeNull();
    expect(screen.queryByTitle('Verification fixture defined')).toBeNull();
  });

  it('shows the plan indicator, verification indicator, and diff badge together', () => {
    render(
      <TodoRow
        todo={{ ...baseTodo, hasPlan: true, hasVerification: true, diff: { commits: 1, files: 2, adds: 3, dels: 1 } }}
        active={false}
        onClick={() => {}}
      />,
    );

    expect(screen.getByTitle('Plan available')).toBeTruthy();
    expect(screen.getByTitle('Verification fixture defined')).toBeTruthy();
    expect(screen.getByTitle('1 commit, 2 files changed')).toBeTruthy();
  });
});
