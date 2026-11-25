import { useCallback, useEffect, useRef, useState } from 'react';
import { AnsiHtml, Badge } from '@flanksource/clicky-ui/data';
import { CopyButton, Modal } from '@flanksource/clicky-ui/components';

interface OutputLine {
  text: string;
  stream: 'stdout' | 'stderr';
}

/** A process the rerun ran; `shell` reruns it from any directory. */
export interface RerunCommand {
  cwd?: string;
  argv: string[];
  line: string;
  shell: string;
}

export type RerunStatus = 'running' | 'success' | 'failed' | 'canceled';

interface StreamSnapshot {
  lines: OutputLine[];
  commands?: RerunCommand[];
  status: RerunStatus;
  command: string;
}

function RerunCommands({ commands }: { commands: RerunCommand[] }) {
  return (
    <div className="flex items-start gap-2 border-b border-border bg-muted px-4 py-2" data-rerun-commands="">
      <div className="min-w-0 flex-1 space-y-0.5 font-mono text-xs">
        {commands.map((cmd, i) => (
          <div key={i} className="truncate" title={cmd.cwd ? `${cmd.line}\ncwd: ${cmd.cwd}` : cmd.line}>
            <span className="text-muted-foreground">$ </span>{cmd.line}
          </div>
        ))}
      </div>
      <CopyButton
        value={() => commands.map(cmd => cmd.shell).join('\n')}
        label={commands.length > 1 ? 'Copy rerun commands' : 'Copy rerun command'}
      />
    </div>
  );
}

const MAX_CONNECT_ATTEMPTS = 20;
const RETRY_MS = 250;

const STATUS_TONE ={ running: 'info', success: 'success', failed: 'danger', canceled: 'warning' } as const;

// A stderr write is tinted red; the text keeps its own ANSI styling inside.
function lineText(line: OutputLine): string {
  return line.stream === 'stderr' ? `\x1b[31m${line.text}\x1b[0m` : line.text;
}

interface Props {
  open: boolean;
  onClose: () => void;
  /** URL of the rerun's SSE stream (/api/rerun/stream under the run's base). */
  streamUrl: string;
}

/**
 * The output of the rerun in progress, streamed from the server. The server
 * sends each write once, and replays them all to a client that reconnects, so
 * every (re)connection starts the output over.
 */
export function RerunDialog({ open, onClose, streamUrl }: Props) {
  const [lines, setLines] = useState<OutputLine[]>([]);
  const [status, setStatus] = useState<RerunStatus>('running');
  const [command, setCommand] = useState('');
  const [commands, setCommands] = useState<RerunCommand[]>([]);
  const [error, setError] = useState<string | undefined>(undefined);
  const scrollRef = useRef<HTMLPreElement>(null);
  const autoScrollRef = useRef(true);
  // The status the stream last reported, which the error handler of the open
  // connection must read: its closure would otherwise keep the first render's.
  const statusRef = useRef<RerunStatus>('running');

  const handleScroll = useCallback(() => {
    const el = scrollRef.current;
    if (!el) return;
    autoScrollRef.current = el.scrollHeight - el.scrollTop - el.clientHeight < 40;
  }, []);

  // attempt counts the connections tried for this rerun: the dialog opens as
  // the rerun is requested, so the server may not have started the output yet.
  const [attempt, setAttempt] = useState(0);
  const retryTimer = useRef<number | null>(null);

  useEffect(() => {
    if (!open) return;
    setStatus('running');
    statusRef.current = 'running';
    setCommand('');
    setCommands([]);
    setError(undefined);
    setAttempt(0);
    autoScrollRef.current = true;
  }, [open, streamUrl]);

  useEffect(() => {
    if (!open) return;
    const es = new EventSource(streamUrl);
    es.addEventListener('open', () => setLines([]));
    es.addEventListener('message', (e: Event) => {
      const snap = JSON.parse((e as MessageEvent<string>).data) as StreamSnapshot;
      if (snap.command) setCommand(snap.command);
      if (snap.commands?.length) setCommands(snap.commands);
      statusRef.current = snap.status;
      setStatus(snap.status);
      if (snap.lines?.length) setLines(prev => [...prev, ...snap.lines]);
    });
    es.addEventListener('done', () => es.close());
    es.onerror = () => {
      if (statusRef.current !== 'running') {
        es.close();
        return;
      }
      // A closed source will not reconnect on its own: the server refused the
      // stream because it has no rerun output yet, so ask again shortly.
      if (es.readyState !== EventSource.CLOSED) return;
      if (attempt >= MAX_CONNECT_ATTEMPTS) {
        setError('The rerun output is not available');
        return;
      }
      retryTimer.current = window.setTimeout(() => setAttempt(n => n + 1), RETRY_MS);
    };

    return () => {
      es.close();
      if (retryTimer.current !== null) window.clearTimeout(retryTimer.current);
    };
  }, [open, streamUrl, attempt]);

  useEffect(() => {
    if (autoScrollRef.current && scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [lines]);

  return (
    <Modal
      open={open}
      onClose={onClose}
      size="xl"
      closeOnBackdrop
      scrollBody={false}
      title={<span className="truncate font-mono text-sm" title={command || 'rerun'}>{command || 'rerun'}</span>}
      headerSlot={<span data-rerun-status={status}><Badge tone={STATUS_TONE[status]} variant="soft" size="xs">{status}</Badge></span>}
    >
      {commands.length > 0 && <RerunCommands commands={commands} />}
      <pre
        ref={scrollRef}
        onScroll={handleScroll}
        className="h-[70vh] overflow-auto bg-zinc-950 p-4 font-mono text-sm leading-relaxed text-zinc-100"
        data-rerun-output=""
      >
        {lines.length === 0 && status === 'running' && !error && <span className="text-zinc-500">Waiting for output...</span>}
        {lines.length === 0 && status !== 'running' && !error && <span className="text-zinc-500">The rerun wrote no output.</span>}
        {error && <span className="text-red-400">{error}</span>}
        <AnsiHtml as="span" text={lines.map(lineText).join('')} />
      </pre>
    </Modal>
  );
}
