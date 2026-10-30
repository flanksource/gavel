import { useCallback, useEffect, useRef, useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import type { EventSourceLike } from '@flanksource/clicky-ui/hooks';
import { openEventStream } from './eventHub';
import { queryKeys } from './query';
import type { GavelResultsSummary, PRDetail, PRInfo, PRItem, PRComment, WorkflowRun } from './types';

export interface PRDetailStreamState {
  detail: PRDetail | null;
  loading: boolean;
  refresh: () => void;
}

export function usePRDetailStream(pr: PRItem | null): PRDetailStreamState {
  const queryClient = useQueryClient();
  const streamRef = useRef<EventSourceLike | null>(null);
  const [loading, setLoading] = useState(false);
  const [revision, setRevision] = useState(0);
  const refresh = useCallback(() => setRevision(current => current + 1), []);
  const detailKey = queryKeys.prDetail(pr?.repo ?? '', pr?.number ?? 0);
  const detailQuery = useQuery<PRDetail>({
    queryKey: detailKey,
    queryFn: async () => queryClient.getQueryData<PRDetail>(detailKey) ?? {},
    enabled: false,
    staleTime: Infinity,
  });

  useEffect(() => {
    if (!pr) {
      setLoading(false);
      return;
    }
    const key = queryKeys.prDetail(pr.repo, pr.number);
    const stream = openEventStream(`/api/prs/detail?repo=${encodeURIComponent(pr.repo)}&number=${pr.number}`);
    streamRef.current = stream;
    setLoading(true);

    const onPR = (event: Event) => {
      try {
        const data = frame((event as MessageEvent<string>).data);
        if (!data.pr || typeof data.pr !== 'object') throw new Error('invalid PR frame');
        // A PR with no actionable comments carries no comments at all — only a
        // present-but-not-a-list value is malformed.
        if (data.comments != null && !Array.isArray(data.comments)) throw new Error('invalid PR frame');
        queryClient.setQueryData<PRDetail>(key, current => ({
          ...current,
          pr: data.pr as PRInfo | undefined,
          comments: (data.comments ?? []) as PRComment[],
        }));
        setLoading(false);
      } catch {
        failFrame(queryClient, key, 'Pull request detail stream received an invalid PR update.');
        setLoading(false);
      }
    };
    const onRuns = (event: Event) => {
      try {
        const data = frame((event as MessageEvent<string>).data);
        if (!data.runs || typeof data.runs !== 'object' || Array.isArray(data.runs)) throw new Error('invalid runs frame');
        queryClient.setQueryData<PRDetail>(key, current => current ? {
          ...current,
          runs: data.runs as Record<string, WorkflowRun> | undefined,
        } : current);
      } catch {
        failFrame(queryClient, key, 'Pull request detail stream received an invalid runs update.');
      }
    };
    const onGavel = (event: Event) => {
      try {
        const data = frame((event as MessageEvent<string>).data);
        if (!Array.isArray(data.gavelResults)) throw new Error('invalid Gavel frame');
        queryClient.setQueryData<PRDetail>(key, current => current ? {
          ...current,
          gavelResults: data.gavelResults as GavelResultsSummary[] | undefined,
        } : current);
      } catch {
        failFrame(queryClient, key, 'Pull request detail stream received an invalid Gavel update.');
      }
    };
    const onErrorFrame = (event: Event) => {
      const messageEvent = event as MessageEvent<string>;
      if (!messageEvent.data) return;
      try {
        const data = frame(messageEvent.data);
        const message = typeof data.error === 'string' && data.error ? data.error : 'Load pull request detail failed';
        queryClient.setQueryData<PRDetail>(key, current => current ?? { error: message });
      } catch {
        failFrame(queryClient, key, 'Pull request detail stream received an invalid error update.');
      }
      setLoading(false);
    };
    const onDone = () => {
      if (streamRef.current !== stream) return;
      setLoading(false);
      stream.close();
      streamRef.current = null;
    };
    stream.addEventListener('pr', onPR);
    stream.addEventListener('runs', onRuns);
    stream.addEventListener('gavel', onGavel);
    stream.addEventListener('error', onErrorFrame);
    stream.addEventListener('done', onDone);
    stream.onerror = () => {
      if (streamRef.current !== stream) return;
      queryClient.setQueryData<PRDetail>(key, current => current ?? { error: 'Connection lost' });
      setLoading(false);
      stream.close();
      streamRef.current = null;
    };

    return () => {
      stream.removeEventListener('pr', onPR);
      stream.removeEventListener('runs', onRuns);
      stream.removeEventListener('gavel', onGavel);
      stream.removeEventListener('error', onErrorFrame);
      stream.removeEventListener('done', onDone);
      if (streamRef.current !== stream) return;
      stream.close();
      streamRef.current = null;
    };
  }, [pr?.number, pr?.repo, queryClient, revision]);

  return {
    detail: pr ? detailQuery.data ?? null : null,
    loading: !!pr && loading,
    refresh,
  };
}

function frame(data: string): Record<string, unknown> {
  const payload: unknown = JSON.parse(data);
  if (!payload || typeof payload !== 'object' || Array.isArray(payload)) throw new Error('invalid frame');
  return payload as Record<string, unknown>;
}

function failFrame(queryClient: ReturnType<typeof useQueryClient>, key: ReturnType<typeof queryKeys.prDetail>, error: string) {
  queryClient.setQueryData<PRDetail>(key, current => ({ ...current, error }));
}
