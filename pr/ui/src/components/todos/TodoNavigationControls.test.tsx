import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { TodoNavigationControls } from './TodoNavigationControls';

function renderControls(overrides: Partial<React.ComponentProps<typeof TodoNavigationControls>> = {}) {
  const onPrevious = vi.fn();
  const onNext = vi.fn();
  render(
    <TodoNavigationControls
      position={2}
      total={3}
      canPrevious
      canNext
      onPrevious={onPrevious}
      onNext={onNext}
      {...overrides}
    />,
  );
  return { onPrevious, onNext };
}

describe('TodoNavigationControls', () => {
  it('renders an accessible position and stops at queue boundaries', () => {
    renderControls({ position: 1, canPrevious: false });
    expect(screen.getByLabelText('Todo 1 of 3')).toBeTruthy();
    expect(screen.getByRole<HTMLButtonElement>('button', { name: 'Previous todo' }).disabled).toBe(true);
    expect(screen.getByRole<HTMLButtonElement>('button', { name: 'Next todo' }).disabled).toBe(false);
  });

  it.each([
    ['j', 'next'],
    ['n', 'next'],
    ['k', 'previous'],
    ['p', 'previous'],
  ] as const)('uses %s for %s navigation', (key, direction) => {
    const callbacks = renderControls();
    fireEvent.keyDown(window, { key });
    expect(direction === 'next' ? callbacks.onNext : callbacks.onPrevious).toHaveBeenCalledOnce();
  });

  it('leaves shortcuts alone while editing or while an overlay is open', () => {
    const callbacks = renderControls();
    const input = document.createElement('input');
    document.body.appendChild(input);
    fireEvent.keyDown(input, { key: 'j' });
    const menu = document.createElement('div');
    menu.setAttribute('role', 'menu');
    document.body.appendChild(menu);
    fireEvent.keyDown(window, { key: 'k' });
    expect(callbacks.onNext).not.toHaveBeenCalled();
    expect(callbacks.onPrevious).not.toHaveBeenCalled();
    input.remove();
    menu.remove();
  });

  it('ignores modified and repeated shortcuts', () => {
    const callbacks = renderControls();
    fireEvent.keyDown(window, { key: 'j', metaKey: true });
    fireEvent.keyDown(window, { key: 'k', repeat: true });
    expect(callbacks.onNext).not.toHaveBeenCalled();
    expect(callbacks.onPrevious).not.toHaveBeenCalled();
  });
});
