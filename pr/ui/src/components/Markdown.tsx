import { Marked, Renderer } from 'marked';

interface Props {
  text: string;
  class?: string;
}

const renderer = new Renderer();

renderer.code = ({ text, lang }: { text: string; lang?: string }) =>
  `<pre class="bg-gray-50 rounded p-2 text-xs overflow-x-auto border border-gray-100 my-1.5"><code${lang ? ` class="language-${lang}"` : ''}>${text}</code></pre>`;

renderer.codespan = ({ text }: { text: string }) =>
  `<code class="bg-gray-100 rounded px-1 text-xs">${text}</code>`;

renderer.heading = ({ text, depth }: { text: string; depth: number }) => {
  const cls = depth <= 1
    ? 'font-bold text-base mt-2 mb-1'
    : depth === 2
      ? 'font-semibold text-sm mt-2 mb-1'
      : 'font-semibold text-sm mt-2 mb-1';
  return `<h${depth} class="${cls}">${text}</h${depth}>`;
};

renderer.link = ({ href, text }: { href: string; text: string }) =>
  `<a href="${href}" target="_blank" rel="noopener" class="text-blue-600 hover:underline">${text}</a>`;

renderer.blockquote = ({ text }: { text: string }) =>
  `<blockquote class="border-l-2 border-gray-300 pl-2 text-gray-500 italic my-1">${text}</blockquote>`;

renderer.table = ({ header, body }: { header: string; body: string }) =>
  `<table class="text-xs border-collapse my-2 w-full"><thead>${header}</thead><tbody>${body}</tbody></table>`;

renderer.tablerow = ({ text }: { text: string }) =>
  `<tr class="border-b border-gray-100">${text}</tr>`;

renderer.tablecell = ({ text, header }: { text: string; header: boolean }) => {
  const tag = header ? 'th' : 'td';
  const cls = header
    ? 'px-2 py-1 text-left font-medium text-gray-600 bg-gray-50 border border-gray-200'
    : 'px-2 py-1 border border-gray-100';
  return `<${tag} class="${cls}">${text}</${tag}>`;
};

renderer.list = ({ body, ordered }: { body: string; ordered: boolean }) => {
  const tag = ordered ? 'ol' : 'ul';
  const cls = ordered ? 'list-decimal ml-4 my-1' : 'list-disc ml-4 my-1';
  return `<${tag} class="${cls}">${body}</${tag}>`;
};

renderer.listitem = ({ text }: { text: string }) =>
  `<li class="my-0.5">${text}</li>`;

renderer.image = ({ href, text }: { href: string; text: string }) =>
  `<img src="${href}" alt="${text}" class="max-w-full rounded my-1" />`;

renderer.hr = () => '<hr class="border-gray-200 my-2" />';

renderer.paragraph = ({ text }: { text: string }) =>
  `<p class="mt-1.5">${text}</p>`;

renderer.html = ({ text }: { text: string }) => text;

const marked = new Marked({
  renderer,
  gfm: true,
  breaks: true,
});

export function Markdown({ text, class: cls }: Props) {
  const html = marked.parse(text) as string;
  return <div class={`prose prose-sm max-w-none ${cls || ''}`} dangerouslySetInnerHTML={{ __html: html }} />;
}
