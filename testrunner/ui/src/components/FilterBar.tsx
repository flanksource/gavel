import { formatCount, frameworkIcon } from '../utils';
import type { FilterMode, FilterState } from '../filterState';
import { cycleFilterState } from '../filterState';

export interface Filters {
  status: FilterState<string>;
  framework: FilterState<string>;
}

interface Props {
  filters: Filters;
  onChange: (f: Filters) => void;
  counts: { passed: number; failed: number; skipped: number; pending: number; timedout: number };
  frameworks: string[];
}

const STATUS_DEFS: { key: string; label: string; badge: string; activeBg: string; activeBorder: string }[] = [
  { key: 'failed', label: 'Failed', badge: 'bg-red-500', activeBg: 'bg-red-50', activeBorder: 'border-red-300' },
  { key: 'timedout', label: 'Timed out', badge: 'bg-amber-500', activeBg: 'bg-amber-50', activeBorder: 'border-amber-300' },
  { key: 'passed', label: 'Passed', badge: 'bg-green-500', activeBg: 'bg-green-50', activeBorder: 'border-green-300' },
  { key: 'skipped', label: 'Skipped', badge: 'bg-yellow-400', activeBg: 'bg-yellow-50', activeBorder: 'border-yellow-300' },
  { key: 'pending', label: 'Pending', badge: 'bg-blue-400', activeBg: 'bg-blue-50', activeBorder: 'border-blue-300' },
];

export function FilterBar({ filters, onChange, counts, frameworks }: Props) {
  const hasActiveFilters = filters.status.size > 0 || filters.framework.size > 0;

  return (
    <div class="flex items-center gap-1.5 flex-wrap">
      {STATUS_DEFS.map(sf => {
        const count = (counts as any)[sf.key] as number;
        if (count === 0) return null;
        const mode = filters.status.get(sf.key);
        return (
          <button
            key={sf.key}
            class={`inline-flex items-center gap-1.5 text-xs px-2 py-0.5 rounded-full border transition-colors ${triStateClasses(mode, sf.activeBg, sf.activeBorder)}`}
            onClick={() => onChange({ ...filters, status: cycleFilterState(filters.status, sf.key) })}
            title={triStateTitle(sf.label, mode)}
          >
            <span
              class={`inline-flex items-center justify-center min-w-[16px] h-[16px] px-1 rounded-full text-[10px] font-bold text-white ${sf.badge}`}
              title={String(count)}
            >
              {formatCount(count)}
            </span>
            <StateMarker mode={mode} />
            {sf.label}
          </button>
        );
      })}

      {frameworks.length > 1 && (
        <>
          <span class="text-gray-300 mx-0.5">|</span>
          {frameworks.map(fw => {
            const icon = frameworkIcon(fw);
            const mode = filters.framework.get(fw);
            return (
              <button
                key={fw}
                class={`inline-flex items-center gap-1 text-xs px-2 py-0.5 rounded-full border transition-colors ${triStateClasses(mode, 'bg-gray-100', 'border-gray-400')}`}
                onClick={() => onChange({ ...filters, framework: cycleFilterState(filters.framework, fw) })}
                title={triStateTitle(fw, mode)}
              >
                <StateMarker mode={mode} />
                {icon && <iconify-icon icon={icon} class="text-sm" />}
                {fw}
              </button>
            );
          })}
        </>
      )}

      {hasActiveFilters && (
        <button
          class="text-xs text-gray-400 hover:text-gray-600 ml-1"
          onClick={() => onChange({ status: new Map(), framework: new Map() })}
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
