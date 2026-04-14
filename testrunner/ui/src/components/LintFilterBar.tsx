import type { LinterResult, Severity } from '../types';
import { countLintBySeverity, countLintByLinter, collectLintLinters } from '../utils';

export type LintGrouping = 'linter-file' | 'file-linter-rule';

export interface LintFilters {
  severity: Set<Severity>;
  linter: Set<string>;
}

interface Props {
  lint: LinterResult[] | undefined;
  grouping: LintGrouping;
  onGroupingChange: (g: LintGrouping) => void;
  filters: LintFilters;
  onFiltersChange: (f: LintFilters) => void;
}

const SEVERITY_DEFS: { key: Severity; label: string; badge: string; activeBg: string; activeBorder: string; icon: string }[] = [
  { key: 'error', label: 'Error', badge: 'bg-red-500', activeBg: 'bg-red-50', activeBorder: 'border-red-300', icon: 'codicon:error' },
  { key: 'warning', label: 'Warning', badge: 'bg-yellow-400', activeBg: 'bg-yellow-50', activeBorder: 'border-yellow-300', icon: 'codicon:warning' },
  { key: 'info', label: 'Info', badge: 'bg-blue-400', activeBg: 'bg-blue-50', activeBorder: 'border-blue-300', icon: 'codicon:info' },
];

function toggle<T>(set: Set<T>, key: T): Set<T> {
  const next = new Set(set);
  if (next.has(key)) next.delete(key);
  else next.add(key);
  return next;
}

export function LintFilterBar({ lint, grouping, onGroupingChange, filters, onFiltersChange }: Props) {
  const severityCounts = countLintBySeverity(lint, filters.linter);
  const linterCounts = countLintByLinter(lint, filters.severity);
  const linters = collectLintLinters(lint);
  const hasActive = filters.severity.size > 0 || filters.linter.size > 0;

  return (
    <div class="flex items-center gap-1.5 flex-wrap">
      <div class="inline-flex rounded-full border border-gray-200 overflow-hidden">
        <button
          class={`text-xs px-2 py-0.5 transition-colors ${
            grouping === 'linter-file' ? 'bg-blue-50 text-blue-700 font-medium' : 'text-gray-500 hover:bg-gray-50'
          }`}
          onClick={() => onGroupingChange('linter-file')}
          title="Group by linter, then file"
        >
          <iconify-icon icon="codicon:list-tree" class="mr-1" />
          Linter → File
        </button>
        <button
          class={`text-xs px-2 py-0.5 border-l border-gray-200 transition-colors ${
            grouping === 'file-linter-rule' ? 'bg-blue-50 text-blue-700 font-medium' : 'text-gray-500 hover:bg-gray-50'
          }`}
          onClick={() => onGroupingChange('file-linter-rule')}
          title="Group by file, then linter, then rule"
        >
          <iconify-icon icon="codicon:files" class="mr-1" />
          File → Linter → Rule
        </button>
      </div>

      <span class="text-gray-300 mx-0.5">|</span>

      {SEVERITY_DEFS.map(sd => {
        const count = severityCounts[sd.key];
        if (count === 0 && filters.severity.size === 0) return null;
        const active = filters.severity.has(sd.key);
        return (
          <button
            key={sd.key}
            class={`inline-flex items-center gap-1.5 text-xs px-2 py-0.5 rounded-full border transition-colors ${
              active
                ? `${sd.activeBg} ${sd.activeBorder} font-medium`
                : 'border-gray-200 text-gray-500 hover:bg-gray-50'
            }`}
            onClick={() => onFiltersChange({ ...filters, severity: toggle(filters.severity, sd.key) })}
            title={active ? `Showing only ${sd.label}` : `Filter to ${sd.label}`}
          >
            <span class={`inline-flex items-center justify-center min-w-[16px] h-[16px] px-1 rounded-full text-[10px] font-bold text-white ${sd.badge}`}>
              {count}
            </span>
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
            const active = filters.linter.has(linter);
            return (
              <button
                key={linter}
                class={`inline-flex items-center gap-1 text-xs px-2 py-0.5 rounded-full border transition-colors ${
                  active
                    ? 'bg-gray-100 border-gray-400 font-medium'
                    : 'border-gray-200 text-gray-500 hover:bg-gray-50'
                }`}
                onClick={() => onFiltersChange({ ...filters, linter: toggle(filters.linter, linter) })}
              >
                <span class="text-[10px] text-gray-500">{count}</span>
                {linter}
              </button>
            );
          })}
        </>
      )}

      {hasActive && (
        <button
          class="text-xs text-gray-400 hover:text-gray-600 ml-1"
          onClick={() => onFiltersChange({ severity: new Set(), linter: new Set() })}
        >
          Clear
        </button>
      )}
    </div>
  );
}
