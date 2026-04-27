import { useState, useEffect, useRef, useCallback } from 'preact/hooks';
import { AnsiHtml } from './AnsiHtml';
import { apiUrl } from '../config';

interface OutputLine {
  text: string;
  stream: 'stdout' | 'stderr';
}

interface StreamSnapshot {
  lines: OutputLine[];
  status: 'running' | 'success' | 'failed' | 'canceled';
  command: string;
}

interface Props {
  open: boolean;
  onClose: () => void;
}

export function RerunDialog({ open, onClose }: Props) {
  const [lines, setLines] = useState<OutputLine[]>([]);
  const [status, setStatus] = useState<'running' | 'success' | 'failed' | 'canceled'>('running');
  const [command, setCommand] = useState('');
  const scrollRef = useRef<HTMLPreElement>(null);
  const autoScrollRef = useRef(true);

  const handleScroll = useCallback(() => {
    const el = scrollRef.current;
    if (!el) return;
    const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 40;
    autoScrollRef.current = atBottom;
  }, []);

  useEffect(() => {
    if (!open) return;
    setLines([]);
    setStatus('running');
    setCommand('');
    autoScrollRef.current = true;

    const es = new EventSource(apiUrl('/api/rerun/stream'));

    es.addEventListener('message', (e: MessageEvent) => {
      const snap: StreamSnapshot = JSON.parse(e.data);
      if (snap.command) setCommand(snap.command);
      setStatus(snap.status);
      if (snap.lines?.length) {
        setLines(prev => [...prev, ...snap.lines]);
      }
    });

    es.addEventListener('done', () => es.close());
    es.onerror = () => {
      if (status === 'running') es.close();
    };

    return () => es.close();
  }, [open]);

  useEffect(() => {
    if (autoScrollRef.current && scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [lines]);

  if (!open) return null;

  const statusIcon = status === 'running'
    ? 'line-md:loading-loop'
    : status === 'canceled'
      ? 'codicon:debug-stop'
    : status === 'success'
      ? 'mdi:check-circle'
      : 'mdi:close-circle';
  const statusColor = status === 'running'
    ? 'text-blue-400'
    : status === 'canceled'
      ? 'text-orange-400'
    : status === 'success'
      ? 'text-green-400'
      : 'text-red-400';

  return (
    <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/60" onClick={onClose}>
      <div
        class="bg-gray-900 rounded-lg shadow-2xl flex flex-col w-[90vw] max-w-5xl h-[80vh]"
        onClick={(e: Event) => e.stopPropagation()}
      >
        {/* Header */}
        <div class="flex items-center justify-between px-4 py-3 border-b border-gray-700">
          <div class="flex items-center gap-2 min-w-0">
            <iconify-icon icon={statusIcon} class={`text-xl ${statusColor}`} />
            <span class="text-sm font-mono text-gray-300 truncate">{command || 'rerun'}</span>
            {status !== 'running' && (
              <span class={`text-xs px-2 py-0.5 rounded ${status === 'success' ? 'bg-green-900 text-green-300' : status === 'canceled' ? 'bg-orange-900 text-orange-300' : 'bg-red-900 text-red-300'}`}>
                {status}
              </span>
            )}
          </div>
          <button
            onClick={onClose}
            class="text-gray-400 hover:text-gray-200 p-1"
            title="Close (rerun continues in background)"
          >
            <iconify-icon icon="mdi:close" class="text-xl" />
          </button>
        </div>

        {/* Output */}
        <pre
          ref={scrollRef}
          onScroll={handleScroll}
          class="flex-1 overflow-auto p-4 text-sm font-mono text-gray-200 leading-relaxed"
        >
          {lines.length === 0 && status === 'running' && (
            <span class="text-gray-500">Waiting for output...</span>
          )}
          {lines.map((line, i) => (
            <div key={i} class={line.stream === 'stderr' ? 'bg-red-950/30' : ''}>
              <AnsiHtml text={line.text} />
            </div>
          ))}
        </pre>
      </div>
    </div>
  );
}
