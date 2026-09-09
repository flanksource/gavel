import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { TodoSortByMenu } from './TodoSortByMenu';

describe('TodoSortByMenu', () => {
  it('changes the timestamp column without changing direction', () => {
    const onChange = vi.fn();
    render(
      <TodoSortByMenu
        sortBy={{ column: 'updated', dir: 'desc' }}
        onChange={onChange}
      />,
    );

    fireEvent.click(screen.getByRole('button', { name: 'Sort: Updated' }));
    fireEvent.click(screen.getByRole('menuitemradio', { name: 'Created' }));

    expect(onChange).toHaveBeenCalledWith({ column: 'created', dir: 'desc' });
  });

  it('toggles direction independently of the selected column', () => {
    const onChange = vi.fn();
    render(
      <TodoSortByMenu
        sortBy={{ column: 'updated', dir: 'desc' }}
        onChange={onChange}
      />,
    );

    fireEvent.click(screen.getByRole('button', { name: 'Sort descending' }));

    expect(onChange).toHaveBeenCalledWith({ column: 'updated', dir: 'asc' });
  });
});
