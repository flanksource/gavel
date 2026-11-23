// The hosted top bar's repository picker: every repository the organization reaches, grouped by
// owner, where a member hides or shows an owner (its future repositories too) or one repository.
import { useEffect, useRef, useState } from 'react';
import { Button, SearchInput } from '@flanksource/clicky-ui/components';
import { UiChevronDown, UiEye, UiEyeClosed, UiFolder, UiOrganization } from '@flanksource/clicky-ui/icons';
import type { Project } from '../types';

interface Props {
  /** The organization's repositories as projects, each with its owner and whether it is hidden. */
  projects: Project[];
  onSetRepoHidden: (dir: string, hidden: boolean) => void;
  onSetOwnerHidden: (owner: string, hidden: boolean) => void;
}

/** The group of pushed repositories, which have no owner to hide. */
const PUSHED = 'Pushed repositories';

export function RepoVisibilityPicker({ projects, onSetRepoHidden, onSetOwnerHidden }: Props) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');
  const rootRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    function onDown(e: MouseEvent) {
      if (rootRef.current && !rootRef.current.contains(e.target as Node)) setOpen(false);
    }
    document.addEventListener('mousedown', onDown);
    return () => document.removeEventListener('mousedown', onDown);
  }, [open]);

  const shown = projects.filter(project => !project.hidden).length;
  const label = shown === projects.length ? 'All repositories' : `${shown} of ${projects.length} repositories`;
  const needle = query.trim().toLowerCase();
  const groups = new Map<string, Project[]>();
  for (const project of projects) {
    if (needle && !project.name.toLowerCase().includes(needle)) continue;
    const owner = project.owner || PUSHED;
    groups.set(owner, [...(groups.get(owner) ?? []), project]);
  }
  const owners = [...groups.keys()].sort((a, b) => (a === PUSHED ? 1 : b === PUSHED ? -1 : a.localeCompare(b)));

  return (
    <div className="relative min-w-0" ref={rootRef}>
      <Button variant="outline" size="sm" className="flex min-w-0 items-center gap-1.5" aria-expanded={open} onClick={() => setOpen(v => !v)}>
        <UiFolder className="shrink-0 text-sm" />
        <span className="truncate font-medium">{label}</span>
        <UiChevronDown className="shrink-0 text-[10px]" />
      </Button>
      {open && (
        <div className="absolute top-full right-0 z-50 mt-1 max-h-[min(32rem,calc(100vh-4rem))] w-80 overflow-y-auto rounded-lg border border-border bg-popover py-1 text-sm shadow-lg">
          <div className="px-2 pb-1">
            <SearchInput value={query} onChange={setQuery} placeholder="Find a repository" shortcut={null} />
          </div>
          {owners.length === 0 && <p className="px-3 py-2 text-xs text-muted-foreground">No repository matches.</p>}
          {owners.map(owner => {
            const repos = groups.get(owner) ?? [];
            const ownerHidden = owner !== PUSHED && repos.every(repo => repo.ownerHidden);
            return (
              <section key={owner} className="py-1">
                <div className="flex items-center gap-2 px-3 py-1 text-xs font-semibold text-muted-foreground">
                  <UiOrganization className="shrink-0" />
                  <span className={`min-w-0 flex-1 truncate ${ownerHidden ? 'line-through' : ''}`}>{owner}</span>
                  {owner !== PUSHED && (
                    <VisibilityToggle name={owner} hidden={ownerHidden} onToggle={() => onSetOwnerHidden(owner, !ownerHidden)} />
                  )}
                </div>
                {repos.map(repo => (
                  <div key={repo.dir} className={`flex items-center gap-2 py-1 pl-8 pr-3 ${repo.hidden ? 'text-muted-foreground' : 'text-foreground'}`}>
                    <span className={`min-w-0 flex-1 truncate ${repo.hidden ? 'line-through' : ''}`}>{repo.name}</span>
                    <VisibilityToggle
                      name={repo.name}
                      hidden={!!repo.hidden}
                      onToggle={() => onSetRepoHidden(repo.dir, !repo.hidden)}
                      {...(repo.ownerHidden && repo.owner ? { disabledReason: `Hidden with ${repo.owner}` } : {})}
                    />
                  </div>
                ))}
              </section>
            );
          })}
        </div>
      )}
    </div>
  );
}

/** An eye that shows name when it is hidden and hides it when it is shown. */
function VisibilityToggle({ name, hidden, onToggle, disabledReason }: { name: string; hidden: boolean; onToggle: () => void; disabledReason?: string }) {
  const Icon = hidden ? UiEyeClosed : UiEye;
  return (
    <Button
      variant="ghost"
      className="h-auto shrink-0 p-0.5 text-muted-foreground hover:text-foreground disabled:opacity-50"
      aria-label={`${hidden ? 'Show' : 'Hide'} ${name}`}
      title={disabledReason ?? `${hidden ? 'Show' : 'Hide'} ${name}`}
      disabled={!!disabledReason}
      onClick={onToggle}
    >
      <Icon className="text-xs" />
    </Button>
  );
}
