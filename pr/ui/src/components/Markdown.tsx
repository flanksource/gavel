interface Props {
  text: string;
  class?: string;
}

export function Markdown({ text, class: cls }: Props) {
  const html = renderMarkdown(text);
  return <div class={`prose prose-sm max-w-none ${cls || ''}`} dangerouslySetInnerHTML={{ __html: html }} />;
}

function renderMarkdown(md: string): string {
  let html = escapeHtml(md);

  // Code blocks (``` ... ```)
  html = html.replace(/```(\w*)\n([\s\S]*?)```/g, '<pre class="bg-gray-50 rounded p-2 text-xs overflow-x-auto border border-gray-100"><code>$2</code></pre>');

  // Inline code
  html = html.replace(/`([^`]+)`/g, '<code class="bg-gray-100 rounded px-1 text-xs">$1</code>');

  // Headers
  html = html.replace(/^### (.+)$/gm, '<h4 class="font-semibold text-sm mt-2 mb-1">$1</h4>');
  html = html.replace(/^## (.+)$/gm, '<h3 class="font-semibold text-sm mt-2 mb-1">$1</h3>');
  html = html.replace(/^# (.+)$/gm, '<h2 class="font-bold text-base mt-2 mb-1">$1</h2>');

  // Bold and italic
  html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
  html = html.replace(/\*(.+?)\*/g, '<em>$1</em>');

  // Links
  html = html.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank" rel="noopener" class="text-blue-600 hover:underline">$1</a>');

  // Blockquotes
  html = html.replace(/^&gt; (.+)$/gm, '<blockquote class="border-l-2 border-gray-300 pl-2 text-gray-500 italic">$1</blockquote>');

  // Unordered lists
  html = html.replace(/^[*-] (.+)$/gm, '<li class="ml-4 list-disc">$1</li>');

  // Line breaks (double newline = paragraph, single = br)
  html = html.replace(/\n\n/g, '</p><p class="mt-1.5">');
  html = html.replace(/\n/g, '<br/>');

  return `<p>${html}</p>`;
}

function escapeHtml(s: string): string {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');
}
