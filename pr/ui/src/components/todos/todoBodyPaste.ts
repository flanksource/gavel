import TurndownService from '@joplin/turndown';
import { gfm } from '@joplin/turndown-plugin-gfm';
import stripAnsi from 'strip-ansi';

export interface NormalizedTodoBodyPaste {
  text: string;
  block: boolean;
}

export interface TodoBodyInsertion {
  value: string;
  caret: number;
}

export function normalizeTodoBodyPaste(
  clipboard: Pick<DataTransfer, 'getData'>,
): NormalizedTodoBodyPaste | null {
  const markdown = clipboard.getData('text/markdown');
  if (markdown) return { text: markdown, block: /[\r\n]/.test(markdown) };

  const html = clipboard.getData('text/html');
  if (html) {
    const text = htmlToMarkdown(html);
    return text ? { text, block: true } : null;
  }

  const plain = stripAnsi(clipboard.getData('text/plain')).replace(/\r\n?/g, '\n');
  if (!plain) return null;
  if (!plain.includes('\n')) return { text: plain, block: false };

  const longestBacktickRun = Math.max(0, ...Array.from(plain.matchAll(/`+/g), match => match[0].length));
  const fence = '`'.repeat(Math.max(3, longestBacktickRun + 1));
  return {
    text: `${fence}text\n${plain}${plain.endsWith('\n') ? '' : '\n'}${fence}`,
    block: true,
  };
}

export function insertTodoBodyPaste(
  value: string,
  start: number,
  end: number,
  paste: NormalizedTodoBodyPaste,
): TodoBodyInsertion {
  const before = value.slice(0, start);
  const after = value.slice(end);
  if (!paste.block) {
    return {
      value: before + paste.text + after,
      caret: before.length + paste.text.length,
    };
  }

  const leading = blockBoundaryBefore(before);
  const trailing = blockBoundaryAfter(after);
  return {
    value: before + leading + paste.text + trailing + after,
    caret: before.length + leading.length + paste.text.length,
  };
}

const htmlConverter = new TurndownService({
  headingStyle: 'atx',
  bulletListMarker: '-',
  codeBlockStyle: 'fenced',
  fence: '```',
});
htmlConverter.use(gfm);
htmlConverter.remove(['script', 'style', 'noscript', 'template']);

function htmlToMarkdown(html: string): string {
  return htmlConverter.turndown(html).trim();
}

function blockBoundaryBefore(value: string): string {
  if (!value || value.endsWith('\n\n')) return '';
  return value.endsWith('\n') ? '\n' : '\n\n';
}

function blockBoundaryAfter(value: string): string {
  if (!value || value.startsWith('\n\n')) return '';
  return value.startsWith('\n') ? '\n' : '\n\n';
}
