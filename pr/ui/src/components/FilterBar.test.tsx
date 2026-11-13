import { cleanup, fireEvent, render, screen, within } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { FilterBar, emptyFilters } from './FilterBar';

afterEach(() => { cleanup(); vi.restoreAllMocks(); });

describe('mobile PR filters', () => {
  it('offers an age filter in the phone sheet even with no results', () => {
    vi.spyOn(window, 'matchMedia').mockImplementation(query => ({
      matches: query.includes('max-width'), media: query, onchange: null,
      addListener: vi.fn(), removeListener: vi.fn(), addEventListener: vi.fn(),
      removeEventListener: vi.fn(), dispatchEvent: () => false,
    }));
    render(<FilterBar filters={emptyFilters()} onChange={vi.fn()} repos={[]} authors={[]} counts={{ open: 0, merged: 0, closed: 0, draft: 0, failing: 0, passing: 0, running: 0 }} />);

    fireEvent.click(screen.getByRole('button', { name: 'Filters' }));

    expect(within(screen.getByRole('dialog', { name: 'Filters' })).getByRole('combobox', { name: 'Age since update' })).toBeDefined();
  });

  it('clears age and author selections together when applying the phone sheet', () => {
    vi.spyOn(window, 'matchMedia').mockImplementation(query => ({
      matches: query.includes('max-width'), media: query, onchange: null,
      addListener: vi.fn(), removeListener: vi.fn(), addEventListener: vi.fn(),
      removeEventListener: vi.fn(), dispatchEvent: () => false,
    }));
    const onChange = vi.fn();
    render(<FilterBar filters={{ ...emptyFilters(), age: { day: 'include' }, authors: { '@bots': 'exclude' } }} onChange={onChange} repos={[]} authors={['@me', '@bots']} counts={{ open: 0, merged: 0, closed: 0, draft: 0, failing: 0, passing: 0, running: 0 }} />);

    fireEvent.click(screen.getByRole('button', { name: /^Filters/ }));
    fireEvent.click(screen.getByRole('button', { name: 'Clear all' }));
    fireEvent.click(screen.getByRole('button', { name: 'Apply' }));

    expect(onChange).toHaveBeenLastCalledWith(emptyFilters());
  });
});
