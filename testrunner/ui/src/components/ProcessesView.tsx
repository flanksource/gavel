import { useEffect, useMemo, useRef, useState } from 'react';
import type { ProcessForestNode, ProcessTraceNode } from '../types';
import {
  buildProcessForest,
  formatBytes,
  formatRunDuration,
  isProcessExited,
  processLabel,
  processStateColor,
  processStateIcon,
  pruneExitedProcesses,
} from '../utils';

interface Props {
  nodes: ProcessTraceNode[];
  selectedKey?: string | null;
  expandAll: boolean | null;
  hideExited: boolean;
  error?: string;
  onHideExitedChange: (hide: boolean) => void;
  onSelect: (key: string) => void;
}

export function ProcessesView({ nodes, selectedKey, expandAll, hideExited, error, onHideExitedChange, onSelect }: Props) {
  const forest = useMemo(() => {
    const full = buildProcessForest(nodes);
    return hideExited ? pruneExitedProcesses(full) : full;
  }, [nodes, hideExited]);

  return (
    <div>
      <div className="flex items-center justify-between gap-2 border-b bg-white px-3 py-1.5 text-xs text-gray-600">
        <span>{nodes.length} {nodes.length === 1 ? 'process' : 'processes'}</span>
        <label className="inline-flex items-center gap-1.5 cursor-pointer">
          <input
            type="checkbox"
            checked={hideExited}
            onChange={(e) => onHideExitedChange((e.target as HTMLInputElement).checked)}
          />
          Hide exited
        </label>
      </div>
      {error && <div className="px-3 py-2 text-xs text-red-700 whitespace-pre-wrap">{error}</div>}
      {nodes.length === 0 && !error && (
        <div className="p-8 text-center text-gray-400">
          <iconify-icon icon="svg-spinners:ring-resize" className="text-3xl text-blue-500" />
          <p className="mt-2">Waiting for process trace...</p>
        </div>
      )}
      {nodes.length > 0 && forest.length === 0 && (
        <div className="p-8 text-center text-gray-400 text-sm">Every process has exited</div>
      )}
      {forest.map(entry => (
        <ProcessTreeNode
          key={entry.node.key}
          entry={entry}
          depth={0}
          selectedKey={selectedKey}
          expandAll={expandAll}
          onSelect={onSelect}
        />
      ))}
    </div>
  );
}

interface TreeNodeProps {
  entry: ProcessForestNode;
  depth: number;
  selectedKey?: string | null;
  expandAll: boolean | null;
  onSelect: (key: string) => void;
}

function ProcessTreeNode({ entry, depth, selectedKey, expandAll, onSelect }: TreeNodeProps) {
  const { node, children } = entry;
  const hasChildren = children.length > 0;
  const [open, setOpen] = useState(true);
  const prevExpandAll = useRef(expandAll);
  const selected = node.key === selectedKey;
  const exited = isProcessExited(node);

  useEffect(() => {
    if (expandAll !== null && expandAll !== prevExpandAll.current) {
      setOpen(expandAll);
    }
    prevExpandAll.current = expandAll;
  }, [expandAll]);

  const rowBg = selected ? 'bg-blue-50 border-l-2 border-blue-500' : 'hover:bg-gray-50';

  return (
    <div>
      <div
        className={`flex items-center gap-1.5 py-1 px-2 cursor-pointer text-sm ${rowBg} ${exited ? 'opacity-60' : ''}`}
        style={{ paddingLeft: `${depth * 16 + 8}px` }}
        onClick={(e) => {
          e.stopPropagation();
          onSelect(node.key);
          if (hasChildren) setOpen(!open);
        }}
      >
        {hasChildren ? (
          <iconify-icon
            icon={open ? 'codicon:chevron-down' : 'codicon:chevron-right'}
            className="text-gray-400 text-xs shrink-0 w-3"
          />
        ) : (
          <span className="w-3 shrink-0" />
        )}
        <iconify-icon
          icon={node.isRoot ? 'codicon:server-process' : 'codicon:debug-alt'}
          className={`text-base shrink-0 ${node.isRoot ? 'text-blue-600' : 'text-gray-400'}`}
        />
        <span className={`shrink-0 ${selected ? 'font-semibold text-blue-900' : 'font-medium text-gray-800'}`}>
          {processLabel(node)}
        </span>
        <span className="text-xs text-gray-400 shrink-0">pid {node.pid}</span>
        <span className="min-w-0 flex-1 truncate text-xs font-mono text-gray-500" title={node.command}>
          {node.command}
        </span>
        <span className={`inline-flex items-center gap-1 text-xs shrink-0 ${processStateColor(node.status)}`}>
          <iconify-icon icon={processStateIcon(node.status)} />
          {node.status}
        </span>
        {exited && (
          <span className="text-xs text-gray-400 shrink-0" title="Lifetime">
            {formatRunDuration(node.startedAt, node.exitedAt || node.lastSeenAt)}
          </span>
        )}
        <span className="text-xs text-gray-400 shrink-0" title={exited ? 'Peak CPU' : 'CPU'}>
          {(exited ? node.peakCpuPercent : node.cpuPercent).toFixed(1)}%
        </span>
        <span className="text-xs text-gray-400 shrink-0" title={exited ? 'Peak RSS' : 'RSS'}>
          {formatBytes(exited ? node.peakRssBytes : node.rssBytes)}
        </span>
      </div>

      {open && hasChildren && children.map(child => (
        <ProcessTreeNode
          key={child.node.key}
          entry={child}
          depth={depth + 1}
          selectedKey={selectedKey}
          expandAll={expandAll}
          onSelect={onSelect}
        />
      ))}
    </div>
  );
}
