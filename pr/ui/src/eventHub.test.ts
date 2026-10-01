import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { EventSourceLike } from '@flanksource/clicky-ui/hooks';

// The hub is a module singleton (one real /api/events connection shared by
// every caller in the tab), so each test needs a fresh module instance —
// vi.resetModules() plus a dynamic re-import gives every test its own hub
// state instead of leaking subs/timers across tests.
async function freshHub() {
  vi.resetModules();
  return import('./eventHub');
}

class FakeRealEventSource {
  static instances: FakeRealEventSource[] = [];
  readonly listeners = new Map<string, Set<(ev: Event) => void>>();
  onerror: ((ev: Event) => void) | null = null;
  readonly close = vi.fn();

  constructor(readonly url: string) {
    FakeRealEventSource.instances.push(this);
  }

  addEventListener(type: string, listener: (ev: Event) => void) {
    let set = this.listeners.get(type);
    if (!set) {
      set = new Set();
      this.listeners.set(type, set);
    }
    set.add(listener);
  }

  removeEventListener(type: string, listener: (ev: Event) => void) {
    this.listeners.get(type)?.delete(listener);
  }

  emit(type: string, data: string, lastEventId?: string) {
    for (const listener of [...(this.listeners.get(type) ?? [])]) listener(new MessageEvent(type, { data, lastEventId }));
  }

  fail() {
    this.onerror?.(new Event('error'));
  }
}

function helloPayload(conn: string, build?: string) {
  return JSON.stringify(build === undefined ? { conn } : { conn, build });
}

function closedPayload(status: number, error?: string) {
  return JSON.stringify(error ? { status, error } : { status });
}

// fetchLog records every POST subscribe / DELETE unsubscribe the hub issues,
// and fetchResponses lets a test script a status per call (default 204).
let fetchLog: { url: string; method: string; body?: unknown }[] = [];
let fetchResponses: (number | (() => number))[] = [];

function installFetchMock() {
  vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input);
    const method = init?.method ?? 'GET';
    fetchLog.push({ url, method, body: init?.body ? JSON.parse(String(init.body)) : undefined });
    const scripted = fetchResponses.shift();
    const status = typeof scripted === 'function' ? scripted() : scripted ?? 204;
    return new Response(null, { status });
  }));
}

async function flush() {
  await Promise.resolve();
  await Promise.resolve();
  await Promise.resolve();
}

// Sets the `<meta name="gavel-ui-build">` tag the Go server stamps into every
// served SPA page, before the module-under-test is (re)imported — eventHub
// reads it once at module init, so it must be in place before freshHub().
function setPageBuildMeta(content: string): void {
  const meta = document.createElement('meta');
  meta.name = 'gavel-ui-build';
  meta.content = content;
  document.head.appendChild(meta);
}

const originalLocationDescriptor = Object.getOwnPropertyDescriptor(window, 'location')!;

function stubReload(): ReturnType<typeof vi.fn> {
  const reload = vi.fn();
  Object.defineProperty(window, 'location', {
    value: { ...window.location, reload },
    writable: true,
    configurable: true,
  });
  return reload;
}

beforeEach(() => {
  FakeRealEventSource.instances = [];
  fetchLog = [];
  fetchResponses = [];
  vi.stubGlobal('EventSource', FakeRealEventSource);
  installFetchMock();
});

afterEach(() => {
  vi.useRealTimers();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
  for (const meta of document.querySelectorAll('meta[name="gavel-ui-build"]')) meta.remove();
  Object.defineProperty(window, 'location', originalLocationDescriptor);
});

describe('eventHub', () => {
  it('opens CONNECTING, subscribes once hello arrives, and dispatches open on 204', async () => {
    const { openEventStream } = await freshHub();
    const stream = openEventStream('/api/prs/stream');
    expect(stream.readyState).toBe(0);
    expect(FakeRealEventSource.instances).toHaveLength(1);
    expect(fetchLog).toHaveLength(0); // no hello yet — nothing subscribed

    const opened = vi.fn();
    stream.onopen = opened;
    FakeRealEventSource.instances[0].emit('__hello', helloPayload('conn-1'));
    await flush();

    expect(fetchLog).toEqual([{ url: '/api/events/conn-1/subs', method: 'POST', body: { id: 's1', path: '/api/prs/stream' } }]);
    expect(stream.readyState).toBe(1);
    expect(opened).toHaveBeenCalledTimes(1);
  });

  it('routes a named event, including multi-line data, only to the matching virtual source', async () => {
    const { openEventStream } = await freshHub();
    const first = openEventStream('/api/prs/stream');
    const second = openEventStream('/api/proc/status/stream');
    const real = FakeRealEventSource.instances[0];
    real.emit('__hello', helloPayload('conn-1'));
    await flush();

    const firstMessages: string[] = [];
    const secondMessages: string[] = [];
    first.addEventListener('message', (event) => firstMessages.push((event as MessageEvent).data));
    second.addEventListener('message', (event) => secondMessages.push((event as MessageEvent).data));

    real.emit('s1/message', 'line one\nline two');
    real.emit('s2/message', 'other stream');

    expect(firstMessages).toEqual(['line one\nline two']);
    expect(secondMessages).toEqual(['other stream']);
  });

  it('multiplexes two virtual sources over a single real connection', async () => {
    const { openEventStream } = await freshHub();
    openEventStream('/api/prs/stream');
    openEventStream('/api/tests/stream');

    expect(FakeRealEventSource.instances).toHaveLength(1);
    FakeRealEventSource.instances[0].emit('__hello', helloPayload('conn-1'));
    await flush();

    expect(fetchLog.filter(call => call.method === 'POST').map(call => (call.body as { path: string }).path)).toEqual([
      '/api/prs/stream',
      '/api/tests/stream',
    ]);
  });

  it('DELETEs the sub and stops dispatching once close() is called', async () => {
    const { openEventStream } = await freshHub();
    const stream = openEventStream('/api/activity/stream');
    const real = FakeRealEventSource.instances[0];
    real.emit('__hello', helloPayload('conn-1'));
    await flush();

    const messages: string[] = [];
    stream.addEventListener('message', (event) => messages.push((event as MessageEvent).data));
    stream.close();
    await flush();

    expect(stream.readyState).toBe(2);
    expect(fetchLog.at(-1)).toEqual({ url: '/api/events/conn-1/subs/s1', method: 'DELETE', body: undefined });

    // A frame for the now-closed sub id must not reach a dead listener — the
    // hub removed its real-connection routing on close().
    real.emit('s1/message', 'late frame');
    expect(messages).toEqual([]);
  });

  it('re-subscribes one second after a 2xx __closed frame, mirroring native reconnect', async () => {
    vi.useFakeTimers();
    const { openEventStream } = await freshHub();
    const stream = openEventStream('/api/prs/stream');
    const real = FakeRealEventSource.instances[0];
    real.emit('__hello', helloPayload('conn-1'));
    await vi.advanceTimersByTimeAsync(0);
    expect(stream.readyState).toBe(1);

    const errored = vi.fn();
    stream.onerror = errored;
    real.emit('s1/__closed', closedPayload(200));

    expect(stream.readyState).toBe(0);
    expect(errored).toHaveBeenCalledTimes(1);
    expect(fetchLog.filter(call => call.method === 'POST')).toHaveLength(1);

    await vi.advanceTimersByTimeAsync(1000);
    expect(fetchLog.filter(call => call.method === 'POST')).toHaveLength(2);
    expect(stream.readyState).toBe(1);
  });

  it('closes without retry on a non-2xx __closed frame and logs the failure', async () => {
    const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    const { openEventStream } = await freshHub();
    const stream = openEventStream('/api/todos/session/stream');
    const real = FakeRealEventSource.instances[0];
    real.emit('__hello', helloPayload('conn-1'));
    await flush();

    const errored = vi.fn();
    stream.onerror = errored;
    real.emit('s1/__closed', closedPayload(500, 'boom'));

    expect(stream.readyState).toBe(2);
    expect(errored).toHaveBeenCalledTimes(1);
    expect(errorSpy).toHaveBeenCalledWith(expect.stringContaining('/api/todos/session/stream'), 'boom');

    await flush();
    expect(fetchLog.filter(call => call.method === 'POST')).toHaveLength(1); // no retry
  });

  it('marks every sub CONNECTING with an error on a real connection drop, then re-subscribes all of them on the next hello', async () => {
    const { openEventStream } = await freshHub();
    const first = openEventStream('/api/prs/stream');
    const second = openEventStream('/api/proc/status/stream');
    const real = FakeRealEventSource.instances[0];
    real.emit('__hello', helloPayload('conn-1'));
    await flush();
    expect(first.readyState).toBe(1);
    expect(second.readyState).toBe(1);

    const firstError = vi.fn();
    const secondError = vi.fn();
    first.onerror = firstError;
    second.onerror = secondError;
    real.fail();

    expect(first.readyState).toBe(0);
    expect(second.readyState).toBe(0);
    expect(firstError).toHaveBeenCalledTimes(1);
    expect(secondError).toHaveBeenCalledTimes(1);

    const firstOpen = vi.fn();
    const secondOpen = vi.fn();
    first.onopen = firstOpen;
    second.onopen = secondOpen;
    real.emit('__hello', helloPayload('conn-2'));
    await flush();

    expect(fetchLog.filter(call => call.method === 'POST').map(call => call.url)).toEqual([
      '/api/events/conn-1/subs',
      '/api/events/conn-1/subs',
      '/api/events/conn-2/subs',
      '/api/events/conn-2/subs',
    ]);
    expect(first.readyState).toBe(1);
    expect(second.readyState).toBe(1);
    expect(firstOpen).toHaveBeenCalledTimes(1);
    expect(secondOpen).toHaveBeenCalledTimes(1);
  });

  it('tears down the real connection only after every sub has been closed for the grace period', async () => {
    vi.useFakeTimers();
    const { openEventStream } = await freshHub();
    const first: EventSourceLike = openEventStream('/api/prs/stream');
    const second = openEventStream('/api/proc/status/stream');
    const real = FakeRealEventSource.instances[0];
    real.emit('__hello', helloPayload('conn-1'));
    await vi.advanceTimersByTimeAsync(0);

    first.close();
    await vi.advanceTimersByTimeAsync(1000);
    expect(real.close).not.toHaveBeenCalled(); // second sub still open

    second.close();
    await vi.advanceTimersByTimeAsync(999);
    expect(real.close).not.toHaveBeenCalled();
    await vi.advanceTimersByTimeAsync(1);
    expect(real.close).toHaveBeenCalledTimes(1);

    // A subscribe after full teardown opens a brand-new real connection.
    openEventStream('/api/tests/stream');
    expect(FakeRealEventSource.instances).toHaveLength(2);
  });

  it('re-opens within the grace window without tearing down the real connection', async () => {
    vi.useFakeTimers();
    const { openEventStream } = await freshHub();
    const first = openEventStream('/api/prs/stream');
    const real = FakeRealEventSource.instances[0];
    real.emit('__hello', helloPayload('conn-1'));
    await vi.advanceTimersByTimeAsync(0);

    first.close();
    openEventStream('/api/prs/stream'); // React StrictMode-style immediate re-open
    await vi.advanceTimersByTimeAsync(1000);

    expect(real.close).not.toHaveBeenCalled();
    expect(FakeRealEventSource.instances).toHaveLength(1);
  });
});

describe('eventHub build-mismatch reload', () => {
  it('reloads once when the hello build differs from the page meta build', async () => {
    setPageBuildMeta('build-a');
    const reload = stubReload();
    const { openEventStream } = await freshHub();
    openEventStream('/api/prs/stream');
    const real = FakeRealEventSource.instances[0];

    real.emit('__hello', helloPayload('conn-1', 'build-b'));
    await flush();
    expect(reload).toHaveBeenCalledTimes(1);

    // A second hello (e.g. the reconnect after a drop, racing the reload)
    // must not trigger a second reload within the same page load.
    real.emit('__hello', helloPayload('conn-2', 'build-b'));
    await flush();
    expect(reload).toHaveBeenCalledTimes(1);
  });

  it('does not reload when the hello build matches the page meta build', async () => {
    setPageBuildMeta('build-a');
    const reload = stubReload();
    const { openEventStream } = await freshHub();
    openEventStream('/api/prs/stream');
    const real = FakeRealEventSource.instances[0];

    real.emit('__hello', helloPayload('conn-1', 'build-a'));
    await flush();

    expect(reload).not.toHaveBeenCalled();
  });

  it('does not reload when the page has no build meta tag', async () => {
    const reload = stubReload();
    const { openEventStream } = await freshHub();
    openEventStream('/api/prs/stream');
    const real = FakeRealEventSource.instances[0];

    real.emit('__hello', helloPayload('conn-1', 'build-b'));
    await flush();

    expect(reload).not.toHaveBeenCalled();
  });

  it('does not reload when the hello omits a build field', async () => {
    setPageBuildMeta('build-a');
    const reload = stubReload();
    const { openEventStream } = await freshHub();
    openEventStream('/api/prs/stream');
    const real = FakeRealEventSource.instances[0];

    real.emit('__hello', helloPayload('conn-1'));
    await flush();

    expect(reload).not.toHaveBeenCalled();
  });
});
