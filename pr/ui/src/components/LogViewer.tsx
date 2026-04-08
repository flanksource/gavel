import { useState, useMemo } from 'preact/hooks';

interface Props {
  logs: string;
  collapsedLines?: number;
  bgClass?: string;
  borderClass?: string;
}

export function LogViewer({ logs, collapsedLines = 5, bgClass = 'bg-gray-50', borderClass = 'border-gray-100' }: Props) {
  const [expanded, setExpanded] = useState(false);
  const lines = useMemo(() => logs.split('\n'), [logs]);
  const hasMore = lines.length > collapsedLines;

  return (
    <div class="relative">
      <pre
        class={`mt-0.5 ml-4 text-[11px] text-gray-500 ${bgClass} rounded p-1.5 whitespace-pre-wrap overflow-y-auto border ${borderClass} transition-all duration-200 ${
          expanded ? 'max-h-[70vh]' : `max-h-[${collapsedLines * 1.4}em]`
        }`}
        style={expanded ? undefined : { maxHeight: `${collapsedLines * 1.4}em` }}
      >
        {expanded ? logs : lines.slice(0, collapsedLines).join('\n')}
      </pre>
      {hasMore && (
        <button
          class={`text-[10px] ml-4 mt-0.5 ${expanded ? 'text-gray-400' : 'text-blue-500 hover:text-blue-700'}`}
          onClick={() => setExpanded(!expanded)}
        >
          {expanded ? `▲ Collapse (${lines.length} lines)` : `▼ Show more (${lines.length} lines)`}
        </button>
      )}
    </div>
  );
}
