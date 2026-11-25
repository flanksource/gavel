import { useCallback, useMemo } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { GitCommitList, type GitCommitItem } from '@flanksource/clicky-ui/data';
import type { PRCommitInfo } from '../types';
import { prCommitDiffQuery } from './oneShotQueries';
import { RelativeTime } from './RelativeTime';

// PRCommitList lists a PR's commits with each commit's diff loaded on expand —
// the PR's Commits tab and the commits a merge lands.
export function PRCommitList({ commits, repo, className, diffMaxHeightClassName }: {
  commits: PRCommitInfo[];
  repo: string;
  className?: string;
  diffMaxHeightClassName?: string;
}) {
  const queryClient = useQueryClient();
  const items = useMemo<GitCommitItem[]>(() => commits.map(commit => ({
    id: commit.oid,
    sha: commit.oid,
    shortSha: commit.oid.slice(0, 7),
    title: commit.messageHeadline,
    body: commit.messageBody,
    authorName: commit.authorName,
    authorLogin: commit.authorLogin,
    authorAvatarUrl: commit.authorAvatarUrl,
    committedAt: commit.committedDate,
    additions: commit.additions,
    deletions: commit.deletions,
    changedFiles: commit.changedFiles,
    href: `https://github.com/${repo}/commit/${commit.oid}`,
  })), [commits, repo]);

  const loadDiff = useCallback(
    (commit: GitCommitItem) => queryClient.fetchQuery(prCommitDiffQuery(repo, commit.sha)),
    [repo, queryClient],
  );

  return (
    <GitCommitList
      commits={items}
      loadDiff={loadDiff}
      renderTime={iso => <RelativeTime iso={iso} />}
      className={className}
      diffMaxHeightClassName={diffMaxHeightClassName}
    />
  );
}
