import type { ComponentChildren } from 'preact';
import type { ProcessNode } from '../types';
import { formatBytes, processLabel } from '../utils';

interface Props {
  process: ProcessNode | null;
  collectBusy?: boolean;
  onCollectStack?: (pid: number) => void | Promise<void>;
}

export function DiagnosticsDetailPanel({ process, collectBusy, onCollectStack }: Props) {
  if (!process) {
    return (
      <div class="flex items-center justify-center h-full text-gray-400 text-sm">
        <div class="text-center">
          <iconify-icon icon="codicon:server-process" class="text-4xl mb-2 block" />
          Select a process to view diagnostics
        </div>
      </div>
    );
  }

  const stack = process.stack_capture;

  return (
    <div class="p-5 space-y-4">
      <div class="flex items-start justify-between gap-3">
        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-2">
            <iconify-icon icon={process.is_root ? 'codicon:server-process' : 'codicon:debug-alt'} class="text-2xl text-blue-600 shrink-0" />
            <h2 class="text-lg font-bold text-gray-900 break-words">{processLabel(process)}</h2>
          </div>
          <div class="mt-1 flex items-center gap-2 flex-wrap text-xs text-gray-500">
            <span class="font-mono">pid {process.pid}</span>
            {process.ppid ? <span class="font-mono">ppid {process.ppid}</span> : null}
            {process.status ? <span>{process.status}</span> : null}
          </div>
        </div>
        {onCollectStack && (
          <button
            class="shrink-0 text-xs px-2 py-1 rounded bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-1"
            onClick={() => onCollectStack(process.pid)}
            disabled={collectBusy}
            title="Collect the latest stack trace"
          >
            <iconify-icon icon={collectBusy ? 'svg-spinners:ring-resize' : 'codicon:debug-alt-small'} />
            {collectBusy ? 'Collecting...' : 'Collect stack trace'}
          </button>
        )}
      </div>

      {process.command && (
        <Section title="Command">
          <pre class="text-sm text-gray-700 whitespace-pre-wrap font-mono bg-blue-50 rounded p-3 break-all">
            {process.command}
          </pre>
        </Section>
      )}

      <Section title="Metrics">
        <div class="grid grid-cols-2 gap-3 text-sm">
          <Metric label="CPU" value={`${(process.cpu_percent || 0).toFixed(1)}%`} />
          <Metric label="RSS" value={formatBytes(process.rss)} />
          <Metric label="Virtual memory" value={formatBytes(process.vms)} />
          <Metric label="Open files" value={process.open_files !== undefined ? String(process.open_files) : 'Unavailable'} />
        </div>
      </Section>

      <Section title="Stack">
        {!stack && (
          <div class="text-sm text-gray-500">No stack trace collected yet.</div>
        )}
        {stack && (
          <div class="space-y-2">
            <div class="flex items-center gap-2 text-xs text-gray-500">
              <span class={`px-2 py-0.5 rounded-full ${
                stack.status === 'ready'
                  ? 'bg-green-100 text-green-700'
                  : stack.status === 'unsupported'
                    ? 'bg-yellow-100 text-yellow-700'
                    : 'bg-red-100 text-red-700'
              }`}>
                {stack.status}
              </span>
              {stack.collected_at && <span>{new Date(stack.collected_at).toLocaleString()}</span>}
            </div>
            {stack.error && (
              <div class="text-sm text-red-700 bg-red-50 rounded p-3 whitespace-pre-wrap">
                {stack.error}
              </div>
            )}
            {stack.text && (
              <pre class="text-xs text-gray-700 whitespace-pre-wrap font-mono bg-gray-50 rounded p-3 max-h-[28rem] overflow-auto">
                {stack.text}
              </pre>
            )}
          </div>
        )}
      </Section>
    </div>
  );
}

function Section({ title, children }: { title: string; children: ComponentChildren }) {
  return (
    <section>
      <div class="text-xs font-semibold uppercase tracking-wide text-gray-500 mb-2">{title}</div>
      {children}
    </section>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div class="border rounded-lg bg-gray-50 px-3 py-2">
      <div class="text-xs uppercase tracking-wide text-gray-500">{label}</div>
      <div class="text-sm font-medium text-gray-800 mt-1">{value}</div>
    </div>
  );
}
