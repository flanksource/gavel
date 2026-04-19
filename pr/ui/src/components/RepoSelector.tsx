import { useState, useEffect, useRef } from 'preact/hooks';

interface RepoInfo {
  repo: string;
  prCount: number;
  selected: boolean;
}

interface Props {
  repos: string[];
  allOrg?: boolean;
  org?: string;
  onChange: (repos: string[]) => void;
}

export function RepoSelector({ repos, allOrg, org, onChange }: Props) {
  const [open, setOpen] = useState(false);
  const [available, setAvailable] = useState<RepoInfo[]>([]);
  const [input, setInput] = useState('');
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (open) {
      fetch('/api/repos')
        .then(r => r.json())
        .then((data: RepoInfo[]) => {
          const sorted = (data || []).sort((a, b) => b.prCount - a.prCount);
          setAvailable(sorted);
        })
        .catch(() => {});
    }
  }, [open]);

  function toggleRepo(repo: string) {
    if (repos.includes(repo)) {
      onChange(repos.filter(r => r !== repo));
    } else {
      onChange([...repos, repo]);
    }
  }

  function addCustom() {
    const val = input.trim();
    if (val && !repos.includes(val)) {
      onChange([...repos, val]);
    }
    setInput('');
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === 'Enter') { e.preventDefault(); addCustom(); }
    if (e.key === 'Escape') { setOpen(false); }
  }

  // The org itself lives in the OrgChooser on the right; this control is
  // just the repo-subset picker. When no specific repos are selected we
  // show "All repos" (org-wide browsing) rather than duplicating the org
  // name — the OrgChooser already owns that label.
  const label = repos.length > 0
    ? `${repos.length} repos`
    : allOrg
      ? 'All repos'
      : 'Current repo';

  return (
    <div class="relative">
      <button
        class="inline-flex items-center gap-1 text-xs px-2 py-1 rounded border border-gray-200 text-gray-600 hover:bg-gray-50 transition-colors"
        onClick={() => { setOpen(!open); setTimeout(() => inputRef.current?.focus(), 50); }}
      >
        <iconify-icon icon="codicon:repo" class="text-sm" />
        {label}
        <iconify-icon icon="codicon:chevron-down" class="text-[10px]" />
      </button>

      {open && (
        <div class="absolute top-full left-0 mt-1 w-80 bg-white rounded-lg shadow-lg border border-gray-200 z-50 p-2">
          {allOrg && repos.length === 0 && (
            <div class="text-xs text-blue-600 bg-blue-50 rounded px-2 py-1 mb-2">
              <iconify-icon icon="codicon:organization" class="mr-1" />
              Showing all repos in <span class="font-medium">{org || 'the selected org'}</span>. Use the chooser (top right) to switch orgs; pick specific repos below to narrow this view.
            </div>
          )}

          {available.length > 0 && (
            <div class="max-h-56 overflow-y-auto mb-2 space-y-0.5">
              {available.map(r => {
                const checked = repos.includes(r.repo);
                const short = r.repo.includes('/') ? r.repo.split('/')[1] : r.repo;
                return (
                  <label key={r.repo}
                    class={`flex items-center gap-2 text-xs px-2 py-1.5 rounded cursor-pointer transition-colors ${
                      checked ? 'bg-blue-50' : 'hover:bg-gray-50'
                    }`}
                  >
                    <input type="checkbox" checked={checked}
                      onChange={() => toggleRepo(r.repo)}
                      class="rounded border-gray-300 text-blue-600 focus:ring-blue-500 h-3.5 w-3.5"
                    />
                    <span class={`font-mono flex-1 truncate ${checked ? 'text-blue-700 font-medium' : 'text-gray-700'}`}>
                      {short}
                    </span>
                    <span class="text-gray-400 tabular-nums">{r.prCount} PRs</span>
                  </label>
                );
              })}
            </div>
          )}

          <div class="flex gap-1 border-t border-gray-100 pt-2">
            <input
              ref={inputRef}
              type="text"
              value={input}
              onInput={(e) => setInput((e.target as HTMLInputElement).value)}
              onKeyDown={handleKeyDown}
              placeholder="owner/repo"
              class="flex-1 text-xs px-2 py-1 border border-gray-200 rounded focus:outline-none focus:border-blue-400"
            />
            <button onClick={addCustom}
              class="text-xs px-2 py-1 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors">
              Add
            </button>
          </div>

          <div class="mt-2 pt-1 flex items-center justify-between">
            {repos.length > 0 && (
              <button class="text-xs text-gray-400 hover:text-gray-600"
                onClick={() => onChange([])}>
                {allOrg ? 'Back to org-wide' : 'Clear all'}
              </button>
            )}
            <button onClick={() => setOpen(false)}
              class="text-xs text-gray-500 hover:text-gray-700 ml-auto">
              Close
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
