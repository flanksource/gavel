import type { LinterResult, Severity } from '../types';
import { countLintBySeverity, countLintByLinter, collectLintLinters, formatCount, lintToolIcon } from '../utils';
import type { FilterMode, FilterState } from '../filterState';
import { cycleFilterState } from '../filterState';

export type LintGrouping = 'linter-file' | 'file-linter-rule';

export interface LintFilters {
  severity: FilterState<Severity>;
  linter: FilterState<string>;
}

interface Props {
  lint: LinterResult[] | undefined;
  filters: LintFilters;
  onFiltersChange: (f: LintFilters) => void;
}

const SEVERITY_DEFS: { key: Severity; label: string; badge: string; activeBg: string; activeBorder: string; icon: string }[] = [
  { key: 'error', label: 'Error', badge: 'bg-red-500', activeBg: 'bg-red-50', activeBorder: 'border-red-300', icon: 'codicon:error' },
  { key: 'warning', label: 'Warning', badge: 'bg-yellow-400', activeBg: 'bg-yellow-50', activeBorder: 'border-yellow-300', icon: 'codicon:warning' },
  { key: 'info', label: 'Info', badge: 'bg-blue-400', activeBg: 'bg-blue-50', activeBorder: 'border-blue-300', icon: 'codicon:info' },
];

export function LintFilterBar({ lint, filters, onFiltersChange }: Props) {
  const severityCounts = countLintBySeverity(lint, filters.linter);
  const linterCounts = countLintByLinter(lint, filters.severity);
  const linters = collectLintLinters(lint);
  const hasActive = filters.severity.size > 0 || filters.linter.size > 0;

  return (
    <div class="flex items-center gap-1.5 flex-wrap">
      <span class="inline-flex items-center gap-1 rounded-full border border-gray-200 bg-gray-50 px-2 py-0.5 text-xs font-medium text-gray-600">
        <iconify-icon icon="codicon:list-tree" />
        Linter → Folder → File
      </span>

      <span class="text-gray-300 mx-0.5">|</span>

      {SEVERITY_DEFS.map(sd => {
        const count = severityCounts[sd.key];
        if (count === 0 && filters.severity.size === 0) return null;
        const mode = filters.severity.get(sd.key);
        return (
          <button
            key={sd.key}
            class={`inline-flex items-center gap-1.5 text-xs px-2 py-0.5 rounded-full border transition-colors ${triStateClasses(mode, sd.activeBg, sd.activeBorder)}`}
            onClick={() => onFiltersChange({ ...filters, severity: cycleFilterState(filters.severity, sd.key) })}
            title={triStateTitle(sd.label, mode)}
          >
            <span
              class={`inline-flex items-center justify-center min-w-[16px] h-[16px] px-1 rounded-full text-[10px] font-bold text-white ${sd.badge}`}
              title={String(count)}
            >
              {formatCount(count)}
            </span>
            <StateMarker mode={mode} />
            <iconify-icon icon={sd.icon} class="text-sm" />
            {sd.label}
          </button>
        );
      })}

      {linters.length > 1 && (
        <>
          <span class="text-gray-300 mx-0.5">|</span>
          {linters.map(linter => {
            const count = linterCounts[linter] || 0;
            const mode = filters.linter.get(linter);
            return (
              <button
                key={linter}
                class={`inline-flex items-center gap-1 text-xs px-2 py-0.5 rounded-full border transition-colors ${triStateClasses(mode, 'bg-gray-100', 'border-gray-400')}`}
                onClick={() => onFiltersChange({ ...filters, linter: cycleFilterState(filters.linter, linter) })}
                title={triStateTitle(linter, mode)}
              >
                <StateMarker mode={mode} />
                <span class="text-[10px] text-gray-500" title={String(count)}>{formatCount(count)}</span>
                <iconify-icon icon={lintToolIcon(linter)} class="text-sm" />
                {linter}
              </button>
            );
          })}
        </>
      )}

      {hasActive && (
        <button
          class="text-xs text-gray-400 hover:text-gray-600 ml-1"
          onClick={() => onFiltersChange({ severity: new Map(), linter: new Map() })}
        >
          Clear
        </button>
      )}
    </div>
  );
}

function triStateClasses(mode: FilterMode | undefined, includeBg: string, includeBorder: string): string {
  if (mode === 'include') {
    return `${includeBg} ${includeBorder} font-medium text-gray-900`;
  }
  if (mode === 'exclude') {
    return 'bg-gray-900 text-white border-gray-900 font-medium';
  }
  return 'border-gray-200 text-gray-500 hover:bg-gray-50';
}

function triStateTitle(label: string, mode: FilterMode | undefined): string {
  if (mode === 'include') return `${label}: included. Click to exclude`;
  if (mode === 'exclude') return `${label}: excluded. Click to clear`;
  return `${label}: neutral. Click to include`;
}

function StateMarker({ mode }: { mode: FilterMode | undefined }) {
  if (mode === 'include') return <iconify-icon icon="codicon:add" class="text-xs" />;
  if (mode === 'exclude') return <iconify-icon icon="codicon:remove" class="text-xs" />;
  return <span class="w-2 h-2 rounded-full bg-current opacity-30" />;
}
