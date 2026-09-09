import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, within } from '@testing-library/react';

import type { TodoItem } from '../../types';
import { TodoLabelsMenu, labelEdit, labelStates, type LabelState } from './TodoLabelsMenu';
import { buildTagIndex } from './tagResolve';

function todo(ref: string, labels: string[]): TodoItem {
  return { ref, title: ref, status: 'pending', priority: 'medium', labels } as TodoItem;
}

const index = buildTagIndex([
  { name: 'bug', color: 'red', scope: 'builtin' },
  { name: 'flaky', color: 'amber', scope: 'workspace' },
  { name: 'docs', color: 'sky', scope: 'builtin' },
]);

describe('labelStates', () => {
  it('reports a label every todo carries as on', () => {
    const states = labelStates([todo('a', ['bug']), todo('b', ['bug'])]);
    expect(states.get('bug')).toBe('on');
  });

  it('reports a label only some carry as mixed', () => {
    const states = labelStates([todo('a', ['bug']), todo('b', [])]);
    expect(states.get('bug')).toBe('mixed');
  });

  it('says nothing about a label nobody carries', () => {
    expect(labelStates([todo('a', ['bug'])]).has('docs')).toBe(false);
  });

  // Lifecycle labels are machine-managed; showing them as editable tags would
  // invite someone to strip a todo's own status.
  it('ignores machine-managed lifecycle labels', () => {
    const states = labelStates([todo('a', ['status:open', 'priority:high', 'bug'])]);
    expect(states.has('status:open')).toBe(false);
    expect(states.has('priority:high')).toBe(false);
    expect(states.get('bug')).toBe('on');
  });

  it('counts a label once even when a todo repeats it', () => {
    const states = labelStates([todo('a', ['bug', 'bug']), todo('b', ['bug'])]);
    expect(states.get('bug')).toBe('on');
  });
});

describe('labelEdit', () => {
  const baseline = new Map<string, LabelState>([['bug', 'on'], ['flaky', 'mixed']]);

  it('adds a label turned on', () => {
    const draft = new Map(baseline).set('docs', 'on');
    expect(labelEdit(baseline, draft)).toEqual({ add: ['docs'], remove: [] });
  });

  it('removes a label turned off', () => {
    const draft = new Map(baseline).set('bug', 'off');
    expect(labelEdit(baseline, draft)).toEqual({ add: [], remove: ['bug'] });
  });

  it('resolves a mixed label in whichever direction it was taken', () => {
    expect(labelEdit(baseline, new Map(baseline).set('flaky', 'on')).add).toEqual(['flaky']);
    expect(labelEdit(baseline, new Map(baseline).set('flaky', 'off')).remove).toEqual(['flaky']);
  });

  // The whole point of applying once: the edit is the net against where things
  // started, not a replay of every click.
  it('ignores a label cycled back to where it started', () => {
    const draft = new Map(baseline).set('bug', 'off').set('bug', 'on');
    expect(labelEdit(baseline, draft)).toEqual({ add: [], remove: [] });
  });

  it('leaves a mixed label alone when it was never touched', () => {
    expect(labelEdit(baseline, new Map(baseline))).toEqual({ add: [], remove: [] });
  });
});

describe('TodoLabelsMenu', () => {
  function open(props: Partial<Parameters<typeof TodoLabelsMenu>[0]> = {}) {
    const onApply = vi.fn();
    render(
      <TodoLabelsMenu
        index={index}
        counts={{ bug: 4, flaky: 2 }}
        selected={[todo('a', ['bug']), todo('b', [])]}
        complete
        busy={false}
        onApply={onApply}
        {...props}
      />,
    );
    return { onApply };
  }

  function box(name: string): HTMLInputElement {
    const chip = document.querySelector(`[data-tag="${name}"]`);
    const row = chip?.closest('label');
    return within(row as HTMLElement).getByRole('checkbox') as HTMLInputElement;
  }

  it('shows a label only some of the selection carries as indeterminate', () => {
    open();
    expect(box('bug').indeterminate).toBe(true);
    expect(box('bug').checked).toBe(false);
  });

  it('shows a label none of the selection carries as unchecked', () => {
    open();
    expect(box('docs').indeterminate).toBe(false);
    expect(box('docs').checked).toBe(false);
  });

  it('applies nothing until there is something to apply', () => {
    open();
    expect(screen.getByRole('button', { name: 'Apply' })).toHaveProperty('disabled', true);
  });

  it('applies the net edit once', () => {
    const { onApply } = open();
    fireEvent.click(box('docs'));
    fireEvent.click(screen.getByRole('button', { name: 'Apply' }));
    expect(onApply).toHaveBeenCalledWith({ add: ['docs'], remove: [] });
  });

  it('resolves a mixed label onto the whole selection in one click', () => {
    const { onApply } = open();
    fireEvent.click(box('bug'));
    fireEvent.click(screen.getByRole('button', { name: 'Apply' }));
    expect(onApply).toHaveBeenCalledWith({ add: ['bug'], remove: [] });
  });

  it('filters the list to the typed substring', () => {
    open();
    fireEvent.change(screen.getByLabelText('Filter labels'), { target: { value: 'fla' } });
    expect(document.querySelector('[data-tag="flaky"]')).toBeTruthy();
    expect(document.querySelector('[data-tag="bug"]')).toBeNull();
  });

  // With a filter-scoped selection the todos behind it were never fetched, so
  // claiming a label is "on all of them" would be a guess.
  it('drops tri-state when the selection is not fully in hand', () => {
    open({ complete: false });
    expect(box('bug').indeterminate).toBe(false);
    expect(box('bug').checked).toBe(false);
  });

  it('locks every control while a bulk call is in flight', () => {
    open({ busy: true });
    expect(box('bug').disabled).toBe(true);
  });
});
