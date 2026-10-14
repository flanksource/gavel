import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { TodoPRLink } from './TodoPRLink';

describe('TODO PR link editing', () => {
  it('saves an explicit URL and unlinks without changing the TODO body', async () => {
    const pr = { number: 27, url: 'https://github.com/acme/service/pull/27' };
    const onSave = vi.fn().mockResolvedValue(true);
    render(<TodoPRLink pr={pr} onSave={onSave} />);
    expect(screen.getByRole('link', { name: 'PR #27' }).getAttribute('href')).toBe(pr.url);
    fireEvent.click(screen.getByRole('button', { name: 'Edit PR link' }));
    fireEvent.change(screen.getByLabelText('Pull request URL'), { target: { value: 'https://dev.azure.com/acme/product/_git/service/pullrequest/28' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save PR link' }));
    await waitFor(() => expect(onSave).toHaveBeenCalledWith({ pr: { number: 28, url: 'https://dev.azure.com/acme/product/_git/service/pullrequest/28' } }));
    await waitFor(() => expect(screen.queryByRole('button', { name: 'Save PR link' })).toBeNull());
    fireEvent.click(screen.getByRole('button', { name: 'Edit PR link' }));
    fireEvent.change(screen.getByLabelText('Pull request URL'), { target: { value: '' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save PR link' }));
    await waitFor(() => expect(onSave).toHaveBeenCalledWith({ clearPR: true }));
  });
});
