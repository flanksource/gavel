import { useEffect, useRef, useState } from 'preact/hooks';
import type { Org, SearchConfig } from '../types';

interface Props {
  config: SearchConfig;
  onChange: (partial: Partial<SearchConfig>) => void;
}

// OrgChooser shows the currently-selected org (or "@me" when no org is
// selected) and lets the user switch between org-wide browsing and
// personal-only browsing. Selecting an org writes `{ org, all: true }`;
// selecting @me clears to `{ org: '', all: false }` so SearchControls'
// @me/all/bots buttons stay in control of that scope.
//
// Orgs are fetched lazily on first open via /api/orgs?include-ignored=1 —
// the server caches the underlying list for 5 minutes, so the dropdown
// stays snappy. Requesting the un-filtered list here lets the chooser
// render an inline "Manage hidden" section so users can unhide without
// juggling a second endpoint.

export function OrgChooser({ config, onChange }: Props) {
  const [open, setOpen] = useState(false);
  const [orgs, setOrgs] = useState<Org[]>([]);
  const [loading, setLoading] = useState(false);
  const [err, setErr] = useState<string | null>(null);
  const [showHidden, setShowHidden] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    setLoading(true);
    // include-ignored=1 so the chooser sees the full membership list —
    // it needs that to render the "Manage hidden" section.
    fetch('/api/orgs?include-ignored=1')
      .then(r => r.ok ? r.json() : Promise.reject(new Error(`HTTP ${r.status}`)))
      .then((data: Org[]) => { setOrgs(data || []); setErr(null); })
      .catch(e => setErr(e?.message || 'fetch failed'))
      .finally(() => setLoading(false));
  }, [open]);

  // Close on outside click — the dropdown is anchored to the button so a
  // stray click should dismiss rather than trap the user.
  useEffect(() => {
    if (!open) return;
    const onDown = (e: MouseEvent) => {
      if (rootRef.current && !rootRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener('mousedown', onDown);
    return () => document.removeEventListener('mousedown', onDown);
  }, [open]);

  const ignoredOrgs = config.ignoredOrgs || [];
  const ignoredSet = new Set(ignoredOrgs);
  const visibleOrgs = orgs.filter(o => !ignoredSet.has(o.login));
  const hiddenOrgs = orgs.filter(o => ignoredSet.has(o.login));

  const activeOrg = config.all ? (config.org || '') : '';
  // Three display modes: a specific org (label = its login), org-wide with
  // no explicit org (label = "All orgs" — the daemon picks the default via
  // ResolveDefaultOrg), or personal scope (label = "@me").
  let label: string;
  if (!config.all) {
    label = '@me';
  } else if (activeOrg) {
    label = activeOrg;
  } else {
    label = 'All orgs';
  }
  const activeOrgMeta = activeOrg ? orgs.find(o => o.login === activeOrg) : undefined;

  function chooseOrg(login: string) {
    // Selecting an org implies org-wide mode. Clear the repo list because
    // old repo filters are almost always from a different org.
    onChange({ org: login, all: true, repos: [] });
    setOpen(false);
  }

  function chooseMe() {
    onChange({ org: '', all: false, repos: [] });
    setOpen(false);
  }

  // Org-wide browsing without pinning a specific org — daemon's
  // ResolveDefaultOrg picks one (skipping ignored orgs) each search.
  function chooseAllOrgs() {
    onChange({ org: '', all: true, repos: [] });
    setOpen(false);
  }

  function hideOrg(login: string, e: MouseEvent) {
    e.stopPropagation(); // don't also select it
    const next = Array.from(new Set([...ignoredOrgs, login]));
    const patch: Partial<SearchConfig> = { ignoredOrgs: next };
    // If the user just hid the currently-selected org, drop back to @me
    // so the next poll doesn't keep fetching PRs from a hidden org.
    if (config.org === login) {
      patch.org = '';
      patch.all = false;
      patch.repos = [];
    }
    onChange(patch);
  }

  function unhideOrg(login: string, e: MouseEvent) {
    e.stopPropagation();
    onChange({ ignoredOrgs: ignoredOrgs.filter(o => o !== login) });
  }

  return (
    <div class="relative" ref={rootRef}>
      <button
        class="inline-flex items-center gap-1.5 text-xs px-2 py-1 rounded border border-gray-200 text-gray-600 hover:bg-gray-50 transition-colors"
        onClick={() => setOpen(!open)}
        title="Switch GitHub org / scope"
      >
        {activeOrgMeta?.avatarUrl ? (
          <img src={activeOrgMeta.avatarUrl} alt={activeOrg} class="w-4 h-4 rounded-sm" />
        ) : (
          <iconify-icon icon="codicon:organization" class="text-sm" />
        )}
        <span class="font-medium">{label}</span>
        <iconify-icon icon="codicon:chevron-down" class="text-[10px]" />
      </button>

      {open && (
        <div class="absolute top-full right-0 mt-1 w-72 bg-white rounded-lg shadow-lg border border-gray-200 z-50 py-1 text-sm">
          <button
            class={`w-full flex items-center gap-2 px-3 py-1.5 text-left transition-colors ${
              !config.all ? 'bg-blue-50 text-blue-700' : 'hover:bg-gray-50 text-gray-700'
            }`}
            onClick={chooseMe}
          >
            <iconify-icon icon="codicon:person" class="text-base" />
            <span class="flex-1">@me (my PRs)</span>
            {!config.all && <iconify-icon icon="codicon:check" class="text-xs" />}
          </button>

          <button
            class={`w-full flex items-center gap-2 px-3 py-1.5 text-left transition-colors ${
              config.all && !activeOrg ? 'bg-blue-50 text-blue-700' : 'hover:bg-gray-50 text-gray-700'
            }`}
            onClick={chooseAllOrgs}
          >
            <iconify-icon icon="codicon:globe" class="text-base" />
            <span class="flex-1">All orgs (default)</span>
            {config.all && !activeOrg && <iconify-icon icon="codicon:check" class="text-xs" />}
          </button>

          <div class="border-t border-gray-100 my-1" />

          {loading && <div class="px-3 py-2 text-xs text-gray-400">Loading orgs…</div>}
          {err && <div class="px-3 py-2 text-xs text-red-500">{err}</div>}
          {!loading && !err && visibleOrgs.length === 0 && hiddenOrgs.length === 0 && (
            <div class="px-3 py-2 text-xs text-gray-400">No orgs — token has no org memberships</div>
          )}
          {visibleOrgs.map(o => {
            const selected = config.all && config.org === o.login;
            return (
              <div
                key={o.login}
                class={`group flex items-center gap-2 px-3 py-1.5 transition-colors ${
                  selected ? 'bg-blue-50 text-blue-700' : 'hover:bg-gray-50 text-gray-700'
                }`}
              >
                <button
                  class="flex-1 flex items-center gap-2 text-left"
                  onClick={() => chooseOrg(o.login)}
                >
                  {o.avatarUrl
                    ? <img src={o.avatarUrl} alt={o.login} class="w-4 h-4 rounded-sm shrink-0" />
                    : <iconify-icon icon="codicon:organization" class="text-base" />}
                  <span class="flex-1 truncate">{o.login}</span>
                  {selected && <iconify-icon icon="codicon:check" class="text-xs" />}
                </button>
                <button
                  class="opacity-0 group-hover:opacity-100 text-gray-400 hover:text-red-500 transition-opacity"
                  title={`Hide ${o.login} from this list`}
                  onClick={(e) => hideOrg(o.login, e)}
                >
                  <iconify-icon icon="codicon:eye-closed" class="text-xs" />
                </button>
              </div>
            );
          })}

          {hiddenOrgs.length > 0 && (
            <>
              <div class="border-t border-gray-100 my-1" />
              <button
                class="w-full flex items-center gap-2 px-3 py-1.5 text-xs text-gray-500 hover:bg-gray-50"
                onClick={() => setShowHidden(v => !v)}
              >
                <iconify-icon icon={showHidden ? 'codicon:chevron-down' : 'codicon:chevron-right'} class="text-[10px]" />
                <span class="flex-1 text-left">Manage hidden ({hiddenOrgs.length})</span>
              </button>
              {showHidden && hiddenOrgs.map(o => (
                <div
                  key={o.login}
                  class="group flex items-center gap-2 px-3 py-1.5 text-xs text-gray-400 hover:bg-gray-50"
                >
                  {o.avatarUrl
                    ? <img src={o.avatarUrl} alt={o.login} class="w-4 h-4 rounded-sm shrink-0 opacity-60" />
                    : <iconify-icon icon="codicon:organization" class="text-base" />}
                  <span class="flex-1 truncate">{o.login}</span>
                  <button
                    class="text-gray-400 hover:text-blue-600 transition-colors"
                    title={`Unhide ${o.login}`}
                    onClick={(e) => unhideOrg(o.login, e)}
                  >
                    <iconify-icon icon="codicon:eye" class="text-xs" />
                  </button>
                </div>
              ))}
            </>
          )}
        </div>
      )}
    </div>
  );
}
