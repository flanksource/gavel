import { expect, test, type Page } from '@playwright/test';

// Driven by hack/session_tail_e2e.go, which seeds a running captain session of
// SESSION_TAIL_SEEDED messages ("tail message 000" …) and appends more on
// POST {SESSION_TAIL_CONTROL_URL}/append?n=. The backend requests this page
// makes are asserted by pr/ui/testdata/session-tail.fixture.md from the HAR.

function env(name: string) {
  const value = process.env[name];
  if (!value) throw new Error(`${name} is not set: run this spec through hack/session_tail_e2e.go`);
  return value;
}

const baseURL = env('SESSION_TAIL_BASE_URL');
const sessionId = env('SESSION_TAIL_ID');
const controlURL = env('SESSION_TAIL_CONTROL_URL');
const seeded = Number(env('SESSION_TAIL_SEEDED'));
const label = (index: number) => `tail message ${String(index).padStart(3, '0')}`;

async function append(count: number) {
  const response = await fetch(`${controlURL}/append?n=${count}`, { method: 'POST' });
  if (!response.ok) throw new Error(`append ${count}: HTTP ${response.status} ${await response.text()}`);
  return ((await response.json()) as { last: number }).last;
}

// Scrolls the transcript's scroll container (the nearest scrollable ancestor
// of a rendered message) to the top, as a reader scrolling back would.
async function scrollTranscriptToTop(page: Page) {
  await page.getByText(/^tail message \d{3}$/).first().evaluate((message) => {
    let node: HTMLElement | null = message.parentElement;
    while (node && !(node.scrollHeight > node.clientHeight && /auto|scroll/.test(getComputedStyle(node).overflowY))) {
      node = node.parentElement;
    }
    if (!node) throw new Error('the transcript has no scrollable container');
    node.scrollTop = 0;
    node.dispatchEvent(new Event('scroll'));
  });
}

test('tails a live session and loads history on scroll-up', async ({ page }) => {
  await page.goto(`${baseURL}/sessions/${sessionId}`);
  await expect(page.getByText(label(seeded - 1), { exact: true })).toBeVisible();

  const last = await append(5);
  await expect(page.getByText(label(last), { exact: true })).toBeVisible();

  await expect(async () => {
    await scrollTranscriptToTop(page);
    await expect(page.getByText(label(0), { exact: true })).toBeAttached({ timeout: 1_000 });
  }).toPass({ timeout: 30_000 });
});
