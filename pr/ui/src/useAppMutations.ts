import { useCallback, useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { mutationJSON, queryKeys } from './query';
import type { PRItem, SearchConfig, Snapshot } from './types';
import { prKey } from './utils';

export interface AppMutations {
  error: string;
  markSeen: (pr: PRItem) => void;
  refresh: () => void;
  togglePause: () => void;
  saveConfig: (config: SearchConfig) => void;
  setIncludeBots: (include: boolean) => void;
  setShowClosed: (show: boolean) => void;
}

export function useAppMutations(): AppMutations {
  const queryClient = useQueryClient();
  const [error, setError] = useState('');
  const clearError = useCallback(() => setError(''), []);
  const setSnapshot = useCallback((updater: (current: Snapshot) => Snapshot) => {
    queryClient.setQueryData<Snapshot>(queryKeys.prSnapshot(), current => current ? updater(current) : current);
  }, [queryClient]);
  const reportError = useCallback((cause: unknown) => {
    setError(cause instanceof Error ? cause.message : 'Application mutation failed');
  }, []);

  const seenMutation = useMutation({
    mutationFn: (pr: PRItem) => mutationJSON({
      url: '/api/prs/seen',
      method: 'POST',
      body: { repo: pr.repo, number: pr.number },
      context: 'Mark pull request seen',
    }),
    onMutate: async pr => {
      clearError();
      await queryClient.cancelQueries({ queryKey: queryKeys.prSnapshot(), exact: true });
      const key = prKey(pr);
      const wasUnread = !!queryClient.getQueryData<Snapshot>(queryKeys.prSnapshot())?.unread?.[key];
      setSnapshot(current => {
        if (!current.unread?.[key]) return current;
        const unread = { ...current.unread };
        delete unread[key];
        return { ...current, unread };
      });
      return { key, wasUnread };
    },
    onError: (cause, _pr, context) => {
      if (context?.wasUnread) {
        setSnapshot(current => ({ ...current, unread: { ...current.unread, [context.key]: true } }));
      }
      reportError(cause);
    },
  });
  const refreshMutation = useMutation({
    mutationFn: () => mutationJSON({
      url: '/api/prs/refresh',
      method: 'POST',
      context: 'Refresh pull requests',
    }),
    onMutate: clearError,
    onSuccess: async () => queryClient.invalidateQueries({ queryKey: queryKeys.prSnapshot(), exact: true }),
    onError: reportError,
  });
  const pauseMutation = useMutation({
    mutationFn: async () => requireBoolean(
      await mutationJSON({ url: '/api/prs/pause', method: 'POST', context: 'Toggle pull request polling' }),
      'paused',
      'Toggle pull request polling',
    ),
    onMutate: clearError,
    onSuccess: paused => setSnapshot(current => ({ ...current, paused })),
    onError: reportError,
  });
  const configMutation = useMutation({
    mutationFn: async (config: SearchConfig) => requireConfig(
      await mutationJSON({ url: '/api/config', method: 'POST', body: config, context: 'Save pull request config' }),
    ),
    onMutate: async config => {
      clearError();
      await queryClient.cancelQueries({ queryKey: queryKeys.prSnapshot(), exact: true });
      const previous = queryClient.getQueryData<Snapshot>(queryKeys.prSnapshot())?.config;
      setSnapshot(current => ({ ...current, config }));
      return { previous };
    },
    onSuccess: config => setSnapshot(current => ({ ...current, config })),
    onError: (cause, _config, context) => {
      if (context?.previous) setSnapshot(current => ({ ...current, config: context.previous! }));
      reportError(cause);
    },
  });
  const botsMutation = useMutation({
    mutationFn: async (include: boolean) => requireBoolean(
      await mutationJSON({
        url: '/api/prs/bots',
        method: 'POST',
        body: { include },
        context: 'Update bot pull request visibility',
      }),
      'includeBots',
      'Update bot pull request visibility',
    ),
    onMutate: clearError,
    onSuccess: includeBots => setSnapshot(current => ({ ...current, includeBots })),
    onError: reportError,
  });
  const closedMutation = useMutation({
    mutationFn: async (show: boolean) => requireBoolean(
      await mutationJSON({
        url: '/api/prs/closed',
        method: 'POST',
        body: { show },
        context: 'Update closed pull request visibility',
      }),
      'showClosed',
      'Update closed pull request visibility',
    ),
    onMutate: clearError,
    onSuccess: showClosed => setSnapshot(current => ({ ...current, showClosed })),
    onError: reportError,
  });

  return {
    error,
    markSeen: seenMutation.mutate,
    refresh: refreshMutation.mutate,
    togglePause: pauseMutation.mutate,
    saveConfig: configMutation.mutate,
    setIncludeBots: botsMutation.mutate,
    setShowClosed: closedMutation.mutate,
  };
}

function requireBoolean(payload: unknown, field: string, context: string): boolean {
  if (!payload || typeof payload !== 'object' || !(field in payload) || typeof payload[field as keyof typeof payload] !== 'boolean') {
    throw new Error(`${context}: invalid response`);
  }
  return payload[field as keyof typeof payload] as boolean;
}

function requireConfig(payload: unknown): SearchConfig {
  if (!payload || typeof payload !== 'object' || !('repos' in payload) || !Array.isArray(payload.repos)) {
    throw new Error('Save pull request config: invalid response');
  }
  return payload as SearchConfig;
}
