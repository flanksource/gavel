import { useState, type ReactNode } from 'react';
import { Tabs } from '@flanksource/clicky-ui/components';
import type { ProcessSample, ProcessTraceNode } from '../types';
import type { ProcessStack } from '../hooks/use-process-trace';
import type { ProcessFacet } from '../hooks/use-process-facet';
import { ProcessFacetView } from './ProcessFacetView';
import type { ProcessStepLink } from '../utils';
import {
  formatBytes,
  formatRunDuration,
  formatRunTimestamp,
  isProcessExited,
  processLabel,
  processStateColor,
  processStateIcon,
} from '../utils';
import { GoroutineStackView } from './GoroutineStackView';

interface Props {
  node: ProcessTraceNode | null;
  samples: ProcessSample[];
  stack?: ProcessStack;
  collecting?: boolean;
  collectMessage?: string;
  stepLink?: ProcessStepLink;
  onCollectStack?: () => void | Promise<void>;
  onOpenStep?: (link: ProcessStepLink) => void;
}

type DetailTab = 'overview' | ProcessFacet;

const detailTabs: Array<{ id: DetailTab; label: string }> = [
  { id: 'overview', label: 'Overview' },
  { id: 'env', label: 'Environment' },
  { id: 'files', label: 'Open files' },
  { id: 'sockets', label: 'Sockets' },
  { id: 'ports', label: 'Ports' },
];

export function ProcessDetailPanel({ node, samples, stack, collecting, collectMessage, stepLink, onCollectStack, onOpenStep }: Props) {
  const [tab, setTab] = useState<DetailTab>('overview');
  if (!node) {
    return (
      <div className="flex items-center justify-center h-full text-gray-400 text-sm">
        <div className="text-center">
          <iconify-icon icon="codicon:server-process" className="text-4xl mb-2 block" />
          Select a process to view its trace
        </div>
      </div>
    );
  }

  const exited = isProcessExited(node);

  return (
    <div className="h-full overflow-y-auto p-4 space-y-3">
      <div className="min-w-0">
        <div className="flex items-center gap-2">
          <iconify-icon icon={node.isRoot ? 'codicon:server-process' : 'codicon:debug-alt'} className="text-2xl text-blue-600 shrink-0" />
          <h2 className="text-lg font-bold text-gray-900 break-words">{processLabel(node)}</h2>
        </div>
        <div className="mt-1 flex items-center gap-2 flex-wrap text-xs text-gray-500">
          <span className="font-mono">pid {node.pid}</span>
          {node.ppid ? <span className="font-mono">ppid {node.ppid}</span> : null}
          <span className={`inline-flex items-center gap-1 rounded-full bg-gray-100 px-2 py-0.5 ${processStateColor(node.status)}`}>
            <iconify-icon icon={processStateIcon(node.status)} />
            {node.status}
          </span>
        </div>
      </div>

      <Tabs tabs={detailTabs} value={tab} onChange={id => setTab(id as DetailTab)} />

      {tab !== 'overview' && <ProcessFacetView facet={tab} pid={node.pid} exited={exited} />}

      {tab === 'overview' && <>
        {stepLink && onOpenStep && (
          <Section title="Step">
            <button
              className="inline-flex items-center gap-1 text-sm text-blue-600 hover:underline"
              onClick={() => onOpenStep(stepLink)}
            >
              <iconify-icon icon="codicon:beaker" />
              {stepLink.name}
            </button>
          </Section>
        )}

        <Section title="Command">
          <pre className="text-xs text-gray-700 whitespace-pre-wrap font-mono bg-blue-50 rounded p-2 break-all">{node.command}</pre>
        </Section>

        <Section title="Process">
          <div className="grid grid-cols-2 gap-2 text-sm">
            <Metric label="Cwd" value={node.cwd || 'Unavailable'} wide />
            <Metric label="Started" value={formatRunTimestamp(node.startedAt)} />
            <Metric label="Exited" value={exited ? formatRunTimestamp(node.exitedAt || node.lastSeenAt) : 'running'} />
            <Metric label="Lifetime" value={formatRunDuration(node.startedAt, node.exitedAt || (exited ? node.lastSeenAt : undefined))} />
          </div>
        </Section>

        <Section title="Resources">
          <div className="grid grid-cols-2 gap-2 text-sm">
            <Metric label="CPU" value={`${node.cpuPercent.toFixed(1)}%`} />
            <Metric label="RSS" value={formatBytes(node.rssBytes)} />
            <Metric label="Peak CPU" value={`${node.peakCpuPercent.toFixed(1)}%`} />
            <Metric label="Peak RSS" value={formatBytes(node.peakRssBytes)} />
          </div>
          <div className="mt-2 grid grid-cols-2 gap-2">
            <Sparkline label="CPU %" values={samples.map(s => s.cpuPercent)} format={v => `${v.toFixed(1)}%`} color="#2563eb" />
            <Sparkline label="RSS" values={samples.map(s => s.rssBytes)} format={formatBytes} color="#7c3aed" />
          </div>
        </Section>

        {node.isRoot && (
          <Section title="Stack">
            <div className="flex items-center justify-between gap-3 mb-2">
              <div className="text-xs text-gray-500">
                {stack?.collectedAt
                  ? `Collected ${formatRunTimestamp(stack.collectedAt)}`
                  : collecting ? 'Collecting stack trace...' : 'No stack trace collected yet.'}
              </div>
              {onCollectStack && (
                <button
                  className="shrink-0 text-[11px] px-2 py-1 rounded bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-1"
                  onClick={() => { void onCollectStack(); }}
                  disabled={collecting}
                  title="Take a fresh observation and dump the goroutine stacks"
                >
                  <iconify-icon icon={collecting ? 'svg-spinners:ring-resize' : 'codicon:debug-alt-small'} />
                  {stack?.text ? 'Refresh stack trace' : 'Collect stack trace'}
                </button>
              )}
            </div>
            {collectMessage && <div className="mb-2 text-xs text-red-700 whitespace-pre-wrap">{collectMessage}</div>}
            {stack?.error && <div className="mb-2 text-xs text-red-700 whitespace-pre-wrap">{stack.error}</div>}
            {stack?.text && <GoroutineStackView text={stack.text} />}
          </Section>
        )}
      </>}
    </div>
  );
}

function Section({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section>
      <div className="text-[11px] font-semibold uppercase tracking-wide text-gray-500 mb-1.5">{title}</div>
      {children}
    </section>
  );
}

function Metric({ label, value, wide }: { label: string; value: string; wide?: boolean }) {
  return (
    <div className={`border rounded-lg bg-gray-50 px-2.5 py-2 ${wide ? 'col-span-2' : ''}`}>
      <div className="text-[10px] uppercase tracking-wide text-gray-500">{label}</div>
      <div className="text-xs font-medium text-gray-800 mt-0.5 break-all">{value}</div>
    </div>
  );
}

const SPARK_WIDTH = 160;
const SPARK_HEIGHT = 36;

function Sparkline({ label, values, format, color }: {
  label: string;
  values: number[];
  format: (value: number) => string;
  color: string;
}) {
  const max = Math.max(...values, 0);
  const points = values.map((value, index) => {
    const x = values.length === 1 ? SPARK_WIDTH : (index / (values.length - 1)) * SPARK_WIDTH;
    const y = max === 0 ? SPARK_HEIGHT - 1 : SPARK_HEIGHT - 1 - (value / max) * (SPARK_HEIGHT - 2);
    return `${x.toFixed(1)},${y.toFixed(1)}`;
  }).join(' ');

  return (
    <div className="border rounded-lg bg-gray-50 px-2.5 py-2">
      <div className="flex items-center justify-between text-[10px] uppercase tracking-wide text-gray-500">
        <span>{label}</span>
        <span>{values.length > 0 ? `max ${format(max)}` : ''}</span>
      </div>
      {values.length < 2 ? (
        <div className="mt-1 text-[11px] text-gray-400">{values.length === 0 ? 'No samples' : `${format(values[0])} (1 sample)`}</div>
      ) : (
        <svg viewBox={`0 0 ${SPARK_WIDTH} ${SPARK_HEIGHT}`} className="mt-1 w-full h-9" preserveAspectRatio="none" role="img" aria-label={`${label} over time`}>
          <polyline points={points} fill="none" stroke={color} strokeWidth="1.5" vectorEffect="non-scaling-stroke" />
        </svg>
      )}
    </div>
  );
}
