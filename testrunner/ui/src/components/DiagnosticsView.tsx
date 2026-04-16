import { useEffect, useRef, useState } from 'preact/hooks';
import type { ProcessNode } from '../types';
import { countProcesses, formatBytes, processLabel } from '../utils';

interface Props {
  root?: ProcessNode;
  selectedPid?: number | null;
  expandAll: boolean | null;
  onSelect: (pid: number) => void;
}

export function DiagnosticsView({ root, selectedPid, expandAll, onSelect }: Props) {
  if (!root) {
    return (
      <div class="p-8 text-center text-gray-400">
        <iconify-icon icon="svg-spinners:ring-resize" class="text-3xl text-blue-500" />
        <p class="mt-2">Waiting for process diagnostics...</p>
      </div>
    );
  }

  return (
    <div>
      <DiagnosticsTreeNode
        node={root}
        depth={0}
        selectedPid={selectedPid}
        expandAll={expandAll}
        onSelect={onSelect}
      />
      {countProcesses(root) === 0 && (
        <div class="p-8 text-center text-gray-400 text-sm">No processes available</div>
      )}
    </div>
  );
}

interface TreeNodeProps {
  node: ProcessNode;
  depth: number;
  selectedPid?: number | null;
  expandAll: boolean | null;
  onSelect: (pid: number) => void;
}

function DiagnosticsTreeNode({ node, depth, selectedPid, expandAll, onSelect }: TreeNodeProps) {
  const hasChildren = (node.children?.length || 0) > 0;
  const [open, setOpen] = useState(node.is_root || depth < 1);
  const prevExpandAll = useRef(expandAll);
  const selected = node.pid === selectedPid;

  useEffect(() => {
    if (expandAll !== null && expandAll !== prevExpandAll.current) {
      setOpen(expandAll);
    }
    prevExpandAll.current = expandAll;
  }, [expandAll]);

  const cpu = node.cpu_percent || 0;
  const rowBg = selected
    ? 'bg-blue-50 border-l-2 border-blue-500'
    : 'hover:bg-gray-50';

  return (
    <div>
      <div
        class={`flex items-center gap-1.5 py-1 px-2 cursor-pointer text-sm ${rowBg}`}
        style={{ paddingLeft: `${depth * 16 + 8}px` }}
        onClick={(e) => {
          e.stopPropagation();
          onSelect(node.pid);
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
        <iconify-icon
          icon={node.is_root ? 'codicon:server-process' : 'codicon:debug-alt'}
          class={`text-base shrink-0 ${node.is_root ? 'text-blue-600' : 'text-gray-400'}`}
        />
        <span class={`truncate ${selected ? 'font-semibold text-blue-900' : 'font-medium text-gray-800'}`}>
          {processLabel(node)}
        </span>
        <span class="text-xs text-gray-400 shrink-0">pid {node.pid}</span>
        <span class="flex-1" />
        {node.status && <span class="text-xs text-gray-400 shrink-0">{node.status}</span>}
        <span class="text-xs text-gray-400 shrink-0">{cpu.toFixed(1)}%</span>
        <span class="text-xs text-gray-400 shrink-0">{formatBytes(node.rss)}</span>
      </div>

      {open && hasChildren && node.children!.map(child => (
        <DiagnosticsTreeNode
          key={child.pid}
          node={child}
          depth={depth + 1}
          selectedPid={selectedPid}
          expandAll={expandAll}
          onSelect={onSelect}
        />
      ))}
    </div>
  );
}
