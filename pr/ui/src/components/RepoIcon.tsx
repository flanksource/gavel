import { useState } from 'react';
import { paletteClass } from '../utils';

export interface RepoIconProps {
  repo: string;
  homepageUrl?: string;
  size: number;
}

// RepoIcon renders a repo's identity badge: its site favicon when a homepage is
// known, falling back to a deterministic letter avatar keyed on the repo name.
// Shared by the PR list's per-repo headers and the Todos sidebar's per-workspace
// headers so both surfaces show the same repo header.
export function RepoIcon({ repo, homepageUrl, size }: RepoIconProps) {
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
