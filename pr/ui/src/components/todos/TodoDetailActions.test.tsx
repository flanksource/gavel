import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { TodoStatus } from '../../types';
import { HeaderActionsMenu } from './TodoDetailActions';
import { emptyTagIndex } from './tagResolve';

vi.mock('./LinkPRDialog', () => ({ LinkPRDialog: () => null }));
vi.mock('./TodoActionSubmenus', () => ({ MoveSubmenu: () => null, SetParentSubmenu: () => null, UnlinkPRSubmenu: () => null }));

describe('TODO header advanced run', () => {
  it.each<{ status: TodoStatus; awaitingHumanAction: boolean; enabled: boolean }>([
    { status: 'review', awaitingHumanAction: true, enabled: true },
    { status: 'ask', awaitingHumanAction: true, enabled: false },
    { status: 'pending', awaitingHumanAction: false, enabled: true },
  ])('allows opening approval from $status without launching a run', ({ status, awaitingHumanAction, enabled }) => {
    const onAdvanced = vi.fn();
    const onRun = vi.fn();
    const unused = vi.fn();
    render(<HeaderActionsMenu
      todo={{ ref: 'review-todo', title: 'Review selected plan', status, priority: 'medium' }}
      dir="/workspace"
      busy={false}
      runBusy={false}
      closed={false}
      sessionInProgress={false}
      awaitingHumanAction={awaitingHumanAction}
      fullTodoId="review-todo"
      labels={[]}
      tags={emptyTagIndex()}
      transferTargets={[]}
      canTransfer={false}
      parentCandidates={[]}
      hasChildren={false}
      onCopy={unused}
      onEditTitle={unused}
      onStatus={unused}
      onPriority={unused}
      onTransfer={unused}
      onResume={unused}
      onRun={onRun}
      onPlan={unused}
      onAdvanced={onAdvanced}
      onVerify={unused}
      onToggleClosed={unused}
      onArchive={unused}
      onPushToGithub={unused}
      onSetParent={unused}
      onRemoveParent={unused}
    />);
    fireEvent.click(screen.getByRole('button', { name: 'Issue actions' }));
    const advanced = screen.getByRole<HTMLButtonElement>('button', { name: /^Advanced run/ });
    expect(advanced.disabled).toBe(!enabled);
    expect(screen.getByRole<HTMLButtonElement>('button', { name: /^Run todo/ }).disabled).toBe(awaitingHumanAction);
    fireEvent.click(advanced);
    expect(onAdvanced.mock.calls).toEqual(enabled ? [['run']] : []);
    expect(onRun).not.toHaveBeenCalled();
  });
});
