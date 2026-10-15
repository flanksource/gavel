import { useEffect, useMemo, useState } from 'react';
import { countGoroutinesByState, parseGoroutineDump, type ParsedGoroutine, type ParsedGoroutineFrame } from '../stacktrace';

// GoroutineStackView renders a Go goroutine dump with a text search, per-state
// filter chips and a hide-runtime-only toggle.
export function GoroutineStackView({ text }: { text: string }) {
  const [search, setSearch] = useState('');
  const [selectedStates, setSelectedStates] = useState<Set<string>>(new Set());
  const [hideRuntimeOnly, setHideRuntimeOnly] = useState(true);
  const parsed = useMemo(() => parseGoroutineDump(text), [text]);
  const stateCounts = useMemo(() => countGoroutinesByState(parsed), [parsed]);
  const filtered = useMemo(() => {
    const needle = search.trim().toLowerCase();
    return parsed.filter(goroutine => {
      if (selectedStates.size > 0 && !selectedStates.has(goroutine.state)) return false;
      if (hideRuntimeOnly && goroutine.userFrameCount === 0) return false;
      if (needle && !goroutine.searchText.includes(needle)) return false;
      return true;
    });
  }, [parsed, search, selectedStates, hideRuntimeOnly]);

  useEffect(() => {
    setSearch('');
    setSelectedStates(new Set());
    setHideRuntimeOnly(true);
  }, [text]);

  if (parsed.length === 0) {
    return <pre className="min-h-0 overflow-auto py-1 text-[11px] text-gray-700 whitespace-pre-wrap font-mono leading-4">{text}</pre>;
  }

  return (
    <div className="min-h-0 flex flex-col">
      <div className="px-1 py-1.5 space-y-2">
        <div className="text-[11px] text-gray-500">{filtered.length} / {parsed.length} goroutines</div>
        <div className="flex items-center gap-1.5 flex-wrap">
          <div className="relative min-w-[14rem] flex-1">
            <iconify-icon icon="codicon:search" className="absolute left-2 top-1/2 -translate-y-1/2 text-gray-400 text-xs" />
            <input
              className="w-full rounded-md border border-gray-200 bg-gray-50 py-1 pl-7 pr-2 text-xs text-gray-700 outline-none focus:border-blue-400 focus:bg-white"
              placeholder="Filter by goroutine id, function, or file"
              value={search}
              onInput={(e) => setSearch((e.target as HTMLInputElement).value)}
            />
          </div>
          <label className="inline-flex items-center gap-1.5 rounded-md border border-gray-200 bg-gray-50 px-2 py-1 text-[11px] text-gray-600">
            <input
              type="checkbox"
              checked={hideRuntimeOnly}
              onChange={(e) => setHideRuntimeOnly((e.target as HTMLInputElement).checked)}
            />
            Hide runtime-only
          </label>
          {(search || selectedStates.size > 0 || !hideRuntimeOnly) && (
            <button
              className="text-[11px] text-gray-500 hover:text-gray-700"
              onClick={() => {
                setSearch('');
                setSelectedStates(new Set());
                setHideRuntimeOnly(true);
              }}
            >
              Clear
            </button>
          )}
        </div>
        <div className="flex items-center gap-1 flex-wrap">
          {Array.from(stateCounts.entries()).sort((a, b) => b[1] - a[1]).map(([state, count]) => {
            const active = selectedStates.has(state);
            return (
              <button
                key={state}
                className={`inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-[11px] transition-colors ${
                  active
                    ? 'border-blue-400 bg-blue-50 text-blue-700'
                    : 'border-gray-200 bg-gray-50 text-gray-600 hover:bg-white'
                }`}
                onClick={() => {
                  setSelectedStates(prev => {
                    const next = new Set(prev);
                    if (next.has(state)) next.delete(state);
                    else next.add(state);
                    return next;
                  });
                }}
              >
                <span className={`h-2 w-2 rounded-full ${goroutineStateDot(state)}`} />
                {state}
                <span className="text-[10px] opacity-70">{count}</span>
              </button>
            );
          })}
        </div>
      </div>
      <div className="flex-1 min-h-0 overflow-auto py-1 space-y-1">
        {filtered.length === 0 && (
          <div className="py-3 text-center text-xs text-gray-500">
            No goroutines match the current filters.
          </div>
        )}
        {filtered.map(goroutine => (
          <GoroutineCard
            key={goroutine.id}
            goroutine={goroutine}
            search={search}
            hideRuntimeOnly={hideRuntimeOnly}
          />
        ))}
      </div>
    </div>
  );
}

function GoroutineCard({ goroutine, search, hideRuntimeOnly }: {
  goroutine: ParsedGoroutine;
  search: string;
  hideRuntimeOnly: boolean;
}) {
  const frames = hideRuntimeOnly
    ? goroutine.frames.filter(frame => !frame.runtime || frame.kind === 'created_by')
    : goroutine.frames;
  const defaultOpen = goroutine.state === 'running' || !!search;

  return (
    <details className="border-0 bg-transparent" open={defaultOpen}>
      <summary className="cursor-pointer list-none px-0 py-1">
        <div className="flex items-center gap-2 flex-wrap">
          <span className="font-mono text-xs font-semibold text-gray-900">g{goroutine.id}</span>
          <span className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] ${goroutineStateBadge(goroutine.state)}`}>
            <span className={`h-2 w-2 rounded-full ${goroutineStateDot(goroutine.state)}`} />
            {goroutine.rawState}
          </span>
          <span className="text-[11px] text-gray-500">{frames.length}f</span>
          {goroutine.userFrameCount > 0 && <span className="text-[11px] text-gray-500">{goroutine.userFrameCount}u</span>}
          {goroutine.topFunction && <span className="truncate text-[11px] text-gray-600">{goroutine.topFunction}</span>}
        </div>
      </summary>
      <div className="pl-3 py-1 space-y-0.5">
        {frames.map((frame, index) => (
          <FrameRow key={`${goroutine.id}-${index}`} frame={frame} />
        ))}
      </div>
    </details>
  );
}

function FrameRow({ frame }: { frame: ParsedGoroutineFrame }) {
  return (
    <div className={`${frame.runtime ? 'text-gray-500' : 'text-gray-800'}`}>
      <div className="flex items-start gap-1.5">
        <iconify-icon icon={frame.kind === 'created_by' ? 'codicon:debug-restart' : frame.runtime ? 'codicon:debug-step-over' : 'codicon:symbol-method'} className="shrink-0 mt-0.5 text-[11px]" />
        <div className="min-w-0">
          <div className="break-all font-mono text-[11px] font-semibold leading-4">
            {frame.displayName}
            {frame.location && <span className="ml-2 text-[10px] font-normal opacity-80">{frame.location}</span>}
          </div>
        </div>
      </div>
    </div>
  );
}

function goroutineStateBadge(state: string): string {
  if (state.includes('running')) return 'bg-green-50 text-green-700';
  if (state.includes('chan') || state.includes('wait')) return 'bg-blue-50 text-blue-700';
  if (state.includes('sleep')) return 'bg-amber-50 text-amber-700';
  if (state.includes('select')) return 'bg-violet-50 text-violet-700';
  return 'bg-gray-100 text-gray-700';
}

function goroutineStateDot(state: string): string {
  if (state.includes('running')) return 'bg-green-500';
  if (state.includes('chan') || state.includes('wait')) return 'bg-blue-500';
  if (state.includes('sleep')) return 'bg-amber-500';
  if (state.includes('select')) return 'bg-violet-500';
  return 'bg-gray-400';
}
