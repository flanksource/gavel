import { act, renderHook } from '@testing-library/react';
import { QueryClient, useQuery } from '@tanstack/react-query';
import { afterEach, expect, it, vi } from 'vitest';
import { queryTestWrapper } from './queryTestWrapper';
import { useWorkspaceChanges } from './useWorkspaceChanges';

const streams = vi.hoisted(() => [] as EventTarget[]);
vi.mock('../../eventHub', () => ({ openEventStream: () => {
  const target = new EventTarget(); streams.push(target);
  return { addEventListener: target.addEventListener.bind(target), removeEventListener: target.removeEventListener.bind(target), close: vi.fn(), onerror: null };
} }));
afterEach(() => { vi.useRealTimers(); streams.length = 0; });

it('invalidates only changed workspace queries and preserves unchanged remounts', async () => {
  vi.useFakeTimers();
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const wrapper = queryTestWrapper(client);
  const first = vi.fn(async () => ({ items: [] }));
  const second = vi.fn(async () => ({ items: [] }));
  const mounted = renderHook(() => {
    useQuery({ queryKey: ['todos', 'list', { dir: '/first' }], queryFn: first, initialData: { items: [] }, staleTime: Infinity });
    useQuery({ queryKey: ['todos', 'list', { dir: '/second' }], queryFn: second, initialData: { items: [] }, staleTime: Infinity });
    return useWorkspaceChanges(true);
  }, { wrapper });
  const emit = (revision: number, epoch = 'boot') => streams.at(-1)!.dispatchEvent(new MessageEvent('message', { data: JSON.stringify({ epoch, catalog: 1, workspaces: { '/first': { revision }, '/second': { revision: 1 } } }) }));
  await act(async () => { emit(1); await vi.advanceTimersByTimeAsync(30_000); });
  expect(first).not.toHaveBeenCalled(); expect(second).not.toHaveBeenCalled();
  await act(async () => { emit(2); emit(2); await vi.advanceTimersByTimeAsync(1001); });
  expect(first).toHaveBeenCalledTimes(1); expect(second).not.toHaveBeenCalled();
  mounted.unmount();
  renderHook(() => useWorkspaceChanges(true), { wrapper });
  await act(async () => { emit(2); await vi.advanceTimersByTimeAsync(1001); });
  expect(client.getQueryState(['todos', 'list', { dir: '/second' }])?.isInvalidated).toBe(false);
  await act(async () => { emit(2, 'reconnected'); await vi.advanceTimersByTimeAsync(1001); });
  expect(client.getQueryState(['todos', 'list', { dir: '/first' }])?.isInvalidated).toBe(true);
  expect(client.getQueryState(['todos', 'list', { dir: '/second' }])?.isInvalidated).toBe(true);
});

it('separates taxonomy and run configuration from TODO usage changes', async () => {
  vi.useFakeTimers();
  const client = new QueryClient();
  const list = vi.fn(async () => ({}));
  const tags = vi.fn(async () => ({}));
  const config = vi.fn(async () => ({}));
  renderHook(() => {
    useQuery({ queryKey: ['todos', 'list', { dir: '/first' }], queryFn: list, initialData: {}, staleTime: Infinity });
    useQuery({ queryKey: ['todos', 'tags', { dir: '/first' }], queryFn: tags, initialData: {}, staleTime: Infinity });
    useQuery({ queryKey: ['todos', 'run-context', '/first'], queryFn: config, initialData: {}, staleTime: Infinity });
    return useWorkspaceChanges(true);
  }, { wrapper: queryTestWrapper(client) });
  const emit = (revision: number, labels: number, config: number) => streams.at(-1)!.dispatchEvent(new MessageEvent('message', {
    data: JSON.stringify({ epoch: 'boot', catalog: 1, workspaces: { '/first': { revision, labels, config } } }),
  }));
  await act(async () => { emit(1, 1, 1); emit(2, 1, 1); await vi.advanceTimersByTimeAsync(1001); });
  expect(list).toHaveBeenCalledTimes(1);
  expect(tags).not.toHaveBeenCalled(); expect(config).not.toHaveBeenCalled();
  await act(async () => { emit(2, 2, 2); await vi.advanceTimersByTimeAsync(1001); });
  expect(tags).toHaveBeenCalledTimes(1); expect(config).toHaveBeenCalledTimes(1);
});

it('refetches again when a change arrives during an in-flight read', async () => {
  vi.useFakeTimers();
  const client = new QueryClient();
  let complete: (value: object) => void = () => { throw new Error('read has not started'); };
  const read = vi.fn().mockImplementationOnce(() => new Promise(resolve => { complete = resolve; })).mockResolvedValue({});
  renderHook(() => {
    useQuery({ queryKey: ['todos', 'list', { dir: '/first' }], queryFn: read, initialData: {}, staleTime: Infinity });
    return useWorkspaceChanges(true);
  }, { wrapper: queryTestWrapper(client) });
  const emit = (revision: number) => streams.at(-1)!.dispatchEvent(new MessageEvent('message', {
    data: JSON.stringify({ epoch: 'boot', catalog: 1, workspaces: { '/first': { revision } } }),
  }));
  await act(async () => { emit(1); emit(2); await vi.advanceTimersByTimeAsync(1001); });
  expect(read).toHaveBeenCalledTimes(1);
  await act(async () => { emit(3); await vi.advanceTimersByTimeAsync(1001); complete({}); await vi.advanceTimersByTimeAsync(1001); });
  expect(read).toHaveBeenCalledTimes(2);
});

it('marks the all-workspace list dirty when any workspace revision or the catalog changes', async () => {
  vi.useFakeTimers();
  const client = new QueryClient();
  const list = vi.fn(async () => ({ counts: {}, items: [] }));
  renderHook(() => {
    useQuery({ queryKey: ['todos', 'list', { exclude: 'hidden-projects' }], queryFn: list, initialData: { counts: {}, items: [] }, staleTime: Infinity });
    return useWorkspaceChanges(true);
  }, { wrapper: queryTestWrapper(client) });
  const emit = (workspaces: Record<string, { revision: number }>, catalog = 1) => streams.at(-1)!.dispatchEvent(new MessageEvent('message', {
    data: JSON.stringify({ epoch: 'boot', catalog, workspaces }),
  }));
  await act(async () => { emit({ '/first': { revision: 1 }, '/second': { revision: 1 } }); await vi.advanceTimersByTimeAsync(30_000); });
  expect(list).not.toHaveBeenCalled();
  await act(async () => { emit({ '/first': { revision: 1 }, '/second': { revision: 2 } }); await vi.advanceTimersByTimeAsync(1001); });
  expect(list).toHaveBeenCalledTimes(1);
  await act(async () => { emit({ '/first': { revision: 1 }, '/second': { revision: 2 } }, 2); await vi.advanceTimersByTimeAsync(1001); });
  expect(list).toHaveBeenCalledTimes(2);
});
