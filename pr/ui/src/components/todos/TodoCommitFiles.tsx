import { useEffect, useState, type ReactNode } from 'react';
import { AnsiHtml } from '@flanksource/clicky-ui/data';
import type { TodoCommitDiffResponse, TodoCommitFile, TodoCommitFilesResponse } from '../../types';
import { GavelIcon } from '../GavelIcon';
import { todoQuery } from './format';

// fileStatusView maps a commit file's change kind to its diff icon, accent
// color, and label, matching how the rest of the dashboard signals add/edit/
// delete state.
function fileStatusView(status: TodoCommitFile['status']) {
  switch (status) {
    case 'added':
      return { icon: 'codicon:diff-added', color: 'text-green-600', label: 'Added' };
    case 'deleted':
      return { icon: 'codicon:diff-removed', color: 'text-red-600', label: 'Deleted' };
    case 'renamed':
      return { icon: 'codicon:diff-renamed', color: 'text-blue-600', label: 'Renamed' };
    default:
      return { icon: 'codicon:diff-modified', color: 'text-amber-600', label: 'Modified' };
  }
}

// splitPath separates a path into its directory prefix (kept muted) and basename
// (kept prominent) so a long path stays scannable when truncated.
function splitPath(path: string): { dir: string; base: string } {
  const i = path.lastIndexOf('/');
  return i < 0 ? { dir: '', base: path } : { dir: path.slice(0, i + 1), base: path.slice(i + 1) };
}

// useCommitFiles fetches the per-file change summary for one commit once its row
// is expanded. Each file carries its repomap scope/language so the rows read as
// a "repomap-based commit status" rather than a flat file list.
function useCommitFiles(dir: string, provider: string, hash: string) {
  const [files, setFiles] = useState<TodoCommitFile[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    let cancelled = false;
    const controller = new AbortController();
    setLoading(true);
    setError('');
    const params = new URLSearchParams(todoQuery(dir, provider));
    params.set('hash', hash);
    fetch(`/api/todos/commits/files?${params.toString()}`, { signal: controller.signal })
      .then(async res => {
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || 'Failed to load files');
        if (!cancelled) setFiles((data as TodoCommitFilesResponse).files ?? []);
      })
      .catch((err: any) => {
        if (!cancelled && err?.name !== 'AbortError') setError(err?.message || 'Failed to load files');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
      controller.abort();
    };
  }, [dir, provider, hash]);

  return { files, loading, error };
}

function Chip({ children, className }: { children: ReactNode; className: string }) {
  return <span className={`inline-flex items-center rounded px-1 text-[10px] font-medium leading-tight ${className}`}>{children}</span>;
}

// FileDiffCard lazily loads and renders a single file's patch (the ANSI-colored
// `git show -- <file>` output). It is the hover/expand payload for a file row,
// shown in the same black terminal panel the full-commit diff used.
function FileDiffCard({ dir, provider, hash, file }: { dir: string; provider: string; hash: string; file: TodoCommitFile }) {
  const [diff, setDiff] = useState('');
  const [truncated, setTruncated] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    let cancelled = false;
    const controller = new AbortController();
    setLoading(true);
    setError('');
    const params = new URLSearchParams(todoQuery(dir, provider));
    params.set('hash', hash);
    params.set('file', file.path);
    fetch(`/api/todos/commits/diff?${params.toString()}`, { signal: controller.signal })
      .then(async res => {
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || 'Failed to load diff');
        if (!cancelled) {
          const payload = data as TodoCommitDiffResponse;
          setDiff(payload.diff ?? '');
          setTruncated(!!payload.truncated);
        }
      })
      .catch((err: any) => {
        if (!cancelled && err?.name !== 'AbortError') setError(err?.message || 'Failed to load diff');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
      controller.abort();
    };
  }, [dir, provider, hash, file.path]);

  return (
    <div className="absolute left-0 right-0 top-full z-30 pt-1">
      <div className="overflow-hidden rounded-md border border-border bg-background shadow-lg">
        <div className="flex items-center gap-2 border-b border-border px-3 py-1.5 text-[11px]">
          <span className="min-w-0 flex-1 truncate font-mono text-foreground" title={file.path}>{file.path}</span>
          {file.adds > 0 && <span className="text-green-600 tabular-nums">+{file.adds}</span>}
          {file.dels > 0 && <span className="text-red-600 tabular-nums">-{file.dels}</span>}
        </div>
        {loading ? (
          <div className="flex items-center gap-2 px-3 py-2 text-[11px] text-muted-foreground">
            <GavelIcon name="svg-spinners:ring-resize" className="text-xs" />
            Loading diff…
          </div>
        ) : error ? (
          <div className="px-3 py-2 text-[11px] text-red-600">{error}</div>
        ) : file.binary ? (
          <div className="px-3 py-2 text-[11px] text-muted-foreground">Binary file — no text diff</div>
        ) : !diff.trim() ? (
          <div className="px-3 py-2 text-[11px] text-muted-foreground">No changes</div>
        ) : (
          <>
            <AnsiHtml
              as="pre"
              text={diff}
              className="max-h-[24rem] overflow-auto whitespace-pre bg-black p-3 text-[11px] leading-snug text-gray-100"
            />
            {truncated && <div className="px-3 py-1 text-[10px] text-muted-foreground">Diff truncated.</div>}
          </>
        )}
      </div>
    </div>
  );
}

// CommitFileRow renders one changed file as a repomap-status row: a change-kind
// icon, the path, scope/language chips, and +/- counts. Hovering the row (or
// clicking to pin it) reveals that file's diff in a card anchored below it.
function CommitFileRow({ dir, provider, hash, file }: { dir: string; provider: string; hash: string; file: TodoCommitFile }) {
  const [hover, setHover] = useState(false);
  const [pinned, setPinned] = useState(false);
  const view = fileStatusView(file.status);
  const { dir: pathDir, base } = splitPath(file.path);
  const open = hover || pinned;

  return (
    <li
      className="relative"
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
    >
      <button
        type="button"
        onClick={() => setPinned(p => !p)}
        aria-expanded={open}
        title={`${view.label} — ${file.path}`}
        className={`flex w-full items-center gap-2 px-3 py-1.5 text-left text-xs hover:bg-muted ${pinned ? 'bg-muted' : ''}`}
      >
        <GavelIcon name={view.icon} className={`shrink-0 text-xs ${view.color}`} />
        <span className="flex min-w-0 flex-1 items-baseline truncate font-mono">
          {pathDir && <span className="truncate text-muted-foreground">{pathDir}</span>}
          <span className="truncate font-medium text-foreground">{base}</span>
        </span>
        {file.language && <Chip className="bg-primary/10 text-primary">{file.language}</Chip>}
        {file.scopes?.map(scope => (
          <Chip key={scope} className="bg-muted text-muted-foreground">{scope}</Chip>
        ))}
        <span className="shrink-0 tabular-nums">
          {file.adds > 0 && <span className="text-green-600">+{file.adds}</span>}
          {file.adds > 0 && file.dels > 0 && <span className="text-muted-foreground"> </span>}
          {file.dels > 0 && <span className="text-red-600">-{file.dels}</span>}
        </span>
      </button>
      {open && <FileDiffCard dir={dir} provider={provider} hash={hash} file={file} />}
    </li>
  );
}

// CommitFiles is the expanded payload of a commit row: the repomap-based status
// of every file the commit touched, each revealing its own diff on hover. It
// replaces the single full-commit diff blob with a per-file breakdown.
export function CommitFiles({ dir, provider, hash }: { dir: string; provider: string; hash: string }) {
  const { files, loading, error } = useCommitFiles(dir, provider, hash);

  if (loading) {
    return (
      <div className="flex items-center gap-2 px-3 py-2 text-[11px] text-muted-foreground">
        <GavelIcon name="svg-spinners:ring-resize" className="text-xs" />
        Loading files…
      </div>
    );
  }
  if (error) {
    return <div className="px-3 py-2 text-[11px] text-red-600">{error}</div>;
  }
  if (files.length === 0) {
    return <div className="px-3 py-2 text-[11px] text-muted-foreground">No file changes</div>;
  }
  return (
    <ul className="border-t border-border bg-muted/30">
      {files.map(file => (
        <CommitFileRow key={file.path} dir={dir} provider={provider} hash={hash} file={file} />
      ))}
    </ul>
  );
}
