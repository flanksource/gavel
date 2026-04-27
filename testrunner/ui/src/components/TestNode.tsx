import { useState, useEffect, useRef } from 'preact/hooks';
import type { Test } from '../types';
import { statusIcon, statusColor, formatCount, formatDuration, sum, hasFailed, frameworkIcon, totalDuration, humanizeName, isLintNode, lintNodeCount, lintBadgeColor } from '../utils';

interface Props {
  test: Test;
  depth: number;
  expandAll: boolean | null;
  selected: Test | null;
  onSelect: (t: Test) => void;
  onRerun?: (t: Test) => void;
  onStop?: (t: Test) => void;
  rerunBusy?: boolean;
  stopBusy?: boolean;
}

function resolveFramework(t: Test): string | undefined {
  if (t.framework) return t.framework;
  for (const c of t.children || []) {
    const f = resolveFramework(c);
    if (f) return f;
  }
  return undefined;
}

export function TestNode({ test: t, depth, expandAll, selected, onSelect, onRerun, onStop, rerunBusy, stopBusy }: Props) {
  const hasChildren = (t.children?.length ?? 0) > 0;
  const failed = hasFailed(t);
  const defaultOpen = failed || depth < 1;
  const [open, setOpen] = useState(defaultOpen);
  const prevExpandAll = useRef(expandAll);
  const isSelected = selected === t;

  useEffect(() => {
    if (expandAll !== null && expandAll !== prevExpandAll.current) {
      setOpen(expandAll);
    }
    prevExpandAll.current = expandAll;
  }, [expandAll]);

  const s = hasChildren ? sum(t) : null;
  const fw = resolveFramework(t);
  const fwIcon = frameworkIcon(fw);
  const isLint = isLintNode(t);
  const lintCount = isLint ? lintNodeCount(t) : 0;
  const isStoppedTask = t.framework === 'task' && typeof (t.context as any)?.status === 'string' && (t.context as any).status.toLowerCase() === 'canceled';
  const canStop = !!onStop && !!t.can_stop && !!t.task_id;

  const rowBg = isSelected
    ? 'bg-blue-50 border-l-2 border-blue-500'
    : isLint ? 'hover:bg-gray-50' : t.failed ? 'bg-red-50/50' : 'hover:bg-gray-50';

  return (
    <div>
      <div
        class={`flex items-center gap-1.5 py-1 px-2 cursor-pointer text-sm ${rowBg}`}
        style={{ paddingLeft: `${depth * 16 + 8}px` }}
        onClick={(e) => {
          e.stopPropagation();
          onSelect(t);
          if (hasChildren) setOpen(!open);
        }}
      >
        {hasChildren ? (
          <iconify-icon
            icon={open ? 'codicon:chevron-down' : 'codicon:chevron-right'}
            class="text-gray-400 text-xs shrink-0 w-3"
          />
        ) : (
          <span class="w-3 shrink-0" />
        )}

        <iconify-icon icon={statusIcon(t)} class={`${isLint ? 'text-gray-500' : statusColor(t)} text-base shrink-0`} />

        {fwIcon && !isLint && (
          <iconify-icon icon={fwIcon} class="text-sm shrink-0 opacity-60" />
        )}

        <span class={`truncate ${isLint ? 'text-gray-800' : isStoppedTask ? 'text-orange-700' : t.pending ? 'text-blue-600' : t.failed ? 'text-red-700' : t.skipped ? 'text-yellow-700' : 'text-gray-800'} ${isSelected ? 'font-semibold' : 'font-medium'}`}>
          {humanizeName(t.name, fw)}
        </span>

        <span class="flex-1" />

        {t.timed_out && (
          <span class="text-xs text-amber-700 bg-amber-50 border border-amber-200 rounded px-1.5 py-0.5 shrink-0 flex items-center gap-1">
            <iconify-icon icon="ion:hourglass-outline" class="text-sm" />
            timed out
          </span>
        )}

        {(() => {
          const dur = t.duration || (hasChildren ? totalDuration(t) : 0);
          return dur > 0 ? <span class="text-xs text-gray-400 shrink-0">{formatDuration(dur)}</span> : null;
        })()}

        {isLint && lintCount > 0 && (
          <span class="flex items-center gap-1 shrink-0">
            <Badge count={lintCount} color={lintBadgeColor(t)} />
          </span>
        )}

        {!isLint && s && s.total > 0 && (
          <span class="flex items-center gap-1 shrink-0">
            {s.passed > 0 && <Badge count={s.passed} color="bg-green-500" />}
            {s.failed > 0 && <Badge count={s.failed} color="bg-red-500" />}
            {s.skipped > 0 && <Badge count={s.skipped} color="bg-yellow-400" />}
            {s.pending > 0 && <Badge count={s.pending} color="bg-blue-400" />}
          </span>
        )}

        {canStop && (
          <button
            class="text-gray-400 hover:text-orange-600 disabled:opacity-30 disabled:hover:text-gray-400 shrink-0 px-1"
            onClick={(e) => { e.stopPropagation(); onStop!(t); }}
            disabled={stopBusy}
            title="Stop"
          >
            <iconify-icon icon="codicon:debug-stop" class="text-sm" />
          </button>
        )}

        {onRerun && t.framework !== 'task' && t.kind !== 'violation' && (
          <button
            class="text-gray-400 hover:text-blue-600 disabled:opacity-30 disabled:hover:text-gray-400 shrink-0 px-1"
            onClick={(e) => { e.stopPropagation(); onRerun(t); }}
            disabled={rerunBusy}
            title="Rerun"
          >
            <iconify-icon icon="codicon:refresh" class="text-sm" />
          </button>
        )}
      </div>

      {open && hasChildren && t.children!.map((child, i) => (
        <TestNode key={i} test={child} depth={depth + 1} expandAll={expandAll} selected={selected} onSelect={onSelect} onRerun={onRerun} onStop={onStop} rerunBusy={rerunBusy} stopBusy={stopBusy} />
      ))}
    </div>
  );
}

function Badge({ count, color }: { count: number; color: string }) {
  return (
    <span
      class={`inline-flex items-center justify-center min-w-[18px] h-[18px] px-1 rounded-full text-[10px] font-bold text-white ${color}`}
      title={String(count)}
    >
      {formatCount(count)}
    </span>
  );
}
