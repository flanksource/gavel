import { beforeEach, describe, expect, it } from 'vitest';
import { defaultLayout, loadLayout, saveLayout } from './todoLayout';

const STORAGE_KEY = 'gavel.pr-ui.todoLayout.v1';

describe('todoLayout', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('defaults to the split (master-detail) layout', () => {
    expect(defaultLayout()).toBe('split');
    expect(loadLayout()).toBe('split');
  });

  it('round-trips a saved layout', () => {
    saveLayout('full');
    expect(loadLayout()).toBe('full');
    saveLayout('split');
    expect(loadLayout()).toBe('split');
  });

  // A value written by an older build (or a hand-edited key) must not put the
  // tab into a layout that does not exist — the toggle would then have no
  // pressed state and neither branch would render.
  it('falls back to the default for an unrecognised stored value', () => {
    localStorage.setItem(STORAGE_KEY, 'tabular');
    expect(loadLayout()).toBe('split');
  });
});
