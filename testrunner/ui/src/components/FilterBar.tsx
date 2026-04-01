import { frameworkIcon } from '../utils';

export interface Filters {
  status: Set<string>;
  framework: Set<string>;
}

interface Props {
  filters: Filters;
  onChange: (f: Filters) => void;
  counts: { passed: number; failed: number; skipped: number; pending: number };
  frameworks: string[];
}

function toggleSet(set: Set<string>, key: string): Set<string> {
  const next = new Set(set);
  if (next.has(key)) next.delete(key);
  else next.add(key);
  return next;
}

const STATUS_DEFS: { key: string; label: string; badge: string; activeBg: string; activeBorder: string }[] = [
  { key: 'failed', label: 'Failed', badge: 'bg-red-500', activeBg: 'bg-red-50', activeBorder: 'border-red-300' },
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
        const active = filters.status.has(sf.key);
        return (
          <button
            key={sf.key}
            class={`inline-flex items-center gap-1.5 text-xs px-2 py-0.5 rounded-full border transition-colors ${
              active
                ? `${sf.activeBg} ${sf.activeBorder} font-medium`
                : 'border-gray-200 text-gray-500 hover:bg-gray-50'
            }`}
            onClick={() => onChange({ ...filters, status: toggleSet(filters.status, sf.key) })}
          >
            <span class={`inline-flex items-center justify-center min-w-[16px] h-[16px] px-1 rounded-full text-[10px] font-bold text-white ${sf.badge}`}>
              {count}
            </span>
            {sf.label}
          </button>
        );
      })}

      {frameworks.length > 1 && (
        <>
          <span class="text-gray-300 mx-0.5">|</span>
          {frameworks.map(fw => {
            const icon = frameworkIcon(fw);
            const active = filters.framework.has(fw);
            return (
              <button
                key={fw}
                class={`inline-flex items-center gap-1 text-xs px-2 py-0.5 rounded-full border transition-colors ${
                  active
                    ? 'bg-gray-100 border-gray-400 font-medium'
                    : 'border-gray-200 text-gray-500 hover:bg-gray-50'
                }`}
                onClick={() => onChange({ ...filters, framework: toggleSet(filters.framework, fw) })}
              >
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
          onClick={() => onChange({ status: new Set(), framework: new Set() })}
        >
          Clear
        </button>
      )}
    </div>
  );
}
