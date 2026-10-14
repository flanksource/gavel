import { chromium } from '@playwright/test';
import type { BrowserContext, ConsoleMessage, Response, WebError } from '@playwright/test';

// Background page watcher for a gavel fixture `playwright` step with
// `background: true`. Run with plain `node` (type stripping), never as a
// Playwright test: it attaches to the fixture's shared Chrome over
// BROWSER_CDP_URL, opens TARGET_URL and streams one NDJSON event per line on
// stdout until SIGTERM:
//
//   {"type":"ready","url":…}                       page loaded, watching
//   {"type":"alert","text":…,"details":…,"url":…}  a role=alert "Session error(s)" banner,
//                                                  or a stream error banner ("banner":true)
//   {"type":"console","text":…,"url":…}            console.error on any page
//   {"type":"pageerror","message":…,"stack":…}     an uncaught page exception
//   {"type":"response","method":…,"url":…,"status":…}  an /api/ response >= 400
//
// The listeners are context-wide, so pages other steps open in the shared
// profile (agent-browser, later Playwright specs) are watched too. The gavel
// kind fails the step at teardown when any non-ready event was seen.
//
// Selectors follow src/components/todos/SessionErrorDetails.tsx: the banner is
// role="alert" with the heading "Session error" / "N session errors" and a
// "Show details" toggle that reveals the error text in a <pre>.

type WatchEvent = { type: 'ready' | 'alert' | 'console' | 'pageerror' | 'response' } & Record<string, unknown>;

function emit(event: WatchEvent) {
  process.stdout.write(`${JSON.stringify({ time: new Date().toISOString(), ...event })}\n`);
}

function requiredEnv(name: string): string {
  const value = process.env[name];
  if (!value) throw new Error(`${name} is not set: run watch-session.ts through a gavel fixture playwright step`);
  return value;
}

// Runs inside every page before its scripts. A MutationObserver (not a poll)
// so a banner that appears and disappears between two polls is still caught.
// Self-contained: Playwright serialises it with Function.prototype.toString.
//
// Two kinds of banner are reported as `alert` events: the Session tab's
// role="alert" "Session error(s)" banner, and the dashboard's stream banners
// ("Pull request stream received an invalid update." and its siblings in
// src/use*.ts), which are plain text with no role and no console output.
function observeSessionAlerts() {
  const pattern = /session errors?/i;
  const streamPattern = /stream received an invalid update|received an invalid update\./i;
  const reported = new WeakSet<Element>();
  const report = (alert: object) =>
    (window as unknown as { __gavelSessionAlert: (alert: object) => void }).__gavelSessionAlert(alert);

  const captureBanner = (element: Element) => {
    if (reported.has(element)) return;
    reported.add(element);
    report({ text: (element.textContent ?? '').trim(), details: '', url: location.href, banner: true });
  };

  const scanText = (root: Node) => {
    const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT);
    for (let node = root.nodeType === Node.TEXT_NODE ? root : walker.nextNode(); node; node = walker.nextNode()) {
      if (streamPattern.test(node.textContent ?? '') && node.parentElement) captureBanner(node.parentElement);
    }
  };

  const capture = (alert: Element) => {
    if (reported.has(alert) || !pattern.test(alert.textContent ?? '')) return;
    reported.add(alert);
    const text = (alert.textContent ?? '').trim();
    alert.querySelector<HTMLButtonElement>('button[aria-label="Show details"]')?.click();
    const started = Date.now();
    const finish = () => {
      const details = alert.querySelector('pre')?.textContent ?? '';
      if (!details && alert.isConnected && Date.now() - started < 2000) {
        setTimeout(finish, 25);
        return;
      }
      report({ text, details, url: location.href, transient: !alert.isConnected });
    };
    finish();
  };

  const scan = (node: Node) => {
    const element = node instanceof Element ? node : node.parentElement;
    if (!element) return;
    const owner = element.closest('[role="alert"]');
    if (owner) capture(owner);
    element.querySelectorAll('[role="alert"]').forEach(capture);
  };

  const observer = new MutationObserver((records) => {
    for (const record of records) {
      scan(record.target);
      record.addedNodes.forEach(scan);
      if (record.type === 'characterData') scanText(record.target);
      record.addedNodes.forEach(scanText);
    }
  });
  const start = () => {
    observer.observe(document.documentElement, { childList: true, subtree: true, characterData: true });
    scan(document.documentElement);
    scanText(document.documentElement);
  };
  if (document.documentElement) start();
  else document.addEventListener('readystatechange', start, { once: true });
}

function isAPIFailure(response: Response): boolean {
  return response.status() >= 400 && new URL(response.url()).pathname.startsWith('/api/');
}

async function watch(context: BrowserContext, targetURL: string) {
  await context.exposeBinding('__gavelSessionAlert', (_source, alert: Record<string, unknown>) => {
    emit({ type: 'alert', ...alert });
  });
  await context.addInitScript(observeSessionAlerts);
  context.on('console', (message: ConsoleMessage) => {
    if (message.type() === 'error') emit({ type: 'console', text: message.text(), url: message.location().url });
  });
  context.on('weberror', (error: WebError) => {
    emit({ type: 'pageerror', message: error.error().message, stack: error.error().stack ?? '' });
  });
  context.on('response', (response: Response) => {
    if (isAPIFailure(response)) {
      emit({ type: 'response', method: response.request().method(), url: response.url(), status: response.status() });
    }
  });

  const page = await context.newPage();
  await page.goto(targetURL);
  emit({ type: 'ready', url: page.url() });
}

const cdpURL = requiredEnv('BROWSER_CDP_URL');
const targetURL = requiredEnv('TARGET_URL');
const browser = await chromium.connectOverCDP(cdpURL);
const [context] = browser.contexts();
if (!context) throw new Error(`Chrome at ${cdpURL} exposes no default browser context`);

// Pipes are asynchronous on macOS: exit only once every queued event line has
// been flushed. Exiting drops the CDP connection only; the shared Chrome
// belongs to the fixture's browser step.
const keepAlive = setInterval(() => {}, 60_000);
for (const signal of ['SIGTERM', 'SIGINT'] as const) {
  process.on(signal, () => {
    clearInterval(keepAlive);
    process.stdout.write('', () => process.exit(0));
  });
}

await watch(context, targetURL);
