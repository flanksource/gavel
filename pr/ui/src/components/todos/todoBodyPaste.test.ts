import { describe, expect, it } from 'vitest';
import { insertTodoBodyPaste, normalizeTodoBodyPaste } from './todoBodyPaste';

function clipboard(values: Partial<Record<'text/markdown' | 'text/html' | 'text/plain', string>>) {
  return { getData: (type: string) => values[type as keyof typeof values] ?? '' };
}

describe('normalizeTodoBodyPaste', () => {
  it('inserts native Markdown verbatim before other clipboard formats', () => {
    const markdown = '# Native\r\n\r\n- exact';

    expect(normalizeTodoBodyPaste(clipboard({
      'text/markdown': markdown,
      'text/html': '<h1>HTML</h1>',
      'text/plain': 'plain',
    }))).toEqual({ text: markdown, block: true });
  });

  it('converts safe HTML structure to GFM and removes executable content', () => {
    const result = normalizeTodoBodyPaste(clipboard({
      'text/html': [
        '<h1>Deploy notes</h1>',
        '<ul><li>Build</li><li><a href="https://example.test/run">Run</a></li></ul>',
        '<table><thead><tr><th>Step</th><th>State</th></tr></thead><tbody><tr><td>lint</td><td>pass</td></tr></tbody></table>',
        '<pre><code class="language-bash">printf "ready\\n"</code></pre>',
        '<a href="javascript:alert(1)">Unsafe link</a>',
        '<script>alert("unsafe")</script><style>body{display:none}</style>',
      ].join(''),
    }));

    expect(result?.block).toBe(true);
    expect(result?.text).toContain('# Deploy notes');
    expect(result?.text).toContain('- Build');
    expect(result?.text).toContain('[Run](https://example.test/run)');
    expect(result?.text).toMatch(/\|\s*Step\s*\|\s*State\s*\|/);
    expect(result?.text).toContain('```bash');
    expect(result?.text).toContain('printf "ready\\n"');
    expect(result?.text).not.toContain('javascript:');
    expect(result?.text).not.toContain('alert("unsafe")');
    expect(result?.text).not.toContain('display:none');
  });

  it('strips ANSI, normalizes CRLF, and fences every multiline plain-text paste', () => {
    expect(normalizeTodoBodyPaste(clipboard({
      'text/plain': '\u001b[31mFAIL\u001b[0m\r\nexpected true\rreceived false',
    }))).toEqual({
      text: '```text\nFAIL\nexpected true\nreceived false\n```',
      block: true,
    });
  });

  it('preserves a trailing newline inside a multiline plain-text fence', () => {
    expect(normalizeTodoBodyPaste(clipboard({ 'text/plain': 'first\nsecond\n' }))).toEqual({
      text: '```text\nfirst\nsecond\n```',
      block: true,
    });
  });

  it('leaves normalized single-line plain text unfenced', () => {
    expect(normalizeTodoBodyPaste(clipboard({ 'text/plain': '\u001b[32mready\u001b[0m' }))).toEqual({
      text: 'ready',
      block: false,
    });
  });

  it('uses a longer fence than backtick runs in terminal output', () => {
    expect(normalizeTodoBodyPaste(clipboard({ 'text/plain': 'before\n```nested```\nafter' }))).toEqual({
      text: '````text\nbefore\n```nested```\nafter\n````',
      block: true,
    });
  });

  it('returns null for an empty textual clipboard', () => {
    expect(normalizeTodoBodyPaste(clipboard({}))).toBeNull();
  });
});

describe('insertTodoBodyPaste', () => {
  it('replaces the selection and separates a Markdown block from adjacent text', () => {
    expect(insertTodoBodyPaste('prefix replace suffix', 7, 14, {
      text: '```text\ncommand output\n```',
      block: true,
    })).toEqual({
      value: 'prefix \n\n```text\ncommand output\n```\n\n suffix',
      caret: 35,
    });
  });

  it('inserts single-line text at the caret without adding block whitespace', () => {
    expect(insertTodoBodyPaste('start end', 6, 6, { text: 'middle ', block: false })).toEqual({
      value: 'start middle end',
      caret: 13,
    });
  });
});
