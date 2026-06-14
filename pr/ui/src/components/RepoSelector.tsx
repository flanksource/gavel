import { useState, useEffect } from 'react';
import { Combobox } from '@flanksource/clicky-ui/components';
import type { ComboboxOption } from '@flanksource/clicky-ui/components';

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

function shortName(repo: string): string {
  return repo.includes('/') ? repo.split('/')[1] : repo;
}

// RepoSelector is a multi-select repo picker backed by clicky-ui's Combobox.
// Options come from /api/repos (with PR counts); freeform `owner/repo` entries
// are committed via Combobox's allowCustomValue (the default).
export function RepoSelector({ repos, allOrg, onChange }: Props) {
  const [available, setAvailable] = useState<RepoInfo[]>([]);

  useEffect(() => {
    fetch('/api/repos')
      .then(r => r.json())
      .then((data: RepoInfo[]) => setAvailable((data || []).sort((a, b) => b.prCount - a.prCount)))
      .catch(() => {});
  }, []);

  const options: ComboboxOption[] = available.map(r => ({
    value: r.repo,
    label: `${shortName(r.repo)} · ${r.prCount} PRs`,
  }));
  // Keep already-selected repos selectable even if they aren't in /api/repos.
  for (const r of repos) {
    if (!options.some(o => o.value === r)) options.push({ value: r, label: shortName(r) });
  }

  return (
    <Combobox
      multiple
      options={options}
      value={repos}
      onChange={(next) => onChange(next as string[])}
      placeholder={allOrg ? 'All repos' : 'repos'}
      ariaLabel="Filter repositories"
    />
  );
}
