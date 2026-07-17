import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { LogViewer } from './LogViewer';

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, onClick, className }: { children: React.ReactNode; onClick: () => void; className?: string }) => (
    <button className={className} onClick={onClick}>{children}</button>
  ),
}));

function renderedLog(container: HTMLElement): HTMLPreElement {
  const pre = container.querySelector('pre');
  if (!pre) throw new Error('expected a log pre element');
  return pre;
}

describe('LogViewer', () => {
  it('shows the final collapsed lines through expand and collapse', () => {
    const logs = Array.from({ length: 8 }, (_, i) => `line ${i + 1}`).join('\n');
    const { container } = render(<LogViewer logs={logs} />);

    expect(renderedLog(container).textContent).toBe('line 4\nline 5\nline 6\nline 7\nline 8');
    fireEvent.click(screen.getByText('▼ Show more (8 lines)'));
    expect(renderedLog(container).textContent).toBe(logs);
    fireEvent.click(screen.getByText('▲ Collapse (8 lines)'));
    expect(renderedLog(container).textContent).toBe('line 4\nline 5\nline 6\nline 7\nline 8');
  });

  it.each([
    ['exact limit', 'line 1\nline 2\nline 3\nline 4\nline 5', 'line 1\nline 2\nline 3\nline 4\nline 5'],
    ['shorter input', 'line 1\nline 2', 'line 1\nline 2'],
    ['final newline', 'line 1\nline 2\nline 3\nline 4\nline 5\nline 6\nline 7\n', 'line 3\nline 4\nline 5\nline 6\nline 7'],
  ])('handles %s without displacing log content', (_, logs, expected) => {
    const { container } = render(<LogViewer logs={logs} />);

    expect(renderedLog(container).textContent).toBe(expected);
  });

  it('renders SGR styles while keeping HTML-looking content as text', () => {
    const logs = '\x1b[1;31mfailed\x1b[0m <script>window.pwned = true</script>';
    const { container } = render(<LogViewer logs={logs} />);
    const pre = renderedLog(container);
    const styled = Array.from(pre.querySelectorAll('span')).find(span => span.textContent === 'failed') as HTMLElement;

    expect(styled.style.color).not.toBe('');
    expect(styled.style.fontWeight).toBe('bold');
    expect(pre.textContent).toBe('failed <script>window.pwned = true</script>');
    expect(pre.textContent).not.toContain('\x1b[');
    expect(pre.querySelector('script')).toBeNull();
  });
});
