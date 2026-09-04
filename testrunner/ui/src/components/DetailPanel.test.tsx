import { afterEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import { DetailPanel } from './DetailPanel';
import type { Test } from '../types';

const editableTest: Test = {
  name: 'works',
  framework: 'vitest',
  file: 'sum.test.ts',
  line: 3,
  failed: true,
};

describe('DetailPanel test edit actions', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('hides edit actions when source edits are unsupported', () => {
    const onTestEdit = vi.fn();
    render(<DetailPanel test={editableTest} testEditSupported={false} onTestEdit={onTestEdit} />);

    expect(screen.queryByText('Skip Test')).toBeNull();
    expect(screen.queryByText('Delete File')).toBeNull();
  });

  it('does not call onTestEdit when confirmation is cancelled', () => {
    const onTestEdit = vi.fn();
    vi.spyOn(window, 'confirm').mockReturnValue(false);
    render(<DetailPanel test={editableTest} testEditSupported onTestEdit={onTestEdit} />);

    fireEvent.click(screen.getByText('Skip Test'));

    expect(window.confirm).toHaveBeenCalledWith('Skip test works?');
    expect(onTestEdit).not.toHaveBeenCalled();
  });

  it('calls onTestEdit with the confirmed action and scope', () => {
    const onTestEdit = vi.fn();
    vi.spyOn(window, 'confirm').mockReturnValue(true);
    render(<DetailPanel test={editableTest} testEditSupported onTestEdit={onTestEdit} />);

    fireEvent.click(screen.getByText('Delete File'));

    expect(window.confirm).toHaveBeenCalledWith('Delete file sum.test.ts?');
    expect(onTestEdit).toHaveBeenCalledWith(editableTest, 'delete', 'file');
  });
});

describe('DetailPanel fixture CEL trace', () => {
  it('renders the native trace instead of the plain expression', () => {
    const trace = 'cel: actual == expected\n     │         │\n     │         └─ int(2)\n     └─ int(1)';
    render(<DetailPanel test={{
      name: 'CEL assertion',
      framework: 'fixture',
      failed: true,
      context: {
        cel_expression: 'actual == expected',
        cel_trace: trace,
      },
    }} />);

    expect(screen.getByText((_, element) => element?.textContent === trace)).toBeTruthy();
    expect(screen.queryByText('actual == expected')).toBeNull();
  });
});
