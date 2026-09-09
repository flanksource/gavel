import { useState, useMemo } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import { AnsiHtml } from '@flanksource/clicky-ui/data';

interface Props {
  logs: string;
  collapsedLines?: number;
}

function logLines(logs: string): string[] {
  const lines = logs.split('\n');
  if (logs.endsWith('\n')) lines.pop();
  return lines;
}

export function LogViewer({ logs, collapsedLines = 5 }: Props) {
  const [expanded, setExpanded] = useState(false);
  const lines = useMemo(() => logLines(logs), [logs]);
  const hasMore = lines.length > collapsedLines;
  const collapsed = collapsedLines > 0 ? lines.slice(-collapsedLines).join('\n') : '';

  return (
    <div className="relative">
      <pre
        className={`mt-0.5 ml-4 text-[11px] text-muted-foreground bg-muted rounded p-1.5 whitespace-pre-wrap overflow-y-auto border border-border transition-all duration-200 ${
          expanded ? 'max-h-[70vh]' : `max-h-[${collapsedLines * 1.4}em]`
        }`}
        style={expanded ? undefined : { maxHeight: `${collapsedLines * 1.4}em` }}
      >
        <AnsiHtml as="span" text={expanded ? logs : collapsed} />
      </pre>
      {hasMore && (
        <Button
          variant="ghost"
          className={`text-[10px] ml-4 mt-0.5 h-auto p-0 ${expanded ? 'text-gray-400' : 'text-blue-500 hover:text-blue-700'}`}
          onClick={() => setExpanded(!expanded)}
        >
          {expanded ? `▲ Collapse (${lines.length} lines)` : `▼ Show more (${lines.length} lines)`}
        </Button>
      )}
    </div>
  );
}
