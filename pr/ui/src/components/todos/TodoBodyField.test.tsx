import { useState } from 'react';
import { createEvent, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { TodoBodyField } from './TodoBodyField';

function clipboard(values: Partial<Record<'text/markdown' | 'text/html' | 'text/plain', string>>) {
  return {
    files: [],
    items: [],
    getData: (type: string) => values[type as keyof typeof values] ?? '',
  };
}

function StatefulField({ initial = 'prefix replace suffix' }: { initial?: string }) {
  const [value, setValue] = useState(initial);
  return (
    <TodoBodyField
      label="Body"
      value={value}
      onChange={setValue}
      placeholder="Details"
    />
  );
}

afterEach(() => vi.restoreAllMocks());

describe('TodoBodyField', () => {
  it('renders a tall, vertically resizable Markdown textarea with paste guidance', () => {
    render(<StatefulField initial="" />);

    const textarea = screen.getByLabelText('Body');
    expect(textarea.className).toContain('h-64');
    expect(textarea.className).toContain('resize-y');
    expect(textarea.className).toContain('font-mono');
    expect(screen.getByText(/HTML.*Markdown/i)).toBeTruthy();
    expect(screen.getByText(/multiline.*code block/i)).toBeTruthy();
  });

  it('replaces the selection with normalized text, prevents native text insertion, and restores the caret', async () => {
    render(<StatefulField />);
    const textarea = screen.getByLabelText('Body') as HTMLTextAreaElement;
    textarea.setSelectionRange(7, 14);
    const event = createEvent.paste(textarea, {
      bubbles: true,
      cancelable: true,
      clipboardData: clipboard({ 'text/plain': '\u001b[31mcommand\u001b[0m\r\noutput' }),
    });

    fireEvent(textarea, event);

    expect(event.defaultPrevented).toBe(true);
    expect(textarea.value).toBe('prefix \n\n```text\ncommand\noutput\n```\n\n suffix');
    await waitFor(() => expect(textarea.selectionStart).toBe(35));
    expect(textarea.selectionEnd).toBe(35);
  });

  it('lets image-only paste bubble to the existing attachment listener', () => {
    render(<StatefulField initial="unchanged" />);
    const textarea = screen.getByLabelText('Body') as HTMLTextAreaElement;
    const onWindowPaste = vi.fn();
    window.addEventListener('paste', onWindowPaste);
    const event = createEvent.paste(textarea, {
      bubbles: true,
      cancelable: true,
      clipboardData: {
        ...clipboard({}),
        files: [new File(['image'], 'screen.png', { type: 'image/png' })],
      },
    });

    fireEvent(textarea, event);

    expect(event.defaultPrevented).toBe(false);
    expect(onWindowPaste).toHaveBeenCalledOnce();
    expect(textarea.value).toBe('unchanged');
    window.removeEventListener('paste', onWindowPaste);
  });

  it('normalizes mixed text and image paste without stopping attachment handling', () => {
    render(<StatefulField initial="" />);
    const textarea = screen.getByLabelText('Body') as HTMLTextAreaElement;
    const onWindowPaste = vi.fn();
    window.addEventListener('paste', onWindowPaste);
    const event = createEvent.paste(textarea, {
      bubbles: true,
      cancelable: true,
      clipboardData: {
        ...clipboard({ 'text/plain': 'caption' }),
        files: [new File(['image'], 'screen.png', { type: 'image/png' })],
      },
    });

    fireEvent(textarea, event);

    expect(event.defaultPrevented).toBe(true);
    expect(onWindowPaste).toHaveBeenCalledOnce();
    expect(textarea.value).toBe('caption');
    window.removeEventListener('paste', onWindowPaste);
  });
});
