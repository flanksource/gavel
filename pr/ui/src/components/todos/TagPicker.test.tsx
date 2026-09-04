import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';

import type { TodoTagDef } from '../../types';
import { TagPicker } from './TagPicker';
import { buildTagIndex } from './tagResolve';

const defs: TodoTagDef[] = [
  { name: 'bug', color: 'red', scope: 'builtin' },
  { name: 'flaky', color: 'amber', scope: 'workspace' },
  { name: 'docs', color: 'sky', scope: 'builtin' },
];
const index = buildTagIndex(defs);

function open(overrides: Partial<Parameters<typeof TagPicker>[0]> = {}) {
  const onPick = vi.fn();
  const onClose = vi.fn();
  render(
    <TagPicker
      candidates={defs}
      counts={{ flaky: 7, bug: 2 }}
      index={index}
      onPick={onPick}
      onClose={onClose}
      {...overrides}
    />,
  );
  return { onPick, onClose };
}

/**
 * The chip labels in the order the picker offers them. Scoped to the options
 * list so the "Create" affordance below it is not counted as a match.
 */
function offered(): string[] {
  return screen.queryAllByRole('option')
    .map(option => option.querySelector('[data-tag]')?.getAttribute('data-tag') ?? '')
    .filter(Boolean);
}

describe('TagPicker', () => {
  it('leads with the tags this project uses most', () => {
    open();
    expect(offered()).toEqual(['flaky', 'bug', 'docs']);
  });

  it('shows the usage count so a common tag is recognisable', () => {
    open();
    expect(screen.getByText('7')).toBeTruthy();
  });

  it('picks the tag that was clicked', () => {
    const { onPick } = open();
    fireEvent.click(screen.getByText('docs'));
    expect(onPick).toHaveBeenCalledWith('docs');
  });

  it('filters to the typed substring', () => {
    open();
    fireEvent.change(screen.getByLabelText('Filter tags'), { target: { value: 'fla' } });
    expect(offered()).toEqual(['flaky']);
  });

  // A partial match must not swallow the create path: "fla" narrows to "flaky"
  // and still offers "fla" itself, which is a legitimate new tag.
  it('offers creation alongside a partial match', () => {
    open();
    fireEvent.change(screen.getByLabelText('Filter tags'), { target: { value: 'fla' } });
    expect(offered()).toEqual(['flaky']);
    expect(screen.getByText('Create')).toBeTruthy();
  });

  it('picks the top match on Enter, so choosing needs no mouse', () => {
    const { onPick } = open();
    const input = screen.getByLabelText('Filter tags');
    fireEvent.change(input, { target: { value: 'doc' } });
    fireEvent.keyDown(input, { key: 'Enter' });
    expect(onPick).toHaveBeenCalledWith('docs');
  });

  it('offers to create a name nothing matches', () => {
    const { onPick } = open();
    fireEvent.change(screen.getByLabelText('Filter tags'), { target: { value: 'brand-new' } });
    fireEvent.click(screen.getByText('Create'));
    expect(onPick).toHaveBeenCalledWith('brand-new');
  });

  it('normalizes a created name the way the server stores it', () => {
    const { onPick } = open();
    fireEvent.change(screen.getByLabelText('Filter tags'), { target: { value: '  MiXeD  ' } });
    fireEvent.click(screen.getByText('Create'));
    expect(onPick).toHaveBeenCalledWith('mixed');
  });

  // The settings picker adopts existing definitions; inventing a name there
  // belongs to the "New tag" row, which also asks for a colour.
  it('cannot create when creation is disabled', () => {
    open({ allowCreate: false });
    fireEvent.change(screen.getByLabelText('Filter tags'), { target: { value: 'brand-new' } });
    expect(screen.queryByText('Create')).toBeNull();
  });

  it('closes on Escape', () => {
    const { onClose } = open();
    fireEvent.keyDown(screen.getByLabelText('Filter tags'), { key: 'Escape' });
    expect(onClose).toHaveBeenCalled();
  });

  it('closes when the pointer goes down outside it', () => {
    const { onClose } = open();
    fireEvent.mouseDown(document.body);
    expect(onClose).toHaveBeenCalled();
  });

  it('stays open when the pointer goes down inside it', () => {
    const { onClose } = open();
    fireEvent.mouseDown(screen.getByLabelText('Filter tags'));
    expect(onClose).not.toHaveBeenCalled();
  });

  it('explains an empty list rather than showing a blank panel', () => {
    open({ candidates: [], allowCreate: false, emptyLabel: 'Nothing left.' });
    expect(screen.getByText('Nothing left.')).toBeTruthy();
  });
});
