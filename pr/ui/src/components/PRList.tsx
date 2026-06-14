import { useState } from 'react';
import type { PRItem, PRSyncStatus, GavelResultsSummary, Project, ProcStatus } from '../types';
import { PRRow } from './PRRow';
import { ProcControl } from './ProcControl';
import { groupByOrg, prKey, paletteClass } from '../utils';
import { Avatar } from './Avatar';

interface Props {
  prs: PRItem[];
  selected: PRItem | null;
  onSelect: (pr: PRItem) => void;
  unread?: Record<string, boolean>;
  syncStatus?: Record<string, PRSyncStatus>;
  gavelResults?: Record<string, GavelResultsSummary>;
  projectsByRepo?: Record<string, Project>;
  procStatus?: Record<string, ProcStatus>;
  onProcChanged?: () => void;
  onProcEdit?: (project: Project) => void;
}

interface RepoIconProps {
  repo: string;
  homepageUrl?: string;
  size: number;
}

function RepoIcon({ repo, homepageUrl, size }: RepoIconProps) {
  const [faviconFailed, setFaviconFailed] = useState(false);
  const showFavicon = !!homepageUrl && !faviconFailed;

  if (showFavicon) {
    const src = `/api/repos/favicon?homepage=${encodeURIComponent(homepageUrl!)}`;
    return (
      <img
        src={src}
        alt={repo}
        title={repo}
        width={size}
        height={size}
        className="inline-block shrink-0 rounded bg-white"
        loading="lazy"
        onError={() => setFaviconFailed(true)}
      />
    );
  }

  const short = repo.split('/').pop() || repo;
  return (
    <span
      className={`inline-flex items-center justify-center shrink-0 rounded font-semibold ${paletteClass(repo)}`}
      style={{ width: size, height: size, fontSize: Math.max(9, Math.floor(size * 0.5)) }}
      title={repo}
    >
      {(short.charAt(0) || '?').toUpperCase()}
    </span>
  );
}

export function PRList({ prs, selected, onSelect, unread, syncStatus, gavelResults, projectsByRepo, procStatus, onProcChanged, onProcEdit }: Props) {
  if (prs.length === 0) {
    return (
      <div className="p-6 text-center text-gray-400">
        <iconify-icon icon="codicon:git-pull-request" className="text-3xl mb-2" />
        <p>No pull requests found</p>
      </div>
    );
  }

  const orgs = groupByOrg(prs);

  return (
    <div className="divide-y divide-gray-100">
      {orgs.map(og => (
        <div key={og.org || '_'}>
          {og.org && (
            <div className="px-3 py-1.5 bg-gray-100 sticky top-0 border-b border-gray-200 flex items-center gap-2 z-20">
              <Avatar
                src={og.orgAvatarUrl}
                alt={og.org}
                size={20}
                rounded="md"
                href={`https://github.com/${og.org}`}
                title={og.org}
                colorKey={og.org}
              />
              <span className="text-sm font-semibold text-gray-800 truncate flex-1">{og.org}</span>
              <span className="text-xs text-gray-500 shrink-0">{og.itemCount}</span>
            </div>
          )}
          {og.repos.map(group => (
            <div key={group.repo}>
              <div className="pl-6 pr-3 py-1.5 bg-gray-50 sticky top-9 border-b border-gray-200 flex items-center gap-2 z-10">
                <a
                  href={`https://github.com/${group.repo}`}
                  target="_blank"
                  rel="noopener"
                  className="inline-flex shrink-0"
                  onClick={(e) => e.stopPropagation()}
                >
                  <RepoIcon repo={group.repo} homepageUrl={group.repoHomepageUrl} size={20} />
                </a>
                <span className="text-sm font-medium text-gray-700 truncate flex-1">{group.repoShort}</span>
                {onProcChanged && (
                  <ProcControl
                    repo={group.repo}
                    project={projectsByRepo?.[group.repo]}
                    status={procStatus?.[group.repo]}
                    onChanged={onProcChanged}
                    onEdit={onProcEdit}
                  />
                )}
                <span className="text-xs text-gray-400 font-normal shrink-0">{group.items.length}</span>
              </div>
              {group.items.map(pr => (
                <PRRow
                  key={prKey(pr)}
                  pr={pr}
                  selected={selected?.repo === pr.repo && selected?.number === pr.number}
                  unread={!!unread?.[prKey(pr)]}
                  syncStatus={syncStatus?.[prKey(pr)]}
                  gavelResults={gavelResults?.[prKey(pr)]}
                  onClick={() => onSelect(pr)}
                />
              ))}
            </div>
          ))}
        </div>
      ))}
    </div>
  );
}
