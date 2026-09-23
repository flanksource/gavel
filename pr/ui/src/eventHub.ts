import type { EventSourceLike } from '@flanksource/clicky-ui/hooks';

// Chrome allows only six HTTP/1.1 connections per host, shared across every
// tab. Each dashboard tab used to hold one native EventSource per topic (PR
// list, proc status, test runs, PR detail, activity, todo session, ...), so a
// second tab or a reload starved ordinary fetches (e.g. Settings) behind a
// queue that never drained.
//
// This hub multiplexes every topic a tab opens onto a single real connection
// to the server's /api/events (see pr/ui/events.go): one native EventSource,
// with each logical stream registered as a "sub" (POST /api/events/{conn}/subs,
// DELETE .../subs/{id}) whose frames arrive prefixed `<subId>/<name>`. Callers
// never see any of this — openEventStream(url) returns an EventSourceLike that
// behaves like `new EventSource(url)` from their point of view.
const REAL_URL = '/api/events';

// The build id this page was served with, read once at module init from the
// `<meta name="gavel-ui-build">` tag the Go server stamps into every served
// SPA HTML page (pr/ui/handler.go). Absent in tests, Storybook, and a bare
// `vite dev` server with no Go backend in front of it — those never reload.
const pageBuildId = typeof document !== 'undefined'
  ? document.querySelector<HTMLMetaElement>('meta[name="gavel-ui-build"]')?.content || null
  : null;

// Guards against a reload loop within one page load: once triggered, further
// __hello frames (e.g. after the reload's own new connection reconnects
// mid-flight) never trigger a second one.
let reloadedForBuildMismatch = false;

// A stale tab still serving an old bundle would otherwise keep working against
// a server that has moved on — including racing the six-connection cap this
// hub exists to close, since an old bundle may hold raw EventSources of its
// own. When the server's /api/events hello reports a different build than the
// one this page was served with, reload once to pick up the current bundle.
function maybeReloadForBuildMismatch(helloBuildId: string | undefined): void {
  if (reloadedForBuildMismatch) return;
  if (!helloBuildId || !pageBuildId) return;
  if (helloBuildId === pageBuildId) return;
  reloadedForBuildMismatch = true;
  window.location.reload();
}

// How long an idle hub (no subs left) keeps the real connection open before
// tearing it down — long enough to survive React StrictMode's mount → unmount
// → mount churn and a component swap that closes one stream and immediately
// opens its replacement.
const TEARDOWN_GRACE_MS = 1000;

// Delay before re-subscribing after a sub's handler ends with a 2xx status
// (server-side stream completed normally and can be re-opened), mirroring
// native EventSource's own reconnect delay.
const RESUBSCRIBE_DELAY_MS = 1000;

const READY_STATE_CONNECTING = 0;
const READY_STATE_OPEN = 1;
const READY_STATE_CLOSED = 2;

type DomListener = EventListenerOrEventListenerObject;

function invokeListener(listener: DomListener, event: Event): void {
  if (typeof listener === 'function') listener(event);
  else listener.handleEvent(event);
}

// Sub is both the internal bookkeeping for one logical stream and the public
// EventSourceLike handed back to the caller — there is no separate wrapper
// object, so `readyState`/`url` always reflect live state.
class Sub implements EventSourceLike {
  readyState = READY_STATE_CONNECTING;
  onopen: ((ev: Event) => unknown) | null = null;
  onmessage: ((ev: MessageEvent) => unknown) | null = null;
  onerror: ((ev: Event) => unknown) | null = null;

  readonly listeners = new Map<string, Set<DomListener>>();
  // Real-connection listeners this sub has registered, keyed by the prefixed
  // event name (`${id}/${name}`), so close() can remove exactly what it added.
  readonly routedHandlers = new Map<string, (ev: Event) => void>();
  resubscribeTimer: ReturnType<typeof setTimeout> | null = null;
  disposed = false;

  constructor(readonly id: string, readonly url: string) {}

  addEventListener(type: string, listener: DomListener): void {
    let set = this.listeners.get(type);
    if (!set) {
      set = new Set();
      this.listeners.set(type, set);
    }
    set.add(listener);
    routeName(this, type);
  }

  removeEventListener(type: string, listener: DomListener): void {
    this.listeners.get(type)?.delete(listener);
  }

  close(): void {
    closeSub(this);
  }
}

function dispatch(sub: Sub, name: string, event: Event): void {
  for (const listener of [...(sub.listeners.get(name) ?? [])]) invokeListener(listener, event);
  // onmessage/onerror/onopen mirror native EventSource: each is just the
  // single-slot form of addEventListener('message'|'error'|'open', ...), so a
  // frame named "error" (server-sent) reaches both a registered
  // addEventListener('error', ...) and .onerror, exactly like the browser.
  if (name === 'message') sub.onmessage?.(event as MessageEvent);
  if (name === 'error') sub.onerror?.(event);
  if (name === 'open') sub.onopen?.(event);
}

// --- hub singleton state -----------------------------------------------

let real: EventSourceLike | null = null;
let connId: string | null = null;
let subCounter = 0;
const subs = new Map<string, Sub>();
let teardownTimer: ReturnType<typeof setTimeout> | null = null;

/**
 * Opens a virtual, EventSource-compatible stream for `url`, multiplexed with
 * every other stream this tab has open over one real connection to
 * /api/events. Behaves like `new EventSource(url)` to the caller: CONNECTING
 * until the subscription is live, dispatches 'open'/'message'/'error' (and
 * any server-named event), and auto-reconnects the way native EventSource
 * does.
 */
export function openEventStream(url: string): EventSourceLike {
  const sub = new Sub(`s${++subCounter}`, url);
  subs.set(sub.id, sub);
  ensureRealOpen();
  // Routing must be registered before the subscribe POST is sent: frames can
  // start arriving on the real connection as soon as the server accepts the
  // subscription, which can race the POST's own response.
  routeName(sub, 'message');
  routeClosed(sub);
  if (connId) void postSub(sub);
  return sub;
}

function ensureRealOpen(): void {
  if (real) {
    if (teardownTimer) {
      clearTimeout(teardownTimer);
      teardownTimer = null;
    }
    return;
  }
  connId = null;
  const es = new EventSource(REAL_URL);
  es.addEventListener('__hello', event => onHello(event as MessageEvent));
  es.onerror = () => onRealError();
  real = es;
}

function onHello(event: MessageEvent): void {
  const payload = JSON.parse(event.data) as { conn: string; build?: string };
  connId = payload.conn;
  maybeReloadForBuildMismatch(payload.build);
  // A hello means every currently-active sub needs a fresh subscription: the
  // first hello of a brand-new connection, or a reconnect hello after a drop,
  // both start from a server side with no subs registered yet.
  for (const sub of subs.values()) {
    if (!sub.disposed) void postSub(sub);
  }
}

function onRealError(): void {
  connId = null;
  for (const sub of subs.values()) {
    if (sub.disposed) continue;
    sub.readyState = READY_STATE_CONNECTING;
    dispatch(sub, 'error', new Event('error'));
  }
  // Native EventSource reconnects on its own; the next __hello re-subscribes
  // everything and dispatches 'open' again.
}

async function postSub(sub: Sub): Promise<void> {
  const conn = connId;
  if (!conn || sub.disposed) return;
  try {
    const res = await fetch(`/api/events/${conn}/subs`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id: sub.id, path: sub.url }),
    });
    if (sub.disposed) return;
    if (res.status === 204) {
      sub.readyState = READY_STATE_OPEN;
      dispatch(sub, 'open', new Event('open'));
      return;
    }
    if (res.status === 404) return; // stale conn id; the next __hello retries
    failSub(sub, `subscribe failed: ${res.status}`);
  } catch (err) {
    if (sub.disposed) return;
    failSub(sub, 'subscribe request failed', err);
  }
}

function failSub(sub: Sub, message: string, cause?: unknown): void {
  sub.readyState = READY_STATE_CLOSED;
  dispatch(sub, 'error', new Event('error'));
  if (cause !== undefined) console.error(`[eventHub] ${message} for ${sub.url}:`, cause);
  else console.error(`[eventHub] ${message} for ${sub.url}`);
}

// routeName registers, at most once per sub+name, a real-connection listener
// that re-dispatches `${sub.id}/${name}` frames to this sub's own listeners.
// "open" is synthetic (dispatched locally on subscribe/reconnect) and never
// arrives on the wire, so it is never routed.
function routeName(sub: Sub, name: string): void {
  if (name === 'open' || sub.routedHandlers.has(eventNameFor(sub, name))) return;
  const eventName = eventNameFor(sub, name);
  const handler = (event: Event) => {
    const message = event as MessageEvent;
    dispatch(sub, name, new MessageEvent(name, { data: message.data, lastEventId: message.lastEventId }));
  };
  real!.addEventListener(eventName, handler);
  sub.routedHandlers.set(eventName, handler);
}

function eventNameFor(sub: Sub, name: string): string {
  return `${sub.id}/${name}`;
}

// routeClosed wires the sub's terminal frame: unlike every other name, its
// payload drives re-subscription rather than being handed to consumers.
function routeClosed(sub: Sub): void {
  const eventName = eventNameFor(sub, '__closed');
  if (sub.routedHandlers.has(eventName)) return;
  const handler = (event: Event) => onSubClosed(sub, event as MessageEvent);
  real!.addEventListener(eventName, handler);
  sub.routedHandlers.set(eventName, handler);
}

function onSubClosed(sub: Sub, event: MessageEvent): void {
  if (sub.disposed) return;
  let payload: { status: number; error?: string };
  try {
    payload = JSON.parse(event.data) as { status: number; error?: string };
  } catch (err) {
    failSub(sub, 'closed frame was malformed', err);
    return;
  }
  if (payload.status >= 200 && payload.status < 300) {
    // The server-side handler ended on its own with a success status — the
    // same shape as a native stream simply reconnecting.
    sub.readyState = READY_STATE_CONNECTING;
    dispatch(sub, 'error', new Event('error'));
    sub.resubscribeTimer = setTimeout(() => {
      sub.resubscribeTimer = null;
      if (!sub.disposed) void postSub(sub);
    }, RESUBSCRIBE_DELAY_MS);
    return;
  }
  failSub(sub, `sub closed with status ${payload.status}`, payload.error);
}

function closeSub(sub: Sub): void {
  if (sub.disposed) return;
  sub.disposed = true;
  sub.readyState = READY_STATE_CLOSED;
  if (sub.resubscribeTimer) {
    clearTimeout(sub.resubscribeTimer);
    sub.resubscribeTimer = null;
  }
  for (const [eventName, handler] of sub.routedHandlers) real?.removeEventListener(eventName, handler);
  sub.routedHandlers.clear();
  subs.delete(sub.id);

  const conn = connId;
  if (conn) {
    fetch(`/api/events/${conn}/subs/${sub.id}`, { method: 'DELETE' })
      .then(res => {
        if (!res.ok && res.status !== 404) console.error(`[eventHub] unsubscribe failed for ${sub.url}: ${res.status}`);
      })
      .catch(err => console.error(`[eventHub] unsubscribe request failed for ${sub.url}:`, err));
  }
  scheduleTeardownIfIdle();
}

function scheduleTeardownIfIdle(): void {
  if (subs.size > 0 || teardownTimer) return;
  teardownTimer = setTimeout(() => {
    teardownTimer = null;
    if (subs.size > 0) return; // something re-subscribed during the grace window
    real?.close();
    real = null;
    connId = null;
  }, TEARDOWN_GRACE_MS);
}
