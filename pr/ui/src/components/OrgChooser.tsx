import { useEffect, useRef, useState } from 'react';
import type { MouseEvent as ReactMouseEvent } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import type { Org, SearchConfig } from '../types';
import { GavelIcon } from './GavelIcon';

interface Props {
  config: SearchConfig;
  onChange: (partial: Partial<SearchConfig>) => void;
}

// OrgChooser shows the currently-selected org (or "@me" when no org is
// selected) and lets the user switch between org-wide browsing and
// personal-only browsing. Selecting an org writes `{ org, all: true }`;
// selecting @me clears to `{ org: '', all: false }`, scoping the fetch to the
// local repo. Author/bot narrowing is handled client-side by the Authors
// filter, not here.
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

  function hideOrg(login: string, e: ReactMouseEvent) {
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

  function unhideOrg(login: string, e: ReactMouseEvent) {
    e.stopPropagation();
    onChange({ ignoredOrgs: ignoredOrgs.filter(o => o !== login) });
  }

  return (
    <div className="relative" ref={rootRef}>
      <Button
        variant="ghost"
        className="inline-flex items-center justify-start gap-1.5 text-xs h-auto px-2 py-1 rounded border border-border text-muted-foreground hover:bg-muted transition-colors"
        onClick={() => setOpen(!open)}
        title="Switch GitHub org / scope"
      >
        {activeOrgMeta?.avatarUrl ? (
          <img src={activeOrgMeta.avatarUrl} alt={activeOrg} className="w-4 h-4 rounded-sm" />
        ) : (
          <GavelIcon name="codicon:organization" className="text-sm" />
        )}
        <span className="font-medium">{label}</span>
        <GavelIcon name="codicon:chevron-down" className="text-[10px]" />
      </Button>

      {open && (
        <div className="absolute top-full right-0 mt-1 w-72 bg-popover rounded-lg shadow-lg border border-border z-50 py-1 text-sm">
          <Button
            variant="ghost"
            className={`w-full flex items-center justify-start gap-2 h-auto px-3 py-1.5 text-left transition-colors ${
              !config.all ? 'bg-primary/10 text-primary' : 'hover:bg-muted text-foreground'
            }`}
            onClick={chooseMe}
          >
            <GavelIcon name="codicon:person" className="text-base" />
            <span className="flex-1">@me (my PRs)</span>
            {!config.all && <GavelIcon name="codicon:check" className="text-xs" />}
          </Button>

          <Button
            variant="ghost"
            className={`w-full flex items-center justify-start gap-2 h-auto px-3 py-1.5 text-left transition-colors ${
              config.all && !activeOrg ? 'bg-primary/10 text-primary' : 'hover:bg-muted text-foreground'
            }`}
            onClick={chooseAllOrgs}
          >
            <GavelIcon name="codicon:globe" className="text-base" />
            <span className="flex-1">All orgs (default)</span>
            {config.all && !activeOrg && <GavelIcon name="codicon:check" className="text-xs" />}
          </Button>

          <div className="border-t border-border my-1" />

          {loading && <div className="px-3 py-2 text-xs text-muted-foreground">Loading orgs…</div>}
          {err && <div className="px-3 py-2 text-xs text-red-500">{err}</div>}
          {!loading && !err && visibleOrgs.length === 0 && hiddenOrgs.length === 0 && (
            <div className="px-3 py-2 text-xs text-muted-foreground">No orgs — token has no org memberships</div>
          )}
          {visibleOrgs.map(o => {
            const selected = config.all && config.org === o.login;
            return (
              <div
                key={o.login}
                className={`group flex items-center gap-2 px-3 py-1.5 transition-colors ${
                  selected ? 'bg-primary/10 text-primary' : 'hover:bg-muted text-foreground'
                }`}
              >
                <Button
                  variant="ghost"
                  className="flex-1 flex items-center justify-start gap-2 h-auto p-0 text-left"
                  onClick={() => chooseOrg(o.login)}
                >
                  {o.avatarUrl
                    ? <img src={o.avatarUrl} alt={o.login} className="w-4 h-4 rounded-sm shrink-0" />
                    : <GavelIcon name="codicon:organization" className="text-base" />}
                  <span className="flex-1 truncate">{o.login}</span>
                  {selected && <GavelIcon name="codicon:check" className="text-xs" />}
                </Button>
                <Button
                  variant="ghost"
                  className="h-auto p-0 opacity-0 group-hover:opacity-100 text-muted-foreground hover:text-red-500 transition-opacity"
                  title={`Hide ${o.login} from this list`}
                  onClick={(e) => hideOrg(o.login, e)}
                >
                  <GavelIcon name="codicon:eye-closed" className="text-xs" />
                </Button>
              </div>
            );
          })}

          {hiddenOrgs.length > 0 && (
            <>
              <div className="border-t border-border my-1" />
              <Button
                variant="ghost"
                className="w-full flex items-center justify-start gap-2 h-auto px-3 py-1.5 text-xs text-muted-foreground hover:bg-muted"
                onClick={() => setShowHidden(v => !v)}
              >
                <GavelIcon name={showHidden ? 'codicon:chevron-down' : 'codicon:chevron-right'} className="text-[10px]" />
                <span className="flex-1 text-left">Manage hidden ({hiddenOrgs.length})</span>
              </Button>
              {showHidden && hiddenOrgs.map(o => (
                <div
                  key={o.login}
                  className="group flex items-center gap-2 px-3 py-1.5 text-xs text-muted-foreground hover:bg-muted"
                >
                  {o.avatarUrl
                    ? <img src={o.avatarUrl} alt={o.login} className="w-4 h-4 rounded-sm shrink-0 opacity-60" />
                    : <GavelIcon name="codicon:organization" className="text-base" />}
                  <span className="flex-1 truncate">{o.login}</span>
                  <Button
                    variant="ghost"
                    className="h-auto p-0 text-muted-foreground hover:text-primary transition-colors"
                    title={`Unhide ${o.login}`}
                    onClick={(e) => unhideOrg(o.login, e)}
                  >
                    <GavelIcon name="codicon:eye" className="text-xs" />
                  </Button>
                </div>
              ))}
            </>
          )}
        </div>
      )}
    </div>
  );
}
